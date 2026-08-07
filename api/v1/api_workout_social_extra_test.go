package v1_test

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
	"unicode/utf8"

	"github.com/solargate/grom/internal/config"
	"github.com/solargate/grom/internal/social"
	"github.com/solargate/grom/internal/workouts"
)

type captureDeliveryRT struct {
	mu       sync.Mutex
	requests []*http.Request
	bodies   []map[string]any
}

func (c *captureDeliveryRT) RoundTrip(req *http.Request) (*http.Response, error) {
	var body []byte
	if req.Body != nil {
		var err error
		body, err = io.ReadAll(req.Body)
		if err != nil {
			return nil, err
		}
	}
	if req.Method == http.MethodPost {
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

	// Minimal ActivityPub actor / webfinger responses for inbox Create metadata fetch.
	path := req.URL.Path
	var payload []byte
	switch {
	case strings.Contains(path, "/.well-known/webfinger"):
		resource := req.URL.Query().Get("resource")
		payload, _ = json.Marshal(map[string]any{
			"subject": resource,
			"links": []any{
				map[string]any{
					"rel":  "self",
					"type": "application/activity+json",
					"href": "https://remote.example/users/bob",
				},
			},
		})
	case strings.HasPrefix(path, "/users/"):
		nick := strings.TrimPrefix(path, "/users/")
		if i := strings.Index(nick, "/"); i >= 0 {
			nick = nick[:i]
		}
		payload, _ = json.Marshal(map[string]any{
			"@context":          "https://www.w3.org/ns/activitystreams",
			"id":                "https://" + req.URL.Host + "/users/" + nick,
			"type":              "Person",
			"preferredUsername": nick,
			"name":              nick,
			"inbox":             "https://" + req.URL.Host + "/users/" + nick + "/inbox",
		})
	default:
		payload = []byte(`{}`)
	}
	return &http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(bytes.NewReader(payload)),
		Header:     http.Header{"Content-Type": []string{"application/activity+json"}},
		Request:    req,
	}, nil
}

func (c *captureDeliveryRT) activities() []map[string]any {
	c.mu.Lock()
	defer c.mu.Unlock()
	out := make([]map[string]any, len(c.bodies))
	copy(out, c.bodies)
	return out
}

func TestWorkoutCommentsBBoltDriver(t *testing.T) {
	ta := setupTestAppWithDriver(t, config.StorageDriverBBolt)
	ta.register(t, "alice", "alice@example.com", "password12")
	ta.register(t, "bob", "bob@example.com", "password12")
	aliceToken, _ := ta.login(t, "alice@example.com", "password12")
	bobToken, _ := ta.login(t, "bob@example.com", "password12")

	w := ta.doJSON(t, http.MethodPost, "/api/v1/social/follow", map[string]string{"handle": "bob"}, aliceToken)
	expectStatus(t, w, http.StatusCreated)

	w = ta.doJSON(t, http.MethodPost, "/api/v1/workouts", map[string]any{
		"name": "Bob run", "sport_type": "Run", "start_date": "2026-07-08T10:00:00Z",
		"duration_seconds": 1800, "distance": 5000,
	}, bobToken)
	expectStatus(t, w, http.StatusCreated)
	id, _ := decodeObject(t, w)["id"].(string)
	path := "/api/v1/workouts/" + id + "/comments?owner=bob"

	w = ta.doJSON(t, http.MethodPost, path, map[string]string{"text": "bbolt hi"}, aliceToken)
	expectStatus(t, w, http.StatusOK)
	if decodeObject(t, w)["count"] != float64(1) {
		t.Fatalf("bbolt comment: %s", w.Body.String())
	}
	w = ta.doJSON(t, http.MethodGet, path, nil, aliceToken)
	expectStatus(t, w, http.StatusOK)
	if decodeObject(t, w)["count"] != float64(1) {
		t.Fatalf("bbolt list: %s", w.Body.String())
	}
}

