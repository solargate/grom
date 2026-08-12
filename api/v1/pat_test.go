package v1_test

import (
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/solargate/grom/internal/auth/pat"
	"github.com/solargate/grom/internal/config"
	"gopkg.in/yaml.v3"
)

func createPAT(t *testing.T, ta *testApp, jwt, name string, scopes []string, extra map[string]any) (rawToken, patID string) {
	t.Helper()
	body := map[string]any{
		"name":   name,
		"scopes": scopes,
	}
	for k, v := range extra {
		body[k] = v
	}
	w := ta.doJSON(t, http.MethodPost, "/api/v1/auth/pat", body, jwt)
	expectStatus(t, w, http.StatusCreated)
	created := decodeObject(t, w)
	rawToken, _ = created["token"].(string)
	patObj, _ := created["pat"].(map[string]any)
	patID, _ = patObj["id"].(string)
	if rawToken == "" || patID == "" {
		t.Fatalf("unexpected create response: %#v", created)
	}
	return rawToken, patID
}

func expirePATsInFileStore(t *testing.T, dataDir string) {
	t.Helper()
	path := filepath.Join(dataDir, "personal_access_tokens.yaml")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read pat file: %v", err)
	}
	var file struct {
		Tokens []map[string]any `yaml:"tokens"`
	}
	if err := yaml.Unmarshal(data, &file); err != nil {
		t.Fatalf("unmarshal pat file: %v", err)
	}
	past := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC).Format(time.RFC3339)
	for i := range file.Tokens {
		file.Tokens[i]["expires_at"] = past
	}
	out, err := yaml.Marshal(file)
	if err != nil {
		t.Fatalf("marshal pat file: %v", err)
	}
	if err := os.WriteFile(path, out, 0600); err != nil {
		t.Fatalf("write pat file: %v", err)
	}
}

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

func TestPATExpiredTokenRejected(t *testing.T) {
	ta := setupTestApp(t)
	ta.register(t, "alice", "alice@example.com", "password12")
	jwt, _ := ta.login(t, "alice@example.com", "password12")

	rawToken, _ := createPAT(t, ta, jwt, "Script", []string{pat.ScopeWorkoutsRead}, nil)
	expirePATsInFileStore(t, ta.dir)

	w := ta.doJSON(t, http.MethodGet, "/api/v1/workouts?scope=own", nil, rawToken)
	expectStatus(t, w, http.StatusUnauthorized)
}

func TestPATListHidesExpired(t *testing.T) {
	ta := setupTestApp(t)
	ta.register(t, "alice", "alice@example.com", "password12")
	jwt, _ := ta.login(t, "alice@example.com", "password12")

	createPAT(t, ta, jwt, "Script", []string{pat.ScopeWorkoutsRead}, nil)
	expirePATsInFileStore(t, ta.dir)

	w := ta.doJSON(t, http.MethodGet, "/api/v1/auth/pat", nil, jwt)
	expectStatus(t, w, http.StatusOK)
	list := decodeList(t, w)
	if len(list) != 0 {
		t.Fatalf("expected empty list, got %#v", list)
	}
}

func TestPATCrossUserRevokeNotFound(t *testing.T) {
	ta := setupTestApp(t)
	ta.register(t, "alice", "alice@example.com", "password12")
	ta.register(t, "bob", "bob@example.com", "password12")
	aliceJWT, _ := ta.login(t, "alice@example.com", "password12")
	bobJWT, _ := ta.login(t, "bob@example.com", "password12")

	_, patID := createPAT(t, ta, aliceJWT, "Script", []string{pat.ScopeWorkoutsRead}, nil)

	w := ta.doJSON(t, http.MethodDelete, "/api/v1/auth/pat/"+patID, nil, bobJWT)
	expectStatus(t, w, http.StatusNotFound)
}

func TestPATCannotAccessOtherOwnerWorkout(t *testing.T) {
	ta := setupTestApp(t)
	ta.register(t, "alice", "alice@example.com", "password12")
	ta.register(t, "bob", "bob@example.com", "password12")
	aliceJWT, _ := ta.login(t, "alice@example.com", "password12")
	bobJWT, _ := ta.login(t, "bob@example.com", "password12")

	w := ta.doJSON(t, http.MethodPost, "/api/v1/workouts", map[string]any{
		"name":       "Bob run",
		"sport_type": "Run",
		"start_date": "2026-07-08T10:00:00Z",
	}, bobJWT)
	expectStatus(t, w, http.StatusCreated)
	bobWorkoutID, _ := decodeObject(t, w)["id"].(string)

	rawToken, _ := createPAT(t, ta, aliceJWT, "Script", []string{pat.ScopeWorkoutsRead}, nil)

	w = ta.doJSON(t, http.MethodGet, "/api/v1/workouts/"+bobWorkoutID+"?owner=bob", nil, rawToken)
	expectStatus(t, w, http.StatusNotFound)
}

