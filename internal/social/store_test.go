package social_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/solargate/grom/internal/social"
	"github.com/solargate/grom/internal/storage/file"
)

const followsFileName = "follows.yaml"

func newTestStore(t *testing.T) *file.SocialStore {
	t.Helper()
	store, err := file.NewSocialStore(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	return store
}

func TestFollowStoreLocalFollow(t *testing.T) {
	dir := t.TempDir()
	store, err := file.NewSocialStore(dir)
	if err != nil {
		t.Fatal(err)
	}

	follow := social.Follow{
		FollowerID:     "user-a",
		TargetHandle:   "bob@localhost",
		TargetNickname: "bob",
		TargetName:     "Bob",
		TargetIsLocal:  true,
		Status:         social.StatusActive,
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
	store := newTestStore(t)

	created, err := store.Create(social.Follow{
		FollowerID:     "user-a",
		TargetHandle:   "bob@remote.test:8445",
		TargetNickname: "bob",
		Status:         social.StatusPending,
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
	store := newTestStore(t)

	follow := social.Follow{
		FollowerID:     "user-a",
		TargetHandle:   "bob@localhost",
		TargetNickname: "bob",
		Status:         social.StatusActive,
	}
	if _, err := store.Create(follow); err != nil {
		t.Fatal(err)
	}
	if _, err := store.Create(follow); err != social.ErrAlreadyFollowing {
		t.Fatalf("expected ErrAlreadyFollowing, got %v", err)
	}
}

func TestFollowStoreListActiveByTarget(t *testing.T) {
	store := newTestStore(t)

	if _, err := store.Create(social.Follow{
		FollowerID:     "user-a",
		TargetHandle:   "bob@localhost",
		TargetNickname: "bob",
		Status:         social.StatusActive,
	}); err != nil {
		t.Fatal(err)
	}
	if _, err := store.Create(social.Follow{
		FollowerID:     "user-c",
		TargetHandle:   "bob@localhost",
		TargetNickname: "bob",
		Status:         social.StatusPending,
	}); err != nil {
		t.Fatal(err)
	}
	if _, err := store.Create(social.Follow{
		FollowerID:     "user-a",
		TargetHandle:   "alice@localhost",
		TargetNickname: "alice",
		Status:         social.StatusActive,
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
