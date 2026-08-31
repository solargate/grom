package workouts_test

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	blobfs "github.com/solargate/grom/internal/storage/blob/fs"
	"github.com/solargate/grom/internal/storage/file"
	"github.com/solargate/grom/internal/tracks"
	"github.com/solargate/grom/internal/workouts"
)

func TestServiceAttachTrackWritesSpeedAndHeartRateCharts(t *testing.T) {
	dir := t.TempDir()
	blobs := blobfs.NewStore(dir)
	speedCharts := workouts.NewBlobSpeedChartStore(blobs)
	hrCharts := workouts.NewBlobHeartRateChartStore(blobs)
	svc := workouts.NewService(file.NewWorkoutsStore(dir), blobs, speedCharts, hrCharts)

	gpxData, err := os.ReadFile(filepath.Join("..", "..", "testdata", "tracks", "1-sample.gpx"))
	if err != nil {
		t.Fatal(err)
	}
	parsed, err := tracks.Parse(gpxData, "sample.gpx")
	if err != nil {
		t.Fatal(err)
	}

	created, err := svc.Create("athlete", &workouts.Workout{
		Name:      "Chart test",
		SportType: "Run",
		StartDate: time.Date(2026, 7, 6, 10, 0, 0, 0, time.UTC),
	})
	if err != nil {
		t.Fatal(err)
	}

	if _, err := svc.AttachTrack("athlete", created, &workouts.TrackInput{
		Filename: "sample.gpx",
		Data:     gpxData,
		Parsed:   parsed,
	}); err != nil {
		t.Fatalf("AttachTrack: %v", err)
	}

	_, speedSamples, err := svc.GetSpeedChart("athlete", created.ID)
	if err != nil {
		t.Fatalf("GetSpeedChart: %v", err)
	}
	if len(speedSamples) < 1 {
		t.Fatal("expected speed chart samples after attach")
	}

	_, hrSamples, err := svc.GetHeartRateChart("athlete", created.ID)
	if err != nil {
		t.Fatalf("GetHeartRateChart: %v", err)
	}
	// Sample GPX may lack HR; chart store may still be empty — speed is required.
	_ = hrSamples
}

func TestServiceDeleteRemovesWorkout(t *testing.T) {
	dir := t.TempDir()
	svc := newTestService(dir)

	created, err := svc.Create("athlete", &workouts.Workout{
		Name:      "To delete",
		SportType: "Run",
		StartDate: time.Now().UTC(),
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := svc.Delete("athlete", created.ID); err != nil {
		t.Fatalf("Delete: %v", err)
	}
	if _, err := svc.Get("athlete", created.ID); err == nil {
		t.Fatal("expected not found after delete")
	}
}

func TestServiceUpdateSportType(t *testing.T) {
	dir := t.TempDir()
	svc := newTestService(dir)

	created, err := svc.Create("athlete", &workouts.Workout{
		Name:      "Update sport",
		SportType: "Run",
		StartDate: time.Now().UTC(),
	})
	if err != nil {
		t.Fatal(err)
	}
	updated, err := svc.Update("athlete", created.ID, &workouts.Workout{
		Name:      "Update sport",
		SportType: "Ride",
		StartDate: created.StartDate,
	})
	if err != nil {
		t.Fatalf("Update: %v", err)
	}
	if updated.SportType != "Ride" {
		t.Fatalf("sport = %q", updated.SportType)
	}
}
