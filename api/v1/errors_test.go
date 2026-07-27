package v1

import (
	"bytes"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestRespondInternalLogsAndReturnsGenericBody(t *testing.T) {
	gin.SetMode(gin.TestMode)

	var buf bytes.Buffer
	prev := slog.Default()
	slog.SetDefault(slog.New(slog.NewJSONHandler(&buf, &slog.HandlerOptions{Level: slog.LevelError})))
	t.Cleanup(func() { slog.SetDefault(prev) })

	router := gin.New()
	router.GET("/boom", func(c *gin.Context) {
		respondInternal(c, "failed to create workout", errors.New("disk full"))
	})

	req := httptest.NewRequest(http.MethodGet, "/boom", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", w.Code)
	}
	var body ErrorResponse
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode body: %v", err)
	}
	if body.Error != "failed to create workout" {
		t.Fatalf("body.Error = %q", body.Error)
	}
	if strings.Contains(w.Body.String(), "disk full") {
		t.Fatal("internal error leaked into response body")
	}

	logLine := buf.String()
	if !strings.Contains(logLine, `"msg":"failed to create workout"`) {
		t.Fatalf("log missing message: %s", logLine)
	}
	if !strings.Contains(logLine, "disk full") {
		t.Fatalf("log missing err: %s", logLine)
	}
	if !strings.Contains(logLine, `"method":"GET"`) {
		t.Fatalf("log missing method: %s", logLine)
	}
}
