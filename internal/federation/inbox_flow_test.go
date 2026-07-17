package federation

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/solargate/grom/internal/config"
	"github.com/solargate/grom/internal/social"
	blobfs "github.com/solargate/grom/internal/storage/blob/fs"
	"github.com/solargate/grom/internal/storage/keys"
	"github.com/solargate/grom/internal/tracks"
	"github.com/solargate/grom/internal/users"
	"github.com/solargate/grom/internal/workouts"
)

type memUsers struct {
	byID map[string]*users.User
}

func (m *memUsers) FindByEmail(string) (*users.User, error) { return nil, users.ErrUserNotFound }
func (m *memUsers) FindByID(id string) (*users.User, error) {
	u, ok := m.byID[id]
	if !ok {
		return nil, users.ErrUserNotFound
	}
	return u, nil
}
func (m *memUsers) FindByNickname(string) (*users.User, error) { return nil, users.ErrUserNotFound }
func (m *memUsers) Search(string, string, int) ([]users.User, error) {
	return nil, nil
}
func (m *memUsers) ListAll() ([]users.User, error) { return nil, nil }
func (m *memUsers) Create(string, string, string, string) (*users.User, error) {
	return nil, errors.New("not implemented")
}
func (m *memUsers) UpdateProfile(string, string) (*users.User, error) {
	return nil, errors.New("not implemented")
}
func (m *memUsers) SetLastEquipmentForSport(string, string, []string) error { return nil }
func (m *memUsers) RemoveEquipmentFromLastSets(string, string) error        { return nil }

type memFollows struct {
	mu      sync.Mutex
	follows map[string]*social.Follow
}

func newMemFollows() *memFollows {
	return &memFollows{follows: map[string]*social.Follow{}}
}

func (m *memFollows) FindByID(id string) (*social.Follow, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	f, ok := m.follows[id]
	if !ok {
		return nil, social.ErrFollowNotFound
	}
	cp := *f
	return &cp, nil
}
func (m *memFollows) ListByFollower(string) ([]social.Follow, error) { return nil, nil }
func (m *memFollows) ListActiveFollowing(string) ([]social.Follow, error) {
	return nil, nil
}
func (m *memFollows) ListActiveByTarget(string) ([]social.Follow, error) { return nil, nil }
func (m *memFollows) FindExisting(string, string) (*social.Follow, error) {
	return nil, social.ErrFollowNotFound
}
func (m *memFollows) Create(follow social.Follow) (*social.Follow, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if follow.ID == "" {
		follow.ID = uuid.NewString()
	}
	cp := follow
	m.follows[cp.ID] = &cp
	out := cp
	return &out, nil
}
func (m *memFollows) UpdateStatus(id, status string) (*social.Follow, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	f, ok := m.follows[id]
	if !ok {
		return nil, social.ErrFollowNotFound
	}
	f.Status = status
	cp := *f
	return &cp, nil
}
func (m *memFollows) FindByFollowActivityID(activityID string) (*social.Follow, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, f := range m.follows {
		if f.FollowActivityID == activityID {
			cp := *f
			return &cp, nil
		}
	}
	return nil, social.ErrFollowNotFound
}
func (m *memFollows) UpdateActivityID(id, activityID string) (*social.Follow, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	f, ok := m.follows[id]
	if !ok {
		return nil, social.ErrFollowNotFound
	}
	f.FollowActivityID = activityID
	cp := *f
	return &cp, nil
}
func (m *memFollows) Delete(id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.follows, id)
	return nil
}

type memFollowers struct {
	mu   sync.Mutex
	by   map[string][]InboundFollower
}

func newMemFollowers() *memFollowers {
	return &memFollowers{by: map[string][]InboundFollower{}}
}

func (m *memFollowers) Add(nickname string, follower InboundFollower) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.by[nickname] = append(m.by[nickname], follower)
	return nil
}
func (m *memFollowers) List(nickname string) ([]InboundFollower, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	out := append([]InboundFollower(nil), m.by[nickname]...)
	return out, nil
}
func (m *memFollowers) ListInboxes(string) ([]string, error) { return nil, nil }
func (m *memFollowers) Remove(nickname, actorURI string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	kept := m.by[nickname][:0]
	for _, f := range m.by[nickname] {
		if f.ActorURI != actorURI {
			kept = append(kept, f)
		}
	}
	m.by[nickname] = kept
	return nil
}

