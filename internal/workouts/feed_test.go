package workouts

import (
	"testing"
	"time"
)

func TestFeedServiceMerge(t *testing.T) {
	dir := t.TempDir()
	store := NewStore(dir)

	start := time.Date(2026, 7, 8, 10, 0, 0, 0, time.UTC)
	_, err := store.Create("alice", &Workout{
		Name:      "Alice run",
		SportType: "Run",
		StartDate: start,
	})
	if err != nil {
		t.Fatal(err)
	}
	_, err = store.Create("bob", &Workout{
		Name:      "Bob ride",
		SportType: "Ride",
		StartDate: start.Add(-time.Hour),
	})
	if err != nil {
		t.Fatal(err)
	}

	feed := NewFeedService(store, "localhost")
	items, err := feed.ListFeed("alice", "Alice", []FeedAuthor{
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
