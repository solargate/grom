package v1_test

import (
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"testing"
	"time"

	"github.com/solargate/grom/internal/auth/reset"
	"github.com/solargate/grom/internal/config"
)

func TestPasswordResetFlow(t *testing.T) {
	ta := setupTestAppWithConfig(t, func(cfg *config.Config) {
		cfg.Server.TLS.Mode = "off"
		cfg.Server.Name = "Grom Test"
		cfg.Federation.Enabled = false
		cfg.Auth.Reset.PublicBaseURL = "https://grom.example.com"
		cfg.Mailer.Driver = "log"
		cfg.Mailer.From = "Grom <noreply@example.com>"
	})

	w := ta.doJSON(t, http.MethodGet, "/api/v1/server-info", nil, "")
	expectStatus(t, w, http.StatusOK)
	info := decodeObject(t, w)
	if info["password_reset_enabled"] != true {
		t.Fatalf("expected password_reset_enabled=true: %#v", info)
	}

	ta.register(t, "alice", "alice@example.com", "password12")

	w = ta.doJSON(t, http.MethodPost, "/api/v1/auth/password/forgot", map[string]string{
		"email": "missing@example.com",
	}, "")
	expectStatus(t, w, http.StatusNoContent)

	w = ta.doJSON(t, http.MethodPost, "/api/v1/auth/password/forgot", map[string]string{
		"email": "alice@example.com",
	}, "")
	expectStatus(t, w, http.StatusNoContent)

	w = ta.doJSON(t, http.MethodPost, "/api/v1/auth/password/reset", map[string]string{
		"token":    "not-a-real-token",
		"password": "newpassword1",
	}, "")
	expectStatus(t, w, http.StatusBadRequest)

	raw := "test-reset-token-raw-value-AAAAAAAA"
	sum := sha256.Sum256([]byte(raw))
	hash := hex.EncodeToString(sum[:])
	user, err := ta.app.Users.FindByEmail("alice@example.com")
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
		"email":    "alice@example.com",
		"password": "password12",
	}, "")
	expectStatus(t, w, http.StatusUnauthorized)

	_, _ = ta.login(t, "alice@example.com", "newpassword1")

	w = ta.doJSON(t, http.MethodPost, "/api/v1/auth/password/reset", map[string]string{
		"token":    raw,
		"password": "anotherpass",
	}, "")
	expectStatus(t, w, http.StatusBadRequest)
}

func TestPasswordResetDisabled(t *testing.T) {
	ta := setupTestApp(t)
	w := ta.doJSON(t, http.MethodGet, "/api/v1/server-info", nil, "")
	expectStatus(t, w, http.StatusOK)
	info := decodeObject(t, w)
	if info["password_reset_enabled"] != false {
		t.Fatalf("expected disabled: %#v", info)
	}
	w = ta.doJSON(t, http.MethodPost, "/api/v1/auth/password/forgot", map[string]string{
		"email": "a@b.c",
	}, "")
	expectStatus(t, w, http.StatusServiceUnavailable)
}