func TestDeliverWorkoutDeletePayload(t *testing.T) {
	var received []map[string]any
	var contentType string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		contentType = r.Header.Get("Content-Type")
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatal(err)
		}
		var activity map[string]any
		if err := json.Unmarshal(body, &activity); err != nil {
			t.Fatal(err)
		}
		received = append(received, activity)
		w.WriteHeader(http.StatusAccepted)
	}))
	defer server.Close()

	prev := config.Cfg
	t.Cleanup(func() { config.Cfg = prev })
	config.Cfg.Federation.Domain = "grom.test"

	delivery := &Delivery{client: server.Client()}
	if err := delivery.DeliverWorkoutDelete("bob", "38472901", []string{server.URL}); err != nil {
		t.Fatalf("DeliverWorkoutDelete() error = %v", err)
	}
	if len(received) != 1 {
		t.Fatalf("expected 1 activity, got %d", len(received))
	}
	act := received[0]
	if act["type"] != "Delete" {
		t.Fatalf("type = %v", act["type"])
	}
	if act["actor"] != actorURL("bob") {
		t.Fatalf("actor = %v", act["actor"])
	}
	if act["@context"] != "https://www.w3.org/ns/activitystreams" {
		t.Fatalf("@context = %v", act["@context"])
	}
	object, _ := act["object"].(string)
	if object != workoutObjectURL("bob", "38472901") {
		t.Fatalf("object = %q", object)
	}
	if contentType != "application/activity+json" {
		t.Fatalf("Content-Type = %q", contentType)
	}
}

func TestDeliverFollowAndUndo(t *testing.T) {
	var types []string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		var activity map[string]any
		_ = json.Unmarshal(body, &activity)
		types = append(types, activity["type"].(string))
		if r.Header.Get("Content-Type") != "application/activity+json" {
			t.Errorf("Content-Type = %q", r.Header.Get("Content-Type"))
		}
		w.WriteHeader(http.StatusAccepted)
	}))
	defer server.Close()

	prev := config.Cfg
	t.Cleanup(func() { config.Cfg = prev })
	config.Cfg.Federation.Domain = "grom.test"

	usersStore := &memUsers{byID: map[string]*users.User{
		"alice-id": {ID: "alice-id", Nickname: "alice"},
	}}
	delivery := &Delivery{client: server.Client(), userStore: usersStore}
	follow := &social.Follow{
		FollowerID:       "alice-id",
		TargetActorURI:   strings.TrimSuffix(server.URL, "/") + "/users/remote",
		FollowActivityID: "https://grom.test/users/alice/follows/1",
	}
	if err := delivery.DeliverFollow(follow); err != nil {
		t.Fatalf("DeliverFollow: %v", err)
	}
	if err := delivery.DeliverUndo(follow); err != nil {
		t.Fatalf("DeliverUndo: %v", err)
	}
	if len(types) != 2 || types[0] != "Follow" || types[1] != "Undo" {
		t.Fatalf("types = %v", types)
	}
}

func TestLoadOrCreateActorKey(t *testing.T) {
	prev := config.Cfg
	t.Cleanup(func() { config.Cfg = prev })
	config.Cfg.Federation.Domain = "grom.test"

	dir := t.TempDir()
	blobs := blobfs.NewStore(dir)

	pub1, keyID1, err := LoadOrCreateActorKey(blobs, "alice")
	if err != nil {
		t.Fatal(err)
	}
	if pub1 == "" || !strings.Contains(pub1, "PUBLIC KEY") {
		t.Fatalf("unexpected public key: %q", pub1)
	}
	if keyID1 != "https://grom.test/users/alice#main-key" {
		t.Fatalf("keyID = %q", keyID1)
	}

	pub2, keyID2, err := LoadOrCreateActorKey(blobs, "alice")
	if err != nil {
		t.Fatal(err)
	}
	if pub1 != pub2 || keyID1 != keyID2 {
		t.Fatal("expected stable key on second load")
	}
	ok, err := blobs.Exists(context.Background(), keys.UserActorKey("alice"))
	if err != nil || !ok {
		t.Fatalf("private key blob missing: %v", err)
	}
}