func TestDeleteWorkoutRemovesComments(t *testing.T) {
	ta := setupTestApp(t)
	ta.register(t, "alice", "alice@example.com", "password12")
	ta.register(t, "bob", "bob@example.com", "password12")
	aliceToken, _ := ta.login(t, "alice@example.com", "password12")
	bobToken, _ := ta.login(t, "bob@example.com", "password12")

	w := ta.doJSON(t, http.MethodPost, "/api/v1/social/follow", map[string]string{"handle": "bob"}, aliceToken)
	expectStatus(t, w, http.StatusCreated)

	w = ta.doJSON(t, http.MethodPost, "/api/v1/workouts", map[string]any{
		"name": "Bob run", "sport_type": "Run", "start_date": "2026-07-08T10:00:00Z",
		"duration_seconds": 1800, "distance": 5000,
	}, bobToken)
	expectStatus(t, w, http.StatusCreated)
	id, _ := decodeObject(t, w)["id"].(string)

	w = ta.doJSON(t, http.MethodPost, "/api/v1/workouts/"+id+"/comments?owner=bob", map[string]string{
		"text": "bye soon",
	}, aliceToken)
	expectStatus(t, w, http.StatusOK)

	w = ta.doJSON(t, http.MethodDelete, "/api/v1/workouts/"+id, nil, bobToken)
	expectStatus(t, w, http.StatusNoContent)

	comments, err := ta.app.Comments.GetLocal("bob", id)
	if err == nil && comments.CommentsNum != 0 {
		t.Fatalf("expected comments cleared after workout delete, got %#v", comments)
	}
}

func TestWorkoutListIncludesSocialSummary(t *testing.T) {
	ta := setupTestApp(t)
	ta.register(t, "alice", "alice@example.com", "password12")
	ta.register(t, "bob", "bob@example.com", "password12")
	aliceToken, _ := ta.login(t, "alice@example.com", "password12")
	bobToken, _ := ta.login(t, "bob@example.com", "password12")

	w := ta.doJSON(t, http.MethodPost, "/api/v1/social/follow", map[string]string{"handle": "bob"}, aliceToken)
	expectStatus(t, w, http.StatusCreated)

	w = ta.doJSON(t, http.MethodPost, "/api/v1/workouts", map[string]any{
		"name": "Bob run", "sport_type": "Run", "start_date": "2026-07-08T10:00:00Z",
		"duration_seconds": 1800, "distance": 5000,
	}, bobToken)
	expectStatus(t, w, http.StatusCreated)
	id, _ := decodeObject(t, w)["id"].(string)

	w = ta.doJSON(t, http.MethodPost, "/api/v1/workouts/"+id+"/likes?owner=bob", nil, aliceToken)
	expectStatus(t, w, http.StatusOK)
	w = ta.doJSON(t, http.MethodPost, "/api/v1/workouts/"+id+"/comments?owner=bob", map[string]string{
		"text": "nice",
	}, aliceToken)
	expectStatus(t, w, http.StatusOK)

	w = ta.doJSON(t, http.MethodGet, "/api/v1/workouts?scope=feed&limit=20", nil, aliceToken)
	expectStatus(t, w, http.StatusOK)
	items, _, _ := decodeWorkoutPage(t, w)
	var found map[string]any
	for _, item := range items {
		if item["id"] == id {
			found = item
			break
		}
	}
	if found == nil {
		t.Fatalf("workout missing from feed: %#v", items)
	}
	if found["likes_count"] != float64(1) || found["liked_by_me"] != true || found["can_like"] != true {
		t.Fatalf("feed likes summary: %#v", found)
	}
	if found["comments_count"] != float64(1) {
		t.Fatalf("feed comments_count: %#v", found)
	}

	w = ta.doJSON(t, http.MethodGet, "/api/v1/workouts?scope=own&limit=20", nil, bobToken)
	expectStatus(t, w, http.StatusOK)
	ownItems, _, _ := decodeWorkoutPage(t, w)
	found = nil
	for _, item := range ownItems {
		if item["id"] == id {
			found = item
			break
		}
	}
	if found == nil {
		t.Fatal("workout missing from own list")
	}
	if found["likes_count"] != float64(1) || found["can_like"] != false {
		t.Fatalf("own list summary: %#v", found)
	}
	if found["comments_count"] != float64(1) {
		t.Fatalf("own list comments_count: %#v", found)
	}
}

