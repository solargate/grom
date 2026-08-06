package v1_test

import (
	"net/http"
	"testing"

	"github.com/solargate/grom/internal/config"
)

func TestLikeUnlikeLocalWorkout(t *testing.T) {
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
	if workoutID == "" {
		t.Fatal("expected workout id")
	}

	path := "/api/v1/workouts/" + workoutID + "/likes?owner=bob"

	w = ta.doJSON(t, http.MethodPost, path, nil, aliceToken)
	expectStatus(t, w, http.StatusOK)
	state := decodeObject(t, w)
	if state["count"] != float64(1) || state["liked_by_me"] != true {
		t.Fatalf("after like: %#v", state)
	}

	w = ta.doJSON(t, http.MethodGet, path, nil, aliceToken)
	expectStatus(t, w, http.StatusOK)
	list := decodeObject(t, w)
	if list["count"] != float64(1) {
		t.Fatalf("list count: %#v", list)
	}
	users, _ := list["users"].([]any)
	if len(users) != 1 {
		t.Fatalf("users: %#v", list["users"])
	}
	user, _ := users[0].(map[string]any)
	if user["nickname"] != "alice" || user["is_local"] != true {
		t.Fatalf("like user: %#v", user)
	}

	w = ta.doJSON(t, http.MethodGet, "/api/v1/workouts/"+workoutID+"?owner=bob", nil, aliceToken)
	expectStatus(t, w, http.StatusOK)
	detail := decodeObject(t, w)
	if detail["likes_count"] != float64(1) || detail["liked_by_me"] != true || detail["can_like"] != true {
		t.Fatalf("workout summary: %#v", detail)
	}

	w = ta.doJSON(t, http.MethodDelete, path, nil, aliceToken)
	expectStatus(t, w, http.StatusOK)
	state = decodeObject(t, w)
	if state["count"] != float64(0) || state["liked_by_me"] != false {
		t.Fatalf("after unlike: %#v", state)
	}
}

func TestCannotLikeOwnWorkout(t *testing.T) {
	ta := setupTestApp(t)
	ta.register(t, "alice", "alice@example.com", "password12")
	token, _ := ta.login(t, "alice@example.com", "password12")

	w := ta.doJSON(t, http.MethodPost, "/api/v1/workouts", map[string]any{
		"name":             "Own run",
		"sport_type":       "Run",
		"start_date":       "2026-07-08T10:00:00Z",
		"duration_seconds": 1800,
		"distance":         5000,
	}, token)
	expectStatus(t, w, http.StatusCreated)
	id, _ := decodeObject(t, w)["id"].(string)

	w = ta.doJSON(t, http.MethodPost, "/api/v1/workouts/"+id+"/likes", nil, token)
	expectStatus(t, w, http.StatusBadRequest)
	errBody := decodeObject(t, w)
	if errBody["error"] != "cannot like your own workout" {
		t.Fatalf("error body: %#v", errBody)
	}

	w = ta.doJSON(t, http.MethodGet, "/api/v1/workouts/"+id, nil, token)
	expectStatus(t, w, http.StatusOK)
	detail := decodeObject(t, w)
	if detail["can_like"] != false {
		t.Fatalf("own workout can_like: %#v", detail)
	}
}

func TestLikeUnlikeIdempotent(t *testing.T) {
	ta := setupTestApp(t)
	ta.register(t, "alice", "alice@example.com", "password12")
	ta.register(t, "bob", "bob@example.com", "password12")
	aliceToken, _ := ta.login(t, "alice@example.com", "password12")
	bobToken, _ := ta.login(t, "bob@example.com", "password12")

	w := ta.doJSON(t, http.MethodPost, "/api/v1/social/follow", map[string]string{"handle": "bob"}, aliceToken)
	expectStatus(t, w, http.StatusCreated)

	w = ta.doJSON(t, http.MethodPost, "/api/v1/workouts", map[string]any{
		"name":             "Bob run",
		"sport_type":       "Run",
		"start_date":       "2026-07-08T10:00:00Z",
		"duration_seconds": 1800,
		"distance":         5000,
	}, bobToken)
	expectStatus(t, w, http.StatusCreated)
	id, _ := decodeObject(t, w)["id"].(string)
	path := "/api/v1/workouts/" + id + "/likes?owner=bob"

	w = ta.doJSON(t, http.MethodDelete, path, nil, aliceToken)
	expectStatus(t, w, http.StatusOK)
	if decodeObject(t, w)["count"] != float64(0) {
		t.Fatalf("unlike when empty: %s", w.Body.String())
	}

	w = ta.doJSON(t, http.MethodPost, path, nil, aliceToken)
	expectStatus(t, w, http.StatusOK)
	w = ta.doJSON(t, http.MethodPost, path, nil, aliceToken)
	expectStatus(t, w, http.StatusOK)
	if decodeObject(t, w)["count"] != float64(1) {
		t.Fatalf("idempotent like: %s", w.Body.String())
	}
}

