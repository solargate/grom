package file_test

import (
	"errors"
	"testing"
	"time"

	"github.com/solargate/grom/internal/auth/reset"
	"github.com/solargate/grom/internal/storage/file"
)

func TestResetTokenStore(t *testing.T) {
	dir := t.TempDir()
	store := file.NewResetTokenStore(dir)
	now := time.Now().UTC()
	rec := reset.TokenRecord{
		TokenHash: "abc",
		UserID:    "u1",
		ExpiresAt: now.Add(time.Hour),
		CreatedAt: now,
	}
	if err := store.ReplaceForUser("u1", rec); err != nil {
		t.Fatal(err)
	}
	got, err := store.GetByHash("abc")
	if err != nil || got.UserID != "u1" {
		t.Fatalf("get: %#v %v", got, err)
	}
	rec2 := rec
	rec2.TokenHash = "def"
	if err := store.ReplaceForUser("u1", rec2); err != nil {
		t.Fatal(err)
	}
	if _, err := store.GetByHash("abc"); !errors.Is(err, reset.ErrInvalidToken) {
		t.Fatalf("old token: %v", err)
	}
	expired := rec2
	expired.TokenHash = "exp"
	expired.ExpiresAt = now.Add(-time.Minute)
	if err := store.ReplaceForUser("u1", expired); err != nil {
		t.Fatal(err)
	}
	if _, err := store.GetByHash("exp"); !errors.Is(err, reset.ErrInvalidToken) {
		t.Fatalf("expired: %v", err)
	}

	rec3 := rec
	rec3.TokenHash = "to-delete"
	if err := store.ReplaceForUser("u2", rec3); err != nil {
		t.Fatal(err)
	}
	if err := store.DeleteByHash("to-delete"); err != nil {
		t.Fatal(err)
	}
	if _, err := store.GetByHash("to-delete"); !errors.Is(err, reset.ErrInvalidToken) {
		t.Fatalf("after delete: %v", err)
	}
}
