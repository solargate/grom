package federation

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/solargate/grom/internal/config"
	"github.com/solargate/grom/internal/workouts"
)

type memLikes struct {
	mu         sync.Mutex
	local      map[string]*workouts.WorkoutLikes
	federated  map[string]*workouts.WorkoutLikes
	activities map[string]string
	callbacks  [][2]string
}

func newMemLikes() *memLikes {
	return &memLikes{
		local:      map[string]*workouts.WorkoutLikes{},
		federated:  map[string]*workouts.WorkoutLikes{},
		activities: map[string]string{},
	}
}

func localLikeKey(owner, workoutID string) string { return owner + "/" + workoutID }
func fedLikeKey(viewer, ownerHandle, workoutID string) string {
	return viewer + "/" + ownerHandle + "/" + workoutID
}

func (m *memLikes) GetLocal(ownerNickname, workoutID string) (*workouts.WorkoutLikes, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if likes, ok := m.local[localLikeKey(ownerNickname, workoutID)]; ok {
		cp := workouts.NormalizeWorkoutLikes(likes)
		return &cp, nil
	}
	empty := workouts.NormalizeWorkoutLikes(nil)
	return &empty, nil
}

func (m *memLikes) PutLocal(ownerNickname, workoutID string, likes *workouts.WorkoutLikes) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	norm := workouts.NormalizeWorkoutLikes(likes)
	m.local[localLikeKey(ownerNickname, workoutID)] = &norm
	return nil
}

func (m *memLikes) DeleteLocal(ownerNickname, workoutID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.local, localLikeKey(ownerNickname, workoutID))
	return nil
}

func (m *memLikes) GetFederated(viewerNickname, ownerHandle, workoutID string) (*workouts.WorkoutLikes, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if likes, ok := m.federated[fedLikeKey(viewerNickname, ownerHandle, workoutID)]; ok {
		cp := workouts.NormalizeWorkoutLikes(likes)
		return &cp, nil
	}
	empty := workouts.NormalizeWorkoutLikes(nil)
	return &empty, nil
}

func (m *memLikes) PutFederated(viewerNickname, ownerHandle, workoutID string, likes *workouts.WorkoutLikes) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	norm := workouts.NormalizeWorkoutLikes(likes)
	m.federated[fedLikeKey(viewerNickname, ownerHandle, workoutID)] = &norm
	return nil
}

func (m *memLikes) DeleteFederated(viewerNickname, ownerHandle, workoutID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.federated, fedLikeKey(viewerNickname, ownerHandle, workoutID))
	return nil
}

func (m *memLikes) GetLikeActivityID(actorNickname, objectID string) (string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.activities[actorNickname+"|"+objectID], nil
}

func (m *memLikes) PutLikeActivityID(actorNickname, objectID, activityID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.activities[actorNickname+"|"+objectID] = activityID
	return nil
}

func (m *memLikes) DeleteLikeActivityID(actorNickname, objectID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.activities, actorNickname+"|"+objectID)
	return nil
}

type captureRoundTripper struct {
	mu       sync.Mutex
	requests []*http.Request
	bodies   []map[string]any
}

func (c *captureRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	body, err := io.ReadAll(req.Body)
	if err != nil {
		return nil, err
	}
	var activity map[string]any
	_ = json.Unmarshal(body, &activity)
	c.mu.Lock()
	c.requests = append(c.requests, req.Clone(req.Context()))
	c.bodies = append(c.bodies, activity)
	c.mu.Unlock()
	return &http.Response{
		StatusCode: http.StatusAccepted,
		Body:       io.NopCloser(bytes.NewReader(nil)),
		Header:     make(http.Header),
		Request:    req,
	}, nil
}

