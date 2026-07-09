package social

import (
	"os"
	"path/filepath"
	"testing"
)

func TestFollowStoreLocalFollow(t *testing.T) {
	dir := t.TempDir()
	store, err := NewStore(dir)
	if err != nil {
		t.Fatal(err)
	}

	follow := Follow{
		FollowerID:     "user-a",
		TargetHandle:   "bob@localhost",
		TargetNickname: "bob",
		TargetName:     "Bob",
		TargetIsLocal:  true,
		Status:         StatusActive,
	}
	created, err := store.Create(follow)
	if err != nil {
		t.Fatal(err)
	}
	if created.ID == "" {
		t.Fatal("expected follow id")
	}

	list, err := store.ListActiveFollowing("user-a")
	if err != nil {
		t.Fatal(err)
	}
	if len(list) != 1 {
		t.Fatalf("expected 1 follow, got %d", len(list))
	}

	if err := store.Delete(created.ID); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(dir, "federation", followsFileName)); err != nil {
		t.Fatal(err)
	}
}

func TestFollowStoreDuplicate(t *testing.T) {
	dir := t.TempDir()
	store, err := NewStore(dir)
	if err != nil {
		t.Fatal(err)
	}

	follow := Follow{
		FollowerID:     "user-a",
		TargetHandle:   "bob@localhost",
		TargetNickname: "bob",
		Status:         StatusActive,
	}
	if _, err := store.Create(follow); err != nil {
		t.Fatal(err)
	}
	if _, err := store.Create(follow); err != ErrAlreadyFollowing {
		t.Fatalf("expected ErrAlreadyFollowing, got %v", err)
	}
}
