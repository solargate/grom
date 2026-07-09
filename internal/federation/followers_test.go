package federation

import (
	"testing"
)

func TestFollowersStoreAddListRemove(t *testing.T) {
	dir := t.TempDir()
	store := NewFollowersStore(dir)

	follower := InboundFollower{
		ActorURI: "https://remote.test/users/alice",
		Inbox:    "https://remote.test/users/alice/inbox",
		Handle:   "alice@remote.test",
	}
	if err := store.Add("bob", follower); err != nil {
		t.Fatal(err)
	}

	list, err := store.List("bob")
	if err != nil {
		t.Fatal(err)
	}
	if len(list) != 1 {
		t.Fatalf("expected 1 follower, got %d", len(list))
	}

	inboxes, err := store.ListInboxes("bob")
	if err != nil {
		t.Fatal(err)
	}
	if len(inboxes) != 1 || inboxes[0] != follower.Inbox {
		t.Fatalf("unexpected inboxes: %v", inboxes)
	}

	if err := store.Remove("bob", follower.ActorURI); err != nil {
		t.Fatal(err)
	}
	list, err = store.List("bob")
	if err != nil {
		t.Fatal(err)
	}
	if len(list) != 0 {
		t.Fatalf("expected 0 followers after remove, got %d", len(list))
	}
}

func TestInboundFollowersAdapter(t *testing.T) {
	dir := t.TempDir()
	store := NewFollowersStore(dir)
	if err := store.Add("bob", InboundFollower{
		ActorURI: "https://remote.test/users/alice",
		Inbox:    "https://remote.test/users/alice/inbox",
		Handle:   "https://remote.test/users/alice",
	}); err != nil {
		t.Fatal(err)
	}

	adapter := NewInboundFollowersAdapter(store)
	list, err := adapter.ListInboundFollowers("bob")
	if err != nil {
		t.Fatal(err)
	}
	if len(list) != 1 {
		t.Fatalf("expected 1 follower, got %d", len(list))
	}
	if list[0].Handle != "alice@remote.test" {
		t.Fatalf("expected handle alice@remote.test, got %q", list[0].Handle)
	}
	if list[0].Nickname != "alice" {
		t.Fatalf("expected nickname alice, got %q", list[0].Nickname)
	}
}