func TestInboxProcessorHandleLikeAndUndo(t *testing.T) {
	prev := config.Cfg
	t.Cleanup(func() { config.Cfg = prev })
	config.Cfg.Federation.Domain = "grom.test"

	likes := newMemLikes()
	var callbacks [][2]string
	processor := NewInboxProcessor(nil, nil, nil, newTestInboxStore(t.TempDir()), nil)
	processor.SetLikes(likes, func(ownerNickname, workoutID string) {
		callbacks = append(callbacks, [2]string{ownerNickname, workoutID})
	})

	workoutID := "38472901"
	objectURL := "https://grom.test/users/alice/workouts/" + workoutID
	likeBody := `{"type":"Like","actor":"https://remote.test/users/bob","object":"` + objectURL + `"}`
	if err := processor.Handle("alice", strings.NewReader(likeBody)); err != nil {
		t.Fatalf("Like: %v", err)
	}
	got, err := likes.GetLocal("alice", workoutID)
	if err != nil {
		t.Fatal(err)
	}
	if got.Likes != 1 || !workouts.LikesContainUser(got, "bob@remote.test") {
		t.Fatalf("after like: %#v", got)
	}
	if len(callbacks) != 1 || callbacks[0] != [2]string{"alice", workoutID} {
		t.Fatalf("callbacks = %#v", callbacks)
	}

	if err := processor.Handle("alice", strings.NewReader(likeBody)); err != nil {
		t.Fatalf("idempotent Like: %v", err)
	}
	got, err = likes.GetLocal("alice", workoutID)
	if err != nil || got.Likes != 1 {
		t.Fatalf("idempotent likes: %#v err=%v", got, err)
	}

	undoBody := map[string]any{
		"type":  "Undo",
		"actor": "https://remote.test/users/bob",
		"object": map[string]any{
			"type":   "Like",
			"actor":  "https://remote.test/users/bob",
			"object": objectURL,
		},
	}
	undoJSON, _ := json.Marshal(undoBody)
	if err := processor.Handle("alice", strings.NewReader(string(undoJSON))); err != nil {
		t.Fatalf("Undo Like: %v", err)
	}
	got, err = likes.GetLocal("alice", workoutID)
	if err != nil || got.Likes != 0 {
		t.Fatalf("after undo: %#v err=%v", got, err)
	}
	if len(callbacks) != 3 {
		t.Fatalf("expected 3 callbacks (like, like, undo), got %d", len(callbacks))
	}
}

func TestInboxProcessorCreateCachesFederatedLikes(t *testing.T) {
	prev := config.Cfg
	t.Cleanup(func() { config.Cfg = prev })
	config.Cfg.Federation.Domain = "grom.test"

	dir := t.TempDir()
	store := newTestInboxStore(dir)
	likes := newMemLikes()
	processor := NewInboxProcessor(nil, nil, nil, store, nil)
	processor.SetLikes(likes, nil)

	createObj := map[string]any{
		"type":  "Create",
		"actor": "https://remote.test/users/bob",
		"object": map[string]any{
			"id":          "https://remote.test/users/bob/workouts/11111111",
			"type":        "Note",
			"name":        "Remote run",
			"sportType":   "Run",
			"startDate":   "2026-07-08T10:00:00Z",
			"durationSeconds": 1200,
			"distance":    3000.0,
			"likesCount":  2,
			"likedUsers": []any{
				map[string]any{
					"handle": "carol@other.test", "nickname": "carol", "name": "Carol", "is_local": false,
				},
				map[string]any{
					"handle": "dave@other.test", "nickname": "dave", "name": "Dave", "is_local": false,
				},
			},
		},
	}
	createJSON, _ := json.Marshal(createObj)
	if err := processor.Handle("alice", strings.NewReader(string(createJSON))); err != nil {
		t.Fatalf("Create: %v", err)
	}

	cached, err := likes.GetFederated("alice", "bob@remote.test", "11111111")
	if err != nil {
		t.Fatal(err)
	}
	if cached.Likes != 2 {
		t.Fatalf("cached likes = %#v", cached)
	}

	updateObj := map[string]any{
		"type":  "Update",
		"actor": "https://remote.test/users/bob",
		"object": map[string]any{
			"id":              "https://remote.test/users/bob/workouts/11111111",
			"type":            "Workout",
			"name":            "Remote run",
			"sportType":       "Run",
			"startDate":       "2026-07-08T10:00:00Z",
			"durationSeconds": 1200,
			"distance":        3000.0,
		},
	}
	updateJSON, _ := json.Marshal(updateObj)
	if err := processor.Handle("alice", strings.NewReader(string(updateJSON))); err != nil {
		t.Fatalf("Update: %v", err)
	}
	cached, err = likes.GetFederated("alice", "bob@remote.test", "11111111")
	if err != nil || cached.Likes != 0 {
		t.Fatalf("empty update should clear cache: %#v err=%v", cached, err)
	}
}

func TestInboxProcessorDeleteClearsFederatedLikes(t *testing.T) {
	dir := t.TempDir()
	store := newTestInboxStore(dir)
	likes := newMemLikes()
	_ = likes.PutFederated("alice", "bob@remote.test", "11111111", &workouts.WorkoutLikes{
		Users: []workouts.WorkoutLikeUser{{Handle: "carol@other.test"}},
	})
	processor := NewInboxProcessor(nil, nil, nil, store, nil)
	processor.SetLikes(likes, nil)

	workout := &workouts.Workout{
		ID: "11111111", Name: "Remote", SportType: "Run",
		StartDate: time.Date(2026, 7, 8, 10, 0, 0, 0, time.UTC),
	}
	if err := store.Save("alice", "bob@remote.test", workout, nil, nil, nil); err != nil {
		t.Fatal(err)
	}

	deleteBody := `{"type":"Delete","actor":"https://remote.test/users/bob","object":"https://remote.test/users/bob/workouts/11111111"}`
	if err := processor.Handle("alice", strings.NewReader(deleteBody)); err != nil {
		t.Fatalf("Delete: %v", err)
	}
	cached, err := likes.GetFederated("alice", "bob@remote.test", "11111111")
	if err != nil || cached.Likes != 0 {
		t.Fatalf("federated likes after delete: %#v err=%v", cached, err)
	}
}

