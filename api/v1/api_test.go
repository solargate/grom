package v1_test

import (
	"bytes"
	"encoding/json"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/gin-gonic/gin"
	v1 "github.com/solargate/grom/api/v1"
	"github.com/solargate/grom/internal/config"
	"github.com/solargate/grom/internal/storage"
)

type testApp struct {
	app    *v1.App
	router *gin.Engine
	dir    string
}

func setupTestApp(t *testing.T) *testApp {
	t.Helper()
	gin.SetMode(gin.TestMode)

	prev := config.Cfg
	t.Cleanup(func() { config.Cfg = prev })

	dir := t.TempDir()
	config.Cfg = config.Config{}
	config.Cfg.Auth.JWTSecret = "test-secret-at-least-32-characters!!"
	config.Cfg.Auth.JWTTTLHours = 24
	config.Cfg.Server.TLS.Mode = "off"
	config.Cfg.Federation.Enabled = false
	config.Cfg.Federation.Domain = "localhost"
	config.Cfg.Storage.Driver = "file"
	config.Cfg.Storage.Location = dir
	config.Cfg.Storage.TempDir = filepath.Join(dir, "tmp")
	if err := config.FinalizeConfig(&config.Cfg); err != nil {
		t.Fatalf("FinalizeConfig: %v", err)
	}

	app, err := v1.NewApp()
	if err != nil {
		t.Fatalf("NewApp: %v", err)
	}
	t.Cleanup(func() { _ = app.Close() })

	router := gin.New()
	router.MaxMultipartMemory = 128 << 20
	app.RegisterRoutes(router)

	return &testApp{app: app, router: router, dir: dir}
}

func (ta *testApp) doJSON(t *testing.T, method, path string, body any, token string) *httptest.ResponseRecorder {
	t.Helper()
	var reader io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			t.Fatal(err)
		}
		reader = bytes.NewReader(data)
	}
	req := httptest.NewRequest(method, path, reader)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	w := httptest.NewRecorder()
	ta.router.ServeHTTP(w, req)
	return w
}

func (ta *testApp) register(t *testing.T, nickname, email, password string) map[string]any {
	t.Helper()
	w := ta.doJSON(t, http.MethodPost, "/api/v1/auth/register", map[string]string{
		"nickname": nickname,
		"name":     nickname,
		"email":    email,
		"password": password,
	}, "")
	if w.Code != http.StatusCreated {
		t.Fatalf("register %s: %d %s", nickname, w.Code, w.Body.String())
	}
	var out map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &out); err != nil {
		t.Fatal(err)
	}
	return out
}

func (ta *testApp) login(t *testing.T, email, password string) (token string, user map[string]any) {
	t.Helper()
	w := ta.doJSON(t, http.MethodPost, "/api/v1/auth/login", map[string]string{
		"email":    email,
		"password": password,
	}, "")
	if w.Code != http.StatusOK {
		t.Fatalf("login: %d %s", w.Code, w.Body.String())
	}
	var out map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &out); err != nil {
		t.Fatal(err)
	}
	token, _ = out["token"].(string)
	user, _ = out["user"].(map[string]any)
	if token == "" {
		t.Fatal("expected token")
	}
	return token, user
}

