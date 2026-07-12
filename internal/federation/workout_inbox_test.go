package federation

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/solargate/grom/internal/tracks"
	"github.com/solargate/grom/internal/workouts"
)

func TestWorkoutInboxStoreSaveTrackAndPreview(t *testing.T) {
	dir := t.TempDir()
	store := NewWorkoutInboxStore(dir)

	gpxData, err := os.ReadFile(filepath.Join("..", "tracks", "testdata", "sample.gpx"))
	if err != nil {
		t.Fatal(err)
	}

	workout := &workouts.Workout{
		ID:              "38472901",
		Name:            "Remote run",
		SportType:       "Run",
		StartDate:       time.Date(2026, 7, 8, 10, 0, 0, 0, time.UTC),
		DurationSeconds: 4200,
		Distance:        10000,
		Track:           tracks.TrackFileGPX,
	}
	ownerHandle := "test2@192.168.1.251:8445"

	if err := store.Save("solarwind", ownerHandle, workout, gpxData, nil, nil); err != nil {
		t.Fatal(err)
	}

	trackData, trackName, err := store.TrackFile("solarwind", "test2", "38472901")
	if err != nil {
		t.Fatalf("TrackFile() error = %v", err)
	}
	if trackName != tracks.TrackFileGPX {
		t.Fatalf("track name = %q", trackName)
	}
	if len(trackData) != len(gpxData) {
		t.Fatalf("track size = %d, want %d", len(trackData), len(gpxData))
	}

	previewPath, err := store.MapPreviewPath("solarwind", "test2", "38472901")
	if err != nil {
		t.Fatalf("MapPreviewPath() error = %v", err)
	}
	if _, err := os.Stat(previewPath); err != nil {
		t.Fatalf("expected map preview file: %v", err)
	}

	items, err := store.List("solarwind")
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(items))
	}
	if !items[0].HasMapPreview {
		t.Fatal("expected has_map_preview true in feed item")
	}
	if items[0].Owner != "test2" {
		t.Fatalf("owner = %q", items[0].Owner)
	}
}
