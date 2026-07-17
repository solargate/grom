package workouts_test

import (
	"testing"
	"time"

	blobfs "github.com/solargate/grom/internal/storage/blob/fs"
	"github.com/solargate/grom/internal/workouts"
)

func TestFeedServiceMerge(t *testing.T) {
	dir := t.TempDir()
	svc := newTestService(dir)
	blobs := blobfs.NewStore(dir)

	start := time.Date(2026, 7, 8, 10, 0, 0, 0, time.UTC)
	_, err := svc.Create("alice", &workouts.Workout{
		Name:      "Alice run",
		SportType: "Run",
		StartDate: start,
	})
	if err != nil {
		t.Fatal(err)
	}
	_, err = svc.Create("bob", &workouts.Workout{
		Name:      "Bob ride",
		SportType: "Ride",
		StartDate: start.Add(-time.Hour),
	})
	if err != nil {
		t.Fatal(err)
	}

	feed := workouts.NewFeedService(svc, blobs, "localhost")
	items, err := feed.ListFeed("alice", "Alice", []workouts.FeedAuthor{
		{Nickname: "bob", Name: "Bob", Handle: "bob@localhost", IsLocal: true},
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 2 {
		t.Fatalf("expected 2 feed items, got %d", len(items))
	}
	if items[0].Name != "Alice run" {
		t.Fatalf("expected newest first, got %s", items[0].Name)
	}
	if items[0].Author.Name != "Alice" {
		t.Fatalf("expected viewer name on own workout, got %q", items[0].Author.Name)
	}
}

func TestFeedServiceListOwn(t *testing.T) {
	dir := t.TempDir()
	svc := newTestService(dir)
	blobs := blobfs.NewStore(dir)

	start := time.Date(2026, 7, 8, 10, 0, 0, 0, time.UTC)
	_, err := svc.Create("alice", &workouts.Workout{
		Name:      "Alice run",
		SportType: "Run",
		StartDate: start,
	})
	if err != nil {
		t.Fatal(err)
	}
	_, err = svc.Create("bob", &workouts.Workout{
		Name:      "Bob ride",
		SportType: "Ride",
		StartDate: start.Add(-time.Hour),
	})
	if err != nil {
		t.Fatal(err)
	}

	feed := workouts.NewFeedService(svc, blobs, "localhost")
	items, err := feed.ListOwn("alice", "Alice")
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 1 {
		t.Fatalf("expected 1 own workout, got %d", len(items))
	}
	if items[0].Name != "Alice run" {
		t.Fatalf("expected Alice run, got %s", items[0].Name)
	}
}
