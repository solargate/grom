package workouts_test

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/solargate/travka/internal/tracks"
	"github.com/solargate/travka/internal/workouts"
)

func TestStoreCreateWithTrack(t *testing.T) {
	dir := t.TempDir()
	store := workouts.NewStore(dir)

	gpxData, err := os.ReadFile(filepath.Join("..", "tracks", "testdata", "sample.gpx"))
	if err != nil {
		t.Fatal(err)
	}
	parsed, err := tracks.Parse(gpxData, "sample.gpx")
	if err != nil {
		t.Fatal(err)
	}

	startDate := time.Date(2026, 7, 6, 10, 0, 0, 0, time.UTC)
	created, err := store.CreateWithTrack("athlete", &workouts.Workout{
		Name:            "GPX workout",
		SportType:       "Run",
		StartDate:       startDate,
		DurationSeconds: 0,
		Distance:        0,
	}, &workouts.TrackInput{
		Filename: "sample.gpx",
		Data:     gpxData,
		Parsed:   parsed,
	})
	if err != nil {
		t.Fatalf("CreateWithTrack() error = %v", err)
	}
	if created.Track != tracks.TrackFileGPX {
		t.Fatalf("track = %q", created.Track)
	}
	if created.Device != workouts.DeviceTravka {
		t.Fatalf("device = %q, want %q", created.Device, workouts.DeviceTravka)
	}
	if created.DurationSeconds != 4200 {
		t.Fatalf("duration = %d, want 4200", created.DurationSeconds)
	}
	if created.Distance <= 0 {
		t.Fatalf("expected distance from track, got %v", created.Distance)
	}

	expectedBase := "2026-07-06T084000Z-" + created.ID
	workoutDir := filepath.Join(dir, "users", "athlete", "workouts", expectedBase)
	if _, err := os.Stat(filepath.Join(workoutDir, tracks.TrackFileGPX)); err != nil {
		t.Fatalf("expected track file: %v", err)
	}

	yamlData, err := os.ReadFile(filepath.Join(workoutDir, expectedBase+".yaml"))
	if err != nil {
		t.Fatal(err)
	}
	if !containsAll(string(yamlData), "track: track.gpx", "duration_seconds: 4200", "device: Travka") {
		t.Fatalf("unexpected yaml: %s", yamlData)
	}

	trackData, trackName, err := store.TrackFile("athlete", created.ID)
	if err != nil {
		t.Fatalf("TrackFile() error = %v", err)
	}
	if trackName != tracks.TrackFileGPX {
		t.Fatalf("track name = %q", trackName)
	}
	if len(trackData) != len(gpxData) {
		t.Fatalf("track size = %d, want %d", len(trackData), len(gpxData))
	}
}

func containsAll(s string, parts ...string) bool {
	for _, part := range parts {
		if !contains(s, part) {
			return false
		}
	}
	return true
}

func contains(s, part string) bool {
	return len(s) >= len(part) && (s == part || len(part) == 0 || indexOf(s, part) >= 0)
}

func indexOf(s, part string) int {
	for i := 0; i+len(part) <= len(s); i++ {
		if s[i:i+len(part)] == part {
			return i
		}
	}
	return -1
}