func TestLikeWithoutFollowReturnsNotFound(t *testing.T) {
	ta := setupTestApp(t)
	ta.register(t, "alice", "alice@example.com", "password12")
	ta.register(t, "bob", "bob@example.com", "password12")
	aliceToken, _ := ta.login(t, "alice@example.com", "password12")
	bobToken, _ := ta.login(t, "bob@example.com", "password12")

	w := ta.doJSON(t, http.MethodPost, "/api/v1/workouts", map[string]any{
		"name": "Bob run", "sport_type": "Run", "start_date": "2026-07-08T10:00:00Z",
	}, bobToken)
	expectStatus(t, w, http.StatusCreated)
	id, _ := decodeObject(t, w)["id"].(string)

	w = ta.doJSON(t, http.MethodPost, "/api/v1/workouts/"+id+"/likes?owner=bob", nil, aliceToken)
	expectStatus(t, w, http.StatusNotFound)
	w = ta.doJSON(t, http.MethodGet, "/api/v1/workouts/"+id+"/likes?owner=bob", nil, aliceToken)
	expectStatus(t, w, http.StatusNotFound)
}

func TestWorkoutCommentRejectsOverlongRunes(t *testing.T) {
	ta := setupTestApp(t)
	ta.register(t, "alice", "alice@example.com", "password12")
	ta.register(t, "bob", "bob@example.com", "password12")
	aliceToken, _ := ta.login(t, "alice@example.com", "password12")
	bobToken, _ := ta.login(t, "bob@example.com", "password12")

	w := ta.doJSON(t, http.MethodPost, "/api/v1/social/follow", map[string]string{"handle": "bob"}, aliceToken)
	expectStatus(t, w, http.StatusCreated)
	w = ta.doJSON(t, http.MethodPost, "/api/v1/workouts", map[string]any{
		"name": "Bob run", "sport_type": "Run", "start_date": "2026-07-08T10:00:00Z",
	}, bobToken)
	expectStatus(t, w, http.StatusCreated)
	id, _ := decodeObject(t, w)["id"].(string)
	path := "/api/v1/workouts/" + id + "/comments?owner=bob"

	exact := strings.Repeat("あ", workouts.MaxCommentTextLength)
	if utf8.RuneCountInString(exact) != workouts.MaxCommentTextLength {
		t.Fatal("fixture length")
	}
	w = ta.doJSON(t, http.MethodPost, path, map[string]string{"text": exact}, aliceToken)
	expectStatus(t, w, http.StatusOK)

	tooLong := exact + "あ"
	w = ta.doJSON(t, http.MethodPost, path, map[string]string{"text": tooLong}, aliceToken)
	expectStatus(t, w, http.StatusBadRequest)
}

func seedRemoteFollow(t *testing.T, ta *testApp, viewerNickname, handle, actorURI string) {
	t.Helper()
	user, err := ta.app.Users.FindByNickname(viewerNickname)
	if err != nil {
		t.Fatal(err)
	}
	nick := handle
	if at := strings.Index(handle, "@"); at >= 0 {
		nick = handle[:at]
	}
	if _, err := ta.app.Backend.Social().Create(social.Follow{
		FollowerID:     user.ID,
		TargetHandle:   handle,
		TargetNickname: nick,
		TargetActorURI: actorURI,
		TargetIsLocal:  false,
		Status:         social.StatusActive,
		CreatedAt:      time.Now().UTC(),
	}); err != nil {
		t.Fatalf("seed remote follow: %v", err)
	}
}

func postFederatedWorkoutCreate(t *testing.T, ta *testApp, viewer, actorURL, workoutID string, extraObject map[string]any) {
	t.Helper()
	obj := map[string]any{
		"id":              actorURL + "/workouts/" + workoutID,
		"type":            "Note",
		"name":            "Remote run",
		"sportType":       "Run",
		"startDate":       "2026-07-08T10:00:00Z",
		"durationSeconds": 1200,
		"distance":        3000.0,
	}
	for k, v := range extraObject {
		obj[k] = v
	}
	createObj := map[string]any{
		"@context": "https://www.w3.org/ns/activitystreams",
		"type":     "Create",
		"actor":    actorURL,
		"object":   obj,
	}
	data, err := json.Marshal(createObj)
	if err != nil {
		t.Fatal(err)
	}
	req := httptest.NewRequest(http.MethodPost, "/users/"+viewer+"/inbox", bytes.NewReader(data))
	req.Header.Set("Content-Type", "application/activity+json")
	w := httptest.NewRecorder()
	ta.router.ServeHTTP(w, req)
	expectStatus(t, w, http.StatusAccepted)
}

