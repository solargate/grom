package v1_test

import (
	"net/http"
	"testing"

	"github.com/solargate/grom/internal/config"
	"github.com/solargate/grom/internal/auth/pat"
)

func TestPATCreateListRevokeAndAPIAccess(t *testing.T) {
	ta := setupTestApp(t)
	ta.register(t, "alice", "alice@example.com", "password12")
	jwt, _ := ta.login(t, "alice@example.com", "password12")

	w := ta.doJSON(t, http.MethodPost, "/api/v1/auth/pat", map[string]any{
		"name":   "Script",
		"scopes": []string{pat.ScopeWorkoutsRead, pat.ScopeEquipmentRead},
	}, jwt)
	expectStatus(t, w, http.StatusCreated)
	created := decodeObject(t, w)
	rawToken, _ := created["token"].(string)
	patObj, _ := created["pat"].(map[string]any)
	patID, _ := patObj["id"].(string)
	if rawToken == "" || patID == "" {
		t.Fatalf("unexpected create response: %#v", created)
	}

	w = ta.doJSON(t, http.MethodGet, "/api/v1/auth/pat", nil, jwt)
	expectStatus(t, w, http.StatusOK)
	list := decodeList(t, w)
	if len(list) != 1 || list[0]["id"] != patID {
		t.Fatalf("unexpected list: %#v", list)
	}

	w = ta.doJSON(t, http.MethodGet, "/api/v1/workouts?scope=own", nil, rawToken)
	expectStatus(t, w, http.StatusOK)

	w = ta.doJSON(t, http.MethodPost, "/api/v1/workouts", map[string]any{
		"name":       "Blocked",
		"sport_type": "Run",
		"start_date": "2026-07-08T10:00:00Z",
	}, rawToken)
	expectStatus(t, w, http.StatusForbidden)

	w = ta.doJSON(t, http.MethodGet, "/api/v1/auth/pat", nil, rawToken)
	expectStatus(t, w, http.StatusUnauthorized)

	w = ta.doJSON(t, http.MethodDelete, "/api/v1/auth/pat/"+patID, nil, jwt)
	expectStatus(t, w, http.StatusNoContent)

	w = ta.doJSON(t, http.MethodGet, "/api/v1/workouts?scope=own", nil, rawToken)
	expectStatus(t, w, http.StatusUnauthorized)
}

func TestPATWriteScope(t *testing.T) {
	ta := setupTestApp(t)
	ta.register(t, "alice", "alice@example.com", "password12")
	jwt, _ := ta.login(t, "alice@example.com", "password12")

	w := ta.doJSON(t, http.MethodPost, "/api/v1/auth/pat", map[string]any{
		"name":   "Writer",
		"scopes": []string{pat.ScopeWorkoutsRead, pat.ScopeWorkoutsWrite},
	}, jwt)
	expectStatus(t, w, http.StatusCreated)
	rawToken, _ := decodeObject(t, w)["token"].(string)

	w = ta.doJSON(t, http.MethodPost, "/api/v1/workouts", map[string]any{
		"name":       "PAT run",
		"sport_type": "Run",
		"start_date": "2026-07-08T10:00:00Z",
	}, rawToken)
	expectStatus(t, w, http.StatusCreated)

	w = ta.doJSON(t, http.MethodGet, "/api/v1/workouts?scope=own", nil, rawToken)
	expectStatus(t, w, http.StatusOK)
}

func TestPATMaxTokens(t *testing.T) {
	ta := setupTestApp(t)
	ta.register(t, "alice", "alice@example.com", "password12")
	jwt, _ := ta.login(t, "alice@example.com", "password12")

	for i := 0; i < pat.MaxTokensPerUser; i++ {
		w := ta.doJSON(t, http.MethodPost, "/api/v1/auth/pat", map[string]any{
			"name":   "Token",
			"scopes": []string{pat.ScopeWorkoutsRead},
		}, jwt)
		expectStatus(t, w, http.StatusCreated)
	}

	w := ta.doJSON(t, http.MethodPost, "/api/v1/auth/pat", map[string]any{
		"name":   "Too many",
		"scopes": []string{pat.ScopeWorkoutsRead},
	}, jwt)
	expectStatus(t, w, http.StatusConflict)
}

func TestPATAcrossStorageDrivers(t *testing.T) {
	for _, driver := range []string{"file", "bbolt"} {
		t.Run(driver, func(t *testing.T) {
			ta := setupTestAppWithDriver(t, config.StorageDriver(driver))
			ta.register(t, "alice", "alice@example.com", "password12")
			jwt, _ := ta.login(t, "alice@example.com", "password12")

			w := ta.doJSON(t, http.MethodPost, "/api/v1/auth/pat", map[string]any{
				"name":   "Script",
				"scopes": []string{pat.ScopeWorkoutsRead},
			}, jwt)
			expectStatus(t, w, http.StatusCreated)
			rawToken, _ := decodeObject(t, w)["token"].(string)

			w = ta.doJSON(t, http.MethodGet, "/api/v1/workouts?scope=own", nil, rawToken)
			expectStatus(t, w, http.StatusOK)
		})
	}
}
