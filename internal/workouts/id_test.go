package workouts_test

import (
	"testing"
	"time"

	"github.com/solargate/grom/internal/workouts"
)

func TestAllocateWorkoutIDUniquePerUser(t *testing.T) {
	dir := t.TempDir()
	svc := newTestService(dir)

	startDate1 := time.Date(2026, 7, 4, 10, 0, 0, 0, time.UTC)
	startDate2 := time.Date(2026, 7, 6, 11, 0, 0, 0, time.UTC)

	first, err := svc.Create("athlete", &workouts.Workout{
		Name: "First", SportType: "Run", StartDate: startDate1,
	})
	if err != nil {
		t.Fatalf("Create first: %v", err)
	}

	second, err := svc.Create("athlete", &workouts.Workout{
		Name: "Second", SportType: "Ride", StartDate: startDate2,
	})
	if err != nil {
		t.Fatalf("Create second: %v", err)
	}

	if first.ID == second.ID {
		t.Fatalf("expected different ids for same user, both are %q", first.ID)
	}

	// Cross-user IDs are independently allocated; creation must succeed even
	// when start dates match another user's workout.
	otherUser, err := svc.Create("other", &workouts.Workout{
		Name: "Other", SportType: "Run", StartDate: startDate1,
	})
	if err != nil {
		t.Fatalf("Create other user: %v", err)
	}
	if otherUser.ID == "" {
		t.Fatal("expected generated id for other user")
	}
}
