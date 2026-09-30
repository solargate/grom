package workouts_test

import (
	"testing"
	"time"

	"github.com/solargate/grom/internal/workouts"
)

func TestDownsampleCadenceSamplesKeepsShortSeries(t *testing.T) {
	samples := make([]workouts.CadenceSample, 10)
	for i := range samples {
		di := float64(i * 10)
		samples[i] = workouts.CadenceSample{
			Time:      time.Date(2026, 7, 8, 10, 0, i, 0, time.UTC),
			Cadence:   70 + float64(i),
			DistanceM: &di,
		}
	}

	got := workouts.DownsampleCadenceSamples(samples, workouts.CadenceChartMaxPoints)
	if len(got) != len(samples) {
		t.Fatalf("len = %d, want %d", len(got), len(samples))
	}
}

func TestDownsampleCadenceSamplesReducesLongSeries(t *testing.T) {
	const n = 5000
	samples := make([]workouts.CadenceSample, n)
	t0 := time.Date(2026, 7, 8, 10, 0, 0, 0, time.UTC)
	for i := range samples {
		samples[i] = workouts.CadenceSample{
			Time:    t0.Add(time.Duration(i) * time.Second),
			Cadence: 70 + float64(i%20),
		}
	}
	got := workouts.DownsampleCadenceSamples(samples, workouts.CadenceChartMaxPoints)
	if len(got) > workouts.CadenceChartMaxPoints {
		t.Fatalf("len = %d, want <= %d", len(got), workouts.CadenceChartMaxPoints)
	}
	if len(got) < workouts.CadenceChartMaxPoints/2 {
		t.Fatalf("len = %d, expected a substantial subset", len(got))
	}
	if !got[0].Time.Equal(samples[0].Time) {
		t.Fatalf("first time = %v, want %v", got[0].Time, samples[0].Time)
	}
	if !got[len(got)-1].Time.Equal(samples[n-1].Time) {
		t.Fatalf("last time = %v, want %v", got[len(got)-1].Time, samples[n-1].Time)
	}
}

func TestDownsampleCadenceSamplesNilOrEmpty(t *testing.T) {
	if got := workouts.DownsampleCadenceSamples(nil, workouts.CadenceChartMaxPoints); got != nil {
		t.Fatalf("nil input: got %#v", got)
	}
	if got := workouts.DownsampleCadenceSamples([]workouts.CadenceSample{}, workouts.CadenceChartMaxPoints); len(got) != 0 {
		t.Fatalf("empty input: got %#v", got)
	}
}
