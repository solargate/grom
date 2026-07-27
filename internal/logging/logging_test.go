package logging_test

import (
	"log/slog"
	"testing"

	"github.com/solargate/grom/internal/config"
	"github.com/solargate/grom/internal/logging"
)

func TestParseLevel(t *testing.T) {
	tests := []struct {
		in   string
		want slog.Level
		ok   bool
	}{
		{"", slog.LevelInfo, true},
		{"info", slog.LevelInfo, true},
		{"DEBUG", slog.LevelDebug, true},
		{"warn", slog.LevelWarn, true},
		{"warning", slog.LevelWarn, true},
		{"error", slog.LevelError, true},
		{"trace", 0, false},
	}
	for _, tt := range tests {
		got, err := logging.ParseLevel(tt.in)
		if tt.ok {
			if err != nil {
				t.Fatalf("ParseLevel(%q): %v", tt.in, err)
			}
			if got != tt.want {
				t.Fatalf("ParseLevel(%q) = %v, want %v", tt.in, got, tt.want)
			}
			continue
		}
		if err == nil {
			t.Fatalf("ParseLevel(%q) expected error", tt.in)
		}
	}
}

func TestNew_TextAndJSON(t *testing.T) {
	for _, format := range []string{"text", "json"} {
		logger, err := logging.New(config.LoggingConfig{Level: "debug", Format: format})
		if err != nil {
			t.Fatalf("New(%s): %v", format, err)
		}
		if logger == nil {
			t.Fatalf("New(%s): nil logger", format)
		}
	}
}

func TestNew_InvalidFormat(t *testing.T) {
	_, err := logging.New(config.LoggingConfig{Level: "info", Format: "yaml"})
	if err == nil {
		t.Fatal("expected error for invalid format")
	}
}
