package workouts_test

import (
	"testing"
	"time"

	"github.com/solargate/grom/internal/workouts"
)

func TestDownsampleSpeedSamplesKeepsShortSeries(t *testing.T) {
	samples := make([]workouts.SpeedSample, 10)
	for i := range samples {
		samples[i] = workouts.SpeedSample{
			Time:      time.Date(2026, 7, 8, 10, 0, i, 0, time.UTC),
			SpeedKmh:  10 + float64(i),
			DistanceM: float64(i * 10),
		}
	}
	got := workouts.DownsampleSpeedSamples(samples, workouts.SpeedChartMaxPoints)
	if len(got) != len(samples) {
		t.Fatalf("len = %d, want %d", len(got), len(samples))
	}
	for i := range samples {
		if got[i].DistanceM != samples[i].DistanceM {
			t.Fatalf("sample %d distance = %v, want %v", i, got[i].DistanceM, samples[i].DistanceM)
		}
	}
}

func TestDownsampleSpeedSamplesReducesLongSeries(t *testing.T) {
	const n = 5000
	samples := make([]workouts.SpeedSample, n)
	for i := range samples {
		samples[i] = workouts.SpeedSample{
			Time:      time.Date(2026, 7, 8, 10, 0, 0, 0, time.UTC).Add(time.Duration(i) * time.Second),
			SpeedKmh:  10 + float64(i%20),
			DistanceM: float64(i),
		}
	}
	got := workouts.DownsampleSpeedSamples(samples, workouts.SpeedChartMaxPoints)
	if len(got) > workouts.SpeedChartMaxPoints {
		t.Fatalf("len = %d, want <= %d", len(got), workouts.SpeedChartMaxPoints)
	}
	if len(got) < workouts.SpeedChartMaxPoints/2 {
		t.Fatalf("len = %d, expected a substantial subset", len(got))
	}
	if got[0].DistanceM != samples[0].DistanceM {
		t.Fatalf("first distance = %v, want %v", got[0].DistanceM, samples[0].DistanceM)
	}
	if got[len(got)-1].DistanceM != samples[n-1].DistanceM {
		t.Fatalf("last distance = %v, want %v", got[len(got)-1].DistanceM, samples[n-1].DistanceM)
	}
}

func TestDownsampleSpeedSamplesNilOrEmpty(t *testing.T) {
	if got := workouts.DownsampleSpeedSamples(nil, workouts.SpeedChartMaxPoints); got != nil {
		t.Fatalf("nil input: got %#v", got)
	}
	if got := workouts.DownsampleSpeedSamples([]workouts.SpeedSample{}, workouts.SpeedChartMaxPoints); len(got) != 0 {
		t.Fatalf("empty input: got %#v", got)
	}
}
