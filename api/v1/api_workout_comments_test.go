package v1_test

import (
	"net/http"
	"testing"
)

func TestWorkoutCommentsCRUD(t *testing.T) {
	ta := setupTestApp(t)
	ta.register(t, "alice", "alice@example.com", "password12")
	ta.register(t, "bob", "bob@example.com", "password12")
	aliceToken, _ := ta.login(t, "alice@example.com", "password12")
	bobToken, _ := ta.login(t, "bob@example.com", "password12")

	w := ta.doJSON(t, http.MethodPost, "/api/v1/social/follow", map[string]string{
		"handle": "bob",
	}, aliceToken)
	expectStatus(t, w, http.StatusCreated)

	w = ta.doJSON(t, http.MethodPost, "/api/v1/workouts", map[string]any{
		"name":             "Bob run",
		"sport_type":       "Run",
		"start_date":       "2026-07-08T10:00:00Z",
		"duration_seconds": 1800,
		"distance":         5000,
	}, bobToken)
	expectStatus(t, w, http.StatusCreated)
	workoutID, _ := decodeObject(t, w)["id"].(string)

	path := "/api/v1/workouts/" + workoutID + "/comments?owner=bob"

	w = ta.doJSON(t, http.MethodPost, path, map[string]string{"text": "  Nice pace!  "}, aliceToken)
	expectStatus(t, w, http.StatusOK)
	created := decodeObject(t, w)
	if created["count"] != float64(1) {
		t.Fatalf("count: %#v", created)
	}
	comment, _ := created["comment"].(map[string]any)
	commentID, _ := comment["id"].(string)
	if commentID == "" || comment["text"] != "Nice pace!" {
		t.Fatalf("comment: %#v", comment)
	}
	if comment["can_delete"] != true {
		t.Fatalf("author can_delete: %#v", comment)
	}

	// Own workout can be commented.
	w = ta.doJSON(t, http.MethodPost, "/api/v1/workouts/"+workoutID+"/comments", map[string]string{
		"text": "Thanks!",
	}, bobToken)
	expectStatus(t, w, http.StatusOK)
	if decodeObject(t, w)["count"] != float64(2) {
		t.Fatalf("bob comment count: %#v", decodeObject(t, w))
	}

	w = ta.doJSON(t, http.MethodGet, path, nil, aliceToken)
	expectStatus(t, w, http.StatusOK)
	list := decodeObject(t, w)
	if list["count"] != float64(2) {
		t.Fatalf("list: %#v", list)
	}

	w = ta.doJSON(t, http.MethodGet, "/api/v1/workouts/"+workoutID+"?owner=bob", nil, aliceToken)
	expectStatus(t, w, http.StatusOK)
	if decodeObject(t, w)["comments_count"] != float64(2) {
		t.Fatalf("summary: %#v", decodeObject(t, w))
	}

	// Empty / too long rejected.
	w = ta.doJSON(t, http.MethodPost, path, map[string]string{"text": "   "}, aliceToken)
	expectStatus(t, w, http.StatusBadRequest)
	long := make([]byte, 1001)
	for i := range long {
		long[i] = 'a'
	}
	w = ta.doJSON(t, http.MethodPost, path, map[string]string{"text": string(long)}, aliceToken)
	expectStatus(t, w, http.StatusBadRequest)

	// Owner can delete alice's comment.
	deletePath := "/api/v1/workouts/" + workoutID + "/comments/" + commentID + "?owner=bob"
	w = ta.doJSON(t, http.MethodDelete, deletePath, nil, bobToken)
	expectStatus(t, w, http.StatusOK)
	if decodeObject(t, w)["count"] != float64(1) {
		t.Fatalf("after owner delete: %#v", decodeObject(t, w))
	}
}

func TestWorkoutCommentDeleteForbidden(t *testing.T) {
	ta := setupTestApp(t)
	ta.register(t, "alice", "alice@example.com", "password12")
	ta.register(t, "bob", "bob@example.com", "password12")
	ta.register(t, "carol", "carol@example.com", "password12")
	aliceToken, _ := ta.login(t, "alice@example.com", "password12")
	bobToken, _ := ta.login(t, "bob@example.com", "password12")
	carolToken, _ := ta.login(t, "carol@example.com", "password12")

	w := ta.doJSON(t, http.MethodPost, "/api/v1/social/follow", map[string]string{"handle": "bob"}, aliceToken)
	expectStatus(t, w, http.StatusCreated)
	w = ta.doJSON(t, http.MethodPost, "/api/v1/social/follow", map[string]string{"handle": "bob"}, carolToken)
	expectStatus(t, w, http.StatusCreated)

	w = ta.doJSON(t, http.MethodPost, "/api/v1/workouts", map[string]any{
		"name":             "Bob run",
		"sport_type":       "Run",
		"start_date":       "2026-07-08T10:00:00Z",
		"duration_seconds": 1800,
		"distance":         5000,
	}, bobToken)
	expectStatus(t, w, http.StatusCreated)
	workoutID, _ := decodeObject(t, w)["id"].(string)
	path := "/api/v1/workouts/" + workoutID + "/comments?owner=bob"

	w = ta.doJSON(t, http.MethodPost, path, map[string]string{"text": "hi"}, aliceToken)
	expectStatus(t, w, http.StatusOK)
	commentID, _ := decodeObject(t, w)["comment"].(map[string]any)["id"].(string)

	w = ta.doJSON(t, http.MethodDelete, "/api/v1/workouts/"+workoutID+"/comments/"+commentID+"?owner=bob", nil, carolToken)
	expectStatus(t, w, http.StatusForbidden)
}

