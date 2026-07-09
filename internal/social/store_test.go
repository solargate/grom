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

func TestFollowStoreFindByFollowActivityID(t *testing.T) {
	dir := t.TempDir()
	store, err := NewStore(dir)
	if err != nil {
		t.Fatal(err)
	}

	created, err := store.Create(Follow{
		FollowerID:     "user-a",
		TargetHandle:   "bob@remote.test:8445",
		TargetNickname: "bob",
		Status:         StatusPending,
	})
	if err != nil {
		t.Fatal(err)
	}

	activityID := "https://local.test/users/alice/follows/abc-123"
	updated, err := store.UpdateActivityID(created.ID, activityID)
	if err != nil {
		t.Fatal(err)
	}
	if updated.FollowActivityID != activityID {
		t.Fatalf("expected activity id %q, got %q", activityID, updated.FollowActivityID)
	}

	found, err := store.FindByFollowActivityID(activityID)
	if err != nil {
		t.Fatal(err)
	}
	if found.ID != created.ID {
		t.Fatalf("expected follow %q, got %q", created.ID, found.ID)
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

func TestFollowStoreListActiveByTarget(t *testing.T) {
	dir := t.TempDir()
	store, err := NewStore(dir)
	if err != nil {
		t.Fatal(err)
	}

	if _, err := store.Create(Follow{
		FollowerID:     "user-a",
		TargetHandle:   "bob@localhost",
		TargetNickname: "bob",
		Status:         StatusActive,
	}); err != nil {
		t.Fatal(err)
	}
	if _, err := store.Create(Follow{
		FollowerID:     "user-c",
		TargetHandle:   "bob@localhost",
		TargetNickname: "bob",
		Status:         StatusPending,
	}); err != nil {
		t.Fatal(err)
	}
	if _, err := store.Create(Follow{
		FollowerID:     "user-a",
		TargetHandle:   "alice@localhost",
		TargetNickname: "alice",
		Status:         StatusActive,
	}); err != nil {
		t.Fatal(err)
	}

	list, err := store.ListActiveByTarget("bob@localhost")
	if err != nil {
		t.Fatal(err)
	}
	if len(list) != 1 {
		t.Fatalf("expected 1 active follower for bob, got %d", len(list))
	}
	if list[0].FollowerID != "user-a" {
		t.Fatalf("expected follower user-a, got %q", list[0].FollowerID)
	}
}
