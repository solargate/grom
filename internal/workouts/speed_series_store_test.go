package workouts_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/solargate/grom/internal/storage/keys"
	"github.com/solargate/grom/internal/tracks"
	"github.com/solargate/grom/internal/workouts"
)

func TestCreateWithTrackWritesSpeedSidecarAndGetLoadsIt(t *testing.T) {
	dir := t.TempDir()
	svc := newTestService(dir)

	gpxData, err := os.ReadFile(filepath.Join("..", "..", "testdata", "tracks", "1-sample.gpx"))
	if err != nil {
		t.Fatal(err)
	}

	created, err := svc.CreateWithTrack("athlete", &workouts.Workout{
		Name:            "Run",
		SportType:       "Run",
		StartDate:       time.Date(2026, 7, 6, 8, 40, 0, 0, time.UTC),
		DurationSeconds: 100,
		Distance:        1000,
	}, &workouts.TrackInput{
		Filename: "1-sample.gpx",
		Data:     gpxData,
	})
	if err != nil {
		t.Fatalf("CreateWithTrack: %v", err)
	}

	dirName := keys.WorkoutDirName(created.StartDate, created.ID)
	speedPath := filepath.Join(dir, "users", "athlete", "workouts", dirName, keys.SpeedFileYAML)
	if _, err := os.Stat(speedPath); err != nil {
		t.Fatalf("expected speed.yaml: %v", err)
	}

	listed, err := svc.List("athlete")
	if err != nil {
		t.Fatal(err)
	}
	if len(listed) != 1 {
		t.Fatalf("list len = %d", len(listed))
	}
	if listed[0].Speed != nil {
		t.Fatalf("list should not load speed, got %d samples", len(listed[0].Speed))
	}

	got, err := svc.Get("athlete", created.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(got.Speed) < 1 {
		t.Fatalf("Get speed len = %d", len(got.Speed))
	}
	raw, err := os.ReadFile(speedPath)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(raw), "speed_kmh:") || !strings.Contains(string(raw), "distance_m:") {
		t.Fatalf("speed.yaml missing fields: %s", raw)
	}
	for i, s := range got.Speed {
		if s.DistanceM < 0 {
			t.Fatalf("DistanceM[%d] = %v", i, s.DistanceM)
		}
		if i > 0 && s.DistanceM < got.Speed[i-1].DistanceM {
			t.Fatalf("DistanceM not non-decreasing: %v then %v", got.Speed[i-1].DistanceM, s.DistanceM)
		}
	}
}

func TestCreateWithoutTrackHasNoSpeedSidecar(t *testing.T) {
	dir := t.TempDir()
	svc := newTestService(dir)

	created, err := svc.Create("athlete", &workouts.Workout{
		Name:            "Gym",
		SportType:       "Workout",
		StartDate:       time.Date(2026, 7, 6, 8, 40, 0, 0, time.UTC),
		DurationSeconds: 100,
	})
	if err != nil {
		t.Fatal(err)
	}

	dirName := keys.WorkoutDirName(created.StartDate, created.ID)
	speedPath := filepath.Join(dir, "users", "athlete", "workouts", dirName, keys.SpeedFileYAML)
	if _, err := os.Stat(speedPath); !os.IsNotExist(err) {
		t.Fatalf("speed.yaml should not exist, err=%v", err)
	}

	got, err := svc.Get("athlete", created.ID)
	if err != nil {
		t.Fatal(err)
	}
	if got.Speed != nil {
		t.Fatalf("expected nil speed, got %v", got.Speed)
	}
}

func TestParseFITPopulatesSpeedSeries(t *testing.T) {
	fitData, err := os.ReadFile(filepath.Join("..", "..", "testdata", "tracks", "1-ride.fit"))
	if err != nil {
		t.Fatal(err)
	}
	parsed, err := tracks.Parse(fitData, "1-ride.fit")
	if err != nil {
		t.Fatal(err)
	}
	if len(parsed.SpeedSeries) == 0 {
		t.Fatal("expected speed series from FIT")
	}
	for _, p := range parsed.SpeedSeries {
		if p.DistanceM < 0 {
			t.Fatalf("DistanceM = %v", p.DistanceM)
		}
	}
}
