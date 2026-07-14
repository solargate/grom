package tracks_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/solargate/grom/internal/tracks"
)

func fitSamplePath(name string) string {
	return filepath.Join("..", "..", "cmd", "grom", name)
}

func TestParseFITBikeSessionMetrics(t *testing.T) {
	data, err := os.ReadFile(fitSamplePath("1.fit"))
	if err != nil {
		t.Fatal(err)
	}
	parsed, err := tracks.Parse(data, "1.fit")
	if err != nil {
		t.Fatal(err)
	}
	if parsed.DurationSeconds == nil || *parsed.DurationSeconds != 2041 {
		t.Fatalf("duration_seconds = %v, want 2041", parsed.DurationSeconds)
	}
	if parsed.DurationTotalSeconds == nil || *parsed.DurationTotalSeconds != 3832 {
		t.Fatalf("duration_total_seconds = %v, want 3832", parsed.DurationTotalSeconds)
	}
	assertFloatClose(t, parsed.Stats.SpeedMaxKmh.Value, 32.39, 0.5)
	assertFloatClose(t, parsed.Stats.ElevationGain.Value, 77, 0)
	if parsed.Stats.HeartRateMax.Value == nil || *parsed.Stats.HeartRateMax.Value != 187 {
		t.Fatalf("heart_rate_max = %v, want 187", parsed.Stats.HeartRateMax.Value)
	}
	if parsed.Stats.Calories.Value == nil || *parsed.Stats.Calories.Value != 415 {
		t.Fatalf("calories = %v, want 415", parsed.Stats.Calories.Value)
	}
	if parsed.Stats.TempAvgKmm.Value == nil {
		t.Fatal("expected temp_avg_kmm")
	}
}

func TestParseFITWalkingSteps(t *testing.T) {
	data, err := os.ReadFile(fitSamplePath("2.fit"))
	if err != nil {
		t.Fatal(err)
	}
	parsed, err := tracks.Parse(data, "2.fit")
	if err != nil {
		t.Fatal(err)
	}
	if parsed.Stats.StepsTotal.Value == nil || *parsed.Stats.StepsTotal.Value != 2583 {
		t.Fatalf("steps_total = %v, want 2583", parsed.Stats.StepsTotal.Value)
	}
}

func TestParseFITPilatesCalories(t *testing.T) {
	data, err := os.ReadFile(fitSamplePath("4.fit"))
	if err != nil {
		t.Fatal(err)
	}
	parsed, err := tracks.Parse(data, "4.fit")
	if err != nil {
		t.Fatal(err)
	}
	if parsed.Stats.Calories.Value == nil || *parsed.Stats.Calories.Value != 125 {
		t.Fatalf("calories = %v, want 125", parsed.Stats.Calories.Value)
	}
	if parsed.Stats.ElevationGain.Value != nil {
		t.Fatalf("expected no elevation_gain, got %v", *parsed.Stats.ElevationGain.Value)
	}
}

func assertFloatClose(t *testing.T, value *float64, want, delta float64) {
	t.Helper()
	if value == nil {
		t.Fatalf("expected %v, got nil", want)
	}
	diff := *value - want
	if diff < 0 {
		diff = -diff
	}
	if diff > delta {
		t.Fatalf("value = %v, want %v ± %v", *value, want, delta)
	}
}
