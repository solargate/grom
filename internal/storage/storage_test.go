package storage_test

import (
	"context"
	"testing"

	"github.com/solargate/grom/internal/config"
	"github.com/solargate/grom/internal/storage"
)

func TestOpenFileBackend(t *testing.T) {
	dir := t.TempDir()

	cfg := config.StorageConfig{
		Driver:           config.StorageDriverFile,
		Location:           dir,
		ResolvedLocation: dir,
	}

	backend, err := storage.Open(cfg)
	if err != nil {
		t.Fatal(err)
	}
	defer backend.Close()

	if err := backend.Ping(context.Background()); err != nil {
		t.Fatalf("Ping() error = %v", err)
	}
	if backend.Users() == nil || backend.Workouts() == nil || backend.Equipment() == nil {
		t.Fatal("backend repositories should be initialized")
	}
	if backend.Social() == nil || backend.Federation() == nil || backend.Blobs() == nil {
		t.Fatal("backend social/federation/blobs should be initialized")
	}
}

func TestOpenUnsupportedDriver(t *testing.T) {
	_, err := storage.Open(config.StorageConfig{Driver: config.StorageDriverPostgres})
	if err == nil {
		t.Fatal("expected error for unsupported driver")
	}
}
