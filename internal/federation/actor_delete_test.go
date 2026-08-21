package federation

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/solargate/grom/internal/config"
	"github.com/solargate/grom/internal/social"
	"github.com/solargate/grom/internal/users"
	"github.com/solargate/grom/internal/workouts"
)

func TestIsActorDelete(t *testing.T) {
	actor := "https://remote.test/users/bob"
	if !isActorDelete(actor, actor) {
		t.Fatal("string object should match")
	}
	if !isActorDelete(map[string]any{"id": actor, "type": "Person"}, actor) {
		t.Fatal("Person object should match")
	}
	if !isActorDelete(map[string]any{"id": actor, "type": "Tombstone"}, actor) {
		t.Fatal("Tombstone object should match")
	}
	if isActorDelete("https://remote.test/users/bob/workouts/1", actor) {
		t.Fatal("workout object must not match")
	}
	if isActorDelete(map[string]any{"id": actor + "/workouts/1", "type": "Note"}, actor) {
		t.Fatal("Note must not match")
	}
}

func TestDeliverActorDelete(t *testing.T) {
	prev := config.Cfg
	t.Cleanup(func() { config.Cfg = prev })
	config.Cfg.Federation.Domain = "grom.test"

	var received []map[string]any
	var inboxes []string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatal(err)
		}
		var activity map[string]any
		if err := json.Unmarshal(body, &activity); err != nil {
			t.Fatal(err)
		}
		received = append(received, activity)
		inboxes = append(inboxes, r.URL.Path)
		w.WriteHeader(http.StatusAccepted)
	}))
	defer server.Close()

	delivery := &Delivery{client: server.Client()}
	targets := []string{server.URL + "/inbox-a", server.URL + "/inbox-b"}
	delivery.DeliverActorDelete("alice", targets)

	if len(received) != 2 {
		t.Fatalf("expected 2 deliveries, got %d (%v)", len(received), inboxes)
	}
	for _, act := range received {
		if act["type"] != "Delete" {
			t.Fatalf("type = %v", act["type"])
		}
		if act["actor"] != "https://grom.test/users/alice" {
			t.Fatalf("actor = %v", act["actor"])
		}
		if act["object"] != "https://grom.test/users/alice" {
			t.Fatalf("object = %v", act["object"])
		}
	}
}

func TestDeliverActorDeleteNoopWhenEmpty(t *testing.T) {
	delivery := &Delivery{client: http.DefaultClient}
	delivery.DeliverActorDelete("alice", nil)
	delivery.DeliverActorDelete("", []string{"https://example/inbox"})
	var nilDelivery *Delivery
	nilDelivery.DeliverActorDelete("alice", []string{"https://example/inbox"})
}

func TestHandleActorDeletePurgesInboxAndCaches(t *testing.T) {
	prev := config.Cfg
	t.Cleanup(func() { config.Cfg = prev })
	config.Cfg.Federation.Domain = "grom.test"

	dir := t.TempDir()
	store := newTestInboxStore(dir)
	likes := newMemLikes()
	comments := newMemComments()
	followers := newMemFollowers()

	viewer := "alice"
	ownerHandle := "bob@remote.test"
	actorURI := "https://remote.test/users/bob"
	workoutID := "38472901"

	workout := &workouts.Workout{
		ID:              workoutID,
		Name:            "Remote run",
		SportType:       "Run",
		StartDate:       time.Date(2026, 7, 8, 10, 0, 0, 0, time.UTC),
		DurationSeconds: 1200,
		Distance:        3000,
	}
	if err := store.Save(viewer, ownerHandle, workout, nil, nil, map[string]any{
		"id":                actorURI,
		"preferredUsername": "bob",
		"name":              "Bob",
	}); err != nil {
		t.Fatal(err)
	}
	fedLikes := workouts.AddWorkoutLikeUser(nil, workouts.WorkoutLikeUser{
		Handle: "carol@other.test", Nickname: "carol",
	})
	if err := likes.PutFederated(viewer, ownerHandle, workoutID, &fedLikes); err != nil {
		t.Fatal(err)
	}
	fedComments := workouts.AddWorkoutComment(nil, workouts.WorkoutComment{
		ID: "c1", User: workouts.WorkoutLikeUser{Handle: "carol@other.test"}, Text: "hi",
	})
	if err := comments.PutFederated(viewer, ownerHandle, workoutID, &fedComments); err != nil {
		t.Fatal(err)
	}
	if err := followers.Add(viewer, InboundFollower{
		ActorURI: actorURI, Inbox: actorURI + "/inbox", Handle: ownerHandle,
	}); err != nil {
		t.Fatal(err)
	}

	follows := newActorDeleteFollows()
	userStore := &actorDeleteUsers{byNick: map[string]*users.User{
		viewer: {ID: "user-alice", Nickname: viewer},
	}}
	if _, err := follows.Create(social.Follow{
		ID: "f1", FollowerID: "user-alice", TargetHandle: ownerHandle,
		TargetActorURI: actorURI, Status: social.StatusActive,
	}); err != nil {
		t.Fatal(err)
	}
	socialSvc := social.NewService(userStore, follows, nil)

	processor := NewInboxProcessor(userStore, socialSvc, nil, store, followers)
	processor.SetLikes(likes, nil)
	processor.SetComments(comments, nil)

	body := strings.NewReader(`{"type":"Delete","actor":"` + actorURI + `","object":"` + actorURI + `"}`)
	if err := processor.Handle(viewer, body); err != nil {
		t.Fatalf("Handle: %v", err)
	}

	items, err := store.List(viewer)
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 0 {
		t.Fatalf("expected inbox empty, got %#v", items)
	}
	gotLikes, err := likes.GetFederated(viewer, ownerHandle, workoutID)
	if err != nil || gotLikes.Likes != 0 {
		t.Fatalf("federated likes: %#v err=%v", gotLikes, err)
	}
	gotComments, err := comments.GetFederated(viewer, ownerHandle, workoutID)
	if err != nil || gotComments.CommentsNum != 0 {
		t.Fatalf("federated comments: %#v err=%v", gotComments, err)
	}
	listed, err := followers.List(viewer)
	if err != nil {
		t.Fatal(err)
	}
	for _, f := range listed {
		if f.ActorURI == actorURI {
			t.Fatalf("expected inbound follower removed: %#v", listed)
		}
	}
	remaining, err := follows.ListByFollower("user-alice")
	if err != nil {
		t.Fatal(err)
	}
	if len(remaining) != 0 {
		t.Fatalf("expected outbound follow removed: %#v", remaining)
	}
}

