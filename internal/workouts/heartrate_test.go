package workouts_test

import (
	"math"
	"strings"
	"testing"
	"time"

	"github.com/solargate/grom/internal/tracks"
	"github.com/solargate/grom/internal/workouts"
)

func TestHeartRateChartRoundTrip(t *testing.T) {
	t0 := time.Date(2026, 7, 14, 8, 0, 1, 0, time.UTC)
	d := 5.2
	samples := []workouts.HeartRateSample{
		{Time: t0, BPM: 120},
		{Time: t0.Add(time.Second), BPM: 130, DistanceM: &d},
	}
	data, err := workouts.MarshalHeartRateChart(samples)
	if err != nil {
		t.Fatal(err)
	}
	raw := string(data)
	if !strings.Contains(raw, `"heart_rate_bpm"`) {
		t.Fatalf("missing field: %s", raw)
	}
	// First sample has no distance → omitempty
	if strings.Count(raw, `"distance_m"`) != 1 {
		t.Fatalf("expected one distance_m, got: %s", raw)
	}
	got, err := workouts.UnmarshalHeartRateChart(data)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 2 || got[0].BPM != 120 || got[1].BPM != 130 {
		t.Fatalf("got %+v", got)
	}
	if got[0].DistanceM != nil {
		t.Fatalf("got[0].DistanceM = %v, want nil", got[0].DistanceM)
	}
	if got[1].DistanceM == nil || *got[1].DistanceM != 5.2 {
		t.Fatalf("got[1].DistanceM = %v", got[1].DistanceM)
	}
}

func TestBuildHeartRateChartSamplesOmitsZeroAndNaN(t *testing.T) {
	t0 := time.Date(2026, 7, 14, 8, 0, 1, 0, time.UTC)
	nan := math.NaN()
	parsed := &tracks.Data{
		HeartRateSeries: []tracks.HeartRatePoint{
			{Time: t0, BPM: 0},
			{Time: t0.Add(time.Second), BPM: nan},
			{Time: t0.Add(2 * time.Second), BPM: 142, HasDistance: true, DistanceM: 10},
		},
	}
	got := workouts.BuildHeartRateChartSamples(parsed)
	if len(got) != 1 || got[0].BPM != 142 {
		t.Fatalf("got %+v", got)
	}
	if got[0].DistanceM == nil || *got[0].DistanceM != 10 {
		t.Fatalf("distance = %v", got[0].DistanceM)
	}
}

func TestBuildHeartRateChartSamplesWithoutGPSOmitsDistance(t *testing.T) {
	t0 := time.Date(2026, 7, 14, 8, 0, 1, 0, time.UTC)
	parsed := &tracks.Data{
		HeartRateSeries: []tracks.HeartRatePoint{
			{Time: t0, BPM: 110},
			{Time: t0.Add(time.Minute), BPM: 140},
		},
	}
	got := workouts.BuildHeartRateChartSamples(parsed)
	if len(got) != 2 {
		t.Fatalf("len = %d", len(got))
	}
	for i, s := range got {
		if s.DistanceM != nil {
			t.Fatalf("sample[%d] DistanceM = %v, want nil", i, *s.DistanceM)
		}
	}
	data, err := workouts.MarshalHeartRateChart(got)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(data), "distance_m") {
		t.Fatalf("distance_m should be omitted: %s", data)
	}
}

func TestDownsampleHeartRateSamplesReducesLongSeries(t *testing.T) {
	const n = 5000
	samples := make([]workouts.HeartRateSample, n)
	t0 := time.Date(2026, 7, 14, 8, 0, 0, 0, time.UTC)
	for i := range samples {
		samples[i] = workouts.HeartRateSample{
			Time: t0.Add(time.Duration(i) * time.Second),
			BPM:  float64(100 + i%50),
		}
	}
	got := workouts.DownsampleHeartRateSamples(samples, workouts.HeartRateChartMaxPoints)
	if len(got) > workouts.HeartRateChartMaxPoints {
		t.Fatalf("len = %d, want <= %d", len(got), workouts.HeartRateChartMaxPoints)
	}
	if len(got) < workouts.HeartRateChartMaxPoints/2 {
		t.Fatalf("len = %d, unexpectedly small", len(got))
	}
}
