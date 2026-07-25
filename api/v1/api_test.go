package v1_test

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/json"
	"encoding/pem"
	"io"
	"math/big"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

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
	return setupTestAppWithConfig(t, func(cfg *config.Config) {
		cfg.Server.TLS.Mode = "off"
		cfg.Federation.Enabled = false
		cfg.Federation.Domain = "localhost"
	})
}

func setupFederationTestApp(t *testing.T) *testApp {
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
	})
}

func setupTestAppWithConfig(t *testing.T, mutate func(*config.Config)) *testApp {
	t.Helper()
	gin.SetMode(gin.TestMode)

	prev := config.Cfg
	t.Cleanup(func() { config.Cfg = prev })

	dir := t.TempDir()
	config.Cfg = config.Config{}
	config.Cfg.Auth.JWTSecret = "test-secret-at-least-32-characters!!"
	config.Cfg.Auth.JWTTTLHours = 24
	config.Cfg.Storage.Driver = "file"
	config.Cfg.Storage.Location = dir
	config.Cfg.Storage.TempDir = filepath.Join(dir, "tmp")
	if mutate != nil {
		mutate(&config.Cfg)
	}
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

func writeTestTLS(t *testing.T, dir string) (certPath, keyPath string) {
	t.Helper()
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("generate key: %v", err)
	}
	tmpl := &x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject:      pkix.Name{CommonName: "localhost"},
		NotBefore:    time.Now().UTC().Add(-time.Hour),
		NotAfter:     time.Now().UTC().Add(24 * time.Hour),
		KeyUsage:     x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment,
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		DNSNames:     []string{"localhost"},
	}
	der, err := x509.CreateCertificate(rand.Reader, tmpl, tmpl, &key.PublicKey, key)
	if err != nil {
		t.Fatalf("create cert: %v", err)
	}
	certPath = filepath.Join(dir, "server.crt")
	keyPath = filepath.Join(dir, "server.key")
	if err := os.WriteFile(certPath, pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: der}), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(keyPath, pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(key)}), 0600); err != nil {
		t.Fatal(err)
	}
	return certPath, keyPath
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

func (ta *testApp) doMultipart(t *testing.T, method, path, token string, fields map[string]string, files map[string][]filePart) *httptest.ResponseRecorder {
	t.Helper()
	var body bytes.Buffer
	mw := multipart.NewWriter(&body)
	for k, v := range fields {
		if err := mw.WriteField(k, v); err != nil {
			t.Fatal(err)
		}
	}
	for field, parts := range files {
		for _, part := range parts {
			w, err := mw.CreateFormFile(field, part.filename)
			if err != nil {
				t.Fatal(err)
			}
			if _, err := w.Write(part.data); err != nil {
				t.Fatal(err)
			}
		}
	}
	if err := mw.Close(); err != nil {
		t.Fatal(err)
	}
	req := httptest.NewRequest(method, path, &body)
	req.Header.Set("Content-Type", mw.FormDataContentType())
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	w := httptest.NewRecorder()
	ta.router.ServeHTTP(w, req)
	return w
}

type filePart struct {
	filename string
	data     []byte
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
	return decodeObject(t, w)
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
	out := decodeObject(t, w)
	token, _ = out["token"].(string)
	user, _ = out["user"].(map[string]any)
	if token == "" {
		t.Fatal("expected token")
	}
	return token, user
}

func decodeObject(t *testing.T, w *httptest.ResponseRecorder) map[string]any {
	t.Helper()
	var out map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &out); err != nil {
		t.Fatalf("decode object: %v body=%s", err, w.Body.String())
	}
	return out
}

func decodeList(t *testing.T, w *httptest.ResponseRecorder) []map[string]any {
	t.Helper()
	var raw []any
	if err := json.Unmarshal(w.Body.Bytes(), &raw); err != nil {
		t.Fatalf("decode list: %v body=%s", err, w.Body.String())
	}
	out := make([]map[string]any, 0, len(raw))
	for _, item := range raw {
		obj, ok := item.(map[string]any)
		if !ok {
			t.Fatalf("list item is not object: %#v", item)
		}
		out = append(out, obj)
	}
	return out
}

func expectStatus(t *testing.T, w *httptest.ResponseRecorder, want int) {
	t.Helper()
	if w.Code != want {
		t.Fatalf("status = %d, want %d body=%s", w.Code, want, w.Body.String())
	}
}

