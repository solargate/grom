package social_test

import (
	"errors"
	"testing"

	"github.com/solargate/grom/internal/config"
	"github.com/solargate/grom/internal/social"
	blobfs "github.com/solargate/grom/internal/storage/blob/fs"
	"github.com/solargate/grom/internal/storage/file"
)

func withSocialConfig(t *testing.T, enabled bool, domain string) {
	t.Helper()
	prev := config.Cfg
	t.Cleanup(func() { config.Cfg = prev })
	config.Cfg = config.Config{}
	config.Cfg.Federation.Enabled = enabled
	config.Cfg.Federation.Domain = domain
}

func newSocialService(t *testing.T, dir string) (*social.Service, *file.UsersStore, *file.SocialStore) {
	t.Helper()
	users, err := file.NewUsersStore(dir)
	if err != nil {
		t.Fatalf("NewUsersStore: %v", err)
	}
	follows, err := file.NewSocialStore(dir)
	if err != nil {
		t.Fatalf("NewSocialStore: %v", err)
	}
	blobs := blobfs.NewStore(dir)
	svc := social.NewService(users, follows, blobs)
	return svc, users, follows
}

func TestParseHandle(t *testing.T) {
	withSocialConfig(t, false, "grom.local")
	svc, _, _ := newSocialService(t, t.TempDir())

	cases := []struct {
		raw       string
		nickname  string
		domain    string
		isLocal   bool
		wantErr   bool
	}{
		{raw: "alice", nickname: "alice", domain: "grom.local", isLocal: true},
		{raw: "@alice", nickname: "alice", domain: "grom.local", isLocal: true},
		{raw: " alice@grom.local ", nickname: "alice", domain: "grom.local", isLocal: true},
		{raw: "alice@GROM.LOCAL", nickname: "alice", domain: "GROM.LOCAL", isLocal: true},
		{raw: "bob@remote.example", nickname: "bob", domain: "remote.example", isLocal: false},
		{raw: "bob@remote.example:8443", nickname: "bob", domain: "remote.example:8443", isLocal: false},
		{raw: "", wantErr: true},
		{raw: "@", wantErr: true},
		{raw: "bad/nick", wantErr: true},
		{raw: "bad\\nick", wantErr: true},
	}

	for _, tc := range cases {
		t.Run(tc.raw, func(t *testing.T) {
			got, err := svc.ParseHandle(tc.raw)
			if tc.wantErr {
				if !errors.Is(err, social.ErrInvalidHandle) {
					t.Fatalf("err = %v, want ErrInvalidHandle", err)
				}
				return
			}
			if err != nil {
				t.Fatalf("ParseHandle(%q): %v", tc.raw, err)
			}
			if got.Nickname != tc.nickname || got.Domain != tc.domain || got.IsLocal != tc.isLocal {
				t.Fatalf("got %+v, want nick=%q domain=%q local=%v", got, tc.nickname, tc.domain, tc.isLocal)
			}
		})
	}
}

func TestFollowLocalAndSelf(t *testing.T) {
	withSocialConfig(t, false, "localhost")
	dir := t.TempDir()
	svc, users, _ := newSocialService(t, dir)

	alice, err := users.Create("alice", "Alice", "alice@example.com", "password12")
	if err != nil {
		t.Fatal(err)
	}
	bob, err := users.Create("bob", "Bob", "bob@example.com", "password12")
	if err != nil {
		t.Fatal(err)
	}

	if _, err := svc.Follow(alice.ID, "alice"); !errors.Is(err, social.ErrCannotFollowSelf) {
		t.Fatalf("self follow err = %v", err)
	}

	follow, err := svc.Follow(alice.ID, "bob")
	if err != nil {
		t.Fatalf("Follow: %v", err)
	}
	if follow.Status != social.StatusActive || !follow.TargetIsLocal {
		t.Fatalf("unexpected follow: %+v", follow)
	}
	if follow.TargetNickname != bob.Nickname {
		t.Fatalf("target = %q", follow.TargetNickname)
	}

	again, err := svc.Follow(alice.ID, "bob@localhost")
	if err != nil {
		t.Fatalf("idempotent Follow: %v", err)
	}
	if again.ID != follow.ID {
		t.Fatalf("expected same follow id, got %q vs %q", again.ID, follow.ID)
	}

	if _, err := svc.Follow(alice.ID, "missing"); !errors.Is(err, social.ErrUserNotFound) {
		t.Fatalf("missing target err = %v", err)
	}
}

type recordingDelivery struct {
	followCalls int
	undoCalls   int
	failFollow  bool
	failUndo    bool
}

func (d *recordingDelivery) DeliverFollow(f *social.Follow) error {
	d.followCalls++
	if d.failFollow {
		return errors.New("delivery failed")
	}
	return nil
}

func (d *recordingDelivery) DeliverUndo(f *social.Follow) error {
	d.undoCalls++
	if d.failUndo {
		return errors.New("undo failed")
	}
	return nil
}

