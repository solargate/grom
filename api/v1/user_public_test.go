package v1_test

import (
	"net/http"
	"net/url"
	"testing"

	"github.com/solargate/grom/internal/auth/pat"
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
	following := decodeList(t, w)
	if following == nil {
		t.Fatal("expected following list (possibly empty)")
	}

	w = ta.doJSON(t, http.MethodGet, "/api/v1/users/alice/followers", nil, bobToken)
	expectStatus(t, w, http.StatusOK)
	followers := decodeList(t, w)
	if len(followers) < 1 {
		t.Fatalf("expected bob among followers: %#v", followers)
	}
	foundBob := false
	for _, f := range followers {
		if f["follower_nickname"] == "bob" {
			foundBob = true
			if f["follower_is_local"] != true {
				t.Fatalf("expected local follower: %#v", f)
			}
			break
		}
	}
	if !foundBob {
		t.Fatalf("bob not in followers: %#v", followers)
	}
}

func TestUserPublicProfileUnauthorized(t *testing.T) {
	ta := setupTestApp(t)
	ta.register(t, "alice", "alice@example.com", "password12")
	w := ta.doJSON(t, http.MethodGet, "/api/v1/users/alice", nil, "")
	expectStatus(t, w, http.StatusUnauthorized)

	w = ta.doJSON(t, http.MethodGet, "/api/v1/users/alice/workouts", nil, "")
	expectStatus(t, w, http.StatusUnauthorized)

	w = ta.doJSON(t, http.MethodGet, "/api/v1/users/alice/following", nil, "")
	expectStatus(t, w, http.StatusUnauthorized)

	w = ta.doJSON(t, http.MethodGet, "/api/v1/users/alice/followers", nil, "")
	expectStatus(t, w, http.StatusUnauthorized)
}

func TestUserPublicProfileNotFoundAndInvalidHandle(t *testing.T) {
	ta := setupTestApp(t)
	ta.register(t, "alice", "alice@example.com", "password12")
	token, _ := ta.login(t, "alice@example.com", "password12")

	w := ta.doJSON(t, http.MethodGet, "/api/v1/users/nobody", nil, token)
	expectStatus(t, w, http.StatusNotFound)

	// Trailing @ → empty domain → invalid handle (slash cannot be used: Gin splits the path).
	invalid := url.PathEscape("alice@")
	w = ta.doJSON(t, http.MethodGet, "/api/v1/users/"+invalid, nil, token)
	expectStatus(t, w, http.StatusBadRequest)

	w = ta.doJSON(t, http.MethodGet, "/api/v1/users/"+invalid+"/workouts", nil, token)
	expectStatus(t, w, http.StatusBadRequest)
}

func TestUserPublicRemoteRequiresFederation(t *testing.T) {
	ta := setupTestApp(t)
	ta.register(t, "alice", "alice@example.com", "password12")
	token, _ := ta.login(t, "alice@example.com", "password12")

	escaped := url.PathEscape("bob@remote.example")
	w := ta.doJSON(t, http.MethodGet, "/api/v1/users/"+escaped, nil, token)
	expectStatus(t, w, http.StatusBadRequest)

	w = ta.doJSON(t, http.MethodGet, "/api/v1/users/"+escaped+"/workouts", nil, token)
	expectStatus(t, w, http.StatusBadRequest)
}

func TestUserPublicWorkoutsRejectsPAT(t *testing.T) {
	ta := setupTestApp(t)
	ta.register(t, "alice", "alice@example.com", "password12")
	ta.register(t, "bob", "bob@example.com", "password12")
	aliceJWT, _ := ta.login(t, "alice@example.com", "password12")
	bobJWT, _ := ta.login(t, "bob@example.com", "password12")

	w := ta.doJSON(t, http.MethodPost, "/api/v1/workouts", map[string]any{
		"name": "Alice run", "sport_type": "Run", "start_date": "2026-07-08T10:00:00Z",
		"duration_seconds": 1800, "distance": 5000,
	}, aliceJWT)
	expectStatus(t, w, http.StatusCreated)

	rawPAT, _ := createPAT(t, ta, bobJWT, "Reader", []string{pat.ScopeWorkoutsRead}, nil)
	// AuthRequired rejects PATs before the handler (JWT-only social/profile routes).
	w = ta.doJSON(t, http.MethodGet, "/api/v1/users/alice/workouts", nil, rawPAT)
	expectStatus(t, w, http.StatusUnauthorized)

	w = ta.doJSON(t, http.MethodGet, "/api/v1/users/alice", nil, rawPAT)
	expectStatus(t, w, http.StatusUnauthorized)
}

func TestUserPublicWorkoutsInvalidLimitAndCursor(t *testing.T) {
	ta := setupTestApp(t)
	ta.register(t, "alice", "alice@example.com", "password12")
	ta.register(t, "bob", "bob@example.com", "password12")
	token, _ := ta.login(t, "bob@example.com", "password12")

	w := ta.doJSON(t, http.MethodGet, "/api/v1/users/alice/workouts?limit=0", nil, token)
	expectStatus(t, w, http.StatusBadRequest)

	w = ta.doJSON(t, http.MethodGet, "/api/v1/users/alice/workouts?cursor=not-a-cursor", nil, token)
	expectStatus(t, w, http.StatusBadRequest)
}

func TestUserPublicLocalFollowingContent(t *testing.T) {
	ta := setupTestApp(t)
	ta.register(t, "alice", "alice@example.com", "password12")
	ta.register(t, "bob", "bob@example.com", "password12")
	ta.register(t, "carol", "carol@example.com", "password12")
	aliceToken, _ := ta.login(t, "alice@example.com", "password12")
	bobToken, _ := ta.login(t, "bob@example.com", "password12")

	w := ta.doJSON(t, http.MethodPost, "/api/v1/social/follow", map[string]string{"handle": "carol"}, aliceToken)
	expectStatus(t, w, http.StatusCreated)

	w = ta.doJSON(t, http.MethodGet, "/api/v1/users/alice/following", nil, bobToken)
	expectStatus(t, w, http.StatusOK)
	following := decodeList(t, w)
	if len(following) != 1 {
		t.Fatalf("expected 1 following, got %#v", following)
	}
	if following[0]["target_nickname"] != "carol" || following[0]["status"] != "active" {
		t.Fatalf("unexpected following entry: %#v", following[0])
	}
}
