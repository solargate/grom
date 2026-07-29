package bbolt_test

import (
	"context"
	"testing"
	"time"

	bolt "go.etcd.io/bbolt"

	storebbolt "github.com/solargate/grom/internal/storage/bbolt"
	"github.com/solargate/grom/internal/workouts"
)

func TestSpeedChartStoreLocalRoundTripAndDelete(t *testing.T) {
	b := openTestBackend(t)
	store := storebbolt.NewSpeedChartStore(b.DB())
	ctx := context.Background()
	start := time.Date(2026, 7, 8, 10, 0, 0, 0, time.UTC)
	samples := []workouts.SpeedSample{
		{Time: start, SpeedKmh: 10, DistanceM: 0},
		{Time: start.Add(time.Minute), SpeedKmh: 12.5, DistanceM: 200},
	}

	if err := store.WriteLocal(ctx, "alice", "dir-1", samples); err != nil {
		t.Fatal(err)
	}
	got, err := store.ReadLocal(ctx, "alice", "dir-1")
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 2 || got[1].SpeedKmh != 12.5 {
		t.Fatalf("got %#v", got)
	}

	if err := store.WriteLocal(ctx, "alice", "dir-1", nil); err != nil {
		t.Fatal(err)
	}
	got, err = store.ReadLocal(ctx, "alice", "dir-1")
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 0 {
		t.Fatalf("expected empty after clear, got %#v", got)
	}
}

func TestHeartRateChartStoreFederatedRoundTrip(t *testing.T) {
	b := openTestBackend(t)
	store := storebbolt.NewHeartRateChartStore(b.DB())
	ctx := context.Background()
	start := time.Date(2026, 7, 8, 10, 0, 0, 0, time.UTC)
	dist := 50.0
	samples := []workouts.HeartRateSample{
		{Time: start, BPM: 110, DistanceM: &dist},
		{Time: start.Add(time.Minute), BPM: 130},
	}

	if err := store.WriteFederated(ctx, "viewer", "owner_key", "wid", samples); err != nil {
		t.Fatal(err)
	}
	got, err := store.ReadFederated(ctx, "viewer", "owner_key", "wid")
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 2 || got[0].BPM != 110 || got[0].DistanceM == nil || got[1].DistanceM != nil {
		t.Fatalf("got %#v", got)
	}

	if err := store.DeleteFederated(ctx, "viewer", "owner_key", "wid"); err != nil {
		t.Fatal(err)
	}
	got, err = store.ReadFederated(ctx, "viewer", "owner_key", "wid")
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 0 {
		t.Fatalf("expected deleted, got %#v", got)
	}
}

func TestSpeedChartStoreRejectsBadMagic(t *testing.T) {
	b := openTestBackend(t)
	ctx := context.Background()
	key := []byte(workouts.LocalSpeedChartKey("alice", "bad"))
	if err := b.DB().Update(func(tx *bolt.Tx) error {
		return tx.Bucket([]byte("speed_charts")).Put(key, []byte("XXXX\x01\x00\x00\x00\x00\x00"))
	}); err != nil {
		t.Fatal(err)
	}

	store := storebbolt.NewSpeedChartStore(b.DB())
	if _, err := store.ReadLocal(ctx, "alice", "bad"); err == nil {
		t.Fatal("expected bad magic error")
	}
}

func TestWorkoutDeleteClearsLocalCharts(t *testing.T) {
	b := openTestBackend(t)
	repo := b.WorkoutsRepo()
	start := time.Date(2026, 7, 8, 10, 0, 0, 0, time.UTC)
	created, err := repo.Create("alice", &workouts.Workout{
		Name: "Run", SportType: "Run", StartDate: start, Track: "track.gpx",
	})
	if err != nil {
		t.Fatal(err)
	}
	dirName, err := repo.WorkoutDirName("alice", created.ID)
	if err != nil {
		t.Fatal(err)
	}

	ctx := context.Background()
	speedStore := storebbolt.NewSpeedChartStore(b.DB())
	hrStore := storebbolt.NewHeartRateChartStore(b.DB())
	if err := speedStore.WriteLocal(ctx, "alice", dirName, []workouts.SpeedSample{
		{Time: start, SpeedKmh: 10, DistanceM: 0},
	}); err != nil {
		t.Fatal(err)
	}
	if err := hrStore.WriteLocal(ctx, "alice", dirName, []workouts.HeartRateSample{
		{Time: start, BPM: 120},
	}); err != nil {
		t.Fatal(err)
	}

	if err := repo.Delete("alice", created.ID); err != nil {
		t.Fatal(err)
	}
	speed, err := speedStore.ReadLocal(ctx, "alice", dirName)
	if err != nil || len(speed) != 0 {
		t.Fatalf("speed after delete: %#v err=%v", speed, err)
	}
	hr, err := hrStore.ReadLocal(ctx, "alice", dirName)
	if err != nil || len(hr) != 0 {
		t.Fatalf("hr after delete: %#v err=%v", hr, err)
	}
}
