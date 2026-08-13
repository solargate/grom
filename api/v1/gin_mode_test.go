package v1

import (
	"testing"

	"github.com/gin-gonic/gin"
)

func TestGinModeForLoggingLevel(t *testing.T) {
	tests := []struct {
		name  string
		level string
		want  string
	}{
		{name: "debug", level: "debug", want: gin.DebugMode},
		{name: "info", level: "info", want: gin.ReleaseMode},
		{name: "warn", level: "warn", want: gin.ReleaseMode},
		{name: "error", level: "error", want: gin.ReleaseMode},
		{name: "empty", level: "", want: gin.ReleaseMode},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ginModeForLoggingLevel(tt.level); got != tt.want {
				t.Fatalf("ginModeForLoggingLevel(%q) = %q, want %q", tt.level, got, tt.want)
			}
		})
	}
}
