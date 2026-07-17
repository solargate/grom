package federation

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/solargate/grom/internal/tracks"
	"github.com/solargate/grom/internal/workouts"
)

func TestInboxProcessorHandleDelete(t *testing.T) {
	dir := t.TempDir()
	store := newTestInboxStore(dir)

	gpxData, err := os.ReadFile(filepath.Join("..", "..", "testdata", "1-sample.gpx"))
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

	processor := NewInboxProcessor(nil, nil, nil, store, nil)
	body := strings.NewReader(`{"type":"Delete","actor":"https://192.168.1.251:8445/users/test2","object":"https://localhost/users/test2/workouts/38472901"}`)
	if err := processor.Handle("solarwind", body); err != nil {
		t.Fatalf("Handle() error = %v", err)
	}

	items, err := store.List("solarwind")
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 0 {
		t.Fatalf("expected 0 items after delete, got %d", len(items))
	}
}
