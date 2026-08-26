package federation

import (
	"testing"

	"github.com/solargate/grom/internal/config"
	"github.com/solargate/grom/internal/social"
	"github.com/solargate/grom/internal/users"
)

func TestResolveSharedInboxRecipientsAcceptUsesPendingFollower(t *testing.T) {
	prev := config.Cfg
	t.Cleanup(func() { config.Cfg = prev })
	config.Cfg.Federation.Domain = "grom.test"

	alice := &users.User{ID: "alice-id", Nickname: "alice", Name: "Alice"}
	usersStore := &memUsers{byID: map[string]*users.User{alice.ID: alice}}
	follows := newMemFollows()
	created, err := follows.Create(social.Follow{
		FollowerID:       alice.ID,
		TargetHandle:     "bob@remote.test",
		TargetActorURI:   "https://remote.test/users/bob",
		Status:           social.StatusPending,
		FollowActivityID: "https://grom.test/users/alice/follows/abc",
	})
	if err != nil {
		t.Fatal(err)
	}
	_ = created

	svc := social.NewService(usersStore, follows, nil)
	proc := NewInboxProcessor(usersStore, svc, nil, nil, nil)

	activity := map[string]any{
		"type":  "Accept",
		"actor": "https://remote.test/users/bob",
		"object": map[string]any{
			"id":     "https://grom.test/users/alice/follows/abc",
			"type":   "Follow",
			"actor":  "https://grom.test/users/alice",
			"object": "https://remote.test/users/bob",
		},
	}
	got := proc.ResolveSharedInboxRecipients(activity)
	if len(got) != 1 || got[0] != "alice" {
		t.Fatalf("recipients = %#v, want [alice]", got)
	}
}

func TestResolveSharedInboxRecipientsAcceptViaToHeader(t *testing.T) {
	prev := config.Cfg
	t.Cleanup(func() { config.Cfg = prev })
	config.Cfg.Federation.Domain = "grom.test"

	proc := NewInboxProcessor(nil, nil, nil, nil, nil)
	activity := map[string]any{
		"type":  "Accept",
		"actor": "https://remote.test/users/bob",
		"to":    []any{"https://grom.test/users/carol"},
		"object": map[string]any{
			"id":    "https://grom.test/users/carol/follows/1",
			"type":  "Follow",
			"actor": "https://grom.test/users/carol",
		},
	}
	got := proc.ResolveSharedInboxRecipients(activity)
	if len(got) != 1 || got[0] != "carol" {
		t.Fatalf("recipients = %#v, want [carol]", got)
	}
}
