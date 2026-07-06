package tracks_test

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/solargate/travka/internal/tracks"
)

func TestParseClientGeneratedGPX(t *testing.T) {
	data, err := os.ReadFile(filepath.Join("testdata", "client_recorded.gpx"))
	if err != nil {
		t.Fatal(err)
	}

	parsed, err := tracks.Parse(data, "track.gpx")
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}
	if !parsed.HasGPS() {
		t.Fatal("expected GPS points")
	}
	if len(parsed.Points) != 2 {
		t.Fatalf("expected 2 points, got %d", len(parsed.Points))
	}
	if parsed.StartTime == nil {
		t.Fatal("expected start time")
	}
	expectedStart := time.Date(2024, 6, 1, 10, 0, 0, 0, time.UTC)
	if !parsed.StartTime.Equal(expectedStart) {
		t.Fatalf("start time = %v, want %v", parsed.StartTime, expectedStart)
	}
	if parsed.DurationSeconds == nil || *parsed.DurationSeconds != 10 {
		t.Fatalf("duration = %v, want 10", parsed.DurationSeconds)
	}
	if parsed.DistanceMeters == nil || *parsed.DistanceMeters <= 0 {
		t.Fatalf("expected positive distance, got %v", parsed.DistanceMeters)
	}
}
