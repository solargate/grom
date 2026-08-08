package v1_test

import (
	"encoding/json"
	"net/http"
	"testing"

	altcha "github.com/altcha-org/altcha-lib-go/v2"
	"github.com/solargate/grom/internal/auth/captcha"
	"github.com/solargate/grom/internal/config"
)

func enableCaptcha(t *testing.T) *testApp {
	t.Helper()
	return setupTestAppWithConfig(t, func(cfg *config.Config) {
		cfg.Server.TLS.Mode = "off"
		cfg.Federation.Enabled = false
		cfg.Auth.Captcha.Enabled = true
		cfg.Auth.Captcha.Cost = 10
		cfg.Auth.Captcha.ExpiresSeconds = 60
		cfg.Auth.Captcha.HMACSecret = "api-test-captcha-hmac-secret!!"
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

func TestCaptchaChallenge_Disabled(t *testing.T) {
	ta := setupTestApp(t)
	w := ta.doJSON(t, http.MethodGet, "/api/v1/captcha/challenge", nil, "")
	if w.Code != http.StatusNotFound {
		t.Fatalf("status = %d", w.Code)
	}
}

func TestCaptcha_RegisterLoginForgot(t *testing.T) {
	ta := enableCaptcha(t)

	w := ta.doJSON(t, http.MethodGet, "/api/v1/server-info", nil, "")
	if w.Code != http.StatusOK {
		t.Fatalf("server-info = %d", w.Code)
	}
	var info map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &info); err != nil {
		t.Fatal(err)
	}
	if info["captcha_enabled"] != true {
		t.Fatalf("captcha_enabled = %#v", info["captcha_enabled"])
	}

	w = ta.doJSON(t, http.MethodPost, "/api/v1/auth/register", map[string]string{
		"nickname": "capuser",
		"name":     "Cap User",
		"email":    "cap@example.com",
		"password": "secret123",
	}, "")
	if w.Code != http.StatusBadRequest {
		t.Fatalf("register without captcha = %d", w.Code)
	}

	payload := solveChallengeFromApp(t, ta)
	w = ta.doJSON(t, http.MethodPost, "/api/v1/auth/register", map[string]string{
		"nickname": "capuser",
		"name":     "Cap User",
		"email":    "cap@example.com",
		"password": "secret123",
		"altcha":   payload,
	}, "")
	if w.Code != http.StatusCreated {
		t.Fatalf("register = %d body=%s", w.Code, w.Body.String())
	}

	w = ta.doJSON(t, http.MethodPost, "/api/v1/auth/login", map[string]string{
		"email":    "cap@example.com",
		"password": "secret123",
	}, "")
	if w.Code != http.StatusBadRequest {
		t.Fatalf("login without captcha = %d", w.Code)
	}

	payload = solveChallengeFromApp(t, ta)
	w = ta.doJSON(t, http.MethodPost, "/api/v1/auth/login", map[string]string{
		"email":    "cap@example.com",
		"password": "secret123",
		"altcha":   payload,
	}, "")
	if w.Code != http.StatusOK {
		t.Fatalf("login = %d body=%s", w.Code, w.Body.String())
	}
}

func TestCaptcha_ForgotRequiresPayload(t *testing.T) {
	ta := setupTestAppWithConfig(t, func(cfg *config.Config) {
		cfg.Server.TLS.Mode = "off"
		cfg.Federation.Enabled = false
		cfg.Auth.Captcha.Enabled = true
		cfg.Auth.Captcha.Cost = 10
		cfg.Auth.Captcha.HMACSecret = "api-test-captcha-hmac-secret!!"
		cfg.Mailer.Driver = "log"
		cfg.Mailer.From = "Grom <noreply@example.com>"
		cfg.Auth.Reset.PublicBaseURL = "https://grom.example.com"
	})

	w := ta.doJSON(t, http.MethodPost, "/api/v1/auth/password/forgot", map[string]string{
		"email": "nobody@example.com",
	}, "")
	if w.Code != http.StatusBadRequest {
		t.Fatalf("forgot without captcha = %d", w.Code)
	}

	payload := solveChallengeFromApp(t, ta)
	w = ta.doJSON(t, http.MethodPost, "/api/v1/auth/password/forgot", map[string]string{
		"email":  "nobody@example.com",
		"altcha": payload,
	}, "")
	if w.Code != http.StatusNoContent {
		t.Fatalf("forgot = %d body=%s", w.Code, w.Body.String())
	}
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
	if w.Code != http.StatusCreated {
		t.Fatalf("first = %d", w.Code)
	}
	body["nickname"] = "replay2"
	body["email"] = "replay2@example.com"
	w = ta.doJSON(t, http.MethodPost, "/api/v1/auth/register", body, "")
	if w.Code != http.StatusBadRequest {
		t.Fatalf("replay = %d", w.Code)
	}
}
