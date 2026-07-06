package workouts

import (
	"testing"
	"time"
)

func TestAllocateWorkoutIDUniquePerUser(t *testing.T) {
	dir := t.TempDir()
	store := NewStore(dir)

	startDate1 := time.Date(2026, 7, 4, 10, 0, 0, 0, time.UTC)
	startDate2 := time.Date(2026, 7, 6, 11, 0, 0, 0, time.UTC)

	first, err := store.Create("athlete", &Workout{
		Name: "First", SportType: "Run", StartDate: startDate1,
	})
	if err != nil {
		t.Fatalf("Create first: %v", err)
	}

	second, err := store.Create("athlete", &Workout{
		Name: "Second", SportType: "Ride", StartDate: startDate2,
	})
	if err != nil {
		t.Fatalf("Create second: %v", err)
	}

	if first.ID == second.ID {
		t.Fatalf("expected different ids, both are %q", first.ID)
	}

	otherUser, err := store.Create("other", &Workout{
		Name: "Other", SportType: "Run", StartDate: startDate1,
	})
	if err != nil {
		t.Fatalf("Create other user: %v", err)
	}

	// IDs may collide across users; only per-user uniqueness is required.
	_ = otherUser
}

func TestSaveWorkoutRejectsDuplicateIDWithDifferentDate(t *testing.T) {
	dir := t.TempDir()
	store := NewStore(dir)

	startDate1 := time.Date(2026, 7, 4, 10, 0, 0, 0, time.UTC)
	startDate2 := time.Date(2026, 7, 6, 11, 0, 0, 0, time.UTC)

	created, err := store.Create("athlete", &Workout{
		Name: "First", SportType: "Run", StartDate: startDate1,
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	_, err = store.saveWorkout("athlete", &Workout{
		ID:        created.ID,
		Name:      "Duplicate ID",
		SportType: "Run",
		StartDate: startDate2,
	})
	if err != ErrWorkoutExists {
		t.Fatalf("expected ErrWorkoutExists, got %v", err)
	}
}
