package notifications

import "testing"

func TestActorDisplayName(t *testing.T) {
	if got := ActorDisplayName("Alice", "alice", "alice@x"); got != "Alice" {
		t.Fatalf("name: %q", got)
	}
	if got := ActorDisplayName("  ", "alice", "alice@x"); got != "alice" {
		t.Fatalf("nickname: %q", got)
	}
	if got := ActorDisplayName("", "", "alice@x"); got != "alice@x" {
		t.Fatalf("handle: %q", got)
	}
}

func TestSlots(t *testing.T) {
	if got := SlotLiked("bob", "w1"); got != "liked:bob:w1" {
		t.Fatalf("liked slot: %q", got)
	}
	if got := SlotCommented("bob", "w1"); got != "commented:bob:w1" {
		t.Fatalf("commented slot: %q", got)
	}
	if got := SlotFollowed("evt-1"); got != "followed:evt-1" {
		t.Fatalf("followed slot: %q", got)
	}
}