func TestPATListWorkoutsIgnoresFeedScope(t *testing.T) {
	ta := setupTestApp(t)
	ta.register(t, "alice", "alice@example.com", "password12")
	ta.register(t, "bob", "bob@example.com", "password12")
	aliceJWT, _ := ta.login(t, "alice@example.com", "password12")
	bobJWT, _ := ta.login(t, "bob@example.com", "password12")

	w := ta.doJSON(t, http.MethodPost, "/api/v1/workouts", map[string]any{
		"name":       "Bob run",
		"sport_type": "Run",
		"start_date": "2026-07-08T10:00:00Z",
	}, bobJWT)
	expectStatus(t, w, http.StatusCreated)

	w = ta.doJSON(t, http.MethodPost, "/api/v1/social/follow", map[string]string{
		"handle": "bob",
	}, aliceJWT)
	expectStatus(t, w, http.StatusCreated)

	w = ta.doJSON(t, http.MethodGet, "/api/v1/workouts?scope=feed&limit=20", nil, aliceJWT)
	expectStatus(t, w, http.StatusOK)
	feedItems, _, _ := decodeWorkoutPage(t, w)
	if len(feedItems) != 1 {
		t.Fatalf("alice JWT feed should include bob workout, got %#v", feedItems)
	}

	rawToken, _ := createPAT(t, ta, aliceJWT, "Script", []string{pat.ScopeWorkoutsRead}, nil)
	w = ta.doJSON(t, http.MethodGet, "/api/v1/workouts?scope=feed&limit=20", nil, rawToken)
	expectStatus(t, w, http.StatusOK)
	ownItems, _, _ := decodeWorkoutPage(t, w)
	if len(ownItems) != 0 {
		t.Fatalf("PAT should only return own workouts, got %#v", ownItems)
	}
}

func TestPATEquipmentReadScope(t *testing.T) {
	ta := setupTestApp(t)
	ta.register(t, "alice", "alice@example.com", "password12")
	jwt, _ := ta.login(t, "alice@example.com", "password12")

	w := ta.doJSON(t, http.MethodPost, "/api/v1/equipment", map[string]any{
		"type": "shoes",
		"name": "Road",
	}, jwt)
	expectStatus(t, w, http.StatusCreated)

	rawToken, _ := createPAT(t, ta, jwt, "Reader", []string{pat.ScopeEquipmentRead}, nil)

	w = ta.doJSON(t, http.MethodGet, "/api/v1/equipment", nil, rawToken)
	expectStatus(t, w, http.StatusOK)
	items := decodeList(t, w)
	if len(items) != 1 {
		t.Fatalf("expected equipment list, got %#v", items)
	}

	w = ta.doJSON(t, http.MethodPost, "/api/v1/equipment", map[string]any{
		"type": "bike",
		"name": "Road bike",
	}, rawToken)
	expectStatus(t, w, http.StatusForbidden)
}

func TestPATEquipmentWriteScope(t *testing.T) {
	ta := setupTestApp(t)
	ta.register(t, "alice", "alice@example.com", "password12")
	jwt, _ := ta.login(t, "alice@example.com", "password12")

	rawToken, _ := createPAT(t, ta, jwt, "Writer", []string{pat.ScopeEquipmentWrite}, nil)

	w := ta.doJSON(t, http.MethodPost, "/api/v1/equipment", map[string]any{
		"type": "shoes",
		"name": "Trail",
	}, rawToken)
	expectStatus(t, w, http.StatusCreated)
	eqID, _ := decodeObject(t, w)["id"].(string)

	w = ta.doJSON(t, http.MethodPut, "/api/v1/equipment/"+eqID, map[string]any{
		"type": "shoes",
		"name": "Trail updated",
	}, rawToken)
	expectStatus(t, w, http.StatusOK)

	w = ta.doJSON(t, http.MethodGet, "/api/v1/equipment", nil, rawToken)
	expectStatus(t, w, http.StatusForbidden)

	w = ta.doJSON(t, http.MethodDelete, "/api/v1/equipment/"+eqID, nil, rawToken)
	expectStatus(t, w, http.StatusNoContent)
}

func TestPATCreateValidation(t *testing.T) {
	ta := setupTestApp(t)
	ta.register(t, "alice", "alice@example.com", "password12")
	jwt, _ := ta.login(t, "alice@example.com", "password12")

	w := ta.doJSON(t, http.MethodPost, "/api/v1/auth/pat", map[string]any{
		"name":   "Script",
		"scopes": []string{"admin:all"},
	}, jwt)
	expectStatus(t, w, http.StatusBadRequest)

	w = ta.doJSON(t, http.MethodPost, "/api/v1/auth/pat", map[string]any{
		"name":   strings.Repeat("a", pat.MaxNameLen+1),
		"scopes": []string{pat.ScopeWorkoutsRead},
	}, jwt)
	expectStatus(t, w, http.StatusBadRequest)

	w = ta.doJSON(t, http.MethodPost, "/api/v1/auth/pat", map[string]any{
		"name":            "Script",
		"scopes":          []string{pat.ScopeWorkoutsRead},
		"expires_in_days": pat.MaxExpiresDays + 1,
	}, jwt)
	expectStatus(t, w, http.StatusBadRequest)

	w = ta.doJSON(t, http.MethodDelete, "/api/v1/auth/pat/missing-id", nil, jwt)
	expectStatus(t, w, http.StatusNotFound)
}
