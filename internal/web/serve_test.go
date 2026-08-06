package web_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/solargate/grom/internal/web"
)

func TestRegisterRoutesServesIndexAndAssets(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	web.RegisterRoutes(router)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/", nil))
	if w.Code != http.StatusOK {
		t.Fatalf("GET / status = %d", w.Code)
	}
	body := w.Body.String()
	if !strings.Contains(body, "<html") && !strings.Contains(strings.ToLower(body), "flutter") && !strings.Contains(body, "DOCTYPE") {
		// index.html from Flutter web build should be HTML
		if len(body) == 0 {
			t.Fatal("expected index.html body")
		}
	}
	ct := w.Header().Get("Content-Type")
	if !strings.Contains(ct, "text/html") {
		t.Fatalf("content-type = %q, want text/html", ct)
	}

	w = httptest.NewRecorder()
	router.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/manifest.json", nil))
	if w.Code != http.StatusOK {
		t.Fatalf("GET /manifest.json status = %d body=%s", w.Code, w.Body.String())
	}
	if !strings.Contains(w.Header().Get("Content-Type"), "json") {
		t.Fatalf("manifest content-type = %q", w.Header().Get("Content-Type"))
	}

	w = httptest.NewRecorder()
	router.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/missing-client-route", nil))
	if w.Code != http.StatusOK {
		t.Fatalf("SPA fallback status = %d", w.Code)
	}
	if !strings.Contains(w.Header().Get("Content-Type"), "text/html") {
		t.Fatalf("SPA fallback content-type = %q", w.Header().Get("Content-Type"))
	}

	w = httptest.NewRecorder()
	router.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/api/v1/status", nil))
	if w.Code != http.StatusNotFound {
		t.Fatalf("API path via UI handler status = %d, want 404", w.Code)
	}
}