type actorDeleteUsers struct {
	byNick map[string]*users.User
}

func (m *actorDeleteUsers) FindByEmail(string) (*users.User, error) { return nil, users.ErrUserNotFound }
func (m *actorDeleteUsers) FindByID(id string) (*users.User, error) {
	for _, u := range m.byNick {
		if u.ID == id {
			return u, nil
		}
	}
	return nil, users.ErrUserNotFound
}
func (m *actorDeleteUsers) FindByNickname(nickname string) (*users.User, error) {
	if u, ok := m.byNick[nickname]; ok {
		return u, nil
	}
	return nil, users.ErrUserNotFound
}
func (m *actorDeleteUsers) Search(string, string, int) ([]users.User, error) { return nil, nil }
func (m *actorDeleteUsers) ListAll() ([]users.User, error)                   { return nil, nil }
func (m *actorDeleteUsers) Create(string, string, string, string) (*users.User, error) {
	return nil, users.ErrUserNotFound
}
func (m *actorDeleteUsers) UpdateProfile(string, string) (*users.User, error) {
	return nil, users.ErrUserNotFound
}
func (m *actorDeleteUsers) UpdatePassword(string, string) error              { return users.ErrUserNotFound }
func (m *actorDeleteUsers) SetLastEquipmentForSport(string, string, []string) error {
	return nil
}
func (m *actorDeleteUsers) RemoveEquipmentFromLastSets(string, string) error { return nil }
func (m *actorDeleteUsers) GetProfile(string) (*users.Profile, error) {
	return &users.Profile{}, nil
}
func (m *actorDeleteUsers) PutProfile(string, users.Profile) error { return nil }
func (m *actorDeleteUsers) SetLastSportType(string, string) error  { return nil }
func (m *actorDeleteUsers) TouchUsedSportType(string, string) error {
	return nil
}
func (m *actorDeleteUsers) PruneUsedSportTypes(string, map[string]struct{}) error {
	return nil
}
func (m *actorDeleteUsers) Delete(string) error { return users.ErrUserNotFound }

type actorDeleteFollows struct {
	mu      sync.Mutex
	follows map[string]*social.Follow
}

func newActorDeleteFollows() *actorDeleteFollows {
	return &actorDeleteFollows{follows: map[string]*social.Follow{}}
}

func (m *actorDeleteFollows) FindByID(id string) (*social.Follow, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	f, ok := m.follows[id]
	if !ok {
		return nil, social.ErrFollowNotFound
	}
	cp := *f
	return &cp, nil
}
func (m *actorDeleteFollows) ListByFollower(followerID string) ([]social.Follow, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	var out []social.Follow
	for _, f := range m.follows {
		if f.FollowerID == followerID {
			out = append(out, *f)
		}
	}
	return out, nil
}
func (m *actorDeleteFollows) ListActiveFollowing(string) ([]social.Follow, error) { return nil, nil }
func (m *actorDeleteFollows) ListActiveByTarget(string) ([]social.Follow, error)  { return nil, nil }
func (m *actorDeleteFollows) FindExisting(string, string) (*social.Follow, error) {
	return nil, social.ErrFollowNotFound
}
func (m *actorDeleteFollows) Create(follow social.Follow) (*social.Follow, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	cp := follow
	m.follows[cp.ID] = &cp
	out := cp
	return &out, nil
}
func (m *actorDeleteFollows) UpdateStatus(string, string) (*social.Follow, error) {
	return nil, social.ErrFollowNotFound
}
func (m *actorDeleteFollows) FindByFollowActivityID(string) (*social.Follow, error) {
	return nil, social.ErrFollowNotFound
}
func (m *actorDeleteFollows) UpdateActivityID(string, string) (*social.Follow, error) {
	return nil, social.ErrFollowNotFound
}
func (m *actorDeleteFollows) Delete(id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.follows, id)
	return nil
}
func (m *actorDeleteFollows) DeleteInvolving(string, string) error { return nil }