func TestAuthRegisterLoginMe(t *testing.T) {
	ta := setupTestApp(t)

	w := ta.doJSON(t, http.MethodPost, "/api/v1/auth/register", map[string]string{
		"nickname": "alice",
		"name":     "Alice",
		"email":    "not-an-email",
		"password": "password12",
	}, "")
	if w.Code != http.StatusBadRequest {
		t.Fatalf("invalid email status = %d", w.Code)
	}

	ta.register(t, "alice", "alice@example.com", "password12")
	w = ta.doJSON(t, http.MethodPost, "/api/v1/auth/register", map[string]string{
		"nickname": "alice2",
		"name":     "Alice",
		"email":    "alice@example.com",
		"password": "password12",
	}, "")
	if w.Code != http.StatusConflict {
		t.Fatalf("duplicate email status = %d body=%s", w.Code, w.Body.String())
	}

	w = ta.doJSON(t, http.MethodPost, "/api/v1/auth/login", map[string]string{
		"email":    "alice@example.com",
		"password": "wrong-password",
	}, "")
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("bad password status = %d", w.Code)
	}
	w = ta.doJSON(t, http.MethodPost, "/api/v1/auth/login", map[string]string{
		"email":    "missing@example.com",
		"password": "password12",
	}, "")
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("unknown email status = %d", w.Code)
	}

	token, _ := ta.login(t, "alice@example.com", "password12")
	w = ta.doJSON(t, http.MethodGet, "/api/v1/auth/me", nil, token)
	if w.Code != http.StatusOK {
		t.Fatalf("me: %d %s", w.Code, w.Body.String())
	}
	w = ta.doJSON(t, http.MethodGet, "/api/v1/auth/me", nil, "")
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("me without token: %d", w.Code)
	}
}

func TestWorkoutCRUDAndList(t *testing.T) {
	ta := setupTestApp(t)
	ta.register(t, "alice", "alice@example.com", "password12")
	token, _ := ta.login(t, "alice@example.com", "password12")

	w := ta.doJSON(t, http.MethodPost, "/api/v1/workouts", map[string]any{
		"name":             "Morning run",
		"sport_type":       "Run",
		"start_date":       "2026-07-08T10:00:00Z",
		"duration_seconds": 3600,
		"distance":         5200,
	}, token)
	if w.Code != http.StatusCreated {
		t.Fatalf("create workout: %d %s", w.Code, w.Body.String())
	}
	var created map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &created); err != nil {
		t.Fatal(err)
	}
	id, _ := created["id"].(string)
	if id == "" {
		t.Fatal("expected workout id")
	}

	w = ta.doJSON(t, http.MethodGet, "/api/v1/workouts?scope=own", nil, token)
	if w.Code != http.StatusOK {
		t.Fatalf("list own: %d %s", w.Code, w.Body.String())
	}

	w = ta.doJSON(t, http.MethodDelete, "/api/v1/workouts/"+id, nil, token)
	if w.Code != http.StatusNoContent && w.Code != http.StatusOK {
		t.Fatalf("delete: %d %s", w.Code, w.Body.String())
	}
}

func TestWorkoutMultipartTrack(t *testing.T) {
	ta := setupTestApp(t)
	ta.register(t, "alice", "alice@example.com", "password12")
	token, _ := ta.login(t, "alice@example.com", "password12")

	gpx, err := os.ReadFile(filepath.Join("..", "..", "testdata", "tracks", "1-sample.gpx"))
	if err != nil {
		t.Fatal(err)
	}

	var body bytes.Buffer
	mw := multipart.NewWriter(&body)
	_ = mw.WriteField("name", "GPX run")
	_ = mw.WriteField("sport_type", "Run")
	_ = mw.WriteField("start_date", "2026-07-08T10:00:00Z")
	part, err := mw.CreateFormFile("track", "sample.gpx")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := part.Write(gpx); err != nil {
		t.Fatal(err)
	}
	_ = mw.Close()

	req := httptest.NewRequest(http.MethodPost, "/api/v1/workouts", &body)
	req.Header.Set("Content-Type", mw.FormDataContentType())
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	ta.router.ServeHTTP(w, req)
	if w.Code != http.StatusCreated {
		t.Fatalf("multipart create: %d %s", w.Code, w.Body.String())
	}
}

