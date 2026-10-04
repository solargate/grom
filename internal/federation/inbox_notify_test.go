package federation

import (
	"strings"
	"testing"

	"github.com/solargate/grom/internal/config"
	"github.com/solargate/grom/internal/workouts"
)

func TestInboxProcessorFollowNotifyOnlyWhenCreated(t *testing.T) {
	prev := config.Cfg
	t.Cleanup(func() { config.Cfg = prev })
	config.Cfg.Federation.Domain = "grom.test"
	config.Cfg.Federation.AutoAcceptFollows = false
	config.Cfg.Federation.Enabled = true

	followers := newMemFollowers()
	var notifyCalls []workouts.WorkoutLikeUser
	processor := NewInboxProcessor(nil, nil, nil, newTestInboxStore(t.TempDir()), followers)
	processor.SetFollowNotify(func(targetNickname string, actor workouts.WorkoutLikeUser) {
		if targetNickname != "alice" {
			t.Fatalf("target = %q", targetNickname)
		}
		notifyCalls = append(notifyCalls, actor)
	})

	followBody := `{"type":"Follow","id":"https://remote.test/follows/1","actor":"https://remote.test/users/bob","object":"https://grom.test/users/alice"}`
	if err := processor.Handle("alice", strings.NewReader(followBody)); err != nil {
		t.Fatalf("Follow: %v", err)
	}
	if len(notifyCalls) != 1 || notifyCalls[0].Handle != "bob@remote.test" || notifyCalls[0].IsLocal {
		t.Fatalf("notify = %#v", notifyCalls)
	}

	if err := processor.Handle("alice", strings.NewReader(followBody)); err != nil {
		t.Fatalf("idempotent Follow: %v", err)
	}
	if len(notifyCalls) != 1 {
		t.Fatalf("idempotent follow must not notify again: %#v", notifyCalls)
	}
}
