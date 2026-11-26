package bbolt_test

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	storebbolt "github.com/solargate/grom/internal/storage/bbolt"
	"github.com/solargate/grom/internal/tracks"
	"github.com/solargate/grom/internal/workouts"
	bolt "go.etcd.io/bbolt"
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

func BenchmarkInboxPageTenThousand(b *testing.B) {
	dir := b.TempDir()
	backend, err := storebbolt.Open(filepath.Join(dir, "grom.db"), dir)
	if err != nil {
		b.Fatal(err)
	}
	defer backend.Close()
	inbox := backend.Federation().Inbox().(*storebbolt.InboxStore)
	start := time.Date(2026, 7, 8, 10, 0, 0, 0, time.UTC)
	for i := range 10000 {
		item := &workouts.Workout{
			ID: fmt.Sprintf("%08d", i), Name: "Remote run", SportType: "Run",
			StartDate: start.Add(time.Duration(i) * time.Second),
		}
		if err := inbox.PutExistingWorkout("alice", "bob_example.org", item); err != nil {
			b.Fatal(err)
		}
	}
	deep := &workouts.Cursor{StartDate: start.Add(2000 * time.Second), ID: "00002000"}
	for _, page := range []struct {
		name   string
		cursor *workouts.Cursor
	}{{"first", nil}, {"deep", deep}} {
		b.Run(page.name+"/full-scan", func(b *testing.B) {
			b.ReportAllocs()
			for b.Loop() {
				all, err := inbox.List("alice")
				if err != nil {
					b.Fatal(err)
				}
				selected := 0
				for _, item := range all {
					if workouts.AfterFeedCursor(item.StartDate, item.ID, item.SortOwner, page.cursor) {
						selected++
					}
					if selected > 20 {
						break
					}
				}
			}
		})
		b.Run(page.name+"/indexed", func(b *testing.B) {
			b.ReportAllocs()
			for b.Loop() {
				items, _, err := inbox.ListPage("alice", page.cursor, 20)
				if err != nil || len(items) != 20 {
					b.Fatalf("got %d workouts, err=%v", len(items), err)
				}
			}
		})
	}
}

func TestInboxOrderIndexBackfillsExistingDatabase(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "grom.db")
	backend, err := storebbolt.Open(dbPath, dir)
	if err != nil {
		t.Fatal(err)
	}
	inbox := backend.Federation().Inbox().(*storebbolt.InboxStore)
	item := &workouts.Workout{ID: "42", Name: "Run", SportType: "Run", StartDate: time.Now().UTC()}
	if err := inbox.PutExistingWorkout("alice", "bob_example.org", item); err != nil {
		t.Fatal(err)
	}
	if err := backend.DB().Update(func(tx *bolt.Tx) error {
		return tx.DeleteBucket([]byte("idx_fed_inbox_order"))
	}); err != nil {
		t.Fatal(err)
	}
	if err := backend.Close(); err != nil {
		t.Fatal(err)
	}

	backend, err = storebbolt.Open(dbPath, dir)
	if err != nil {
		t.Fatal(err)
	}
	defer backend.Close()
	items, more, err := backend.Federation().Inbox().ListPage("alice", nil, 1)
	if err != nil || more || len(items) != 1 || items[0].ID != item.ID {
		t.Fatalf("page after migration = %v, more=%v, err=%v", items, more, err)
	}
}
