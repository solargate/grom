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
	ErrEmptyComment         = errors.New("comment text is empty")
	ErrCommentTooLong       = errors.New("comment text is too long")
	ErrCommentNotFound      = errors.New("comment not found")
	ErrCannotDeleteComment  = errors.New("cannot delete this comment")
)