func TestFederatedLikeAndCommentViaAPI(t *testing.T) {
	ta := setupFederationTestApp(t)
	transport := &captureDeliveryRT{}
	ta.app.SetFederationHTTPClient(&http.Client{Transport: transport})

	ta.register(t, "alice", "alice@example.com", "password12")
	aliceToken, _ := ta.login(t, "alice@example.com", "password12")

	actorURL := "https://remote.example/users/bob"
	seedRemoteFollow(t, ta, "alice", "bob@remote.example", actorURL)
	workoutID := "22222222"
	postFederatedWorkoutCreate(t, ta, "alice", actorURL, workoutID, nil)

	likePath := "/api/v1/workouts/" + workoutID + "/likes?owner=bob"
	w := ta.doJSON(t, http.MethodPost, likePath, nil, aliceToken)
	expectStatus(t, w, http.StatusOK)
	state := decodeObject(t, w)
	if state["count"] != float64(1) || state["liked_by_me"] != true {
		t.Fatalf("federated like state: %#v", state)
	}

	cachedLikes, err := ta.app.Likes.GetFederated("alice", "bob@remote.example", workoutID)
	if err != nil || cachedLikes.Likes != 1 {
		t.Fatalf("federated likes cache: %#v err=%v", cachedLikes, err)
	}
	acts := transport.activities()
	if len(acts) < 1 || acts[0]["type"] != "Like" {
		t.Fatalf("expected Like delivery, got %#v", acts)
	}

	commentPath := "/api/v1/workouts/" + workoutID + "/comments?owner=bob"
	w = ta.doJSON(t, http.MethodPost, commentPath, map[string]string{"text": "remote hi"}, aliceToken)
	expectStatus(t, w, http.StatusOK)
	created := decodeObject(t, w)
	if created["count"] != float64(1) {
		t.Fatalf("federated comment: %#v", created)
	}
	commentID, _ := created["comment"].(map[string]any)["id"].(string)
	if commentID == "" {
		t.Fatal("expected comment id")
	}

	cachedComments, err := ta.app.Comments.GetFederated("alice", "bob@remote.example", workoutID)
	if err != nil || cachedComments.CommentsNum != 1 {
		t.Fatalf("federated comments cache: %#v err=%v", cachedComments, err)
	}

	w = ta.doJSON(t, http.MethodGet, "/api/v1/workouts/"+workoutID+"?owner=bob", nil, aliceToken)
	expectStatus(t, w, http.StatusOK)
	detail := decodeObject(t, w)
	if detail["likes_count"] != float64(1) || detail["comments_count"] != float64(1) {
		t.Fatalf("federated detail summary: %#v", detail)
	}

	w = ta.doJSON(t, http.MethodDelete, "/api/v1/workouts/"+workoutID+"/comments/"+commentID+"?owner=bob", nil, aliceToken)
	expectStatus(t, w, http.StatusOK)
	foundDelete := false
	for _, act := range transport.activities() {
		if act["type"] == "Delete" {
			foundDelete = true
			break
		}
	}
	if !foundDelete {
		t.Fatalf("expected Delete Note delivery, got %#v", transport.activities())
	}
}

