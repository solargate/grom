package web_test

import (
	"io/fs"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/solargate/grom/internal/web"
)

func TestRegisterRoutesServesUIAndSkipsAPI(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	web.RegisterRoutes(router)

	// UI handler must not claim /api/* (even with an empty embed).
	w := httptest.NewRecorder()
	router.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/api/v1/status", nil))
	if w.Code != http.StatusNotFound {
		t.Fatalf("API path via UI handler status = %d, want 404", w.Code)
	}

	dist, err := fs.Sub(web.Assets, "dist")
	if err != nil {
		t.Fatalf("embed dist: %v", err)
	}
	_, err = fs.Stat(dist, "index.html")
	if err != nil {
		// CI checkout only has dist/.gitkeep; Flutter web assets are build artifacts.
		w = httptest.NewRecorder()
		router.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/", nil))
		if w.Code != http.StatusNotFound {
			t.Fatalf("GET / without embedded index.html status = %d, want 404", w.Code)
		}
		return
	}

	w = httptest.NewRecorder()
	router.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/", nil))
	if w.Code != http.StatusOK {
		t.Fatalf("GET / status = %d", w.Code)
	}
	if !strings.Contains(w.Header().Get("Content-Type"), "text/html") {
		t.Fatalf("content-type = %q, want text/html", w.Header().Get("Content-Type"))
	}
	if len(w.Body.Bytes()) == 0 {
		t.Fatal("expected index.html body")
	}

	w = httptest.NewRecorder()
	router.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/missing-client-route", nil))
	if w.Code != http.StatusOK {
		t.Fatalf("SPA fallback status = %d", w.Code)
	}
	if !strings.Contains(w.Header().Get("Content-Type"), "text/html") {
		t.Fatalf("SPA fallback content-type = %q", w.Header().Get("Content-Type"))
	}
}
