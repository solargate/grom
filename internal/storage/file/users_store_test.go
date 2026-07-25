package file

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/solargate/grom/internal/users"
)

func TestUsersStoreCreateAndReload(t *testing.T) {
	dir := t.TempDir()
	store, err := NewUsersStore(dir)
	if err != nil {
		t.Fatalf("NewUsersStore: %v", err)
	}

	created, err := store.Create("Alice", "Alice Name", "Alice@Example.com", "password123")
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if created.Nickname != "Alice" {
		t.Fatalf("nickname = %q", created.Nickname)
	}

	reloaded, err := NewUsersStore(dir)
	if err != nil {
		t.Fatalf("reload: %v", err)
	}
	byEmail, err := reloaded.FindByEmail("alice@example.com")
	if err != nil {
		t.Fatalf("FindByEmail: %v", err)
	}
	if byEmail.ID != created.ID {
		t.Fatalf("reloaded id = %q, want %q", byEmail.ID, created.ID)
	}
	byNick, err := reloaded.FindByNickname("alice")
	if err != nil {
		t.Fatalf("FindByNickname: %v", err)
	}
	if byNick.Email != "Alice@Example.com" {
		t.Fatalf("email = %q", byNick.Email)
	}
}

func TestUsersStoreDuplicateAndInvalidNickname(t *testing.T) {
	store, err := NewUsersStore(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	if _, err := store.Create("alice", "Alice", "alice@example.com", "password123"); err != nil {
		t.Fatal(err)
	}

	if _, err := store.Create("bob", "Bob", "alice@example.com", "password123"); !errors.Is(err, users.ErrEmailTaken) {
		t.Fatalf("duplicate email: %v", err)
	}
	if _, err := store.Create("alice", "Alice", "other@example.com", "password123"); !errors.Is(err, users.ErrNicknameTaken) {
		t.Fatalf("duplicate nickname: %v", err)
	}
	if _, err := store.Create("../evil", "X", "x@example.com", "password123"); !errors.Is(err, users.ErrInvalidNickname) {
		t.Fatalf("invalid nickname: %v", err)
	}
	if _, err := store.Create("bad/name", "X", "y@example.com", "password123"); !errors.Is(err, users.ErrInvalidNickname) {
		t.Fatalf("slash nickname: %v", err)
	}
}

func TestUsersStoreSearchAndUpdateProfile(t *testing.T) {
	store, err := NewUsersStore(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	alice, err := store.Create("alice", "Alice", "alice@example.com", "password123")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := store.Create("bob", "Bobby", "bob@example.com", "password123"); err != nil {
		t.Fatal(err)
	}

	empty, err := store.Search("", "", 10)
	if err != nil || empty != nil {
		t.Fatalf("empty query: got %#v err=%v", empty, err)
	}

	found, err := store.Search("bo", alice.ID, 10)
	if err != nil {
		t.Fatal(err)
	}
	if len(found) != 1 || found[0].Nickname != "bob" {
		t.Fatalf("search results: %#v", found)
	}

	updated, err := store.UpdateProfile(alice.ID, "Alice Updated")
	if err != nil {
		t.Fatal(err)
	}
	if updated.Name != "Alice Updated" {
		t.Fatalf("name = %q", updated.Name)
	}
	if _, err := store.UpdateProfile("missing", "X"); !errors.Is(err, users.ErrUserNotFound) {
		t.Fatalf("missing update: %v", err)
	}
}

func TestUsersStoreCorruptFile(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "users.yaml"), []byte(":::not-yaml"), 0600); err != nil {
		t.Fatal(err)
	}
	if _, err := NewUsersStore(dir); err == nil {
		t.Fatal("expected corrupt users.yaml error")
	}
}
