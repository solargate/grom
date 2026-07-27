package logging

import (
	"fmt"
	"log/slog"
	"os"
	"strings"

	"github.com/solargate/grom/internal/config"
	"github.com/solargate/grom/internal/version"
)

// New builds a slog.Logger from logging config (level + format).
func New(cfg config.LoggingConfig) (*slog.Logger, error) {
	level, err := ParseLevel(cfg.Level)
	if err != nil {
		return nil, err
	}
	format := strings.ToLower(strings.TrimSpace(cfg.Format))
	if format == "" {
		format = "json"
	}

	opts := &slog.HandlerOptions{Level: level}
	var handler slog.Handler
	switch format {
	case "text":
		handler = slog.NewTextHandler(os.Stderr, opts)
	case "json":
		handler = slog.NewJSONHandler(os.Stderr, opts)
	default:
		return nil, fmt.Errorf("logging.format must be text or json (got %q)", cfg.Format)
	}

	return slog.New(handler).With(
		"service", "grom",
		"version", version.Version,
	), nil
}

// SetupDefault creates a logger from config and installs it as slog's default.
func SetupDefault(cfg config.LoggingConfig) error {
	logger, err := New(cfg)
	if err != nil {
		return err
	}
	slog.SetDefault(logger)
	return nil
}

// ParseLevel maps config strings to slog levels.
func ParseLevel(s string) (slog.Level, error) {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "", "info":
		return slog.LevelInfo, nil
	case "debug":
		return slog.LevelDebug, nil
	case "warn", "warning":
		return slog.LevelWarn, nil
	case "error":
		return slog.LevelError, nil
	default:
		return 0, fmt.Errorf("logging.level must be one of debug, info, warn, error (got %q)", s)
	}
}
