package bbolt_test

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/solargate/grom/internal/tracks"
	"github.com/solargate/grom/internal/workouts"
)

func TestInboxStoreSaveListDelete(t *testing.T) {
	b := openTestBackend(t)
	store := b.Federation().Inbox()

	gpxData, err := os.ReadFile(filepath.Join("..", "..", "..", "testdata", "tracks", "1-sample.gpx"))
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
	ownerHandle := "bob@remote.test"

	if err := store.Save("alice", ownerHandle, workout, gpxData, nil, nil); err != nil {
		t.Fatalf("Save: %v", err)
	}

	items, err := store.List("alice")
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 1 {
		t.Fatalf("list len = %d", len(items))
	}
	if items[0].ID != workout.ID || items[0].Owner != "bob" {
		t.Fatalf("item = %#v", items[0])
	}

	got, err := store.Get("alice", "bob", workout.ID)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if got.Name != "Remote run" {
		t.Fatalf("name = %q", got.Name)
	}

	if err := store.Delete("alice", ownerHandle, workout.ID); err != nil {
		t.Fatalf("Delete: %v", err)
	}
	items, err = store.List("alice")
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 0 {
		t.Fatalf("after delete len = %d", len(items))
	}
}
