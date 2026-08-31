package v1_test

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/solargate/grom/internal/config"
	"github.com/solargate/grom/internal/federation"
	"github.com/solargate/grom/internal/federation/httpsig"
)

func setupFederationTestAppDefaultAF(t *testing.T) *testApp {
	t.Helper()
	tlsDir := t.TempDir()
	certPath, keyPath := writeTestTLS(t, tlsDir)
	return setupTestAppWithConfig(t, func(cfg *config.Config) {
		cfg.Server.TLS.Mode = "static"
		cfg.Server.TLS.CertFile = certPath
		cfg.Server.TLS.KeyFile = keyPath
		cfg.Federation.Enabled = true
		cfg.Federation.Domain = "localhost"
		cfg.Federation.AutoAcceptFollows = false
		// AuthorizedFetch nil → default true
	})
}

func (ta *testApp) signedInstanceGET(t *testing.T, path string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(http.MethodGet, path, nil)
	req.Header.Set("Accept", "application/activity+json")
	ak, err := federation.LoadOrCreateInstanceActorKey(ta.app.Blobs)
	if err != nil {
		t.Fatalf("LoadOrCreateInstanceActorKey: %v", err)
	}
	if err := httpsig.SignGET(req, ak.Private, ak.KeyID); err != nil {
		t.Fatalf("SignGET: %v", err)
	}
	w := httptest.NewRecorder()
	ta.router.ServeHTTP(w, req)
	return w
}

func (ta *testApp) installInstanceActorKey(t *testing.T) {
	t.Helper()
	ak, err := federation.LoadOrCreateInstanceActorKey(ta.app.Blobs)
	if err != nil {
		t.Fatalf("LoadOrCreateInstanceActorKey: %v", err)
	}
	owner := "https://" + config.Cfg.Federation.Domain + "/actor"
	ta.app.SetFederationKeyResolver(federation.StaticKeyResolver{
		Keys: map[string]federation.StaticKey{
			ak.KeyID: {Public: &ak.Private.PublicKey, Owner: owner},
		},
	})
}

func TestAuthorizedFetchDefaultRequiresSignatureOnActor(t *testing.T) {
	ta := setupFederationTestAppDefaultAF(t)
	ta.register(t, "alice", "alice@example.com", "password12")
	ta.installInstanceActorKey(t)

	req := httptest.NewRequest(http.MethodGet, "/users/alice", nil)
	req.Header.Set("Accept", "application/activity+json")
	w := httptest.NewRecorder()
	ta.router.ServeHTTP(w, req)
	expectStatus(t, w, http.StatusUnauthorized)

	w = ta.signedInstanceGET(t, "/users/alice")
	expectStatus(t, w, http.StatusOK)
	actor := decodeObject(t, w)
	if actor["type"] != "Person" {
		t.Fatalf("actor: %#v", actor)
	}

	w = ta.doJSON(t, http.MethodGet, "/actor", nil, "")
	expectStatus(t, w, http.StatusNotFound) // needs Accept header

	req = httptest.NewRequest(http.MethodGet, "/actor", nil)
	req.Header.Set("Accept", "application/activity+json")
	w = httptest.NewRecorder()
	ta.router.ServeHTTP(w, req)
	expectStatus(t, w, http.StatusOK)
}

func TestAuthorizedFetchDefaultRequiresSignatureOnOutbox(t *testing.T) {
	ta := setupFederationTestAppDefaultAF(t)
	ta.register(t, "alice", "alice@example.com", "password12")
	ta.installInstanceActorKey(t)

	req := httptest.NewRequest(http.MethodGet, "/users/alice/outbox", nil)
	req.Header.Set("Accept", "application/activity+json")
	w := httptest.NewRecorder()
	ta.router.ServeHTTP(w, req)
	expectStatus(t, w, http.StatusUnauthorized)

	w = ta.signedInstanceGET(t, "/users/alice/outbox")
	expectStatus(t, w, http.StatusOK)
}

func TestSharedInboxLikeRoutesToWorkoutOwner(t *testing.T) {
	ta := setupFederationTestApp(t)
	ta.register(t, "alice", "alice@example.com", "password12")
	token, _ := ta.login(t, "alice@example.com", "password12")

	w := ta.doJSON(t, http.MethodPost, "/api/v1/workouts", map[string]any{
		"name":             "Local run",
		"sport_type":       "Run",
		"start_date":       time.Date(2026, 8, 1, 10, 0, 0, 0, time.UTC).Format(time.RFC3339),
		"duration_seconds": 1800,
		"distance":         5000,
	}, token)
	expectStatus(t, w, http.StatusCreated)
	created := decodeObject(t, w)
	workoutID, _ := created["id"].(string)
	if workoutID == "" {
		t.Fatal("missing workout id")
	}

	priv, keyID := newRemoteTestKey(t)
	ta.installRemoteKey(t, priv, keyID, remoteTestActor)

	objectID := "https://localhost/users/alice/workouts/" + workoutID
	likeObj := map[string]any{
		"@context": "https://www.w3.org/ns/activitystreams",
		"type":     "Like",
		"actor":    remoteTestActor,
		"object":   objectID,
		"id":       "https://remote.example/activities/like-1",
	}
	w = ta.postSignedActivity(t, "/inbox", likeObj, priv, keyID)
	expectStatus(t, w, http.StatusAccepted)

	w = ta.doJSON(t, http.MethodGet, "/api/v1/workouts/"+workoutID+"/likes", nil, token)
	expectStatus(t, w, http.StatusOK)
	likes := decodeObject(t, w)
	if likes["count"].(float64) != 1 {
		t.Fatalf("count = %#v", likes["count"])
	}
}

func TestSharedInboxUnsignedRejected(t *testing.T) {
	ta := setupFederationTestApp(t)
	body := []byte(`{"type":"Follow","actor":"https://remote.example/users/bob"}`)
	req := httptest.NewRequest(http.MethodPost, "/inbox", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/activity+json")
	w := httptest.NewRecorder()
	ta.router.ServeHTTP(w, req)
	expectStatus(t, w, http.StatusUnauthorized)
}
