package workouts_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	blobfs "github.com/solargate/grom/internal/storage/blob/fs"
	"github.com/solargate/grom/internal/storage/file"
	"github.com/solargate/grom/internal/tracks"
	"github.com/solargate/grom/internal/workouts"
)

func TestStoreCreateWithFITWritesStatsYAML(t *testing.T) {
	dir := t.TempDir()
	blobs := blobfs.NewStore(dir)
	charts := workouts.NewBlobSpeedChartStore(blobs)
	hrCharts := workouts.NewBlobHeartRateChartStore(blobs)
	svc := workouts.NewService(file.NewWorkoutsStore(dir), blobs, charts, hrCharts)

	fitData, err := os.ReadFile(filepath.Join("..", "..", "testdata", "tracks", "1-ride.fit"))
	if err != nil {
		t.Fatal(err)
	}
	parsed, err := tracks.Parse(fitData, "1-ride.fit")
	if err != nil {
		t.Fatal(err)
	}

	startDate := time.Date(2026, 7, 6, 10, 0, 0, 0, time.UTC)
	created, err := svc.CreateWithTrack("athlete", &workouts.Workout{
		Name:            "Bike ride",
		SportType:       "Ride",
		StartDate:       startDate,
		DurationSeconds: 0,
		Distance:        0,
		SpeedMaxKmh:     floatPtr(10),
	}, &workouts.TrackInput{
		Filename: "1-ride.fit",
		Data:     fitData,
		Parsed:   parsed,
	})
	if err != nil {
		t.Fatalf("CreateWithTrack() error = %v", err)
	}
	if created.DurationSeconds != 2041 {
		t.Fatalf("duration_seconds = %d, want 2041", created.DurationSeconds)
	}
	if created.DurationTotalSeconds != 3832 {
		t.Fatalf("duration_total_seconds = %d, want 3832", created.DurationTotalSeconds)
	}
	if created.SpeedMaxKmh == nil || *created.SpeedMaxKmh < 30 {
		t.Fatalf("expected track speed_max_kmh, got %v", created.SpeedMaxKmh)
	}
	if created.HeartRateMax == nil {
		t.Fatal("expected heart_rate_max")
	}
	if created.TempAvgKmm == nil || *created.TempAvgKmm == "" {
		t.Fatal("expected temp_avg_kmm")
	}

	expectedBase := created.StartDate.UTC().Format("2006-01-02T15:04:05Z")
	expectedBase = strings.ReplaceAll(expectedBase, ":", "") + "-" + created.ID
	yamlData, err := os.ReadFile(filepath.Join(dir, "users", "athlete", "workouts", expectedBase, expectedBase+".yaml"))
	if err != nil {
		t.Fatal(err)
	}
	content := string(yamlData)
	for _, needle := range []string{
		"duration_total_seconds: 3832",
		"speed_max_kmh:",
		"heart_rate_max:",
		"temp_avg_kmm:",
		"elevation_gain:",
	} {
		if !strings.Contains(content, needle) {
			t.Fatalf("yaml missing %q:\n%s", needle, content)
		}
	}
}

func floatPtr(v float64) *float64 {
	return &v
}
