package v1_test

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	altcha "github.com/altcha-org/altcha-lib-go/v2"
	"github.com/solargate/grom/internal/auth/captcha"
	"github.com/solargate/grom/internal/auth/reset"
	"github.com/solargate/grom/internal/config"
)

const apiTestCaptchaHMAC = "api-test-captcha-hmac-secret!!"

func enableCaptcha(t *testing.T) *testApp {
	t.Helper()
	return setupTestAppWithConfig(t, func(cfg *config.Config) {
		cfg.Server.TLS.Mode = "off"
		cfg.Federation.Enabled = false
		cfg.Auth.Captcha.Enabled = true
		cfg.Auth.Captcha.Cost = 10
		cfg.Auth.Captcha.ExpiresSeconds = 60
		cfg.Auth.Captcha.HMACSecret = apiTestCaptchaHMAC
	})
}

func enableCaptchaWithReset(t *testing.T) *testApp {
	t.Helper()
	return setupTestAppWithConfig(t, func(cfg *config.Config) {
		cfg.Server.TLS.Mode = "off"
		cfg.Federation.Enabled = false
		cfg.Auth.Captcha.Enabled = true
		cfg.Auth.Captcha.Cost = 10
		cfg.Auth.Captcha.ExpiresSeconds = 60
		cfg.Auth.Captcha.HMACSecret = apiTestCaptchaHMAC
		cfg.Mailer.Driver = "log"
		cfg.Mailer.From = "Grom <noreply@example.com>"
		cfg.Auth.Reset.PublicBaseURL = "https://grom.example.com"
	})
}

func solveChallengeFromApp(t *testing.T, ta *testApp) string {
	t.Helper()
	w := ta.doJSON(t, http.MethodGet, "/api/v1/captcha/challenge", nil, "")
	if w.Code != http.StatusOK {
		t.Fatalf("challenge status = %d body=%s", w.Code, w.Body.String())
	}
	var challenge altcha.Challenge
	if err := json.Unmarshal(w.Body.Bytes(), &challenge); err != nil {
		t.Fatalf("decode challenge: %v", err)
	}
	solution, err := altcha.SolveChallenge(altcha.SolveChallengeOptions{
		Challenge: challenge,
		DeriveKey: altcha.DeriveKeyPBKDF2(),
	})
	if err != nil || solution == nil {
		t.Fatalf("SolveChallenge: %v %#v", err, solution)
	}
	encoded, err := captcha.EncodePayload(altcha.Payload{
		Challenge: challenge,
		Solution:  *solution,
	})
	if err != nil {
		t.Fatalf("EncodePayload: %v", err)
	}
	return encoded
}

func expiredCaptchaPayload(t *testing.T) string {
	t.Helper()
	expired := time.Now().Add(-time.Minute)
	counter := 10
	challenge, err := altcha.CreateChallenge(altcha.CreateChallengeOptions{
		Algorithm:           "PBKDF2/SHA-256",
		DeriveKey:           altcha.DeriveKeyPBKDF2(),
		HMACSignatureSecret: apiTestCaptchaHMAC,
		Cost:                10,
		KeyLength:           32,
		Counter:             &counter,
		ExpiresAt:           &expired,
	})
	if err != nil {
		t.Fatalf("CreateChallenge: %v", err)
	}
	encoded, err := captcha.EncodePayload(altcha.Payload{
		Challenge: challenge,
		Solution:  altcha.Solution{Counter: 0, DerivedKey: "unused"},
	})
	if err != nil {
		t.Fatalf("EncodePayload: %v", err)
	}
	return encoded
}

func expectCaptchaError(t *testing.T, w *httptest.ResponseRecorder, wantStatus int, wantMsg string) {
	t.Helper()
	expectStatus(t, w, wantStatus)
	out := decodeObject(t, w)
	if out["error"] != wantMsg {
		t.Fatalf("error = %#v, want %q", out["error"], wantMsg)
	}
}

func TestCaptchaChallenge_Disabled(t *testing.T) {
	ta := setupTestApp(t)
	w := ta.doJSON(t, http.MethodGet, "/api/v1/captcha/challenge", nil, "")
	expectCaptchaError(t, w, http.StatusNotFound, "captcha is disabled")
}

func TestCaptcha_RegisterLoginForgot(t *testing.T) {
	ta := enableCaptcha(t)

	w := ta.doJSON(t, http.MethodGet, "/api/v1/server-info", nil, "")
	expectStatus(t, w, http.StatusOK)
	info := decodeObject(t, w)
	if info["captcha_enabled"] != true {
		t.Fatalf("captcha_enabled = %#v", info["captcha_enabled"])
	}

	w = ta.doJSON(t, http.MethodPost, "/api/v1/auth/register", map[string]string{
		"nickname": "capuser",
		"name":     "Cap User",
		"email":    "cap@example.com",
		"password": "secret123",
	}, "")
	expectCaptchaError(t, w, http.StatusBadRequest, "captcha is required")

	payload := solveChallengeFromApp(t, ta)
	w = ta.doJSON(t, http.MethodPost, "/api/v1/auth/register", map[string]string{
		"nickname": "capuser",
		"name":     "Cap User",
		"email":    "cap@example.com",
		"password": "secret123",
		"altcha":   payload,
	}, "")
	expectStatus(t, w, http.StatusCreated)

	w = ta.doJSON(t, http.MethodPost, "/api/v1/auth/login", map[string]string{
		"email":    "cap@example.com",
		"password": "secret123",
	}, "")
	expectCaptchaError(t, w, http.StatusBadRequest, "captcha is required")

	payload = solveChallengeFromApp(t, ta)
	w = ta.doJSON(t, http.MethodPost, "/api/v1/auth/login", map[string]string{
		"email":    "cap@example.com",
		"password": "secret123",
		"altcha":   payload,
	}, "")
	expectStatus(t, w, http.StatusOK)
}