func TestWorkoutCommentAuthorDeleteAndCanDeleteFlags(t *testing.T) {
	ta := setupTestApp(t)
	ta.register(t, "alice", "alice@example.com", "password12")
	ta.register(t, "bob", "bob@example.com", "password12")
	ta.register(t, "carol", "carol@example.com", "password12")
	aliceToken, _ := ta.login(t, "alice@example.com", "password12")
	bobToken, _ := ta.login(t, "bob@example.com", "password12")
	carolToken, _ := ta.login(t, "carol@example.com", "password12")

	w := ta.doJSON(t, http.MethodPost, "/api/v1/social/follow", map[string]string{"handle": "bob"}, aliceToken)
	expectStatus(t, w, http.StatusCreated)
	w = ta.doJSON(t, http.MethodPost, "/api/v1/social/follow", map[string]string{"handle": "bob"}, carolToken)
	expectStatus(t, w, http.StatusCreated)

	w = ta.doJSON(t, http.MethodPost, "/api/v1/workouts", map[string]any{
		"name":             "Bob run",
		"sport_type":       "Run",
		"start_date":       "2026-07-08T10:00:00Z",
		"duration_seconds": 1800,
		"distance":         5000,
	}, bobToken)
	expectStatus(t, w, http.StatusCreated)
	workoutID, _ := decodeObject(t, w)["id"].(string)
	path := "/api/v1/workouts/" + workoutID + "/comments?owner=bob"

	w = ta.doJSON(t, http.MethodPost, path, map[string]string{"text": "from alice"}, aliceToken)
	expectStatus(t, w, http.StatusOK)
	commentID, _ := decodeObject(t, w)["comment"].(map[string]any)["id"].(string)

	w = ta.doJSON(t, http.MethodGet, path, nil, aliceToken)
	expectStatus(t, w, http.StatusOK)
	aliceList := decodeObject(t, w)
	aliceComments, _ := aliceList["comments"].([]any)
	if len(aliceComments) != 1 {
		t.Fatalf("alice list: %#v", aliceList)
	}
	if aliceComments[0].(map[string]any)["can_delete"] != true {
		t.Fatalf("author can_delete: %#v", aliceComments[0])
	}

	w = ta.doJSON(t, http.MethodGet, path, nil, bobToken)
	expectStatus(t, w, http.StatusOK)
	bobComments, _ := decodeObject(t, w)["comments"].([]any)
	if bobComments[0].(map[string]any)["can_delete"] != true {
		t.Fatalf("owner can_delete: %#v", bobComments[0])
	}

	w = ta.doJSON(t, http.MethodGet, path, nil, carolToken)
	expectStatus(t, w, http.StatusOK)
	carolComments, _ := decodeObject(t, w)["comments"].([]any)
	if carolComments[0].(map[string]any)["can_delete"] != false {
		t.Fatalf("stranger can_delete: %#v", carolComments[0])
	}

	w = ta.doJSON(t, http.MethodDelete, "/api/v1/workouts/"+workoutID+"/comments/"+commentID+"?owner=bob", nil, aliceToken)
	expectStatus(t, w, http.StatusOK)
	if decodeObject(t, w)["count"] != float64(0) {
		t.Fatalf("after author delete: %#v", decodeObject(t, w))
	}
}

func TestWorkoutCommentErrors(t *testing.T) {
	ta := setupTestApp(t)
	ta.register(t, "alice", "alice@example.com", "password12")
	ta.register(t, "bob", "bob@example.com", "password12")
	aliceToken, _ := ta.login(t, "alice@example.com", "password12")
	bobToken, _ := ta.login(t, "bob@example.com", "password12")

	w := ta.doJSON(t, http.MethodPost, "/api/v1/workouts", map[string]any{
		"name":             "Bob run",
		"sport_type":       "Run",
		"start_date":       "2026-07-08T10:00:00Z",
		"duration_seconds": 1800,
		"distance":         5000,
	}, bobToken)
	expectStatus(t, w, http.StatusCreated)
	workoutID, _ := decodeObject(t, w)["id"].(string)

	w = ta.doJSON(t, http.MethodGet, "/api/v1/workouts/"+workoutID+"/comments?owner=bob", nil, "")
	expectStatus(t, w, http.StatusUnauthorized)

	w = ta.doJSON(t, http.MethodPost, "/api/v1/workouts/"+workoutID+"/comments?owner=bob", map[string]string{
		"text": "no follow",
	}, aliceToken)
	expectStatus(t, w, http.StatusNotFound)

	w = ta.doJSON(t, http.MethodGet, "/api/v1/workouts/zzzzzzzz/comments?owner=bob", nil, aliceToken)
	expectStatus(t, w, http.StatusNotFound)

	w = ta.doJSON(t, http.MethodPost, "/api/v1/social/follow", map[string]string{"handle": "bob"}, aliceToken)
	expectStatus(t, w, http.StatusCreated)

	path := "/api/v1/workouts/" + workoutID + "/comments?owner=bob"
	w = ta.doJSON(t, http.MethodPost, path, "not-json", aliceToken)
	expectStatus(t, w, http.StatusBadRequest)

	w = ta.doJSON(t, http.MethodPost, path, map[string]string{"text": "ok"}, aliceToken)
	expectStatus(t, w, http.StatusOK)

	w = ta.doJSON(t, http.MethodDelete, "/api/v1/workouts/"+workoutID+"/comments/missing-comment-id?owner=bob", nil, aliceToken)
	expectStatus(t, w, http.StatusNotFound)
}
