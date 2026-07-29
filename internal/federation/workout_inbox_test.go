package federation

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/solargate/grom/internal/storage/keys"
	"github.com/solargate/grom/internal/tracks"
	"github.com/solargate/grom/internal/workouts"
)

func TestWorkoutInboxStoreSaveTrackAndPreview(t *testing.T) {
	dir := t.TempDir()
	store := newTestInboxStore(dir)

	gpxData, err := os.ReadFile(filepath.Join("..", "..", "testdata", "tracks", "1-sample.gpx"))
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

	trackData, trackName, _, err := store.TrackFile("solarwind", "test2", "38472901")
	if err != nil {
		t.Fatalf("TrackFile() error = %v", err)
	}
	if trackName != tracks.TrackFileGPX {
		t.Fatalf("track name = %q", trackName)
	}
	if len(trackData) != len(gpxData) {
		t.Fatalf("track size = %d, want %d", len(trackData), len(gpxData))
	}

	previewData, err := store.MapPreview("solarwind", "test2", "38472901")
	if err != nil {
		t.Fatalf("MapPreview() error = %v", err)
	}
	if len(previewData) == 0 {
		t.Fatal("expected map preview bytes")
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

	got, err := store.Get("solarwind", "test2", "38472901")
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}
	if got.ID != "38472901" || got.Owner != "test2" || got.Author.IsLocal {
		t.Fatalf("unexpected Get item: %#v", got)
	}

	_, samples, err := store.GetSpeedChart("solarwind", "test2", "38472901")
	if err != nil {
		t.Fatalf("GetSpeedChart() error = %v", err)
	}
	if len(samples) < 1 {
		t.Fatalf("expected speed chart samples, got %d", len(samples))
	}

	ownerKey := OwnerKeyFromHandle(ownerHandle)
	speedKey := keys.FederatedInboxSpeed("solarwind", ownerKey, "38472901", keys.SpeedChartFileJSON)
	speedPath := filepath.Join(dir, speedKey)
	if _, err := os.Stat(speedPath); err != nil {
		t.Fatalf("expected federated speed-chart.json: %v", err)
	}
}

func TestWorkoutInboxStoreSaveFITWritesHeartRateChart(t *testing.T) {
	dir := t.TempDir()
	store := newTestInboxStore(dir)

	fitData, err := os.ReadFile(filepath.Join("..", "..", "testdata", "tracks", "1-ride.fit"))
	if err != nil {
		t.Fatal(err)
	}

	workout := &workouts.Workout{
		ID:              "87654321",
		Name:            "Remote ride",
		SportType:       "Ride",
		StartDate:       time.Date(2026, 7, 8, 10, 0, 0, 0, time.UTC),
		DurationSeconds: 3600,
		Distance:        20000,
		Track:           tracks.TrackFileFIT,
	}
	ownerHandle := "rider@remote.test"

	if err := store.Save("viewer", ownerHandle, workout, fitData, nil, nil); err != nil {
		t.Fatal(err)
	}

	_, samples, err := store.GetHeartRateChart("viewer", "rider", "87654321")
	if err != nil {
		t.Fatalf("GetHeartRateChart() error = %v", err)
	}
	if len(samples) < 1 {
		t.Fatalf("expected heart rate chart samples, got %d", len(samples))
	}

	ownerKey := OwnerKeyFromHandle(ownerHandle)
	hrKey := keys.FederatedInboxSpeed("viewer", ownerKey, "87654321", keys.HeartRateChartFileJSON)
	if _, err := os.Stat(filepath.Join(dir, hrKey)); err != nil {
		t.Fatalf("expected federated heartrate-chart.json: %v", err)
	}
}

func TestWorkoutInboxStoreDelete(t *testing.T) {
	dir := t.TempDir()
	store := newTestInboxStore(dir)

	gpxData, err := os.ReadFile(filepath.Join("..", "..", "testdata", "tracks", "1-sample.gpx"))
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

	if err := store.Delete("solarwind", ownerHandle, "38472901"); err != nil {
		t.Fatalf("Delete() error = %v", err)
	}

	items, err := store.List("solarwind")
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 0 {
		t.Fatalf("expected 0 items after delete, got %d", len(items))
	}

	if err := store.Delete("solarwind", ownerHandle, "38472901"); err != nil {
		t.Fatalf("second Delete() error = %v", err)
	}
}
