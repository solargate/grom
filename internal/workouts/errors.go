package workouts

import "errors"

const WorkoutIDLength = 8

var (
	ErrInvalidSportType     = errors.New("invalid sport type")
	ErrInvalidWorkout       = errors.New("invalid workout")
	ErrWorkoutExists        = errors.New("workout already exists")
	ErrWorkoutNotFound      = errors.New("workout not found")
	ErrExternalIDExists     = errors.New("workout with this external_id already exists")
	ErrCannotLikeOwnWorkout = errors.New("cannot like your own workout")
)