func TestDeliverWorkoutLikeAndUndo(t *testing.T) {
	prev := config.Cfg
	t.Cleanup(func() { config.Cfg = prev })
	config.Cfg.Federation.Domain = "grom.test"

	transport := &captureRoundTripper{}
	delivery := &Delivery{client: &http.Client{Transport: transport}}

	objectID := "https://remote.test/users/bob/workouts/38472901"
	activityID, err := delivery.DeliverWorkoutLike("alice", "bob@remote.test", objectID)
	if err != nil {
		t.Fatalf("DeliverWorkoutLike: %v", err)
	}
	if activityID == "" || !strings.HasPrefix(activityID, "https://grom.test/users/alice/activities/") {
		t.Fatalf("activityID = %q", activityID)
	}
	if len(transport.bodies) != 1 {
		t.Fatalf("expected 1 request, got %d", len(transport.bodies))
	}
	like := transport.bodies[0]
	if like["type"] != "Like" {
		t.Fatalf("type = %v", like["type"])
	}
	if like["actor"] != actorURL("alice") {
		t.Fatalf("actor = %v", like["actor"])
	}
	if like["object"] != objectID {
		t.Fatalf("object = %v", like["object"])
	}
	if like["id"] != activityID {
		t.Fatalf("id = %v", like["id"])
	}
	if !strings.HasSuffix(transport.requests[0].URL.String(), "/users/bob/inbox") {
		t.Fatalf("inbox url = %s", transport.requests[0].URL.String())
	}

	if err := delivery.DeliverWorkoutUndoLike("alice", "bob@remote.test", objectID, activityID); err != nil {
		t.Fatalf("DeliverWorkoutUndoLike: %v", err)
	}
	if len(transport.bodies) != 2 {
		t.Fatalf("expected 2 requests, got %d", len(transport.bodies))
	}
	undo := transport.bodies[1]
	if undo["type"] != "Undo" {
		t.Fatalf("undo type = %v", undo["type"])
	}
	obj, ok := undo["object"].(map[string]any)
	if !ok || obj["type"] != "Like" || obj["id"] != activityID || obj["object"] != objectID {
		t.Fatalf("undo object = %#v", undo["object"])
	}

	_, err = delivery.DeliverWorkoutLike("alice", "invalid", objectID)
	if err == nil {
		t.Fatal("expected error for invalid handle")
	}
}

func TestDeliverWorkoutIncludesLikes(t *testing.T) {
	var received []map[string]any
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
		w.WriteHeader(http.StatusAccepted)
	}))
	defer server.Close()

	prev := config.Cfg
	t.Cleanup(func() { config.Cfg = prev })
	config.Cfg.Federation.Domain = "grom.test"

	workout := &workouts.Workout{
		ID:              "38472901",
		Name:            "Morning run",
		SportType:       "Run",
		StartDate:       time.Date(2026, 7, 8, 10, 0, 0, 0, time.UTC),
		DurationSeconds: 3600,
		Distance:        10000,
		LikesCount:      1,
		LikedUsers: []workouts.WorkoutLikeUser{
			{Handle: "carol@other.test", Nickname: "carol", Name: "Carol", IsLocal: false, AvatarURL: "https://other.test/avatar.png"},
		},
	}
	delivery := &Delivery{client: server.Client()}
	if err := delivery.DeliverWorkout("bob", workout, []string{server.URL}, nil, nil); err != nil {
		t.Fatalf("DeliverWorkout: %v", err)
	}
	if len(received) != 1 {
		t.Fatalf("expected 1 activity, got %d", len(received))
	}
	object, ok := received[0]["object"].(map[string]any)
	if !ok {
		t.Fatalf("object type = %T", received[0]["object"])
	}
	if object["likesCount"] != float64(1) {
		t.Fatalf("likesCount = %v", object["likesCount"])
	}
	users, ok := object["likedUsers"].([]any)
	if !ok || len(users) != 1 {
		t.Fatalf("likedUsers = %#v", object["likedUsers"])
	}
	user, _ := users[0].(map[string]any)
	if user["handle"] != "carol@other.test" || user["avatarUrl"] != "https://other.test/avatar.png" {
		t.Fatalf("liked user = %#v", user)
	}
}
