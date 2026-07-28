package workouts_test

import (
	"math"
	"strings"
	"testing"
	"time"

	"github.com/solargate/grom/internal/workouts"
)

func TestSpeedSidecarRoundTripYAML(t *testing.T) {
	t0 := time.Date(2026, 7, 14, 8, 0, 1, 0, time.UTC)
	samples := []workouts.SpeedSample{
		{Time: t0, SpeedKmh: 18.4, DistanceM: 0},
		{Time: t0.Add(time.Second), SpeedKmh: 18.7, DistanceM: 5.2},
	}
	data, err := workouts.MarshalSpeedSidecar(workouts.SpeedSidecarYAML, samples)
	if err != nil {
		t.Fatal(err)
	}
	text := string(data)
	if !strings.Contains(text, "2026-07-14T08:00:01Z") ||
		!strings.Contains(text, "speed_kmh: 18.4") ||
		!strings.Contains(text, "distance_m: 5.2") {
		t.Fatalf("yaml = %s", data)
	}
	got, err := workouts.UnmarshalSpeedSidecar(workouts.SpeedSidecarYAML, data)
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

func TestSpeedSidecarRoundTripJSON(t *testing.T) {
	t0 := time.Date(2026, 7, 14, 8, 0, 1, 0, time.UTC)
	samples := []workouts.SpeedSample{
		{Time: t0, SpeedKmh: 18.4, DistanceM: 12.5},
	}
	data, err := workouts.MarshalSpeedSidecar(workouts.SpeedSidecarJSON, samples)
	if err != nil {
		t.Fatal(err)
	}
	text := string(data)
	if !strings.Contains(text, `"2026-07-14T08:00:01Z"`) ||
		!strings.Contains(text, `"speed_kmh": 18.4`) ||
		!strings.Contains(text, `"distance_m": 12.5`) {
		t.Fatalf("json = %s", data)
	}
	got, err := workouts.UnmarshalSpeedSidecar(workouts.SpeedSidecarJSON, data)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 || !got[0].Time.Equal(t0) || got[0].DistanceM != 12.5 {
		t.Fatalf("got %+v", got)
	}
}

func TestSpeedSidecarOmitsZeroAndNaN(t *testing.T) {
	t0 := time.Date(2026, 7, 14, 8, 0, 1, 0, time.UTC)
	samples := []workouts.SpeedSample{
		{Time: t0, SpeedKmh: 0, DistanceM: 0},
		{Time: t0.Add(time.Second), SpeedKmh: math.NaN(), DistanceM: 1},
		{Time: t0.Add(2 * time.Second), SpeedKmh: 18.4, DistanceM: 5},
	}
	data, err := workouts.MarshalSpeedSidecar(workouts.SpeedSidecarYAML, samples)
	if err != nil {
		t.Fatal(err)
	}
	text := string(data)
	if strings.Contains(text, ".nan") || strings.Contains(text, "speed_kmh: 0") {
		t.Fatalf("yaml should omit zero/NaN: %s", data)
	}
	got, err := workouts.UnmarshalSpeedSidecar(workouts.SpeedSidecarYAML, data)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 || got[0].SpeedKmh != 18.4 {
		t.Fatalf("got %+v", got)
	}
}

func TestSpeedFileName(t *testing.T) {
	if workouts.SpeedFileName(workouts.SpeedSidecarYAML) != "speed.yaml" {
		t.Fatal("yaml name")
	}
	if workouts.SpeedFileName(workouts.SpeedSidecarJSON) != "speed.json" {
		t.Fatal("json name")
	}
}
