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
	"github.com/solargate/grom/internal/workouts"
)

type memComments struct {
	mu         sync.Mutex
	local      map[string]*workouts.WorkoutComments
	federated  map[string]*workouts.WorkoutComments
	activities map[string]string
}

func newMemComments() *memComments {
	return &memComments{
		local:      map[string]*workouts.WorkoutComments{},
		federated:  map[string]*workouts.WorkoutComments{},
		activities: map[string]string{},
	}
}

func localCommentKey(owner, workoutID string) string { return owner + "/" + workoutID }
func fedCommentKey(viewer, ownerHandle, workoutID string) string {
	return viewer + "/" + ownerHandle + "/" + workoutID
}

func (m *memComments) GetLocal(ownerNickname, workoutID string) (*workouts.WorkoutComments, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if comments, ok := m.local[localCommentKey(ownerNickname, workoutID)]; ok {
		cp := workouts.NormalizeWorkoutComments(comments)
		return &cp, nil
	}
	empty := workouts.NormalizeWorkoutComments(nil)
	return &empty, nil
}

func (m *memComments) PutLocal(ownerNickname, workoutID string, comments *workouts.WorkoutComments) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	norm := workouts.NormalizeWorkoutComments(comments)
	m.local[localCommentKey(ownerNickname, workoutID)] = &norm
	return nil
}

func (m *memComments) DeleteLocal(ownerNickname, workoutID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.local, localCommentKey(ownerNickname, workoutID))
	return nil
}

func (m *memComments) GetFederated(viewerNickname, ownerHandle, workoutID string) (*workouts.WorkoutComments, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if comments, ok := m.federated[fedCommentKey(viewerNickname, ownerHandle, workoutID)]; ok {
		cp := workouts.NormalizeWorkoutComments(comments)
		return &cp, nil
	}
	empty := workouts.NormalizeWorkoutComments(nil)
	return &empty, nil
}

func (m *memComments) PutFederated(viewerNickname, ownerHandle, workoutID string, comments *workouts.WorkoutComments) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	norm := workouts.NormalizeWorkoutComments(comments)
	m.federated[fedCommentKey(viewerNickname, ownerHandle, workoutID)] = &norm
	return nil
}

func (m *memComments) DeleteFederated(viewerNickname, ownerHandle, workoutID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.federated, fedCommentKey(viewerNickname, ownerHandle, workoutID))
	return nil
}

func (m *memComments) GetCommentActivityID(actorNickname, noteID string) (string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.activities[actorNickname+"|"+noteID], nil
}

func (m *memComments) PutCommentActivityID(actorNickname, noteID, activityID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.activities[actorNickname+"|"+noteID] = activityID
	return nil
}

func (m *memComments) DeleteCommentActivityID(actorNickname, noteID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.activities, actorNickname+"|"+noteID)
	return nil
}

func TestInboxProcessorHandleCommentCreateAndDelete(t *testing.T) {
	prev := config.Cfg
	t.Cleanup(func() { config.Cfg = prev })
	config.Cfg.Federation.Domain = "grom.test"

	comments := newMemComments()
	var callbacks [][2]string
	processor := NewInboxProcessor(nil, nil, nil, newTestInboxStore(t.TempDir()), nil)
	processor.SetComments(comments, func(ownerNickname, workoutID string) {
		callbacks = append(callbacks, [2]string{ownerNickname, workoutID})
	})

	workoutID := "38472901"
	noteID := "https://remote.test/users/bob/notes/550e8400-e29b-41d4-a716-446655440000"
	inReplyTo := "https://grom.test/users/alice/workouts/" + workoutID
	createBody := map[string]any{
		"type":  "Create",
		"actor": "https://remote.test/users/bob",
		"object": map[string]any{
			"id":        noteID,
			"type":      "Note",
			"content":   "  Nice run!  ",
			"published": "2026-08-06T12:00:00Z",
			"inReplyTo": inReplyTo,
		},
	}
	createJSON, _ := json.Marshal(createBody)
	if err := processor.Handle("alice", strings.NewReader(string(createJSON))); err != nil {
		t.Fatalf("Create comment: %v", err)
	}
	got, err := comments.GetLocal("alice", workoutID)
	if err != nil {
		t.Fatal(err)
	}
	if got.CommentsNum != 1 {
		t.Fatalf("after create: %#v", got)
	}
	c := got.Comments[0]
	if c.ID != "550e8400-e29b-41d4-a716-446655440000" {
		t.Fatalf("comment id = %q", c.ID)
	}
	if c.Text != "Nice run!" || c.NoteID != noteID || c.User.Handle != "bob@remote.test" {
		t.Fatalf("comment = %#v", c)
	}
	if len(callbacks) != 1 || callbacks[0] != [2]string{"alice", workoutID} {
		t.Fatalf("callbacks = %#v", callbacks)
	}

	if err := processor.Handle("alice", strings.NewReader(string(createJSON))); err != nil {
		t.Fatalf("idempotent Create: %v", err)
	}
	got, err = comments.GetLocal("alice", workoutID)
	if err != nil || got.CommentsNum != 1 {
		t.Fatalf("idempotent comments: %#v err=%v", got, err)
	}

	deleteBody := map[string]any{
		"type":  "Delete",
		"actor": "https://remote.test/users/bob",
		"object": map[string]any{
			"id":        noteID,
			"type":      "Note",
			"inReplyTo": inReplyTo,
		},
	}
	deleteJSON, _ := json.Marshal(deleteBody)
	if err := processor.Handle("alice", strings.NewReader(string(deleteJSON))); err != nil {
		t.Fatalf("Delete comment: %v", err)
	}
	got, err = comments.GetLocal("alice", workoutID)
	if err != nil || got.CommentsNum != 0 {
		t.Fatalf("after delete: %#v err=%v", got, err)
	}
	if len(callbacks) != 3 {
		t.Fatalf("expected 3 callbacks (create, create, delete), got %d", len(callbacks))
	}
}

