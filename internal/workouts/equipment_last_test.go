package workouts_test

import (
	"testing"
	"time"

	"github.com/solargate/grom/internal/workouts"
)

func TestLastEquipmentIDsForSportPicksNewestSameType(t *testing.T) {
	older := time.Date(2026, 7, 1, 10, 0, 0, 0, time.UTC)
	newer := time.Date(2026, 7, 8, 10, 0, 0, 0, time.UTC)
	items := []workouts.Workout{
		{
			ID: "1", SportType: "Run", StartDate: older,
			Equipment: []workouts.WorkoutEquipment{{ID: "old-shoes"}},
		},
		{
			ID: "2", SportType: "Ride", StartDate: newer,
			Equipment: []workouts.WorkoutEquipment{{ID: "bike"}},
		},
		{
			ID: "3", SportType: "Run", StartDate: newer,
			Equipment: []workouts.WorkoutEquipment{{ID: "new-shoes"}, {ID: "watch"}},
		},
	}

	got := workouts.LastEquipmentIDsForSport(items, "Run")
	if len(got) != 2 || got[0] != "new-shoes" || got[1] != "watch" {
		t.Fatalf("got %#v, want [new-shoes watch]", got)
	}
}

func TestLastEquipmentIDsForSportEmptyWhenMissingOrNoGear(t *testing.T) {
	start := time.Date(2026, 7, 8, 10, 0, 0, 0, time.UTC)
	items := []workouts.Workout{
		{ID: "1", SportType: "Run", StartDate: start},
	}
	if got := workouts.LastEquipmentIDsForSport(items, "Ride"); got != nil {
		t.Fatalf("expected nil for missing sport, got %#v", got)
	}
	if got := workouts.LastEquipmentIDsForSport(items, "Run"); got != nil {
		t.Fatalf("expected nil when previous has no gear, got %#v", got)
	}
	if got := workouts.LastEquipmentIDsForSport(nil, "Run"); got != nil {
		t.Fatalf("expected nil for empty list, got %#v", got)
	}
}

func TestLastEquipmentIDsForSportTieBreakByID(t *testing.T) {
	start := time.Date(2026, 7, 8, 10, 0, 0, 0, time.UTC)
	items := []workouts.Workout{
		{
			ID: "aaa", SportType: "Run", StartDate: start,
			Equipment: []workouts.WorkoutEquipment{{ID: "eq-a"}},
		},
		{
			ID: "zzz", SportType: "Run", StartDate: start,
			Equipment: []workouts.WorkoutEquipment{{ID: "eq-z"}},
		},
	}
	got := workouts.LastEquipmentIDsForSport(items, "Run")
	if len(got) != 1 || got[0] != "eq-z" {
		t.Fatalf("got %#v, want [eq-z] (higher id wins)", got)
	}
}