func TestGetLikesUnauthorizedAndNotFound(t *testing.T) {
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
	id, _ := decodeObject(t, w)["id"].(string)

	w = ta.doJSON(t, http.MethodGet, "/api/v1/workouts/"+id+"/likes?owner=bob", nil, "")
	expectStatus(t, w, http.StatusUnauthorized)

	w = ta.doJSON(t, http.MethodPost, "/api/v1/workouts/"+id+"/likes?owner=bob", nil, aliceToken)
	expectStatus(t, w, http.StatusNotFound)

	w = ta.doJSON(t, http.MethodGet, "/api/v1/workouts/zzzzzzzz/likes?owner=bob", nil, aliceToken)
	expectStatus(t, w, http.StatusNotFound)
}

func TestDeleteWorkoutRemovesLikes(t *testing.T) {
	ta := setupTestApp(t)
	ta.register(t, "alice", "alice@example.com", "password12")
	ta.register(t, "bob", "bob@example.com", "password12")
	aliceToken, _ := ta.login(t, "alice@example.com", "password12")
	bobToken, _ := ta.login(t, "bob@example.com", "password12")

	w := ta.doJSON(t, http.MethodPost, "/api/v1/social/follow", map[string]string{"handle": "bob"}, aliceToken)
	expectStatus(t, w, http.StatusCreated)

	w = ta.doJSON(t, http.MethodPost, "/api/v1/workouts", map[string]any{
		"name":             "Bob run",
		"sport_type":       "Run",
		"start_date":       "2026-07-08T10:00:00Z",
		"duration_seconds": 1800,
		"distance":         5000,
	}, bobToken)
	expectStatus(t, w, http.StatusCreated)
	id, _ := decodeObject(t, w)["id"].(string)

	w = ta.doJSON(t, http.MethodPost, "/api/v1/workouts/"+id+"/likes?owner=bob", nil, aliceToken)
	expectStatus(t, w, http.StatusOK)

	w = ta.doJSON(t, http.MethodDelete, "/api/v1/workouts/"+id, nil, bobToken)
	expectStatus(t, w, http.StatusNoContent)

	likes, err := ta.app.Likes.GetLocal("bob", id)
	if err == nil && likes.Likes != 0 {
		t.Fatalf("expected likes cleared after workout delete, got %#v", likes)
	}
}

func TestLikeLocalWorkoutBBoltDriver(t *testing.T) {
	ta := setupTestAppWithDriver(t, config.StorageDriverBBolt)
	ta.register(t, "alice", "alice@example.com", "password12")
	ta.register(t, "bob", "bob@example.com", "password12")
	aliceToken, _ := ta.login(t, "alice@example.com", "password12")
	bobToken, _ := ta.login(t, "bob@example.com", "password12")

	w := ta.doJSON(t, http.MethodPost, "/api/v1/social/follow", map[string]string{"handle": "bob"}, aliceToken)
	expectStatus(t, w, http.StatusCreated)

	w = ta.doJSON(t, http.MethodPost, "/api/v1/workouts", map[string]any{
		"name":             "Bob run",
		"sport_type":       "Run",
		"start_date":       "2026-07-08T10:00:00Z",
		"duration_seconds": 1800,
		"distance":         5000,
	}, bobToken)
	expectStatus(t, w, http.StatusCreated)
	id, _ := decodeObject(t, w)["id"].(string)

	path := "/api/v1/workouts/" + id + "/likes?owner=bob"
	w = ta.doJSON(t, http.MethodPost, path, nil, aliceToken)
	expectStatus(t, w, http.StatusOK)
	if decodeObject(t, w)["count"] != float64(1) {
		t.Fatalf("bbolt like: %s", w.Body.String())
	}
	w = ta.doJSON(t, http.MethodDelete, path, nil, aliceToken)
	expectStatus(t, w, http.StatusOK)
	if decodeObject(t, w)["count"] != float64(0) {
		t.Fatalf("bbolt unlike: %s", w.Body.String())
	}
}