func TestInboxProcessorCreateCachesFederatedComments(t *testing.T) {
	prev := config.Cfg
	t.Cleanup(func() { config.Cfg = prev })
	config.Cfg.Federation.Domain = "grom.test"

	dir := t.TempDir()
	store := newTestInboxStore(dir)
	comments := newMemComments()
	processor := NewInboxProcessor(nil, nil, nil, store, nil)
	processor.SetComments(comments, nil)

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
			"commentsCount":   1,
			"comments": []any{
				map[string]any{
					"id":       "c1",
					"datetime": "2026-08-06T12:00:00Z",
					"text":     "Great!",
					"noteId":   "https://other.test/users/carol/notes/c1",
					"user": map[string]any{
						"handle": "carol@other.test", "nickname": "carol", "name": "Carol", "is_local": false,
					},
				},
			},
		},
	}
	createJSON, _ := json.Marshal(createObj)
	if err := processor.Handle("alice", strings.NewReader(string(createJSON))); err != nil {
		t.Fatalf("Create: %v", err)
	}

	cached, err := comments.GetFederated("alice", "bob@remote.test", "11111111")
	if err != nil {
		t.Fatal(err)
	}
	if cached.CommentsNum != 1 || cached.Comments[0].Text != "Great!" {
		t.Fatalf("cached comments = %#v", cached)
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
	cached, err = comments.GetFederated("alice", "bob@remote.test", "11111111")
	if err != nil || cached.CommentsNum != 0 {
		t.Fatalf("empty update should clear cache: %#v err=%v", cached, err)
	}
}

func TestInboxProcessorDeleteClearsFederatedComments(t *testing.T) {
	dir := t.TempDir()
	store := newTestInboxStore(dir)
	comments := newMemComments()
	_ = comments.PutFederated("alice", "bob@remote.test", "11111111", &workouts.WorkoutComments{
		Comments: []workouts.WorkoutComment{{
			ID: "c1", User: workouts.WorkoutLikeUser{Handle: "carol@other.test"}, Text: "hi",
		}},
	})
	processor := NewInboxProcessor(nil, nil, nil, store, nil)
	processor.SetComments(comments, nil)

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
	cached, err := comments.GetFederated("alice", "bob@remote.test", "11111111")
	if err != nil || cached.CommentsNum != 0 {
		t.Fatalf("federated comments after delete: %#v err=%v", cached, err)
	}
}

func TestInboxProcessorHandleCommentDeleteFederatedCache(t *testing.T) {
	comments := newMemComments()
	_ = comments.PutFederated("alice", "bob@remote.test", "11111111", &workouts.WorkoutComments{
		Comments: []workouts.WorkoutComment{
			{
				ID: "c1", User: workouts.WorkoutLikeUser{Handle: "carol@other.test"},
				Text: "hi", NoteID: "https://other.test/users/carol/notes/c1",
			},
			{
				ID: "c2", User: workouts.WorkoutLikeUser{Handle: "dave@other.test"},
				Text: "yo", NoteID: "https://other.test/users/dave/notes/c2",
			},
		},
	})
	processor := NewInboxProcessor(nil, nil, nil, nil, nil)
	processor.SetComments(comments, nil)

	// Comment author deletes; owner is inferred from inReplyTo workout URL.
	deleteBody := map[string]any{
		"type":  "Delete",
		"actor": "https://other.test/users/carol",
		"object": map[string]any{
			"id":        "https://other.test/users/carol/notes/c1",
			"type":      "Note",
			"inReplyTo": "https://remote.test/users/bob/workouts/11111111",
		},
	}
	deleteJSON, _ := json.Marshal(deleteBody)
	if err := processor.Handle("alice", strings.NewReader(string(deleteJSON))); err != nil {
		t.Fatalf("Delete federated comment: %v", err)
	}
	cached, err := comments.GetFederated("alice", "bob@remote.test", "11111111")
	if err != nil || cached.CommentsNum != 1 || cached.Comments[0].ID != "c2" {
		t.Fatalf("after author delete: %#v err=%v", cached, err)
	}

	deleteBody2 := map[string]any{
		"type":  "Delete",
		"actor": "https://remote.test/users/bob",
		"object": map[string]any{
			"id":        "https://other.test/users/dave/notes/c2",
			"type":      "Note",
			"inReplyTo": "https://remote.test/users/bob/workouts/11111111",
		},
	}
	deleteJSON2, _ := json.Marshal(deleteBody2)
	if err := processor.Handle("alice", strings.NewReader(string(deleteJSON2))); err != nil {
		t.Fatalf("Delete last federated comment: %v", err)
	}
	cached, err = comments.GetFederated("alice", "bob@remote.test", "11111111")
	if err != nil || cached.CommentsNum != 0 {
		t.Fatalf("after last delete should clear: %#v err=%v", cached, err)
	}
}

