package federation

import "testing"

func TestDeduplicateDeliveryInboxesPrefersShared(t *testing.T) {
	followers := []InboundFollower{
		{ActorURI: "https://a.example/users/u1", Inbox: "https://a.example/users/u1/inbox", SharedInbox: "https://a.example/inbox"},
		{ActorURI: "https://a.example/users/u2", Inbox: "https://a.example/users/u2/inbox", SharedInbox: "https://a.example/inbox"},
		{ActorURI: "https://b.example/users/u3", Inbox: "https://b.example/users/u3/inbox"},
	}
	got := DeduplicateDeliveryInboxes(followers)
	if len(got) != 2 {
		t.Fatalf("got %#v", got)
	}
	if got[0] != "https://a.example/inbox" || got[1] != "https://b.example/users/u3/inbox" {
		t.Fatalf("got %#v", got)
	}
}

func TestExtractActorEndpoints(t *testing.T) {
	ep := ExtractActorEndpoints(map[string]any{
		"inbox": "https://x/users/a/inbox",
		"endpoints": map[string]any{
			"sharedInbox": "https://x/inbox",
		},
	})
	if PreferDeliveryInbox(ep) != "https://x/inbox" {
		t.Fatalf("%#v", ep)
	}
}