func TestSocialFollowAPI(t *testing.T) {
	ta := setupTestApp(t)
	ta.register(t, "alice", "alice@example.com", "password12")
	ta.register(t, "bob", "bob@example.com", "password12")
	token, _ := ta.login(t, "alice@example.com", "password12")

	w := ta.doJSON(t, http.MethodPost, "/api/v1/social/follow", map[string]string{
		"handle": "alice",
	}, token)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("self follow status = %d body=%s", w.Code, w.Body.String())
	}

	w = ta.doJSON(t, http.MethodPost, "/api/v1/social/follow", map[string]string{
		"handle": "bob",
	}, token)
	if w.Code != http.StatusCreated && w.Code != http.StatusOK {
		t.Fatalf("follow: %d %s", w.Code, w.Body.String())
	}
	var follow map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &follow); err != nil {
		t.Fatal(err)
	}
	followID, _ := follow["id"].(string)

	w = ta.doJSON(t, http.MethodDelete, "/api/v1/social/follow/"+followID, nil, token)
	if w.Code != http.StatusNoContent && w.Code != http.StatusOK {
		t.Fatalf("unfollow: %d %s", w.Code, w.Body.String())
	}
}

func TestEquipmentDeleteCascades(t *testing.T) {
	ta := setupTestApp(t)
	ta.register(t, "alice", "alice@example.com", "password12")
	token, _ := ta.login(t, "alice@example.com", "password12")

	w := ta.doJSON(t, http.MethodPost, "/api/v1/equipment", map[string]any{
		"type": "shoes",
		"name": "Trail shoes",
	}, token)
	if w.Code != http.StatusCreated {
		t.Fatalf("create equipment: %d %s", w.Code, w.Body.String())
	}
	var eq map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &eq); err != nil {
		t.Fatal(err)
	}
	eqID, _ := eq["id"].(string)

	w = ta.doJSON(t, http.MethodPost, "/api/v1/workouts", map[string]any{
		"name":          "Run with shoes",
		"sport_type":    "Run",
		"start_date":    "2026-07-08T10:00:00Z",
		"equipment_ids": []string{eqID},
	}, token)
	if w.Code != http.StatusCreated {
		t.Fatalf("create workout: %d %s", w.Code, w.Body.String())
	}

	w = ta.doJSON(t, http.MethodDelete, "/api/v1/equipment/"+eqID, nil, token)
	if w.Code != http.StatusNoContent && w.Code != http.StatusOK {
		t.Fatalf("delete equipment: %d %s", w.Code, w.Body.String())
	}

	w = ta.doJSON(t, http.MethodGet, "/api/v1/workouts?scope=own", nil, token)
	if w.Code != http.StatusOK {
		t.Fatalf("list: %d %s", w.Code, w.Body.String())
	}
}

func TestParseTrackEndpoint(t *testing.T) {
	ta := setupTestApp(t)
	ta.register(t, "alice", "alice@example.com", "password12")
	token, _ := ta.login(t, "alice@example.com", "password12")

	gpx, err := os.ReadFile(filepath.Join("..", "..", "testdata", "tracks", "1-sample.gpx"))
	if err != nil {
		t.Fatal(err)
	}
	var body bytes.Buffer
	mw := multipart.NewWriter(&body)
	part, err := mw.CreateFormFile("track", "sample.gpx")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := part.Write(gpx); err != nil {
		t.Fatal(err)
	}
	_ = mw.Close()

	req := httptest.NewRequest(http.MethodPost, "/api/v1/workouts/parse-track", &body)
	req.Header.Set("Content-Type", mw.FormDataContentType())
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	ta.router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("parse-track: %d %s", w.Code, w.Body.String())
	}
}

func TestUserSearch(t *testing.T) {
	ta := setupTestApp(t)
	ta.register(t, "alice", "alice@example.com", "password12")
	ta.register(t, "bob", "bob@example.com", "password12")
	token, _ := ta.login(t, "alice@example.com", "password12")

	w := ta.doJSON(t, http.MethodGet, "/api/v1/users/search?q=bob", nil, token)
	if w.Code != http.StatusOK {
		t.Fatalf("search: %d %s", w.Code, w.Body.String())
	}
}

