package v1_test

import (
	"net/http"
	"testing"
)

func TestUserPublicProfileAndWorkoutsWithoutFollow(t *testing.T) {
	ta := setupTestApp(t)
	ta.register(t, "alice", "alice@example.com", "password12")
	ta.register(t, "bob", "bob@example.com", "password12")
	aliceToken, _ := ta.login(t, "alice@example.com", "password12")
	bobToken, _ := ta.login(t, "bob@example.com", "password12")

	w := ta.doJSON(t, http.MethodPost, "/api/v1/workouts", map[string]any{
		"name": "Alice run", "sport_type": "Run", "start_date": "2026-07-08T10:00:00Z",
		"duration_seconds": 1800, "distance": 5000,
	}, aliceToken)
	expectStatus(t, w, http.StatusCreated)

	w = ta.doJSON(t, http.MethodGet, "/api/v1/users/alice", nil, bobToken)
	expectStatus(t, w, http.StatusOK)
	got := decodeObject(t, w)
	if got["nickname"] != "alice" || got["is_local"] != true {
		t.Fatalf("unexpected profile: %#v", got)
	}
	if got["viewer_follow"] != nil {
		t.Fatalf("expected no viewer_follow, got %#v", got["viewer_follow"])
	}

	w = ta.doJSON(t, http.MethodPost, "/api/v1/social/follow", map[string]string{"handle": "alice"}, bobToken)
	expectStatus(t, w, http.StatusCreated)

	w = ta.doJSON(t, http.MethodGet, "/api/v1/users/alice", nil, bobToken)
	expectStatus(t, w, http.StatusOK)
	got = decodeObject(t, w)
	vf, _ := got["viewer_follow"].(map[string]any)
	if vf == nil || vf["status"] != "active" {
		t.Fatalf("expected viewer_follow active: %#v", got["viewer_follow"])
	}

	w = ta.doJSON(t, http.MethodGet, "/api/v1/users/alice/workouts", nil, bobToken)
	expectStatus(t, w, http.StatusOK)
	page := decodeObject(t, w)
	items, _ := page["items"].([]any)
	if len(items) != 1 {
		t.Fatalf("expected 1 workout, got %#v", page)
	}
	item, _ := items[0].(map[string]any)
	if item["name"] != "Alice run" {
		t.Fatalf("unexpected item: %#v", item)
	}
	if item["object_id"] == nil || item["object_id"] == "" {
		t.Fatalf("expected object_id: %#v", item)
	}

	w = ta.doJSON(t, http.MethodGet, "/api/v1/users/alice/following", nil, bobToken)
	expectStatus(t, w, http.StatusOK)
	w = ta.doJSON(t, http.MethodGet, "/api/v1/users/alice/followers", nil, bobToken)
	expectStatus(t, w, http.StatusOK)
	followers := decodeList(t, w)
	if len(followers) < 1 {
		t.Fatalf("expected bob among followers: %#v", followers)
	}
}

func TestUserPublicProfileUnauthorized(t *testing.T) {
	ta := setupTestApp(t)
	ta.register(t, "alice", "alice@example.com", "password12")
	w := ta.doJSON(t, http.MethodGet, "/api/v1/users/alice", nil, "")
	expectStatus(t, w, http.StatusUnauthorized)
}