func TestFederationInboxCreateCachesLikesAndComments(t *testing.T) {
	ta := setupFederationTestApp(t)
	ta.register(t, "alice", "alice@example.com", "password12")
	aliceToken, _ := ta.login(t, "alice@example.com", "password12")

	actorURL := "https://remote.example/users/bob"
	seedRemoteFollow(t, ta, "alice", "bob@remote.example", actorURL)
	workoutID := "33333333"
	postFederatedWorkoutCreate(t, ta, "alice", actorURL, workoutID, map[string]any{
		"likesCount": 2,
		"likedUsers": []any{
			map[string]any{
				"handle": "carol@other.test", "nickname": "carol", "name": "Carol", "is_local": false,
			},
			map[string]any{
				"handle": "dave@other.test", "nickname": "dave", "name": "Dave", "is_local": false,
			},
		},
		"commentsCount": 1,
		"comments": []any{
			map[string]any{
				"id": "c1", "datetime": "2026-08-06T12:00:00Z", "text": "Great!",
				"noteId": "https://other.test/users/carol/notes/c1",
				"user": map[string]any{
					"handle": "carol@other.test", "nickname": "carol", "name": "Carol", "is_local": true,
				},
			},
		},
	})

	w := ta.doJSON(t, http.MethodGet, "/api/v1/workouts/"+workoutID+"?owner=bob", nil, aliceToken)
	expectStatus(t, w, http.StatusOK)
	detail := decodeObject(t, w)
	if detail["likes_count"] != float64(2) {
		t.Fatalf("likes_count from snapshot: %#v", detail)
	}
	if detail["comments_count"] != float64(1) {
		t.Fatalf("comments_count from snapshot: %#v", detail)
	}

	cachedLikes, err := ta.app.Likes.GetFederated("alice", "bob@remote.example", workoutID)
	if err != nil || cachedLikes.Likes != 2 {
		t.Fatalf("likes cache: %#v err=%v", cachedLikes, err)
	}
	cachedComments, err := ta.app.Comments.GetFederated("alice", "bob@remote.example", workoutID)
	if err != nil || cachedComments.CommentsNum != 1 || cachedComments.Comments[0].User.IsLocal {
		t.Fatalf("comments cache (is_local recomputed): %#v err=%v", cachedComments, err)
	}
}

func TestOwnerDeleteRemoteCommentDeliversDeleteNote(t *testing.T) {
	ta := setupFederationTestApp(t)
	transport := &captureDeliveryRT{}
	ta.app.SetFederationHTTPClient(&http.Client{Transport: transport})

	ta.register(t, "alice", "alice@example.com", "password12")
	aliceToken, _ := ta.login(t, "alice@example.com", "password12")

	w := ta.doJSON(t, http.MethodPost, "/api/v1/workouts", map[string]any{
		"name": "Alice run", "sport_type": "Run", "start_date": "2026-07-08T10:00:00Z",
	}, aliceToken)
	expectStatus(t, w, http.StatusCreated)
	workoutID, _ := decodeObject(t, w)["id"].(string)

	noteID := "https://remote.example/users/carol/notes/c-remote-1"
	if err := ta.app.Comments.PutLocal("alice", workoutID, &workouts.WorkoutComments{
		Comments: []workouts.WorkoutComment{{
			ID: "c-remote-1",
			User: workouts.WorkoutLikeUser{
				Handle: "carol@remote.example", Nickname: "carol", Name: "Carol", IsLocal: false,
			},
			Datetime: time.Date(2026, 8, 6, 12, 0, 0, 0, time.UTC),
			Text:     "from remote",
			NoteID:   noteID,
		}},
	}); err != nil {
		t.Fatal(err)
	}

	w = ta.doJSON(t, http.MethodDelete, "/api/v1/workouts/"+workoutID+"/comments/c-remote-1", nil, aliceToken)
	expectStatus(t, w, http.StatusOK)
	if decodeObject(t, w)["count"] != float64(0) {
		t.Fatalf("after owner delete: %s", w.Body.String())
	}

	acts := transport.activities()
	if len(acts) != 1 || acts[0]["type"] != "Delete" {
		t.Fatalf("expected Delete delivery to remote author, got %#v", acts)
	}
	obj, _ := acts[0]["object"].(map[string]any)
	if obj["id"] != noteID || obj["type"] != "Note" {
		t.Fatalf("delete object: %#v", acts[0]["object"])
	}
	if !strings.Contains(transport.requests[0].URL.String(), "/users/carol/inbox") {
		t.Fatalf("inbox url = %s", transport.requests[0].URL.String())
	}
}
