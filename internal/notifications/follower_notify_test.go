package notifications

import (
	"testing"

	"github.com/solargate/grom/internal/workouts"
)

func TestNotifyNewFollowerNilSafe(t *testing.T) {
	var n *Notifier
	n.NotifyNewFollower("alice", workouts.WorkoutLikeUser{
		Handle: "bob@localhost", Nickname: "bob", Name: "Bob", IsLocal: true,
	})

	n = &Notifier{}
	n.NotifyNewFollower("alice", workouts.WorkoutLikeUser{
		Handle: "bob@localhost", Nickname: "bob", Name: "Bob", IsLocal: true,
	})
	n.NotifyNewFollower("alice", workouts.WorkoutLikeUser{
		Handle: "alice@localhost", Nickname: "alice", Name: "Alice", IsLocal: true,
	})
}