func readTestdata(t *testing.T, rel string) []byte {
	t.Helper()
	data, err := os.ReadFile(filepath.Join("..", "..", "testdata", rel))
	if err != nil {
		t.Fatal(err)
	}
	return data
}

func TestAuthRegisterLoginMe(t *testing.T) {
	ta := setupTestApp(t)

	w := ta.doJSON(t, http.MethodPost, "/api/v1/auth/register", map[string]string{
		"nickname": "alice",
		"name":     "Alice",
		"email":    "not-an-email",
		"password": "password12",
	}, "")
	expectStatus(t, w, http.StatusBadRequest)

	registered := ta.register(t, "alice", "alice@example.com", "password12")
	if registered["nickname"] != "alice" || registered["email"] != "alice@example.com" {
		t.Fatalf("unexpected register body: %#v", registered)
	}

	w = ta.doJSON(t, http.MethodPost, "/api/v1/auth/register", map[string]string{
		"nickname": "alice2",
		"name":     "Alice",
		"email":    "alice@example.com",
		"password": "password12",
	}, "")
	expectStatus(t, w, http.StatusConflict)

	w = ta.doJSON(t, http.MethodPost, "/api/v1/auth/login", map[string]string{
		"email":    "alice@example.com",
		"password": "wrong-password",
	}, "")
	expectStatus(t, w, http.StatusUnauthorized)

	w = ta.doJSON(t, http.MethodPost, "/api/v1/auth/login", map[string]string{
		"email":    "missing@example.com",
		"password": "password12",
	}, "")
	expectStatus(t, w, http.StatusUnauthorized)

	token, _ := ta.login(t, "alice@example.com", "password12")
	w = ta.doJSON(t, http.MethodGet, "/api/v1/auth/me", nil, token)
	expectStatus(t, w, http.StatusOK)
	me := decodeObject(t, w)
	if me["nickname"] != "alice" || me["email"] != "alice@example.com" {
		t.Fatalf("unexpected me body: %#v", me)
	}

	w = ta.doJSON(t, http.MethodGet, "/api/v1/auth/me", nil, "")
	expectStatus(t, w, http.StatusUnauthorized)
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
	expectStatus(t, w, http.StatusCreated)
	created := decodeObject(t, w)
	id, _ := created["id"].(string)
	if id == "" {
		t.Fatal("expected workout id")
	}
	if created["name"] != "Morning run" || created["sport_type"] != "Run" {
		t.Fatalf("unexpected create body: %#v", created)
	}
	if created["duration_seconds"].(float64) != 3600 || created["distance"].(float64) != 5200 {
		t.Fatalf("unexpected metrics: %#v", created)
	}

	w = ta.doJSON(t, http.MethodPut, "/api/v1/workouts/"+id, map[string]any{
		"name":             "Evening run",
		"description":      "Updated",
		"sport_type":       "Run",
		"start_date":       "2026-07-09T18:00:00Z",
		"duration_seconds": 4000,
		"distance":         6000,
	}, token)
	expectStatus(t, w, http.StatusOK)
	updated := decodeObject(t, w)
	if updated["name"] != "Evening run" || updated["description"] != "Updated" {
		t.Fatalf("unexpected update body: %#v", updated)
	}
	if updated["duration_seconds"].(float64) != 4000 || updated["distance"].(float64) != 6000 {
		t.Fatalf("unexpected updated metrics: %#v", updated)
	}
	if updated["owner"] != "alice" {
		t.Fatalf("expected owner alice, got %#v", updated["owner"])
	}
	if updated["start_date"] != "2026-07-09T18:00:00Z" {
		t.Fatalf("unexpected start_date: %#v", updated["start_date"])
	}

	w = ta.doJSON(t, http.MethodGet, "/api/v1/workouts?scope=own", nil, token)
	expectStatus(t, w, http.StatusOK)
	list := decodeList(t, w)
	if len(list) != 1 || list[0]["id"] != id || list[0]["name"] != "Evening run" {
		t.Fatalf("unexpected list after update: %#v", list)
	}

	w = ta.doJSON(t, http.MethodPut, "/api/v1/workouts/missingid", map[string]any{
		"name":       "Nope",
		"sport_type": "Run",
		"start_date": "2026-07-09T18:00:00Z",
	}, token)
	expectStatus(t, w, http.StatusNotFound)

	w = ta.doJSON(t, http.MethodDelete, "/api/v1/workouts/"+id, nil, token)
	expectStatus(t, w, http.StatusNoContent)

	w = ta.doJSON(t, http.MethodGet, "/api/v1/workouts?scope=own", nil, token)
	expectStatus(t, w, http.StatusOK)
	if list = decodeList(t, w); len(list) != 0 {
		t.Fatalf("expected empty list after delete, got %#v", list)
	}
}

