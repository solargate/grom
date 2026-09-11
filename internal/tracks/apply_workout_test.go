package tracks_test

import (
	"testing"
	"time"

	"github.com/solargate/grom/internal/tracks"
)

func TestApplyToWorkoutPreservesClientMetrics(t *testing.T) {
	start := time.Date(2026, 7, 6, 10, 0, 0, 0, time.UTC)
	trackStart := time.Date(2026, 7, 6, 11, 0, 0, 0, time.UTC)
	dur := 1000
	dist := 2000.0
	data := &tracks.Data{
		StartTime:       &trackStart,
		DurationSeconds: &dur,
		DistanceMeters:  &dist,
	}

	clientStart := start
	clientDur := 3600
	clientDist := 10000.0
	data.ApplyToWorkout(&clientStart, &clientDur, &clientDist)

	if !clientStart.Equal(start) {
		t.Fatalf("start_date = %v, want %v", clientStart, start)
	}
	if clientDur != 3600 {
		t.Fatalf("duration = %d, want 3600", clientDur)
	}
	if clientDist != 10000 {
		t.Fatalf("distance = %v, want 10000", clientDist)
	}
}

func TestApplyToWorkoutFillsEmptyMetrics(t *testing.T) {
	trackStart := time.Date(2026, 7, 6, 11, 0, 0, 0, time.UTC)
	dur := 1000
	dist := 2000.0
	data := &tracks.Data{
		StartTime:       &trackStart,
		DurationSeconds: &dur,
		DistanceMeters:  &dist,
	}

	var start time.Time
	var clientDur int
	var clientDist float64
	data.ApplyToWorkout(&start, &clientDur, &clientDist)

	if !start.Equal(trackStart) {
		t.Fatalf("start_date = %v, want track", start)
	}
	if clientDur != 1000 {
		t.Fatalf("duration = %d, want 1000", clientDur)
	}
	if clientDist != 2000 {
		t.Fatalf("distance = %v, want 2000", clientDist)
	}
}

func TestApplyDurationTotalPreservesClient(t *testing.T) {
	total := 5000
	data := &tracks.Data{DurationTotalSeconds: &total}
	client := 2000
	data.ApplyDurationTotal(&client)
	if client != 2000 {
		t.Fatalf("duration_total = %d, want 2000", client)
	}
}

func TestApplyDurationTotalFillsEmpty(t *testing.T) {
	total := 5000
	data := &tracks.Data{DurationTotalSeconds: &total}
	client := 0
	data.ApplyDurationTotal(&client)
	if client != 5000 {
		t.Fatalf("duration_total = %d, want 5000", client)
	}
}
