package workouts_test

import (
	"math"
	"testing"
	"time"

	"github.com/solargate/grom/internal/tracks"
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

func TestBuildSpeedChartSamplesKeepsZeroOmitsNaN(t *testing.T) {
	t0 := time.Date(2026, 7, 14, 8, 0, 1, 0, time.UTC)
	parsed := &tracks.Data{
		SpeedSeries: []tracks.SpeedPoint{
			{Time: t0, Kmh: 0, DistanceM: 0},
			{Time: t0.Add(time.Second), Kmh: math.NaN(), DistanceM: 1},
			{Time: t0.Add(2 * time.Second), Kmh: 18.4, DistanceM: 5},
		},
	}
	got := workouts.BuildSpeedChartSamples(parsed)
	if len(got) != 2 {
		t.Fatalf("len = %d, want 2 (zero kept, NaN dropped)", len(got))
	}
	if got[0].SpeedKmh != 0 || got[1].SpeedKmh != 18.4 {
		t.Fatalf("got %+v", got)
	}
}

func TestBuildSpeedChartSamplesOmitsAllZeroSeries(t *testing.T) {
	t0 := time.Date(2026, 7, 14, 8, 0, 1, 0, time.UTC)
	parsed := &tracks.Data{
		SpeedSeries: []tracks.SpeedPoint{
			{Time: t0, Kmh: 0, DistanceM: 0},
			{Time: t0.Add(time.Second), Kmh: 0, DistanceM: 0},
			{Time: t0.Add(2 * time.Second), Kmh: 0, DistanceM: 0},
		},
	}
	if got := workouts.BuildSpeedChartSamples(parsed); got != nil {
		t.Fatalf("got %+v, want nil for all-zero series", got)
	}
}

func TestMeaningfulSpeedChartSamples(t *testing.T) {
	t0 := time.Date(2026, 7, 14, 8, 0, 1, 0, time.UTC)
	allZero := []workouts.SpeedSample{
		{Time: t0, SpeedKmh: 0, DistanceM: 0},
		{Time: t0.Add(time.Second), SpeedKmh: 0, DistanceM: 0},
	}
	if got := workouts.MeaningfulSpeedChartSamples(allZero); got != nil {
		t.Fatalf("all-zero: got %+v", got)
	}
	withMotion := []workouts.SpeedSample{
		{Time: t0, SpeedKmh: 0, DistanceM: 0},
		{Time: t0.Add(time.Second), SpeedKmh: 12, DistanceM: 3},
		{Time: t0.Add(2 * time.Second), SpeedKmh: 0, DistanceM: 3},
	}
	got := workouts.MeaningfulSpeedChartSamples(withMotion)
	if len(got) != 3 {
		t.Fatalf("with motion: len = %d", len(got))
	}
}