func TestDeliverWorkoutCommentAndDelete(t *testing.T) {
	prev := config.Cfg
	t.Cleanup(func() { config.Cfg = prev })
	config.Cfg.Federation.Domain = "grom.test"

	transport := &captureRoundTripper{}
	delivery := &Delivery{client: &http.Client{Transport: transport}}

	objectID := "https://remote.test/users/bob/workouts/38472901"
	noteID := "https://grom.test/users/alice/notes/comment-1"
	published := time.Date(2026, 8, 6, 12, 0, 0, 0, time.UTC)
	activityID, err := delivery.DeliverWorkoutComment("alice", "bob@remote.test", objectID, noteID, "Nice!", published)
	if err != nil {
		t.Fatalf("DeliverWorkoutComment: %v", err)
	}
	if activityID == "" || !strings.HasPrefix(activityID, "https://grom.test/users/alice/activities/") {
		t.Fatalf("activityID = %q", activityID)
	}
	if len(transport.bodies) != 1 {
		t.Fatalf("expected 1 request, got %d", len(transport.bodies))
	}
	create := transport.bodies[0]
	if create["type"] != "Create" {
		t.Fatalf("type = %v", create["type"])
	}
	if create["actor"] != actorURL("alice") {
		t.Fatalf("actor = %v", create["actor"])
	}
	obj, ok := create["object"].(map[string]any)
	if !ok || obj["type"] != "Note" || obj["id"] != noteID || obj["content"] != "Nice!" || obj["inReplyTo"] != objectID {
		t.Fatalf("object = %#v", create["object"])
	}
	if !strings.HasSuffix(transport.requests[0].URL.String(), "/users/bob/inbox") {
		t.Fatalf("inbox url = %s", transport.requests[0].URL.String())
	}

	if err := delivery.DeliverWorkoutCommentDeleteWithReply("alice", "bob@remote.test", noteID, objectID); err != nil {
		t.Fatalf("DeliverWorkoutCommentDeleteWithReply: %v", err)
	}
	if len(transport.bodies) != 2 {
		t.Fatalf("expected 2 requests, got %d", len(transport.bodies))
	}
	del := transport.bodies[1]
	if del["type"] != "Delete" {
		t.Fatalf("delete type = %v", del["type"])
	}
	delObj, ok := del["object"].(map[string]any)
	if !ok || delObj["type"] != "Note" || delObj["id"] != noteID || delObj["inReplyTo"] != objectID {
		t.Fatalf("delete object = %#v", del["object"])
	}

	_, err = delivery.DeliverWorkoutComment("alice", "invalid", objectID, noteID, "x", published)
	if err == nil {
		t.Fatal("expected error for invalid handle")
	}
}

func TestDeliverWorkoutIncludesComments(t *testing.T) {
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
		CommentsCount:   1,
		Comments: []workouts.WorkoutComment{
			{
				ID: "c1",
				User: workouts.WorkoutLikeUser{
					Handle: "carol@other.test", Nickname: "carol", Name: "Carol",
					IsLocal: false, AvatarURL: "https://other.test/avatar.png",
				},
				Datetime: time.Date(2026, 8, 6, 12, 0, 0, 0, time.UTC),
				Text:     "Nice!",
				NoteID:   "https://other.test/users/carol/notes/c1",
			},
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
	if object["commentsCount"] != float64(1) {
		t.Fatalf("commentsCount = %v", object["commentsCount"])
	}
	items, ok := object["comments"].([]any)
	if !ok || len(items) != 1 {
		t.Fatalf("comments = %#v", object["comments"])
	}
	item, _ := items[0].(map[string]any)
	if item["text"] != "Nice!" || item["noteId"] != "https://other.test/users/carol/notes/c1" {
		t.Fatalf("comment entry = %#v", item)
	}
	user, _ := item["user"].(map[string]any)
	if user["handle"] != "carol@other.test" || user["avatarUrl"] != "https://other.test/avatar.png" {
		t.Fatalf("comment user = %#v", user)
	}
}
