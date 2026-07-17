package workouts

import (
	"fmt"
	"strings"
)

func validateWorkout(workout *Workout) error {
	if workout == nil {
		return ErrInvalidWorkout
	}
	if strings.TrimSpace(workout.Name) == "" {
		return fmt.Errorf("%w: name is required", ErrInvalidWorkout)
	}
	if !IsValidSportType(workout.SportType) {
		return ErrInvalidSportType
	}
	if workout.DurationSeconds < 0 {
		return fmt.Errorf("%w: duration must be non-negative", ErrInvalidWorkout)
	}
	if workout.Distance < 0 {
		return fmt.Errorf("%w: distance must be non-negative", ErrInvalidWorkout)
	}
	if workout.StartDate.IsZero() {
		return fmt.Errorf("%w: start_date is required", ErrInvalidWorkout)
	}
	return nil
}

func TrimWorkoutName(name string) string {
	return strings.TrimSpace(name)
}

func TrimWorkoutDescription(description string) string {
	return strings.TrimSpace(description)
}

func trimWorkoutName(name string) string {
	return TrimWorkoutName(name)
}

func trimWorkoutDescription(description string) string {
	return TrimWorkoutDescription(description)
}