func TestCaptcha_ForgotRequiresPayload(t *testing.T) {
	ta := enableCaptchaWithReset(t)

	w := ta.doJSON(t, http.MethodPost, "/api/v1/auth/password/forgot", map[string]string{
		"email": "nobody@example.com",
	}, "")
	expectCaptchaError(t, w, http.StatusBadRequest, "captcha is required")

	payload := solveChallengeFromApp(t, ta)
	w = ta.doJSON(t, http.MethodPost, "/api/v1/auth/password/forgot", map[string]string{
		"email":  "nobody@example.com",
		"altcha": payload,
	}, "")
	expectStatus(t, w, http.StatusNoContent)
}

func TestCaptcha_ResetDoesNotRequirePayload(t *testing.T) {
	ta := enableCaptchaWithReset(t)

	payload := solveChallengeFromApp(t, ta)
	w := ta.doJSON(t, http.MethodPost, "/api/v1/auth/register", map[string]string{
		"nickname": "resetcap",
		"name":     "Reset Cap",
		"email":    "resetcap@example.com",
		"password": "password12",
		"altcha":   payload,
	}, "")
	expectStatus(t, w, http.StatusCreated)

	raw := "test-reset-token-captcha-AAAAAAAA"
	sum := sha256.Sum256([]byte(raw))
	hash := hex.EncodeToString(sum[:])
	user, err := ta.app.Users.FindByEmail("resetcap@example.com")
	if err != nil {
		t.Fatal(err)
	}
	now := time.Now().UTC()
	if err := ta.app.Backend.ResetTokens().ReplaceForUser(user.ID, reset.TokenRecord{
		TokenHash: hash,
		UserID:    user.ID,
		ExpiresAt: now.Add(time.Hour),
		CreatedAt: now,
	}); err != nil {
		t.Fatal(err)
	}

	w = ta.doJSON(t, http.MethodPost, "/api/v1/auth/password/reset", map[string]string{
		"token":    raw,
		"password": "newpassword1",
	}, "")
	expectStatus(t, w, http.StatusNoContent)

	w = ta.doJSON(t, http.MethodPost, "/api/v1/auth/login", map[string]string{
		"email":    "resetcap@example.com",
		"password": "newpassword1",
	}, "")
	expectCaptchaError(t, w, http.StatusBadRequest, "captcha is required")

	payload = solveChallengeFromApp(t, ta)
	w = ta.doJSON(t, http.MethodPost, "/api/v1/auth/login", map[string]string{
		"email":    "resetcap@example.com",
		"password": "newpassword1",
		"altcha":   payload,
	}, "")
	expectStatus(t, w, http.StatusOK)
}

func TestCaptcha_ReplayRejected(t *testing.T) {
	ta := enableCaptcha(t)
	payload := solveChallengeFromApp(t, ta)
	body := map[string]string{
		"nickname": "replay1",
		"name":     "R",
		"email":    "replay1@example.com",
		"password": "secret123",
		"altcha":   payload,
	}
	w := ta.doJSON(t, http.MethodPost, "/api/v1/auth/register", body, "")
	expectStatus(t, w, http.StatusCreated)
	body["nickname"] = "replay2"
	body["email"] = "replay2@example.com"
	w = ta.doJSON(t, http.MethodPost, "/api/v1/auth/register", body, "")
	expectCaptchaError(t, w, http.StatusBadRequest, "captcha already used")
}

func TestCaptcha_InvalidAndExpiredMessages(t *testing.T) {
	ta := enableCaptcha(t)

	w := ta.doJSON(t, http.MethodPost, "/api/v1/auth/login", map[string]string{
		"email":    "nobody@example.com",
		"password": "secret123",
		"altcha":   "not-a-valid-payload",
	}, "")
	expectCaptchaError(t, w, http.StatusBadRequest, "invalid captcha")

	w = ta.doJSON(t, http.MethodPost, "/api/v1/auth/login", map[string]string{
		"email":    "nobody@example.com",
		"password": "secret123",
		"altcha":   expiredCaptchaPayload(t),
	}, "")
	expectCaptchaError(t, w, http.StatusBadRequest, "captcha expired")
}

func TestCaptchaChallenge_RateLimited(t *testing.T) {
	ta := enableCaptcha(t)
	for i := 0; i < 60; i++ {
		w := ta.doJSON(t, http.MethodGet, "/api/v1/captcha/challenge", nil, "")
		if w.Code != http.StatusOK {
			t.Fatalf("challenge %d status = %d body=%s", i, w.Code, w.Body.String())
		}
	}
	w := ta.doJSON(t, http.MethodGet, "/api/v1/captcha/challenge", nil, "")
	expectStatus(t, w, http.StatusTooManyRequests)
	out := decodeObject(t, w)
	if out["error"] != "too many requests, try again later" {
		t.Fatalf("error = %#v", out["error"])
	}
	if w.Header().Get("Retry-After") == "" {
		t.Fatal("expected Retry-After header")
	}
}
