package bbolt_test

import (
	"path/filepath"
	"testing"
	"time"

	"github.com/solargate/grom/internal/auth/pat"
	"github.com/solargate/grom/internal/storage/bbolt"
)

func TestPATStoreRoundTrip(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "meta.db")
	backend, err := bbolt.Open(dbPath, dir)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = backend.Close() })

	store := backend.PAT()
	expires := time.Date(2026, 11, 1, 0, 0, 0, 0, time.UTC)
	rec := pat.TokenRecord{
		ID:          "id-1",
		TokenHash:   "hash-1",
		TokenPrefix: "grom_pat_ab",
		UserID:      "user-1",
		Name:        "Script",
		Scopes:      []string{pat.ScopeWorkoutsRead},
		CreatedAt:   time.Now().UTC(),
		ExpiresAt:   &expires,
	}
	if err := store.Create(rec); err != nil {
		t.Fatal(err)
	}
	got, err := store.GetByHash("hash-1")
	if err != nil {
		t.Fatal(err)
	}
	if got.UserID != "user-1" {
		t.Fatalf("user_id = %q", got.UserID)
	}
}
