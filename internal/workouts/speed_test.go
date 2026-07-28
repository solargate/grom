package workouts_test

import (
	"math"
	"testing"
	"time"

	"github.com/solargate/grom/internal/workouts"
)

func TestSpeedChartRoundTrip(t *testing.T) {
	t0 := time.Date(2026, 7, 14, 8, 0, 1, 0, time.UTC)
	samples := []workouts.SpeedSample{
		{Time: t0, SpeedKmh: 18.4, DistanceM: 0},
		{Time: t0.Add(time.Second), SpeedKmh: 18.7, DistanceM: 5.2},
	}
	data, err := workouts.MarshalSpeedChart(samples)
	if err != nil {
		t.Fatal(err)
	}
	got, err := workouts.UnmarshalSpeedChart(data)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 2 || got[0].SpeedKmh != 18.4 || got[1].SpeedKmh != 18.7 {
		t.Fatalf("got %+v", got)
	}
	if got[0].DistanceM != 0 || got[1].DistanceM != 5.2 {
		t.Fatalf("distances %+v", got)
	}
}

func TestBuildSpeedChartSamplesOmitsZeroAndNaN(t *testing.T) {
	t0 := time.Date(2026, 7, 14, 8, 0, 1, 0, time.UTC)
	samples := []workouts.SpeedSample{
		{Time: t0, SpeedKmh: 0, DistanceM: 0},
		{Time: t0.Add(time.Second), SpeedKmh: math.NaN(), DistanceM: 1},
		{Time: t0.Add(2 * time.Second), SpeedKmh: 18.4, DistanceM: 5},
	}
	filtered := make([]workouts.SpeedSample, 0, len(samples))
	for _, s := range samples {
		if s.SpeedKmh > 0 && !math.IsNaN(s.SpeedKmh) {
			filtered = append(filtered, s)
		}
	}
	got := workouts.DownsampleSpeedSamples(filtered, workouts.SpeedChartMaxPoints)
	if len(got) != 1 || got[0].SpeedKmh != 18.4 {
		t.Fatalf("got %+v", got)
	}
}
