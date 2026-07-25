package file

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestOpenRequiresLocation(t *testing.T) {
	if _, err := Open(""); err == nil {
		t.Fatal("expected error for empty location")
	}
}

func TestOpenWiresRepositoriesAndPing(t *testing.T) {
	dir := t.TempDir()
	backend, err := Open(dir)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	t.Cleanup(func() { _ = backend.Close() })

	if backend.Location() != dir {
		t.Fatalf("Location() = %q, want %q", backend.Location(), dir)
	}
	if backend.Users() == nil || backend.Workouts() == nil || backend.Equipment() == nil {
		t.Fatal("expected users/workouts/equipment repositories")
	}
	if backend.Social() == nil || backend.Federation() == nil || backend.Blobs() == nil {
		t.Fatal("expected social/federation/blobs")
	}

	if err := backend.Ping(context.Background()); err != nil {
		t.Fatalf("Ping: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if err := backend.Ping(ctx); err == nil {
		t.Fatal("expected cancelled context error")
	}

	filePath := filepath.Join(t.TempDir(), "not-a-dir")
	if err := os.WriteFile(filePath, []byte("x"), 0600); err != nil {
		t.Fatal(err)
	}
	bad, err := Open(filePath)
	if err != nil {
		// Open may succeed even for a file path until Ping; accept either behavior.
		return
	}
	if err := bad.Ping(context.Background()); err == nil {
		t.Fatal("expected Ping failure for non-directory")
	}
}

func mustTime(t *testing.T, value string) time.Time {
	t.Helper()
	parsed, err := time.Parse(time.RFC3339, value)
	if err != nil {
		t.Fatal(err)
	}
	return parsed
}