func TestWorkoutMultipartTrack(t *testing.T) {
	ta := setupTestApp(t)
	ta.register(t, "alice", "alice@example.com", "password12")
	token, _ := ta.login(t, "alice@example.com", "password12")

	gpx := readTestdata(t, "tracks/1-sample.gpx")
	w := ta.doMultipart(t, http.MethodPost, "/api/v1/workouts", token,
		map[string]string{
			"name":       "GPX run",
			"sport_type": "Run",
			"start_date": "2026-07-08T10:00:00Z",
		},
		map[string][]filePart{
			"track": {{filename: "sample.gpx", data: gpx}},
		},
	)
	expectStatus(t, w, http.StatusCreated)
	created := decodeObject(t, w)
	if created["track"] == nil || created["track"] == "" {
		t.Fatalf("expected track filename, got %#v", created)
	}
	if created["has_map_preview"] != true {
		t.Fatalf("expected has_map_preview, got %#v", created)
	}
	if dist, _ := created["distance"].(float64); dist <= 0 {
		t.Fatalf("expected distance from GPX, got %#v", created["distance"])
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
	expectStatus(t, w, http.StatusBadRequest)

	w = ta.doJSON(t, http.MethodPost, "/api/v1/social/follow", map[string]string{
		"handle": "bob",
	}, token)
	expectStatus(t, w, http.StatusCreated)
	follow := decodeObject(t, w)
	followID, _ := follow["id"].(string)
	if followID == "" || follow["target_nickname"] != "bob" || follow["status"] != "active" {
		t.Fatalf("unexpected follow body: %#v", follow)
	}

	w = ta.doJSON(t, http.MethodDelete, "/api/v1/social/follow/"+followID, nil, token)
	expectStatus(t, w, http.StatusNoContent)

	w = ta.doJSON(t, http.MethodGet, "/api/v1/social/following", nil, token)
	expectStatus(t, w, http.StatusOK)
	if following := decodeList(t, w); len(following) != 0 {
		t.Fatalf("expected empty following after unfollow, got %#v", following)
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
	expectStatus(t, w, http.StatusCreated)
	eq := decodeObject(t, w)
	eqID, _ := eq["id"].(string)
	if eqID == "" {
		t.Fatal("expected equipment id")
	}

	w = ta.doJSON(t, http.MethodPost, "/api/v1/workouts", map[string]any{
		"name":          "Run with shoes",
		"sport_type":    "Run",
		"start_date":    "2026-07-08T10:00:00Z",
		"equipment_ids": []string{eqID},
	}, token)
	expectStatus(t, w, http.StatusCreated)
	created := decodeObject(t, w)
	equipment, _ := created["equipment"].([]any)
	if len(equipment) != 1 {
		t.Fatalf("expected equipment on workout, got %#v", created["equipment"])
	}

	w = ta.doJSON(t, http.MethodDelete, "/api/v1/equipment/"+eqID, nil, token)
	expectStatus(t, w, http.StatusNoContent)

	w = ta.doJSON(t, http.MethodGet, "/api/v1/workouts?scope=own", nil, token)
	expectStatus(t, w, http.StatusOK)
	list := decodeList(t, w)
	if len(list) != 1 {
		t.Fatalf("expected workout to remain, got %#v", list)
	}
	if eqList, ok := list[0]["equipment"].([]any); ok && len(eqList) != 0 {
		t.Fatalf("expected equipment cleared from workout, got %#v", list[0]["equipment"])
	}
	if list[0]["name"] != "Run with shoes" {
		t.Fatalf("workout fields should remain: %#v", list[0])
	}

	w = ta.doJSON(t, http.MethodGet, "/api/v1/equipment", nil, token)
	expectStatus(t, w, http.StatusOK)
	if items := decodeList(t, w); len(items) != 0 {
		t.Fatalf("expected empty equipment list, got %#v", items)
	}
}

func TestParseTrackEndpoint(t *testing.T) {
	ta := setupTestApp(t)
	ta.register(t, "alice", "alice@example.com", "password12")
	token, _ := ta.login(t, "alice@example.com", "password12")

	gpx := readTestdata(t, "tracks/1-sample.gpx")
	w := ta.doMultipart(t, http.MethodPost, "/api/v1/workouts/parse-track", token, nil,
		map[string][]filePart{
			"track": {{filename: "sample.gpx", data: gpx}},
		},
	)
	expectStatus(t, w, http.StatusOK)
	parsed := decodeObject(t, w)
	if dist, _ := parsed["distance"].(float64); dist <= 0 {
		t.Fatalf("expected parsed distance, got %#v", parsed)
	}
	if _, ok := parsed["has_gps"]; !ok {
		t.Fatalf("expected has_gps field, got %#v", parsed)
	}
}

func TestUserSearch(t *testing.T) {
	ta := setupTestApp(t)
	ta.register(t, "alice", "alice@example.com", "password12")
	ta.register(t, "bob", "bob@example.com", "password12")
	token, _ := ta.login(t, "alice@example.com", "password12")

	w := ta.doJSON(t, http.MethodGet, "/api/v1/users/search?q=bob", nil, token)
	expectStatus(t, w, http.StatusOK)
	results := decodeList(t, w)
	if len(results) != 1 || results[0]["nickname"] != "bob" {
		t.Fatalf("unexpected search results: %#v", results)
	}
}

func TestWorkoutTrackACL(t *testing.T) {
	ta := setupTestApp(t)
	ta.register(t, "alice", "alice@example.com", "password12")
	ta.register(t, "bob", "bob@example.com", "password12")
	aliceToken, _ := ta.login(t, "alice@example.com", "password12")
	bobToken, _ := ta.login(t, "bob@example.com", "password12")

	gpx := readTestdata(t, "tracks/1-sample.gpx")
	w := ta.doMultipart(t, http.MethodPost, "/api/v1/workouts", aliceToken,
		map[string]string{
			"name":       "Alice GPX",
			"sport_type": "Run",
			"start_date": "2026-07-08T10:00:00Z",
		},
		map[string][]filePart{
			"track": {{filename: "sample.gpx", data: gpx}},
		},
	)
	expectStatus(t, w, http.StatusCreated)
	id, _ := decodeObject(t, w)["id"].(string)

	w = ta.doJSON(t, http.MethodGet, "/api/v1/workouts/"+id+"/track?owner=alice&format=gpx", nil, bobToken)
	expectStatus(t, w, http.StatusNotFound)

	w = ta.doJSON(t, http.MethodPost, "/api/v1/social/follow", map[string]string{"handle": "alice"}, bobToken)
	expectStatus(t, w, http.StatusCreated)

	w = ta.doJSON(t, http.MethodGet, "/api/v1/workouts/"+id+"/track?owner=alice&format=gpx", nil, bobToken)
	expectStatus(t, w, http.StatusOK)

	w = ta.doJSON(t, http.MethodGet, "/api/v1/workouts/"+id+"/track?owner=alice", nil, bobToken)
	expectStatus(t, w, http.StatusForbidden)
}

func TestAvatarUploadAPI(t *testing.T) {
	ta := setupTestApp(t)
	ta.register(t, "alice", "alice@example.com", "password12")
	token, _ := ta.login(t, "alice@example.com", "password12")

	pngData := readTestdata(t, "images/avatar-square.png")
	w := ta.doMultipart(t, http.MethodPut, "/api/v1/auth/me/avatar", token, nil,
		map[string][]filePart{
			"avatar": {{filename: "avatar.png", data: pngData}},
		},
	)
	expectStatus(t, w, http.StatusOK)
	user := decodeObject(t, w)
	if user["has_avatar"] != true {
		t.Fatalf("expected has_avatar after upload: %#v", user)
	}
	avatarURL, _ := user["avatar_url"].(string)
	if avatarURL == "" {
		t.Fatalf("expected avatar_url: %#v", user)
	}

	w = ta.doJSON(t, http.MethodGet, "/api/v1/users/alice/avatar", nil, token)
	expectStatus(t, w, http.StatusOK)
	if ct := w.Header().Get("Content-Type"); ct != "image/webp" {
		t.Fatalf("avatar content-type = %q", ct)
	}
	if len(w.Body.Bytes()) == 0 {
		t.Fatal("expected avatar bytes")
	}

	bad := readTestdata(t, "images/avatar-nonsquare.png")
	w = ta.doMultipart(t, http.MethodPut, "/api/v1/auth/me/avatar", token, nil,
		map[string][]filePart{
			"avatar": {{filename: "bad.png", data: bad}},
		},
	)
	expectStatus(t, w, http.StatusBadRequest)
}

// Ensure storage.Open path used by NewApp is exercised via setup.
var _ = storage.Open
