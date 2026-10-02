package v1

import (
	"github.com/solargate/grom/internal/workouts"
)

func (a *App) notifyWorkoutLiked(ownerNickname, workoutID string, actor workouts.WorkoutLikeUser) {
	if a == nil || a.Notifier == nil {
		return
	}
	a.Notifier.NotifyWorkoutLiked(ownerNickname, workoutID, actor)
}

func (a *App) notifyWorkoutCommented(ownerNickname, workoutID string, actor workouts.WorkoutLikeUser) {
	if a == nil || a.Notifier == nil {
		return
	}
	a.Notifier.NotifyWorkoutCommented(ownerNickname, workoutID, actor)
}

func (a *App) notifyNewFollower(targetNickname string, actor workouts.WorkoutLikeUser) {
	if a == nil || a.Notifier == nil {
		return
	}
	a.Notifier.NotifyNewFollower(targetNickname, actor)
}
