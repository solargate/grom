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
	"github.com/solargate/grom/internal/federation"
	"github.com/solargate/grom/internal/federation/httpsig"
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
	afOff := false
	return setupTestAppWithConfig(t, func(cfg *config.Config) {
		cfg.Server.TLS.Mode = "static"
		cfg.Server.TLS.CertFile = certPath
		cfg.Server.TLS.KeyFile = keyPath
		cfg.Federation.Enabled = true
		cfg.Federation.Domain = "localhost"
		cfg.Federation.AutoAcceptFollows = false
		cfg.Federation.AuthorizedFetch = &afOff
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

// remoteTestActor is a stable actor URI used by signed federation inbox tests.
const remoteTestActor = "https://remote.example/users/bob"

func newRemoteTestKey(t *testing.T) (*rsa.PrivateKey, string) {
	t.Helper()
	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}
	return priv, remoteTestActor + "#main-key"
}

func (ta *testApp) installRemoteKey(t *testing.T, priv *rsa.PrivateKey, keyID, owner string) {
	t.Helper()
	ta.app.SetFederationKeyResolver(federation.StaticKeyResolver{
		Keys: map[string]federation.StaticKey{
			keyID: {Public: &priv.PublicKey, Owner: owner},
		},
	})
}

func (ta *testApp) postSignedActivity(t *testing.T, path string, activity map[string]any, priv *rsa.PrivateKey, keyID string) *httptest.ResponseRecorder {
	t.Helper()
	data, err := json.Marshal(activity)
	if err != nil {
		t.Fatal(err)
	}
	req := httptest.NewRequest(http.MethodPost, path, bytes.NewReader(data))
	req.Header.Set("Content-Type", "application/activity+json")
	req.Host = "localhost"
	if err := httpsig.SignPOST(req, data, priv, keyID); err != nil {
		t.Fatalf("SignPOST: %v", err)
	}
	w := httptest.NewRecorder()
	ta.router.ServeHTTP(w, req)
	return w
}