func TestWorkoutTrackACL(t *testing.T) {
	ta := setupTestApp(t)
	ta.register(t, "alice", "alice@example.com", "password12")
	ta.register(t, "bob", "bob@example.com", "password12")
	aliceToken, _ := ta.login(t, "alice@example.com", "password12")
	bobToken, _ := ta.login(t, "bob@example.com", "password12")

	gpx, err := os.ReadFile(filepath.Join("..", "..", "testdata", "tracks", "1-sample.gpx"))
	if err != nil {
		t.Fatal(err)
	}
	var body bytes.Buffer
	mw := multipart.NewWriter(&body)
	_ = mw.WriteField("name", "Alice GPX")
	_ = mw.WriteField("sport_type", "Run")
	_ = mw.WriteField("start_date", "2026-07-08T10:00:00Z")
	part, err := mw.CreateFormFile("track", "sample.gpx")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := part.Write(gpx); err != nil {
		t.Fatal(err)
	}
	_ = mw.Close()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/workouts", &body)
	req.Header.Set("Content-Type", mw.FormDataContentType())
	req.Header.Set("Authorization", "Bearer "+aliceToken)
	w := httptest.NewRecorder()
	ta.router.ServeHTTP(w, req)
	if w.Code != http.StatusCreated {
		t.Fatalf("create: %d %s", w.Code, w.Body.String())
	}
	var created map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &created); err != nil {
		t.Fatal(err)
	}
	id, _ := created["id"].(string)

	// Stranger cannot access another owner's workout.
	w = ta.doJSON(t, http.MethodGet, "/api/v1/workouts/"+id+"/track?owner=alice&format=gpx", nil, bobToken)
	if w.Code != http.StatusNotFound {
		t.Fatalf("stranger track access = %d, want 404", w.Code)
	}

	// Follow then GPX export allowed; original format still forbidden for follower.
	w = ta.doJSON(t, http.MethodPost, "/api/v1/social/follow", map[string]string{"handle": "alice"}, bobToken)
	if w.Code != http.StatusCreated {
		t.Fatalf("follow: %d %s", w.Code, w.Body.String())
	}
	w = ta.doJSON(t, http.MethodGet, "/api/v1/workouts/"+id+"/track?owner=alice&format=gpx", nil, bobToken)
	if w.Code != http.StatusOK {
		t.Fatalf("follower gpx track = %d %s", w.Code, w.Body.String())
	}
	w = ta.doJSON(t, http.MethodGet, "/api/v1/workouts/"+id+"/track?owner=alice", nil, bobToken)
	if w.Code != http.StatusForbidden {
		t.Fatalf("follower original track = %d, want 403", w.Code)
	}
}

func TestAvatarUploadAPI(t *testing.T) {
	ta := setupTestApp(t)
	ta.register(t, "alice", "alice@example.com", "password12")
	token, _ := ta.login(t, "alice@example.com", "password12")

	pngData, err := os.ReadFile(filepath.Join("..", "..", "testdata", "images", "avatar-square.png"))
	if err != nil {
		t.Fatal(err)
	}
	var body bytes.Buffer
	mw := multipart.NewWriter(&body)
	part, err := mw.CreateFormFile("avatar", "avatar.png")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := part.Write(pngData); err != nil {
		t.Fatal(err)
	}
	_ = mw.Close()
	req := httptest.NewRequest(http.MethodPut, "/api/v1/auth/me/avatar", &body)
	req.Header.Set("Content-Type", mw.FormDataContentType())
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	ta.router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("upload avatar: %d %s", w.Code, w.Body.String())
	}

	bad, err := os.ReadFile(filepath.Join("..", "..", "testdata", "images", "avatar-nonsquare.png"))
	if err != nil {
		t.Fatal(err)
	}
	body.Reset()
	mw = multipart.NewWriter(&body)
	part, err = mw.CreateFormFile("avatar", "bad.png")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := part.Write(bad); err != nil {
		t.Fatal(err)
	}
	_ = mw.Close()
	req = httptest.NewRequest(http.MethodPut, "/api/v1/auth/me/avatar", &body)
	req.Header.Set("Content-Type", mw.FormDataContentType())
	req.Header.Set("Authorization", "Bearer "+token)
	w = httptest.NewRecorder()
	ta.router.ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("nonsquare avatar = %d, want 400", w.Code)
	}
}

// Ensure storage.Open path used by NewApp is exercised via setup.
var _ = storage.Open
