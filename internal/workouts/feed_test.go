package workouts_test

import (
	"sort"
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

func TestFeedServiceCanAccessWorkout(t *testing.T) {
	feed := workouts.NewFeedService(nil, nil, "localhost")
	if !feed.CanAccessWorkout("alice", nil, "alice") {
		t.Fatal("owner should access own workout")
	}
	if !feed.CanAccessWorkout("alice", []string{"bob"}, "bob") {
		t.Fatal("follower should access followed workout")
	}
	if feed.CanAccessWorkout("alice", []string{"bob"}, "charlie") {
		t.Fatal("should not access stranger workout")
	}
}

type stubFederatedSource struct {
	items []workouts.FeedWorkout
	err   error
}

func (s stubFederatedSource) ListFederated(viewerNickname string) ([]workouts.FeedWorkout, error) {
	return s.items, s.err
}

func (s stubFederatedSource) ListFederatedPage(viewerNickname string, cursor *workouts.Cursor, limit int) ([]workouts.FeedWorkout, bool, error) {
	if s.err != nil {
		return nil, false, s.err
	}
	limit = workouts.ClampLimit(limit)
	items := append([]workouts.FeedWorkout{}, s.items...)
	sort.Slice(items, func(i, j int) bool {
		return workouts.FeedNewer(items[i].StartDate, items[i].ID, items[j].StartDate, items[j].ID)
	})
	out := make([]workouts.FeedWorkout, 0, limit)
	for i := range items {
		if !workouts.AfterCursor(items[i].StartDate, items[i].ID, cursor) {
			continue
		}
		out = append(out, items[i])
		if len(out) > limit {
			break
		}
	}
	hasMore := len(out) > limit
	if hasMore {
		out = out[:limit]
	}
	return out, hasMore, nil
}

func TestFeedServiceListOwnPage(t *testing.T) {
	dir := t.TempDir()
	svc := newTestService(dir)
	blobs := blobfs.NewStore(dir)

	start := time.Date(2026, 7, 8, 10, 0, 0, 0, time.UTC)
	for i := 0; i < 5; i++ {
		_, err := svc.Create("alice", &workouts.Workout{
			Name:      "Run",
			SportType: "Run",
			StartDate: start.Add(-time.Duration(i) * time.Hour),
		})
		if err != nil {
			t.Fatal(err)
		}
	}

	feed := workouts.NewFeedService(svc, blobs, "localhost")
	page1, err := feed.ListOwnPage("alice", "Alice", nil, 2, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(page1.Items) != 2 || !page1.HasMore || page1.NextCursor == "" {
		t.Fatalf("page1 = %#v", page1)
	}

	cursor, err := workouts.DecodeCursor(page1.NextCursor)
	if err != nil {
		t.Fatal(err)
	}
	page2, err := feed.ListOwnPage("alice", "Alice", cursor, 2, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(page2.Items) != 2 || !page2.HasMore {
		t.Fatalf("page2 = %#v", page2)
	}
	if page2.Items[0].ID == page1.Items[0].ID || page2.Items[0].ID == page1.Items[1].ID {
		t.Fatalf("pages overlap: %v vs %v", page1.Items, page2.Items)
	}

	cursor2, err := workouts.DecodeCursor(page2.NextCursor)
	if err != nil {
		t.Fatal(err)
	}
	page3, err := feed.ListOwnPage("alice", "Alice", cursor2, 2, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(page3.Items) != 1 || page3.HasMore {
		t.Fatalf("page3 = %#v", page3)
	}
}

func TestFeedServiceListOwnPageSportTypesFilter(t *testing.T) {
	dir := t.TempDir()
	svc := newTestService(dir)
	blobs := blobfs.NewStore(dir)

	start := time.Date(2026, 7, 8, 10, 0, 0, 0, time.UTC)
	types := []string{"Run", "Ride", "Run", "Walk", "Ride"}
	for i, sport := range types {
		_, err := svc.Create("alice", &workouts.Workout{
			Name:      sport,
			SportType: sport,
			StartDate: start.Add(-time.Duration(i) * time.Hour),
		})
		if err != nil {
			t.Fatal(err)
		}
	}

	feed := workouts.NewFeedService(svc, blobs, "localhost")
	allow := map[string]struct{}{"Run": {}, "Walk": {}}
	page1, err := feed.ListOwnPage("alice", "Alice", nil, 2, allow)
	if err != nil {
		t.Fatal(err)
	}
	if len(page1.Items) != 2 || !page1.HasMore {
		t.Fatalf("page1 = %#v", page1)
	}
	if page1.Items[0].SportType != "Run" || page1.Items[1].SportType != "Run" {
		t.Fatalf("unexpected sports: %#v", page1.Items)
	}

	cursor, err := workouts.DecodeCursor(page1.NextCursor)
	if err != nil {
		t.Fatal(err)
	}
	page2, err := feed.ListOwnPage("alice", "Alice", cursor, 2, allow)
	if err != nil {
		t.Fatal(err)
	}
	if len(page2.Items) != 1 || page2.HasMore {
		t.Fatalf("page2 = %#v", page2)
	}
	if page2.Items[0].SportType != "Walk" {
		t.Fatalf("expected Walk, got %#v", page2.Items[0])
	}
}

func TestFeedServiceListFeedPageMergesSources(t *testing.T) {
	dir := t.TempDir()
	svc := newTestService(dir)
	blobs := blobfs.NewStore(dir)

	start := time.Date(2026, 7, 8, 12, 0, 0, 0, time.UTC)
	_, err := svc.Create("alice", &workouts.Workout{
		Name: "Alice late", SportType: "Run", StartDate: start,
	})
	if err != nil {
		t.Fatal(err)
	}
	_, err = svc.Create("bob", &workouts.Workout{
		Name: "Bob mid", SportType: "Ride", StartDate: start.Add(-time.Hour),
	})
	if err != nil {
		t.Fatal(err)
	}

	feed := workouts.NewFeedService(svc, blobs, "localhost")
	feed.SetFederatedSource(stubFederatedSource{items: []workouts.FeedWorkout{{
		Workout: workouts.Workout{
			ID: "99999999", Name: "Remote early", SportType: "Ride",
			StartDate: start.Add(-2 * time.Hour),
		},
		Author: workouts.FeedAuthor{Nickname: "remote", Handle: "remote@other", IsLocal: false},
		Owner:  "remote@other",
	}}})

	page, err := feed.ListFeedPage("alice", "Alice", []workouts.FeedAuthor{
		{Nickname: "bob", Name: "Bob", Handle: "bob@localhost", IsLocal: true},
	}, nil, 2)
	if err != nil {
		t.Fatal(err)
	}
	if len(page.Items) != 2 || !page.HasMore {
		t.Fatalf("expected 2 items with more, got %#v", page)
	}
	if page.Items[0].Name != "Alice late" || page.Items[1].Name != "Bob mid" {
		t.Fatalf("unexpected order: %#v", page.Items)
	}

	cursor, err := workouts.DecodeCursor(page.NextCursor)
	if err != nil {
		t.Fatal(err)
	}
	page2, err := feed.ListFeedPage("alice", "Alice", []workouts.FeedAuthor{
		{Nickname: "bob", Name: "Bob", Handle: "bob@localhost", IsLocal: true},
	}, cursor, 2)
	if err != nil {
		t.Fatal(err)
	}
	if len(page2.Items) != 1 || page2.HasMore {
		t.Fatalf("page2 = %#v", page2)
	}
	if page2.Items[0].Name != "Remote early" {
		t.Fatalf("expected remote leftover, got %#v", page2.Items[0])
	}
}

func TestFeedServiceMergesFederated(t *testing.T) {
	dir := t.TempDir()
	svc := newTestService(dir)
	blobs := blobfs.NewStore(dir)

	start := time.Date(2026, 7, 8, 10, 0, 0, 0, time.UTC)
	_, err := svc.Create("alice", &workouts.Workout{
		Name: "Alice run", SportType: "Run", StartDate: start,
	})
	if err != nil {
		t.Fatal(err)
	}

	feed := workouts.NewFeedService(svc, blobs, "localhost")
	feed.SetFederatedSource(stubFederatedSource{items: []workouts.FeedWorkout{{
		Workout: workouts.Workout{
			ID: "99999999", Name: "Remote", SportType: "Ride",
			StartDate: start.Add(time.Hour),
		},
		Author: workouts.FeedAuthor{Nickname: "remote", Handle: "remote@other", IsLocal: false},
		Owner:  "remote@other",
	}}})

	items, err := feed.ListFeed("alice", "Alice", nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 2 {
		t.Fatalf("expected 2 items, got %d", len(items))
	}
	if items[0].Name != "Remote" {
		t.Fatalf("expected federated newest first, got %q", items[0].Name)
	}
}
