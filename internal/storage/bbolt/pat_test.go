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
		CreatedAt:   time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC),
		ExpiresAt:   &expires,
	}
	if err := store.Create(rec); err != nil {
		t.Fatal(err)
	}

	got, err := store.GetByHash("hash-1")
	if err != nil {
		t.Fatal(err)
	}
	if got.Name != "Script" || got.TokenPrefix != "grom_pat_ab" {
		t.Fatalf("unexpected record: %#v", got)
	}

	list, err := store.ListByUser("user-1")
	if err != nil || len(list) != 1 {
		t.Fatalf("list: %#v err=%v", list, err)
	}

	if count, err := store.CountByUser("user-1"); err != nil || count != 1 {
		t.Fatalf("count = %d err=%v", count, err)
	}

	at := time.Date(2026, 8, 2, 0, 0, 0, 0, time.UTC)
	if err := store.UpdateLastUsed("id-1", at); err != nil {
		t.Fatal(err)
	}
	got, err = store.GetByHash("hash-1")
	if err != nil || got.LastUsedAt == nil || !got.LastUsedAt.Equal(at) {
		t.Fatalf("last used: %#v err=%v", got, err)
	}

	if err := store.DeleteByUserAndID("user-1", "id-1"); err != nil {
		t.Fatal(err)
	}
	if _, err := store.GetByHash("hash-1"); err == nil {
		t.Fatal("expected missing token")
	}
}
