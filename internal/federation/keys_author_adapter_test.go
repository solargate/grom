package federation

import (
	"strings"
	"testing"

	"github.com/solargate/grom/internal/config"
	blobfs "github.com/solargate/grom/internal/storage/blob/fs"
	"github.com/solargate/grom/internal/social"
)

func TestLoadOrCreateActorKeyIdempotent(t *testing.T) {
	dir := t.TempDir()
	blobs := blobfs.NewStore(dir)

	pub1, keyID1, err := LoadOrCreateActorKey(blobs, "alice")
	if err != nil {
		t.Fatal(err)
	}
	if pub1 == "" || keyID1 == "" {
		t.Fatalf("empty key: pub=%q id=%q", pub1, keyID1)
	}
	if !strings.Contains(pub1, "BEGIN PUBLIC KEY") {
		t.Fatalf("expected PEM public key, got %q", pub1)
	}

	pub2, keyID2, err := LoadOrCreateActorKey(blobs, "alice")
	if err != nil {
		t.Fatal(err)
	}
	if pub1 != pub2 || keyID1 != keyID2 {
		t.Fatalf("second call changed key")
	}
	if ActorKeyID("alice") != keyID1 {
		t.Fatalf("ActorKeyID = %q want %q", ActorKeyID("alice"), keyID1)
	}
}

func TestUpdateAuthorMetaMergesActorFields(t *testing.T) {
	prev := config.Cfg
	t.Cleanup(func() { config.Cfg = prev })
	config.Cfg.Federation.Domain = "grom.test"

	meta := AuthorMeta{}
	UpdateAuthorMeta(&meta, "bob@remote.test", "bob", map[string]any{
		"name": "Bob Remote",
		"icon": map[string]any{"url": "https://cdn.example/bob.png"},
	}, nil, false, nil, "alice", "bob_remote.test")

	if meta.Handle != "bob@remote.test" || meta.Nickname != "bob" {
		t.Fatalf("identity: %#v", meta)
	}
	if meta.Name != "Bob Remote" {
		t.Fatalf("name = %q", meta.Name)
	}
	if meta.RemoteAvatarURL != "https://cdn.example/bob.png" {
		t.Fatalf("remote avatar = %q", meta.RemoteAvatarURL)
	}

	UpdateAuthorMeta(&meta, "bob@remote.test", "bob", map[string]any{
		"name": "Bob Updated",
	}, nil, false, nil, "alice", "bob_remote.test")
	if meta.Name != "Bob Updated" {
		t.Fatalf("name after update = %q", meta.Name)
	}

	ownerDir := t.TempDir()
	if err := writeAuthorMeta(ownerDir, meta); err != nil {
		t.Fatal(err)
	}
	got, err := readAuthorMeta(ownerDir)
	if err != nil {
		t.Fatal(err)
	}
	if got.Name != "Bob Updated" || got.Handle != "bob@remote.test" {
		t.Fatalf("round-trip: %#v", got)
	}
}

func TestInboundFollowersAdapterNormalizesActorURLs(t *testing.T) {
	store := stubFollowersRepo{items: []InboundFollower{
		{Handle: "carol@other.test"},
		{Handle: "https://remote.test/users/bob"},
		{Handle: ""},
	}}
	adapter := NewInboundFollowersAdapter(store)
	got, err := adapter.ListInboundFollowers("alice")
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 2 {
		t.Fatalf("got %#v", got)
	}
	if got[0] != (social.InboundFollowerInfo{Handle: "carol@other.test", Nickname: "carol"}) {
		t.Fatalf("first = %#v", got[0])
	}
	if got[1].Handle != "bob@remote.test" || got[1].Nickname != "bob" {
		t.Fatalf("actor url normalized = %#v", got[1])
	}
}

type stubFollowersRepo struct {
	items []InboundFollower
}

func (s stubFollowersRepo) List(string) ([]InboundFollower, error) { return s.items, nil }
func (s stubFollowersRepo) Add(string, InboundFollower) error      { return nil }
func (s stubFollowersRepo) Remove(string, string) error            { return nil }
func (s stubFollowersRepo) ListInboxes(string) ([]string, error) {
	return nil, nil
}
