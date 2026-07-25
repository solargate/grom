package file

import (
	"errors"
	"testing"
	"time"

	"github.com/solargate/grom/internal/social"
)

func TestSocialStoreRejectedAllowsRefollowAndReload(t *testing.T) {
	dir := t.TempDir()
	store, err := NewSocialStore(dir)
	if err != nil {
		t.Fatalf("NewSocialStore: %v", err)
	}

	created, err := store.Create(social.Follow{
		FollowerID:     "alice",
		TargetHandle:   "bob",
		TargetNickname: "bob",
		Status:         social.StatusRejected,
		CreatedAt:      time.Now().UTC(),
	})
	if err != nil {
		t.Fatal(err)
	}

	if _, err := store.FindExisting("alice", "bob"); !errors.Is(err, social.ErrFollowNotFound) {
		t.Fatalf("rejected should not count as existing: %v", err)
	}

	again, err := store.Create(social.Follow{
		FollowerID:     "alice",
		TargetHandle:   "bob",
		TargetNickname: "bob",
		Status:         social.StatusPending,
		CreatedAt:      time.Now().UTC(),
	})
	if err != nil {
		t.Fatalf("re-follow after reject: %v", err)
	}

	updated, err := store.UpdateStatus(again.ID, social.StatusActive)
	if err != nil {
		t.Fatal(err)
	}
	if updated.Status != social.StatusActive {
		t.Fatalf("status = %q", updated.Status)
	}

	pending, err := store.Create(social.Follow{
		FollowerID:     "alice",
		TargetHandle:   "carol",
		TargetNickname: "carol",
		Status:         social.StatusPending,
		CreatedAt:      time.Now().UTC(),
	})
	if err != nil {
		t.Fatal(err)
	}
	all, err := store.ListByFollower("alice")
	if err != nil {
		t.Fatal(err)
	}
	if len(all) < 2 {
		t.Fatalf("ListByFollower = %#v", all)
	}
	active, err := store.ListActiveFollowing("alice")
	if err != nil {
		t.Fatal(err)
	}
	for _, f := range active {
		if f.ID == pending.ID {
			t.Fatal("pending follow must not appear in ListActiveFollowing")
		}
	}

	reloaded, err := NewSocialStore(dir)
	if err != nil {
		t.Fatal(err)
	}
	got, err := reloaded.FindByID(created.ID)
	if err != nil {
		t.Fatal(err)
	}
	if got.Status != social.StatusRejected {
		t.Fatalf("reloaded rejected status = %q", got.Status)
	}

	if _, err := store.FindByID("missing"); !errors.Is(err, social.ErrFollowNotFound) {
		t.Fatalf("missing find: %v", err)
	}
	if _, err := store.UpdateStatus("missing", social.StatusActive); !errors.Is(err, social.ErrFollowNotFound) {
		t.Fatalf("missing update: %v", err)
	}
}
