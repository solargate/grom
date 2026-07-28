package workouts_test

import (
	"strings"
	"testing"
	"time"

	"github.com/solargate/grom/internal/workouts"
)

func TestSpeedSidecarRoundTripYAML(t *testing.T) {
	t0 := time.Date(2026, 7, 14, 8, 0, 1, 0, time.UTC)
	samples := []workouts.SpeedSample{
		{Time: t0, SpeedKmh: 18.4},
		{Time: t0.Add(time.Second), SpeedKmh: 18.7},
	}
	data, err := workouts.MarshalSpeedSidecar(workouts.SpeedSidecarYAML, samples)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), "2026-07-14T08:00:01Z") || !strings.Contains(string(data), "18.4") {
		t.Fatalf("yaml = %s", data)
	}
	got, err := workouts.UnmarshalSpeedSidecar(workouts.SpeedSidecarYAML, data)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 2 || got[0].SpeedKmh != 18.4 || got[1].SpeedKmh != 18.7 {
		t.Fatalf("got %+v", got)
	}
}

func TestSpeedSidecarRoundTripJSON(t *testing.T) {
	t0 := time.Date(2026, 7, 14, 8, 0, 1, 0, time.UTC)
	samples := []workouts.SpeedSample{
		{Time: t0, SpeedKmh: 18.4},
	}
	data, err := workouts.MarshalSpeedSidecar(workouts.SpeedSidecarJSON, samples)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), `"2026-07-14T08:00:01Z": 18.4`) {
		t.Fatalf("json = %s", data)
	}
	got, err := workouts.UnmarshalSpeedSidecar(workouts.SpeedSidecarJSON, data)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 || !got[0].Time.Equal(t0) {
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
