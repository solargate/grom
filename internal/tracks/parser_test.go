package tracks_test

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/solargate/grom/internal/tracks"
)

func TestParseGPX(t *testing.T) {
	data, err := os.ReadFile(filepath.Join("..", "..", "testdata", "sample.gpx"))
	if err != nil {
		t.Fatal(err)
	}

	parsed, err := tracks.Parse(data, "activity.gpx")
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}
	if !parsed.HasGPS() {
		t.Fatal("expected GPS points")
	}
	if len(parsed.Points) != 3 {
		t.Fatalf("expected 3 points, got %d", len(parsed.Points))
	}
	if parsed.StartTime == nil {
		t.Fatal("expected start time")
	}
	expectedStart := time.Date(2026, 7, 6, 8, 40, 0, 0, time.UTC)
	if !parsed.StartTime.Equal(expectedStart) {
		t.Fatalf("start time = %v, want %v", parsed.StartTime, expectedStart)
	}
	if parsed.DurationSeconds == nil || *parsed.DurationSeconds != 4200 {
		t.Fatalf("duration = %v, want 4200", parsed.DurationSeconds)
	}
	if parsed.DurationTotalSeconds == nil || *parsed.DurationTotalSeconds != 4200 {
		t.Fatalf("duration_total = %v, want 4200", parsed.DurationTotalSeconds)
	}
	if parsed.DistanceMeters == nil || *parsed.DistanceMeters <= 0 {
		t.Fatalf("expected positive distance, got %v", parsed.DistanceMeters)
	}
}

func TestParseRejectsInvalidExtension(t *testing.T) {
	_, err := tracks.Parse([]byte("data"), "file.txt")
	if err != tracks.ErrInvalidFormat {
		t.Fatalf("expected ErrInvalidFormat, got %v", err)
	}
}

func TestTrackFileName(t *testing.T) {
	name, err := tracks.TrackFileName("run.GPX")
	if err != nil {
		t.Fatal(err)
	}
	if name != tracks.TrackFileGPX {
		t.Fatalf("got %q", name)
	}
}

func TestSimplifyForRender(t *testing.T) {
	points := make([]tracks.LatLng, 2000)
	for i := range points {
		points[i] = tracks.LatLng{Lat: 55.0 + float64(i)*0.0001, Lng: 37.0}
	}
	simplified := tracks.SimplifyForRender(points)
	if len(simplified) > 500 {
		t.Fatalf("expected <= 500 points, got %d", len(simplified))
	}
}
