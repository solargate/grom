package storage

import (
	"context"
	"fmt"
	"strings"

	"github.com/solargate/grom/internal/config"
	"github.com/solargate/grom/internal/storage/file"
)

func Open(cfg config.StorageConfig) (Backend, error) {
	driver := strings.TrimSpace(string(cfg.Driver))
	if driver == "" {
		driver = string(config.StorageDriverFile)
	}

	switch config.StorageDriver(driver) {
	case config.StorageDriverFile:
		return file.Open(cfg.ResolvedLocation)
	default:
		return nil, fmt.Errorf("unsupported storage driver %q (supported: file)", driver)
	}
}

func MustOpen(cfg config.StorageConfig) Backend {
	backend, err := Open(cfg)
	if err != nil {
		panic(err)
	}
	return backend
}

// Ping verifies the backend is reachable. Convenience wrapper around Backend.Ping.
func Ping(ctx context.Context, backend Backend) error {
	if backend == nil {
		return fmt.Errorf("storage backend is nil")
	}
	return backend.Ping(ctx)
}
