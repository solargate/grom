package file

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/solargate/grom/internal/federation"
)

func TestFederationFollowersStoreIdempotentAndInboxes(t *testing.T) {
	dir := t.TempDir()
	store := NewFederationFollowersStore(dir)

	follower := federation.InboundFollower{
		ActorURI: "https://remote.test/users/bob",
		Inbox:    "https://remote.test/users/bob/inbox",
		Handle:   "bob@remote.test",
	}
	if err := store.Add("alice", follower); err != nil {
		t.Fatal(err)
	}
	if err := store.Add("alice", follower); err != nil {
		t.Fatalf("idempotent Add: %v", err)
	}
	if err := store.Add("alice", federation.InboundFollower{
		ActorURI: "https://remote.test/users/carol",
		Inbox:    "",
		Handle:   "carol@remote.test",
	}); err != nil {
		t.Fatal(err)
	}

	list, err := store.List("alice")
	if err != nil {
		t.Fatal(err)
	}
	if len(list) != 2 {
		t.Fatalf("list len = %d", len(list))
	}

	inboxes, err := store.ListInboxes("alice")
	if err != nil {
		t.Fatal(err)
	}
	if len(inboxes) != 1 || inboxes[0] != follower.Inbox {
		t.Fatalf("inboxes = %#v", inboxes)
	}

	path := filepath.Join(dir, "users", "alice", "federation", "followers.yaml")
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("followers file missing: %v", err)
	}

	if err := store.Remove("alice", "https://missing/actor"); err != nil {
		t.Fatalf("remove missing actor: %v", err)
	}
	if err := store.Remove("nobody", follower.ActorURI); err != nil {
		t.Fatalf("remove missing file: %v", err)
	}
	if err := store.Remove("alice", follower.ActorURI); err != nil {
		t.Fatal(err)
	}
	list, err = store.List("alice")
	if err != nil {
		t.Fatal(err)
	}
	if len(list) != 1 {
		t.Fatalf("after remove len = %d", len(list))
	}
}
