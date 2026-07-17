package file

import (
	"testing"
	"time"

	"github.com/solargate/grom/internal/workouts"
)

func TestWorkoutsStoreRejectsDuplicateIDWithDifferentDate(t *testing.T) {
	dir := t.TempDir()
	store := NewWorkoutsStore(dir)

	startDate1 := time.Date(2026, 7, 4, 10, 0, 0, 0, time.UTC)
	startDate2 := time.Date(2026, 7, 6, 11, 0, 0, 0, time.UTC)

	created, err := store.saveNewWorkout("athlete", &workouts.Workout{
		ID: "12345678", Name: "First", SportType: "Run", StartDate: startDate1,
	})
	if err != nil {
		t.Fatalf("saveNewWorkout: %v", err)
	}

	_, err = store.saveNewWorkout("athlete", &workouts.Workout{
		ID: created.ID, Name: "Duplicate ID", SportType: "Run", StartDate: startDate2,
	})
	if err != workouts.ErrWorkoutExists {
		t.Fatalf("expected ErrWorkoutExists, got %v", err)
	}
}

func TestWorkoutsStoreCreateFailsIfWorkoutDirExists(t *testing.T) {
	dir := t.TempDir()
	store := NewWorkoutsStore(dir)

	startDate := time.Date(2026, 7, 5, 14, 30, 0, 0, time.UTC)
	created, err := store.saveNewWorkout("solarwind", &workouts.Workout{
		ID: "87654321", Name: "First", SportType: "Run", StartDate: startDate,
	})
	if err != nil {
		t.Fatalf("saveNewWorkout: %v", err)
	}

	_, err = store.saveNewWorkout("solarwind", &workouts.Workout{
		ID: created.ID, Name: "Duplicate", SportType: "Run", StartDate: startDate,
	})
	if err != workouts.ErrWorkoutExists {
		t.Fatalf("expected ErrWorkoutExists, got %v", err)
	}
}
