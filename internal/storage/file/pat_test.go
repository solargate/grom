package file_test

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/solargate/grom/internal/auth/pat"
	"github.com/solargate/grom/internal/storage/file"
)

func TestPATStoreRoundTrip(t *testing.T) {
	dir := t.TempDir()
	store := file.NewPATStore(dir)

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

	all, err := store.ListAll()
	if err != nil || len(all) != 1 || all[0].ID != "id-1" {
		t.Fatalf("ListAll: %#v err=%v", all, err)
	}

	updated := rec
	updated.Name = "Script 2"
	if err := store.Import(updated); err != nil {
		t.Fatal(err)
	}
	got, err := store.GetByHash("hash-1")
	if err != nil || got.Name != "Script 2" || got.TokenPrefix != "grom_pat_ab" {
		t.Fatalf("Import upsert: %#v err=%v", got, err)
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

	path := filepath.Join(dir, "personal_access_tokens.yaml")
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("expected personal_access_tokens.yaml: %v", err)
	}
}