func TestFollowRemoteRollbackOnDeliveryFailure(t *testing.T) {
	withSocialConfig(t, true, "localhost")
	dir := t.TempDir()
	svc, users, follows := newSocialService(t, dir)

	alice, err := users.Create("alice", "Alice", "alice@example.com", "password12")
	if err != nil {
		t.Fatal(err)
	}

	delivery := &recordingDelivery{failFollow: true}
	svc.SetDelivery(delivery)

	_, err = svc.Follow(alice.ID, "remote@other.example")
	if err == nil {
		t.Fatal("expected delivery error")
	}
	listed, err := follows.ListByFollower(alice.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(listed) != 0 {
		t.Fatalf("expected rollback, still have %d follows", len(listed))
	}
	if delivery.followCalls != 1 {
		t.Fatalf("followCalls = %d", delivery.followCalls)
	}
}

func TestFollowRemoteDisabled(t *testing.T) {
	withSocialConfig(t, false, "localhost")
	dir := t.TempDir()
	svc, users, _ := newSocialService(t, dir)
	alice, err := users.Create("alice", "Alice", "alice@example.com", "password12")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := svc.Follow(alice.ID, "bob@remote.example"); !errors.Is(err, social.ErrRemoteNotReady) {
		t.Fatalf("err = %v, want ErrRemoteNotReady", err)
	}
}

func TestUnfollowOwnershipAndLocal(t *testing.T) {
	withSocialConfig(t, false, "localhost")
	dir := t.TempDir()
	svc, users, _ := newSocialService(t, dir)

	alice, err := users.Create("alice", "Alice", "alice@example.com", "password12")
	if err != nil {
		t.Fatal(err)
	}
	_, err = users.Create("bob", "Bob", "bob@example.com", "password12")
	if err != nil {
		t.Fatal(err)
	}
	charlie, err := users.Create("charlie", "Charlie", "charlie@example.com", "password12")
	if err != nil {
		t.Fatal(err)
	}

	follow, err := svc.Follow(alice.ID, "bob")
	if err != nil {
		t.Fatal(err)
	}
	if err := svc.Unfollow(charlie.ID, follow.ID); !errors.Is(err, social.ErrFollowNotFound) {
		t.Fatalf("foreign unfollow err = %v", err)
	}
	if err := svc.Unfollow(alice.ID, follow.ID); err != nil {
		t.Fatalf("Unfollow: %v", err)
	}
	listed, err := svc.ListFollowing(alice.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(listed) != 0 {
		t.Fatalf("expected empty following, got %d", len(listed))
	}
}

func TestListFollowersMergesLocal(t *testing.T) {
	withSocialConfig(t, false, "localhost")
	dir := t.TempDir()
	svc, users, _ := newSocialService(t, dir)

	alice, err := users.Create("alice", "Alice", "alice@example.com", "password12")
	if err != nil {
		t.Fatal(err)
	}
	bob, err := users.Create("bob", "Bob", "bob@example.com", "password12")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := svc.Follow(bob.ID, "alice"); err != nil {
		t.Fatal(err)
	}

	followers, err := svc.ListFollowers(alice.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(followers) != 1 || followers[0].FollowerNickname != "bob" {
		t.Fatalf("followers = %+v", followers)
	}
}

func TestHandleIncomingAcceptIsNoop(t *testing.T) {
	withSocialConfig(t, true, "localhost")
	svc, _, _ := newSocialService(t, t.TempDir())
	if err := svc.HandleIncomingAccept("any-id"); err != nil {
		t.Fatalf("expected nil stub, got %v", err)
	}
}

func TestActivateFollowByActivityID(t *testing.T) {
	withSocialConfig(t, true, "localhost")
	dir := t.TempDir()
	svc, users, follows := newSocialService(t, dir)
	svc.SetDelivery(&recordingDelivery{})

	alice, err := users.Create("alice", "Alice", "alice@example.com", "password12")
	if err != nil {
		t.Fatal(err)
	}
	follow, err := svc.Follow(alice.ID, "remote@other.example")
	if err != nil {
		t.Fatal(err)
	}
	if follow.Status != social.StatusPending {
		t.Fatalf("status = %q", follow.Status)
	}
	if err := svc.ActivateFollowByActivityID(follow.FollowActivityID); err != nil {
		t.Fatal(err)
	}
	updated, err := follows.FindByID(follow.ID)
	if err != nil {
		t.Fatal(err)
	}
	if updated.Status != social.StatusActive {
		t.Fatalf("status = %q, want active", updated.Status)
	}
}

func TestSearchLocalRemoteWhenFederationDisabled(t *testing.T) {
	withSocialConfig(t, false, "localhost")
	svc, _, _ := newSocialService(t, t.TempDir())

	_, err := svc.SearchLocal("bob@remote.example", "")
	if !errors.Is(err, social.ErrRemoteNotReady) {
		t.Fatalf("err = %v, want ErrRemoteNotReady", err)
	}
}

func TestSearchLocalByNickname(t *testing.T) {
	withSocialConfig(t, false, "localhost")
	dir := t.TempDir()
	svc, users, _ := newSocialService(t, dir)

	alice, err := users.Create("alice", "Alice", "alice@example.com", "password12")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := users.Create("bob", "Bob", "bob@example.com", "password12"); err != nil {
		t.Fatal(err)
	}

	results, err := svc.SearchLocal("bob@localhost", alice.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 1 || results[0].Nickname != "bob" {
		t.Fatalf("results = %#v", results)
	}

	self, err := svc.SearchLocal("alice@localhost", alice.ID)
	if err != nil {
		t.Fatal(err)
	}
	if self != nil {
		t.Fatalf("expected nil when searching self, got %#v", self)
	}
}

func TestSearchLocalPrefix(t *testing.T) {
	withSocialConfig(t, false, "localhost")
	dir := t.TempDir()
	svc, users, _ := newSocialService(t, dir)

	if _, err := users.Create("alice", "Alice", "alice@example.com", "password12"); err != nil {
		t.Fatal(err)
	}
	if _, err := users.Create("bob", "Bob", "bob@example.com", "password12"); err != nil {
		t.Fatal(err)
	}

	results, err := svc.SearchLocal("bo", "")
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 1 || results[0].Nickname != "bob" {
		t.Fatalf("results = %#v", results)
	}
}
