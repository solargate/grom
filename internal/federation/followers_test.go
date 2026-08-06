package federation

import "testing"

func TestInboundFollowersAdapter(t *testing.T) {
	store := newMemFollowers()
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

func TestInboundFollowersAdapterKeepsPlainHandle(t *testing.T) {
	store := newMemFollowers()
	if err := store.Add("bob", InboundFollower{
		ActorURI: "https://remote.test/users/alice",
		Inbox:    "https://remote.test/users/alice/inbox",
		Handle:   "alice@remote.test",
	}); err != nil {
		t.Fatal(err)
	}

	adapter := NewInboundFollowersAdapter(store)
	list, err := adapter.ListInboundFollowers("bob")
	if err != nil {
		t.Fatal(err)
	}
	if len(list) != 1 || list[0].Handle != "alice@remote.test" || list[0].Nickname != "alice" {
		t.Fatalf("unexpected list: %#v", list)
	}
}
