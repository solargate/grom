package v1_test

import (
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/solargate/grom/internal/data"
)

func TestDeleteAccountRequiresPasswordAndPurges(t *testing.T) {
	ta := setupTestApp(t)
	ta.register(t, "alice", "alice@example.com", "password12")
	token, _ := ta.login(t, "alice@example.com", "password12")

	w := ta.doJSON(t, http.MethodPost, "/api/v1/workouts", map[string]any{
		"name":            "Morning run",
		"sport_type":      "Run",
		"start_date":      time.Now().UTC().Format(time.RFC3339),
		"duration_seconds": 600,
	}, token)
	expectStatus(t, w, http.StatusCreated)

	w = ta.doJSON(t, http.MethodDelete, "/api/v1/auth/me", map[string]string{
		"password": "wrong-password",
	}, token)
	expectStatus(t, w, http.StatusForbidden)

	w = ta.doJSON(t, http.MethodGet, "/api/v1/auth/me", nil, token)
	expectStatus(t, w, http.StatusOK)

	w = ta.doJSON(t, http.MethodDelete, "/api/v1/auth/me", map[string]string{
		"password": "password12",
	}, token)
	expectStatus(t, w, http.StatusNoContent)

	w = ta.doJSON(t, http.MethodGet, "/api/v1/auth/me", nil, token)
	expectStatus(t, w, http.StatusUnauthorized)

	w = ta.doJSON(t, http.MethodPost, "/api/v1/auth/login", map[string]string{
		"email":    "alice@example.com",
		"password": "password12",
	}, "")
	expectStatus(t, w, http.StatusUnauthorized)

	userDir := data.UserDir(ta.dir, "alice")
	if _, err := os.Stat(userDir); !os.IsNotExist(err) {
		t.Fatalf("expected user dir removed, stat err=%v path=%s", err, userDir)
	}

	registered := ta.register(t, "alice", "alice@example.com", "password12")
	if registered["nickname"] != "alice" {
		t.Fatalf("re-register failed: %#v", registered)
	}
}

func TestDeleteAccountRejectsPAT(t *testing.T) {
	ta := setupTestApp(t)
	ta.register(t, "alice", "alice@example.com", "password12")
	token, _ := ta.login(t, "alice@example.com", "password12")

	w := ta.doJSON(t, http.MethodPost, "/api/v1/auth/pat", map[string]any{
		"name":   "cli",
		"scopes": []string{"workouts:read"},
	}, token)
	expectStatus(t, w, http.StatusCreated)
	patToken, _ := decodeObject(t, w)["token"].(string)
	if patToken == "" {
		t.Fatal("expected pat token")
	}

	w = ta.doJSON(t, http.MethodDelete, "/api/v1/auth/me", map[string]string{
		"password": "password12",
	}, patToken)
	expectStatus(t, w, http.StatusUnauthorized)
}

func TestDeleteAccountRemovesLikesOnOthersWorkouts(t *testing.T) {
	ta := setupTestApp(t)
	ta.register(t, "alice", "alice@example.com", "password12")
	ta.register(t, "bob", "bob@example.com", "password12")
	aliceToken, _ := ta.login(t, "alice@example.com", "password12")
	bobToken, _ := ta.login(t, "bob@example.com", "password12")

	w := ta.doJSON(t, http.MethodPost, "/api/v1/workouts", map[string]any{
		"name":             "Alice run",
		"sport_type":       "Run",
		"start_date":       time.Now().UTC().Format(time.RFC3339),
		"duration_seconds": 600,
	}, aliceToken)
	expectStatus(t, w, http.StatusCreated)
	workoutID, _ := decodeObject(t, w)["id"].(string)

	w = ta.doJSON(t, http.MethodPost, "/api/v1/social/follow", map[string]string{
		"handle": "alice",
	}, bobToken)
	expectStatus(t, w, http.StatusCreated)

	w = ta.doJSON(t, http.MethodPost, "/api/v1/workouts/"+workoutID+"/likes?owner=alice", nil, bobToken)
	expectStatus(t, w, http.StatusOK)

	w = ta.doJSON(t, http.MethodDelete, "/api/v1/auth/me", map[string]string{
		"password": "password12",
	}, bobToken)
	expectStatus(t, w, http.StatusNoContent)

	w = ta.doJSON(t, http.MethodGet, "/api/v1/workouts/"+workoutID+"/likes?owner=alice", nil, aliceToken)
	expectStatus(t, w, http.StatusOK)
	likes := decodeObject(t, w)
	if likes["count"] != float64(0) {
		t.Fatalf("expected likes cleared, got %#v", likes)
	}
}
