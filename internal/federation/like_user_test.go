package federation

import (
	"testing"

	"github.com/solargate/grom/internal/config"
	"github.com/solargate/grom/internal/workouts"
)

func TestHandleIsLocalAndParseFederatedLikeUser(t *testing.T) {
	prev := config.Cfg
	t.Cleanup(func() { config.Cfg = prev })
	config.Cfg.Federation.Domain = "grom.test"

	if !HandleIsLocal("alice@grom.test") {
		t.Fatal("expected local handle")
	}
	if HandleIsLocal("bob@other.test") {
		t.Fatal("expected remote handle")
	}

	// Origin marks its own user as local; receiver must recompute from handle domain.
	remote := parseFederatedLikeUser(map[string]any{
		"handle":    "carol@other.test",
		"nickname":  "carol",
		"name":      "Carol",
		"is_local":  true,
		"avatarUrl": "/api/v1/users/carol/avatar",
	})
	if remote.IsLocal {
		t.Fatalf("remote user IsLocal = true: %#v", remote)
	}
	if remote.AvatarURL != "https://other.test/users/carol/avatar" {
		t.Fatalf("avatar URL = %q", remote.AvatarURL)
	}

	local := parseFederatedLikeUser(map[string]any{
		"handle":    "alice@grom.test",
		"nickname":  "alice",
		"name":      "Alice",
		"is_local":  false,
		"avatarUrl": "https://grom.test/users/alice/avatar",
	})
	if !local.IsLocal {
		t.Fatalf("local user IsLocal = false: %#v", local)
	}
	if local.AvatarURL != "" {
		t.Fatalf("local avatar should be cleared for store lookup, got %q", local.AvatarURL)
	}
}

func TestExportLikeUserAvatarURL(t *testing.T) {
	got := ExportLikeUserAvatarURL(workouts.WorkoutLikeUser{
		Handle:    "alice@grom.test",
		Nickname:  "alice",
		AvatarURL: "/api/v1/users/alice/avatar",
	})
	if got != "https://grom.test/users/alice/avatar" {
		t.Fatalf("export relative = %q", got)
	}
	got = ExportLikeUserAvatarURL(workouts.WorkoutLikeUser{
		Handle:    "bob@other.test",
		Nickname:  "bob",
		AvatarURL: "https://cdn.example/bob.png",
	})
	if got != "https://cdn.example/bob.png" {
		t.Fatalf("export absolute = %q", got)
	}
	got = ExportLikeUserAvatarURL(workouts.WorkoutLikeUser{
		Handle: "bob@other.test", Nickname: "bob",
	})
	if got != "" {
		t.Fatalf("export empty = %q", got)
	}
}

func TestParseFederatedCommentsIgnoresOriginIsLocal(t *testing.T) {
	prev := config.Cfg
	t.Cleanup(func() { config.Cfg = prev })
	config.Cfg.Federation.Domain = "viewer.test"

	comments := parseFederatedComments([]any{
		map[string]any{
			"id":   "c1",
			"text": "hi",
			"user": map[string]any{
				"handle":    "owner@origin.test",
				"nickname":  "owner",
				"name":      "Owner",
				"is_local":  true,
				"avatar_url": "",
			},
		},
	})
	if len(comments) != 1 {
		t.Fatalf("comments = %#v", comments)
	}
	u := comments[0].User
	if u.IsLocal {
		t.Fatal("expected IsLocal false for origin author on viewer instance")
	}
	if u.AvatarURL != "https://origin.test/users/owner/avatar" {
		t.Fatalf("avatar URL = %q", u.AvatarURL)
	}
}
