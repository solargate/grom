package workouts_test

import (
	"testing"
	"time"

	"github.com/solargate/grom/internal/workouts"
)

func TestNewestSportTypePicksNewestByStartDate(t *testing.T) {
	older := time.Date(2026, 7, 1, 10, 0, 0, 0, time.UTC)
	newer := time.Date(2026, 7, 8, 10, 0, 0, 0, time.UTC)
	items := []workouts.Workout{
		{ID: "1", SportType: "Run", StartDate: older},
		{ID: "2", SportType: "Ride", StartDate: newer},
		{ID: "3", SportType: "Walk", StartDate: older},
	}
	if got := workouts.NewestSportType(items); got != "Ride" {
		t.Fatalf("got %q, want Ride", got)
	}
}

func TestNewestSportTypeEmpty(t *testing.T) {
	if got := workouts.NewestSportType(nil); got != "" {
		t.Fatalf("got %q, want empty", got)
	}
}

func TestNewestSportTypeTieBreakByID(t *testing.T) {
	start := time.Date(2026, 7, 8, 10, 0, 0, 0, time.UTC)
	items := []workouts.Workout{
		{ID: "aaa", SportType: "Run", StartDate: start},
		{ID: "zzz", SportType: "Ride", StartDate: start},
	}
	if got := workouts.NewestSportType(items); got != "Ride" {
		t.Fatalf("got %q, want Ride (higher id wins)", got)
	}
}
