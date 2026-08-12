package auth_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/solargate/grom/internal/auth"
	"github.com/solargate/grom/internal/auth/pat"
	"github.com/solargate/grom/internal/config"
)

type mockPATAuthenticator struct {
	record *pat.TokenRecord
	err    error
}

func (m *mockPATAuthenticator) Authenticate(string) (*pat.TokenRecord, error) {
	return m.record, m.err
}

func withTestAuthConfig(t *testing.T) {
	t.Helper()
	prev := config.Cfg
	t.Cleanup(func() { config.Cfg = prev })
	config.Cfg = config.Config{}
	config.Cfg.Auth.JWTSecret = "test-secret-at-least-32-characters!!"
	config.Cfg.Auth.JWTTTLHours = 24
}

func TestGenerateAndValidateToken(t *testing.T) {
	withTestAuthConfig(t)

	token, expiresAt, err := auth.GenerateToken("user-1", "a@example.com")
	if err != nil {
		t.Fatalf("GenerateToken: %v", err)
	}
	if token == "" {
		t.Fatal("expected token")
	}
	if expiresAt.Before(time.Now().UTC()) {
		t.Fatal("expiresAt should be in the future")
	}

	claims, err := auth.ValidateToken(token)
	if err != nil {
		t.Fatalf("ValidateToken: %v", err)
	}
	if claims.Subject != "user-1" {
		t.Fatalf("subject = %q", claims.Subject)
	}
	if claims.Email != "a@example.com" {
		t.Fatalf("email = %q", claims.Email)
	}
	if claims.Issuer != "grom" {
		t.Fatalf("issuer = %q", claims.Issuer)
	}
}

func TestValidateTokenRejectsTampered(t *testing.T) {
	withTestAuthConfig(t)

	token, _, err := auth.GenerateToken("user-1", "a@example.com")
	if err != nil {
		t.Fatal(err)
	}
	tampered := token[:len(token)-2] + "xx"
	if _, err := auth.ValidateToken(tampered); err == nil {
		t.Fatal("expected error for tampered token")
	}
}

func TestValidateTokenRejectsExpired(t *testing.T) {
	withTestAuthConfig(t)

	claims := auth.Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   "user-1",
			ExpiresAt: jwt.NewNumericDate(time.Now().UTC().Add(-time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now().UTC().Add(-2 * time.Hour)),
			Issuer:    "grom",
		},
		Email: "a@example.com",
	}
	signed, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).
		SignedString([]byte(config.Cfg.Auth.JWTSecret))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := auth.ValidateToken(signed); err == nil {
		t.Fatal("expected error for expired token")
	}
}

func TestValidateTokenRejectsWrongAlg(t *testing.T) {
	withTestAuthConfig(t)

	claims := auth.Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   "user-1",
			ExpiresAt: jwt.NewNumericDate(time.Now().UTC().Add(time.Hour)),
			Issuer:    "grom",
		},
		Email: "a@example.com",
	}
	// Unsigned "none" is rejected by WithValidMethods / method check.
	token := jwt.NewWithClaims(jwt.SigningMethodNone, claims)
	signed, err := token.SignedString(jwt.UnsafeAllowNoneSignatureType)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := auth.ValidateToken(signed); err == nil {
		t.Fatal("expected error for none algorithm")
	}
}

func TestHashAndCheckPassword(t *testing.T) {
	hash, err := auth.HashPassword("secret123")
	if err != nil {
		t.Fatal(err)
	}
	if !auth.CheckPassword(hash, "secret123") {
		t.Fatal("expected password to match")
	}
	if auth.CheckPassword(hash, "wrong") {
		t.Fatal("expected password mismatch")
	}
}

func TestAuthRequiredMiddleware(t *testing.T) {
	withTestAuthConfig(t)
	gin.SetMode(gin.TestMode)

	token, _, err := auth.GenerateToken("user-42", "u@example.com")
	if err != nil {
		t.Fatal(err)
	}

	cases := []struct {
		name   string
		header string
		want   int
		wantID string
	}{
		{name: "missing", want: http.StatusUnauthorized},
		{name: "malformed", header: "Token abc", want: http.StatusUnauthorized},
		{name: "invalid", header: "Bearer not-a-token", want: http.StatusUnauthorized},
		{name: "valid", header: "Bearer " + token, want: http.StatusOK, wantID: "user-42"},
		{name: "valid lower bearer", header: "bearer " + token, want: http.StatusOK, wantID: "user-42"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			r := gin.New()
			r.GET("/x", auth.AuthRequired(), func(c *gin.Context) {
				id, _ := c.Get(auth.ContextUserIDKey)
				c.String(http.StatusOK, "%v", id)
			})
			req := httptest.NewRequest(http.MethodGet, "/x", nil)
			if tc.header != "" {
				req.Header.Set("Authorization", tc.header)
			}
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)
			if w.Code != tc.want {
				t.Fatalf("status = %d, want %d, body=%s", w.Code, tc.want, w.Body.String())
			}
			if tc.wantID != "" && w.Body.String() != tc.wantID {
				t.Fatalf("body = %q, want %q", w.Body.String(), tc.wantID)
			}
		})
	}
}

func TestAuthRequiredRejectsPAT(t *testing.T) {
	withTestAuthConfig(t)
	gin.SetMode(gin.TestMode)

	r := gin.New()
	r.GET("/x", auth.AuthRequired(), func(c *gin.Context) {
		c.Status(http.StatusOK)
	})
	req := httptest.NewRequest(http.MethodGet, "/x", nil)
	req.Header.Set("Authorization", "Bearer grom_pat_testsecretvalue")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d body=%s", w.Code, http.StatusUnauthorized, w.Body.String())
	}
}

func TestAuthAPIWithPATInsufficientScope(t *testing.T) {
	withTestAuthConfig(t)
	gin.SetMode(gin.TestMode)

	authenticator := &mockPATAuthenticator{
		record: &pat.TokenRecord{
			UserID: "user-1",
			Scopes: []string{pat.ScopeWorkoutsRead},
		},
	}

	r := gin.New()
	r.GET("/x", auth.AuthAPI(authenticator, pat.ScopeWorkoutsWrite), func(c *gin.Context) {
		c.Status(http.StatusOK)
	})
	req := httptest.NewRequest(http.MethodGet, "/x", nil)
	req.Header.Set("Authorization", "Bearer grom_pat_testsecretvalue")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want %d body=%s", w.Code, http.StatusForbidden, w.Body.String())
	}
}

func TestAuthAPIWithPATValidScope(t *testing.T) {
	withTestAuthConfig(t)
	gin.SetMode(gin.TestMode)

	authenticator := &mockPATAuthenticator{
		record: &pat.TokenRecord{
			UserID: "user-42",
			ID:     "pat-1",
			Scopes: []string{pat.ScopeWorkoutsRead},
		},
	}

	r := gin.New()
	r.GET("/x", auth.AuthAPI(authenticator, pat.ScopeWorkoutsRead), func(c *gin.Context) {
		id, _ := c.Get(auth.ContextUserIDKey)
		c.String(http.StatusOK, "%v", id)
	})
	req := httptest.NewRequest(http.MethodGet, "/x", nil)
	req.Header.Set("Authorization", "Bearer grom_pat_testsecretvalue")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d body=%s", w.Code, http.StatusOK, w.Body.String())
	}
	if w.Body.String() != "user-42" {
		t.Fatalf("body = %q, want user-42", w.Body.String())
	}
}
