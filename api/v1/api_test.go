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
	"github.com/solargate/grom/internal/avatars"
	"github.com/solargate/grom/internal/config"
	"github.com/solargate/grom/internal/federation"
	"github.com/solargate/grom/internal/storage"
	"github.com/solargate/grom/internal/storage/keys"
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

func setupTestAppWithDriver(t *testing.T, driver config.StorageDriver) *testApp {
	t.Helper()
	return setupTestAppWithConfig(t, func(cfg *config.Config) {
		cfg.Server.TLS.Mode = "off"
		cfg.Federation.Enabled = false
		cfg.Federation.Domain = "localhost"
		cfg.Storage.Driver = driver
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

func decodeWorkoutPage(t *testing.T, w *httptest.ResponseRecorder) (items []map[string]any, nextCursor string, hasMore bool) {
	t.Helper()
	obj := decodeObject(t, w)
	rawItems, _ := obj["items"].([]any)
	items = make([]map[string]any, 0, len(rawItems))
	for _, item := range rawItems {
		m, ok := item.(map[string]any)
		if !ok {
			t.Fatalf("page item is not object: %#v", item)
		}
		items = append(items, m)
	}
	nextCursor, _ = obj["next_cursor"].(string)
	hasMore, _ = obj["has_more"].(bool)
	return items, nextCursor, hasMore
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
	list, _, _ := decodeWorkoutPage(t, w)
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
	if list, _, _ = decodeWorkoutPage(t, w); len(list) != 0 {
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
	list, _, _ := decodeWorkoutPage(t, w)
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
	if name, _ := parsed["name"].(string); name != "Test track" {
		t.Fatalf("expected name %q, got %#v", "Test track", parsed["name"])
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

func TestGetWorkoutSpeed(t *testing.T) {
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
	created := decodeObject(t, w)
	id, _ := created["id"].(string)
	if _, ok := created["speed_max_kmh"].(float64); !ok {
		t.Fatalf("expected speed_max_kmh on create: %#v", created)
	}

	w = ta.doJSON(t, http.MethodGet, "/api/v1/workouts/"+id+"/speed", nil, aliceToken)
	expectStatus(t, w, http.StatusOK)
	body := decodeObject(t, w)
	samples, _ := body["samples"].([]any)
	if len(samples) < 2 {
		t.Fatalf("expected speed samples from GPX, got %#v", body)
	}
	first, _ := samples[0].(map[string]any)
	if _, ok := first["speed_kmh"].(float64); !ok {
		t.Fatalf("expected speed_kmh in sample: %#v", first)
	}
	if _, ok := first["distance_m"].(float64); !ok {
		t.Fatalf("expected distance_m in sample: %#v", first)
	}
	if _, ok := first["t"].(string); !ok {
		t.Fatalf("expected t in sample: %#v", first)
	}

	w = ta.doJSON(t, http.MethodGet, "/api/v1/workouts/"+id+"/speed?owner=alice", nil, bobToken)
	expectStatus(t, w, http.StatusNotFound)

	w = ta.doJSON(t, http.MethodPost, "/api/v1/social/follow", map[string]string{"handle": "alice"}, bobToken)
	expectStatus(t, w, http.StatusCreated)

	w = ta.doJSON(t, http.MethodGet, "/api/v1/workouts/"+id+"/speed?owner=alice", nil, bobToken)
	expectStatus(t, w, http.StatusOK)
	body = decodeObject(t, w)
	samples, _ = body["samples"].([]any)
	if len(samples) < 2 {
		t.Fatalf("expected followed workout speed samples, got %#v", body)
	}

	w = ta.doJSON(t, http.MethodPost, "/api/v1/workouts", map[string]any{
		"name":             "No track",
		"sport_type":       "Run",
		"start_date":       "2026-07-08T11:00:00Z",
		"duration_seconds": 600,
		"distance":         1000,
	}, aliceToken)
	expectStatus(t, w, http.StatusCreated)
	noTrackID, _ := decodeObject(t, w)["id"].(string)

	w = ta.doJSON(t, http.MethodGet, "/api/v1/workouts/"+noTrackID+"/speed", nil, aliceToken)
	expectStatus(t, w, http.StatusOK)
	body = decodeObject(t, w)
	samples, _ = body["samples"].([]any)
	if len(samples) != 0 {
		t.Fatalf("expected empty samples without track, got %#v", body)
	}
}

func TestGetWorkoutHeartRate(t *testing.T) {
	ta := setupTestApp(t)
	ta.register(t, "alice", "alice@example.com", "password12")
	ta.register(t, "bob", "bob@example.com", "password12")
	aliceToken, _ := ta.login(t, "alice@example.com", "password12")
	bobToken, _ := ta.login(t, "bob@example.com", "password12")

	fit := readTestdata(t, "tracks/1-ride.fit")
	w := ta.doMultipart(t, http.MethodPost, "/api/v1/workouts", aliceToken,
		map[string]string{
			"name":       "Alice FIT",
			"sport_type": "Ride",
			"start_date": "2026-07-08T10:00:00Z",
		},
		map[string][]filePart{
			"track": {{filename: "ride.fit", data: fit}},
		},
	)
	expectStatus(t, w, http.StatusCreated)
	created := decodeObject(t, w)
	id, _ := created["id"].(string)
	if _, ok := created["heart_rate_max"].(float64); !ok {
		t.Fatalf("expected heart_rate_max on create: %#v", created)
	}

	w = ta.doJSON(t, http.MethodGet, "/api/v1/workouts/"+id+"/heartrate", nil, aliceToken)
	expectStatus(t, w, http.StatusOK)
	body := decodeObject(t, w)
	samples, _ := body["samples"].([]any)
	if len(samples) < 2 {
		t.Fatalf("expected heart rate samples from FIT, got %#v", body)
	}
	first, _ := samples[0].(map[string]any)
	if _, ok := first["heart_rate_bpm"].(float64); !ok {
		t.Fatalf("expected heart_rate_bpm in sample: %#v", first)
	}
	if _, ok := first["t"].(string); !ok {
		t.Fatalf("expected t in sample: %#v", first)
	}
	if hasGPS, ok := body["has_gps"].(bool); !ok || !hasGPS {
		t.Fatalf("expected has_gps true for GPS FIT, got %#v", body["has_gps"])
	}
	if _, ok := first["distance_m"].(float64); !ok {
		t.Fatalf("expected distance_m in GPS sample: %#v", first)
	}

	w = ta.doJSON(t, http.MethodGet, "/api/v1/workouts/"+id+"/heartrate?owner=alice", nil, bobToken)
	expectStatus(t, w, http.StatusNotFound)

	w = ta.doJSON(t, http.MethodPost, "/api/v1/social/follow", map[string]string{"handle": "alice"}, bobToken)
	expectStatus(t, w, http.StatusCreated)

	w = ta.doJSON(t, http.MethodGet, "/api/v1/workouts/"+id+"/heartrate?owner=alice", nil, bobToken)
	expectStatus(t, w, http.StatusOK)
	body = decodeObject(t, w)
	samples, _ = body["samples"].([]any)
	if len(samples) < 2 {
		t.Fatalf("expected followed workout heart rate samples, got %#v", body)
	}

	w = ta.doJSON(t, http.MethodGet, "/api/v1/workouts/"+id, nil, aliceToken)
	expectStatus(t, w, http.StatusOK)
	detail := decodeObject(t, w)
	if _, ok := detail["heart_rate_avg"].(float64); !ok {
		t.Fatalf("expected heart_rate_avg on workout response: %#v", detail)
	}
	if _, ok := detail["heart_rate_max"].(float64); !ok {
		t.Fatalf("expected heart_rate_max on workout response: %#v", detail)
	}

	w = ta.doJSON(t, http.MethodPost, "/api/v1/workouts", map[string]any{
		"name":             "No track",
		"sport_type":       "Workout",
		"start_date":       "2026-07-08T11:00:00Z",
		"duration_seconds": 600,
	}, aliceToken)
	expectStatus(t, w, http.StatusCreated)
	noTrackID, _ := decodeObject(t, w)["id"].(string)

	w = ta.doJSON(t, http.MethodGet, "/api/v1/workouts/"+noTrackID+"/heartrate", nil, aliceToken)
	expectStatus(t, w, http.StatusOK)
	body = decodeObject(t, w)
	samples, _ = body["samples"].([]any)
	if len(samples) != 0 {
		t.Fatalf("expected empty samples without track, got %#v", body)
	}
}

func TestGetWorkout(t *testing.T) {
	ta := setupTestApp(t)
	ta.register(t, "alice", "alice@example.com", "password12")
	ta.register(t, "bob", "bob@example.com", "password12")
	aliceToken, _ := ta.login(t, "alice@example.com", "password12")
	bobToken, _ := ta.login(t, "bob@example.com", "password12")

	w := ta.doJSON(t, http.MethodPost, "/api/v1/workouts", map[string]any{
		"name":            "Solo run",
		"sport_type":      "Run",
		"start_date":      "2026-07-08T10:00:00Z",
		"duration_seconds": 1800,
		"distance":        5000,
	}, aliceToken)
	expectStatus(t, w, http.StatusCreated)
	created := decodeObject(t, w)
	id, _ := created["id"].(string)

	w = ta.doJSON(t, http.MethodGet, "/api/v1/workouts/"+id, nil, aliceToken)
	expectStatus(t, w, http.StatusOK)
	got := decodeObject(t, w)
	if got["id"] != id || got["name"] != "Solo run" || got["owner"] != "alice" {
		t.Fatalf("unexpected own workout: %#v", got)
	}
	author, _ := got["author"].(map[string]any)
	if author == nil || author["nickname"] != "alice" || author["is_local"] != true {
		t.Fatalf("unexpected author: %#v", got["author"])
	}

	w = ta.doJSON(t, http.MethodGet, "/api/v1/workouts/"+id+"?owner=alice", nil, bobToken)
	expectStatus(t, w, http.StatusNotFound)

	w = ta.doJSON(t, http.MethodPost, "/api/v1/social/follow", map[string]string{"handle": "alice"}, bobToken)
	expectStatus(t, w, http.StatusCreated)

	w = ta.doJSON(t, http.MethodGet, "/api/v1/workouts/"+id+"?owner=alice", nil, bobToken)
	expectStatus(t, w, http.StatusOK)
	got = decodeObject(t, w)
	if got["id"] != id || got["owner"] != "alice" || got["name"] != "Solo run" {
		t.Fatalf("unexpected followed workout: %#v", got)
	}

	w = ta.doJSON(t, http.MethodGet, "/api/v1/workouts/missingid", nil, aliceToken)
	expectStatus(t, w, http.StatusNotFound)
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

func TestListWorkoutsPagination(t *testing.T) {
	ta := setupTestApp(t)
	ta.register(t, "alice", "alice@example.com", "password12")
	token, _ := ta.login(t, "alice@example.com", "password12")

	start := time.Date(2026, 7, 8, 12, 0, 0, 0, time.UTC)
	for i := 0; i < 5; i++ {
		w := ta.doJSON(t, http.MethodPost, "/api/v1/workouts", map[string]any{
			"name":       "Run",
			"sport_type": "Run",
			"start_date": start.Add(-time.Duration(i) * time.Hour).Format(time.RFC3339),
		}, token)
		expectStatus(t, w, http.StatusCreated)
	}

	w := ta.doJSON(t, http.MethodGet, "/api/v1/workouts?scope=own&limit=2", nil, token)
	expectStatus(t, w, http.StatusOK)
	items, cursor, hasMore := decodeWorkoutPage(t, w)
	if len(items) != 2 || !hasMore || cursor == "" {
		t.Fatalf("page1 items=%d hasMore=%v cursor=%q", len(items), hasMore, cursor)
	}

	w = ta.doJSON(t, http.MethodGet, "/api/v1/workouts?scope=own&limit=2&cursor="+cursor, nil, token)
	expectStatus(t, w, http.StatusOK)
	items2, cursor2, hasMore2 := decodeWorkoutPage(t, w)
	if len(items2) != 2 || !hasMore2 || cursor2 == "" {
		t.Fatalf("page2 items=%d hasMore=%v cursor=%q", len(items2), hasMore2, cursor2)
	}
	if items2[0]["id"] == items[0]["id"] || items2[0]["id"] == items[1]["id"] {
		t.Fatalf("pages overlap: %#v %#v", items, items2)
	}

	w = ta.doJSON(t, http.MethodGet, "/api/v1/workouts?scope=own&limit=2&cursor="+cursor2, nil, token)
	expectStatus(t, w, http.StatusOK)
	items3, _, hasMore3 := decodeWorkoutPage(t, w)
	if len(items3) != 1 || hasMore3 {
		t.Fatalf("page3 items=%d hasMore=%v", len(items3), hasMore3)
	}

	w = ta.doJSON(t, http.MethodGet, "/api/v1/workouts?scope=own&cursor=not-a-valid-cursor", nil, token)
	expectStatus(t, w, http.StatusBadRequest)

	w = ta.doJSON(t, http.MethodGet, "/api/v1/workouts?scope=own&limit=0", nil, token)
	expectStatus(t, w, http.StatusBadRequest)

	w = ta.doJSON(t, http.MethodGet, "/api/v1/workouts?scope=own&limit=999", nil, token)
	expectStatus(t, w, http.StatusOK)
	clamped, _, _ := decodeWorkoutPage(t, w)
	if len(clamped) > 100 {
		t.Fatalf("expected clamped page size, got %d", len(clamped))
	}
}

func TestWorkoutChartsAcrossStorageDrivers(t *testing.T) {
	for _, driver := range []config.StorageDriver{config.StorageDriverFile, config.StorageDriverBBolt} {
		t.Run(string(driver), func(t *testing.T) {
			ta := setupTestAppWithDriver(t, driver)
			ta.register(t, "alice", "alice@example.com", "password12")
			token, _ := ta.login(t, "alice@example.com", "password12")

			gpx := readTestdata(t, "tracks/1-sample.gpx")
			w := ta.doMultipart(t, http.MethodPost, "/api/v1/workouts", token,
				map[string]string{
					"name":       "Chart run",
					"sport_type": "Run",
					"start_date": "2026-07-08T10:00:47Z",
				},
				map[string][]filePart{
					"track": {{filename: "sample.gpx", data: gpx}},
				},
			)
			expectStatus(t, w, http.StatusCreated)
			created := decodeObject(t, w)
			id, _ := created["id"].(string)

			w = ta.doJSON(t, http.MethodGet, "/api/v1/workouts/"+id+"/speed", nil, token)
			expectStatus(t, w, http.StatusOK)
			body := decodeObject(t, w)
			samples, _ := body["samples"].([]any)
			if len(samples) < 2 {
				t.Fatalf("speed samples: %#v", body)
			}

			w = ta.doJSON(t, http.MethodPut, "/api/v1/workouts/"+id, map[string]any{
				"name":       "Chart run renamed",
				"sport_type": "Run",
				"start_date": "2026-07-09T10:00:47Z",
			}, token)
			expectStatus(t, w, http.StatusOK)

			w = ta.doJSON(t, http.MethodGet, "/api/v1/workouts/"+id+"/speed", nil, token)
			expectStatus(t, w, http.StatusOK)
			body = decodeObject(t, w)
			samples, _ = body["samples"].([]any)
			if len(samples) < 2 {
				t.Fatalf("speed samples after rename: %#v", body)
			}

			fit := readTestdata(t, "tracks/1-ride.fit")
			w = ta.doMultipart(t, http.MethodPost, "/api/v1/workouts", token,
				map[string]string{
					"name":       "Chart ride",
					"sport_type": "Ride",
					"start_date": "2026-07-10T10:00:00Z",
				},
				map[string][]filePart{
					"track": {{filename: "ride.fit", data: fit}},
				},
			)
			expectStatus(t, w, http.StatusCreated)
			rideID, _ := decodeObject(t, w)["id"].(string)
			w = ta.doJSON(t, http.MethodGet, "/api/v1/workouts/"+rideID+"/heartrate", nil, token)
			expectStatus(t, w, http.StatusOK)
			hrBody := decodeObject(t, w)
			hrSamples, _ := hrBody["samples"].([]any)
			if len(hrSamples) < 2 {
				t.Fatalf("hr samples: %#v", hrBody)
			}
		})
	}
}

func TestAuthShortPasswordAndDuplicateNickname(t *testing.T) {
	ta := setupTestApp(t)
	ta.register(t, "alice", "alice@example.com", "password12")

	w := ta.doJSON(t, http.MethodPost, "/api/v1/auth/register", map[string]string{
		"nickname": "bob",
		"name":     "Bob",
		"email":    "bob@example.com",
		"password": "short",
	}, "")
	expectStatus(t, w, http.StatusBadRequest)

	w = ta.doJSON(t, http.MethodPost, "/api/v1/auth/register", map[string]string{
		"nickname": "alice",
		"name":     "Other",
		"email":    "other@example.com",
		"password": "password12",
	}, "")
	expectStatus(t, w, http.StatusConflict)
}

func TestWorkoutUpdateEquipmentDistanceAndForeignACL(t *testing.T) {
	ta := setupTestApp(t)
	ta.register(t, "alice", "alice@example.com", "password12")
	ta.register(t, "bob", "bob@example.com", "password12")
	aliceToken, _ := ta.login(t, "alice@example.com", "password12")
	bobToken, _ := ta.login(t, "bob@example.com", "password12")

	w := ta.doJSON(t, http.MethodPost, "/api/v1/equipment", map[string]any{
		"type": "shoes", "name": "Road",
	}, aliceToken)
	expectStatus(t, w, http.StatusCreated)
	eqID, _ := decodeObject(t, w)["id"].(string)

	w = ta.doJSON(t, http.MethodPost, "/api/v1/workouts", map[string]any{
		"name": "Run", "sport_type": "Run", "start_date": "2026-07-08T10:00:00Z",
		"duration_seconds": 1800, "distance": 5000, "equipment_ids": []string{eqID},
	}, aliceToken)
	expectStatus(t, w, http.StatusCreated)
	id, _ := decodeObject(t, w)["id"].(string)

	if err := ta.app.EquipmentDistance.RecalculateForIDs("alice", []string{eqID}); err != nil {
		t.Fatal(err)
	}
	w = ta.doJSON(t, http.MethodGet, "/api/v1/equipment", nil, aliceToken)
	expectStatus(t, w, http.StatusOK)
	items := decodeList(t, w)
	if len(items) != 1 || items[0]["distance"].(float64) != 5000 {
		t.Fatalf("equipment distance after create: %#v", items)
	}

	w = ta.doJSON(t, http.MethodPut, "/api/v1/workouts/"+id, map[string]any{
		"name": "Run", "sport_type": "Run", "start_date": "2026-07-08T10:00:00Z",
		"duration_seconds": 1800, "distance": 7500, "equipment_ids": []string{eqID},
	}, aliceToken)
	expectStatus(t, w, http.StatusOK)
	if err := ta.app.EquipmentDistance.RecalculateForIDs("alice", []string{eqID}); err != nil {
		t.Fatal(err)
	}
	w = ta.doJSON(t, http.MethodGet, "/api/v1/equipment", nil, aliceToken)
	expectStatus(t, w, http.StatusOK)
	items = decodeList(t, w)
	if items[0]["distance"].(float64) != 7500 {
		t.Fatalf("equipment distance after update: %#v", items)
	}

	w = ta.doJSON(t, http.MethodPut, "/api/v1/workouts/"+id, map[string]any{
		"name": "Nope", "sport_type": "Run", "start_date": "2026-07-08T10:00:00Z",
	}, bobToken)
	expectStatus(t, w, http.StatusNotFound)
}

func TestCreateWorkoutCopiesEquipmentFromPreviousSameSport(t *testing.T) {
	ta := setupTestApp(t)
	ta.register(t, "alice", "alice@example.com", "password12")
	token, _ := ta.login(t, "alice@example.com", "password12")

	w := ta.doJSON(t, http.MethodPost, "/api/v1/equipment", map[string]any{
		"type": "shoes",
		"name": "Road shoes",
	}, token)
	expectStatus(t, w, http.StatusCreated)
	eqID, _ := decodeObject(t, w)["id"].(string)

	w = ta.doJSON(t, http.MethodPost, "/api/v1/workouts", map[string]any{
		"name": "Earlier run", "sport_type": "Run", "start_date": "2026-07-01T10:00:00Z",
		"equipment_ids": []string{eqID},
	}, token)
	expectStatus(t, w, http.StatusCreated)

	w = ta.doJSON(t, http.MethodPost, "/api/v1/workouts", map[string]any{
		"name": "Ride with bike", "sport_type": "Ride", "start_date": "2026-07-08T10:00:00Z",
	}, token)
	expectStatus(t, w, http.StatusCreated)
	ride := decodeObject(t, w)
	if eqList, ok := ride["equipment"].([]any); ok && len(eqList) != 0 {
		t.Fatalf("Ride should not copy Run equipment, got %#v", ride["equipment"])
	}

	w = ta.doJSON(t, http.MethodPost, "/api/v1/workouts", map[string]any{
		"name": "Later run", "sport_type": "Run", "start_date": "2026-07-09T10:00:00Z",
	}, token)
	expectStatus(t, w, http.StatusCreated)
	created := decodeObject(t, w)
	equipment, _ := created["equipment"].([]any)
	if len(equipment) != 1 {
		t.Fatalf("expected equipment copied from previous Run, got %#v", created["equipment"])
	}
	item, _ := equipment[0].(map[string]any)
	if item["id"] != eqID {
		t.Fatalf("unexpected equipment: %#v", item)
	}

	w = ta.doJSON(t, http.MethodGet, "/api/v1/auth/me", nil, token)
	expectStatus(t, w, http.StatusOK)
	me := decodeObject(t, w)
	lastBySport, _ := me["last_equipment_by_sport"].(map[string]any)
	runIDs, _ := lastBySport["Run"].([]any)
	if len(runIDs) != 1 || runIDs[0] != eqID {
		t.Fatalf("expected last_equipment_by_sport Run updated, got %#v", lastBySport)
	}
}

func TestCreateWorkoutExplicitEmptyEquipmentSkipsDefault(t *testing.T) {
	ta := setupTestApp(t)
	ta.register(t, "alice", "alice@example.com", "password12")
	token, _ := ta.login(t, "alice@example.com", "password12")

	w := ta.doJSON(t, http.MethodPost, "/api/v1/equipment", map[string]any{
		"type": "shoes",
		"name": "Road shoes",
	}, token)
	expectStatus(t, w, http.StatusCreated)
	eqID, _ := decodeObject(t, w)["id"].(string)

	w = ta.doJSON(t, http.MethodPost, "/api/v1/workouts", map[string]any{
		"name": "Earlier run", "sport_type": "Run", "start_date": "2026-07-01T10:00:00Z",
		"equipment_ids": []string{eqID},
	}, token)
	expectStatus(t, w, http.StatusCreated)

	w = ta.doJSON(t, http.MethodPost, "/api/v1/workouts", map[string]any{
		"name": "Barefoot", "sport_type": "Run", "start_date": "2026-07-09T10:00:00Z",
		"equipment_ids": []string{},
	}, token)
	expectStatus(t, w, http.StatusCreated)
	created := decodeObject(t, w)
	if eqList, ok := created["equipment"].([]any); ok && len(eqList) != 0 {
		t.Fatalf("explicit empty equipment_ids should not copy previous, got %#v", created["equipment"])
	}

	w = ta.doMultipart(t, http.MethodPost, "/api/v1/workouts", token, map[string]string{
		"name": "Barefoot multipart", "sport_type": "Run", "start_date": "2026-07-10T10:00:00Z",
		"duration_seconds": "600", "distance": "1000",
		"equipment_ids": "[]",
	}, nil)
	expectStatus(t, w, http.StatusCreated)
	created = decodeObject(t, w)
	if eqList, ok := created["equipment"].([]any); ok && len(eqList) != 0 {
		t.Fatalf("multipart explicit [] should not copy previous, got %#v", created["equipment"])
	}
}

func TestCreateWorkoutMultipartOmitsEquipmentUsesDefault(t *testing.T) {
	ta := setupTestApp(t)
	ta.register(t, "alice", "alice@example.com", "password12")
	token, _ := ta.login(t, "alice@example.com", "password12")

	w := ta.doJSON(t, http.MethodPost, "/api/v1/equipment", map[string]any{
		"type": "shoes",
		"name": "Trail shoes",
	}, token)
	expectStatus(t, w, http.StatusCreated)
	eqID, _ := decodeObject(t, w)["id"].(string)

	w = ta.doJSON(t, http.MethodPost, "/api/v1/workouts", map[string]any{
		"name": "Earlier run", "sport_type": "Run", "start_date": "2026-07-01T10:00:00Z",
		"equipment_ids": []string{eqID},
	}, token)
	expectStatus(t, w, http.StatusCreated)

	w = ta.doMultipart(t, http.MethodPost, "/api/v1/workouts", token, map[string]string{
		"name": "Imported run", "sport_type": "Run", "start_date": "2026-07-11T10:00:00Z",
		"duration_seconds": "1800", "distance": "5000",
	}, nil)
	expectStatus(t, w, http.StatusCreated)
	created := decodeObject(t, w)
	equipment, _ := created["equipment"].([]any)
	if len(equipment) != 1 {
		t.Fatalf("multipart omit should copy previous equipment, got %#v", created["equipment"])
	}
	item, _ := equipment[0].(map[string]any)
	if item["id"] != eqID {
		t.Fatalf("unexpected equipment: %#v", item)
	}
}

func TestUpdateWorkoutDoesNotCopyPreviousEquipment(t *testing.T) {
	ta := setupTestApp(t)
	ta.register(t, "alice", "alice@example.com", "password12")
	token, _ := ta.login(t, "alice@example.com", "password12")

	w := ta.doJSON(t, http.MethodPost, "/api/v1/equipment", map[string]any{
		"type": "shoes",
		"name": "Road shoes",
	}, token)
	expectStatus(t, w, http.StatusCreated)
	eqID, _ := decodeObject(t, w)["id"].(string)

	w = ta.doJSON(t, http.MethodPost, "/api/v1/workouts", map[string]any{
		"name": "Earlier run", "sport_type": "Run", "start_date": "2026-07-01T10:00:00Z",
		"equipment_ids": []string{eqID},
	}, token)
	expectStatus(t, w, http.StatusCreated)

	w = ta.doJSON(t, http.MethodPost, "/api/v1/workouts", map[string]any{
		"name": "Later run", "sport_type": "Run", "start_date": "2026-07-09T10:00:00Z",
		"equipment_ids": []string{eqID},
	}, token)
	expectStatus(t, w, http.StatusCreated)
	id, _ := decodeObject(t, w)["id"].(string)

	w = ta.doJSON(t, http.MethodPut, "/api/v1/workouts/"+id, map[string]any{
		"name": "Later run", "sport_type": "Run", "start_date": "2026-07-09T10:00:00Z",
		"equipment_ids": []string{},
	}, token)
	expectStatus(t, w, http.StatusOK)
	updated := decodeObject(t, w)
	if eqList, ok := updated["equipment"].([]any); ok && len(eqList) != 0 {
		t.Fatalf("update with empty equipment_ids should clear, got %#v", updated["equipment"])
	}
}

func TestFederatedAuthorAvatarAPI(t *testing.T) {
	ta := setupTestApp(t)
	ta.register(t, "alice", "alice@example.com", "password12")
	token, _ := ta.login(t, "alice@example.com", "password12")

	ownerKey := federation.OwnerKeyFromHandle("bob@remote.test")
	key := keys.FederatedInboxAvatar("alice", ownerKey)
	png := readTestdata(t, "images/avatar-square.png")
	if err := avatars.SaveKey(ta.app.Blobs, key, png); err != nil {
		t.Fatal(err)
	}

	w := ta.doJSON(t, http.MethodGet, "/api/v1/federation/authors/"+ownerKey+"/avatar", nil, token)
	expectStatus(t, w, http.StatusOK)
	if ct := w.Header().Get("Content-Type"); ct != "image/webp" {
		t.Fatalf("content-type = %q", ct)
	}

	w = ta.doJSON(t, http.MethodGet, "/api/v1/federation/authors/missing_key/avatar", nil, token)
	expectStatus(t, w, http.StatusNotFound)
}

// Ensure storage.Open path used by NewApp is exercised via setup.
var _ = storage.Open