func TestInboxProcessorHandleFollowAcceptCreate(t *testing.T) {
	prev := config.Cfg
	t.Cleanup(func() { config.Cfg = prev })
	config.Cfg.Federation.Domain = "grom.test"
	config.Cfg.Federation.AutoAcceptFollows = false
	config.Cfg.Federation.Enabled = true

	dir := t.TempDir()
	store := newTestInboxStore(dir)
	followers := newMemFollowers()
	usersStore := &memUsers{byID: map[string]*users.User{
		"alice-id": {ID: "alice-id", Nickname: "alice", Name: "Alice"},
	}}
	follows := newMemFollows()
	socialSvc := social.NewService(usersStore, follows, blobfs.NewStore(dir))

	created, err := follows.Create(social.Follow{
		FollowerID:       "alice-id",
		TargetHandle:     "remote@other.test",
		TargetNickname:   "remote",
		Status:           social.StatusPending,
		FollowActivityID: "https://other.test/follows/abc",
		CreatedAt:        time.Now().UTC(),
	})
	if err != nil {
		t.Fatal(err)
	}

	processor := NewInboxProcessor(usersStore, socialSvc, nil, store, followers)

	followBody := `{"type":"Follow","id":"https://remote.test/follows/1","actor":"https://remote.test/users/bob","object":"https://grom.test/users/alice"}`
	if err := processor.Handle("alice", strings.NewReader(followBody)); err != nil {
		t.Fatalf("Follow: %v", err)
	}
	list, err := followers.List("alice")
	if err != nil {
		t.Fatal(err)
	}
	if len(list) != 1 {
		t.Fatalf("expected 1 inbound follower, got %d", len(list))
	}

	acceptBody := `{"type":"Accept","object":{"id":"https://other.test/follows/abc","type":"Follow"}}`
	if err := processor.Handle("alice", strings.NewReader(acceptBody)); err != nil {
		t.Fatalf("Accept: %v", err)
	}
	updated, err := follows.FindByID(created.ID)
	if err != nil {
		t.Fatal(err)
	}
	if updated.Status != social.StatusActive {
		t.Fatalf("status = %q", updated.Status)
	}

	createObj := map[string]any{
		"type":  "Create",
		"actor": "https://remote.test/users/bob",
		"object": map[string]any{
			"id":              "https://remote.test/users/bob/workouts/11111111",
			"type":            "Note",
			"name":            "Remote run",
			"sportType":       "Run",
			"startDate":       "2026-07-08T10:00:00Z",
			"durationSeconds": 1200,
			"distance":        3000.0,
		},
	}
	createJSON, _ := json.Marshal(createObj)
	if err := processor.Handle("alice", strings.NewReader(string(createJSON))); err != nil {
		t.Fatalf("Create: %v", err)
	}
	items, err := store.List("alice")
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 1 {
		t.Fatalf("expected 1 federated workout after Create, got %d", len(items))
	}

	undoBody := `{"type":"Undo","object":{"type":"Follow","actor":"https://remote.test/users/bob"}}`
	if err := processor.Handle("alice", strings.NewReader(undoBody)); err != nil {
		t.Fatalf("Undo: %v", err)
	}
	list, err = followers.List("alice")
	if err != nil {
		t.Fatal(err)
	}
	if len(list) != 0 {
		t.Fatalf("expected 0 followers after undo, got %d", len(list))
	}
}

func TestInboxProcessorHandleDeleteRemovesBlobs(t *testing.T) {
	dir := t.TempDir()
	store := newTestInboxStore(dir)
	blobs := blobfs.NewStore(dir)

	gpxData, err := os.ReadFile(filepath.Join("..", "..", "testdata", "tracks", "1-sample.gpx"))
	if err != nil {
		t.Fatal(err)
	}

	workout := &workouts.Workout{
		ID:              "38472901",
		Name:            "Remote run",
		SportType:       "Run",
		StartDate:       time.Date(2026, 7, 8, 10, 0, 0, 0, time.UTC),
		DurationSeconds: 4200,
		Distance:        10000,
		Track:           tracks.TrackFileGPX,
	}
	ownerHandle := "test2@192.168.1.251:8445"
	if err := store.Save("solarwind", ownerHandle, workout, gpxData, nil, nil); err != nil {
		t.Fatal(err)
	}

	ownerKey := OwnerKeyFromHandle(ownerHandle)
	trackKey := keys.FederatedInboxTrack("solarwind", ownerKey, workout.ID, tracks.TrackFileGPX)
	if ok, _ := blobs.Exists(context.Background(), trackKey); !ok {
		t.Fatalf("expected track blob at %q", trackKey)
	}

	processor := NewInboxProcessor(nil, nil, nil, store, nil)
	body := strings.NewReader(`{"type":"Delete","actor":"https://192.168.1.251:8445/users/test2","object":"https://localhost/users/test2/workouts/38472901"}`)
	if err := processor.Handle("solarwind", body); err != nil {
		t.Fatalf("Handle() error = %v", err)
	}

	items, err := store.List("solarwind")
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 0 {
		t.Fatalf("expected 0 items after delete, got %d", len(items))
	}
	if ok, _ := blobs.Exists(context.Background(), trackKey); ok {
		t.Fatal("expected track blob removed after delete")
	}
}
