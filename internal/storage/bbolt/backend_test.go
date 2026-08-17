package bbolt_test

import (
	"context"
	"errors"
	"path/filepath"
	"testing"
	"time"

	"github.com/solargate/grom/internal/config"
	"github.com/solargate/grom/internal/federation"
	"github.com/solargate/grom/internal/storage"
	storebbolt "github.com/solargate/grom/internal/storage/bbolt"
	"github.com/solargate/grom/internal/users"
	"github.com/solargate/grom/internal/workouts"
)

func openTestBackend(t *testing.T) *storebbolt.Backend {
	t.Helper()
	dir := t.TempDir()
	backend, err := storebbolt.Open(filepath.Join(dir, "grom.db"), dir)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	t.Cleanup(func() { _ = backend.Close() })
	return backend
}

func TestOpenBBoltBackend(t *testing.T) {
	dir := t.TempDir()
	cfg := config.StorageConfig{
		Driver:            config.StorageDriverBBolt,
		ResolvedLocation:  dir,
		ResolvedBBoltPath: filepath.Join(dir, "grom.db"),
	}
	backend, err := storage.Open(cfg)
	if err != nil {
		t.Fatal(err)
	}
	defer backend.Close()
	if err := backend.Ping(context.Background()); err != nil {
		t.Fatal(err)
	}
}

func TestUsersStoreCreateAndFind(t *testing.T) {
	b := openTestBackend(t)
	created, err := b.Users().Create("Alice", "Alice Name", "Alice@Example.com", "password123")
	if err != nil {
		t.Fatal(err)
	}
	byEmail, err := b.Users().FindByEmail("alice@example.com")
	if err != nil {
		t.Fatal(err)
	}
	if byEmail.ID != created.ID || byEmail.PasswordHash == "" {
		t.Fatalf("unexpected user: %#v", byEmail)
	}
	if _, err := b.Users().Create("bob", "Bob", "alice@example.com", "password123"); !errors.Is(err, users.ErrEmailTaken) {
		t.Fatalf("duplicate email: %v", err)
	}
}

func TestWorkoutsStoreCRUDAndListPage(t *testing.T) {
	b := openTestBackend(t)
	repo := b.WorkoutsRepo()

	w1, err := repo.Create("alice", &workouts.Workout{
		Name: "Morning", SportType: "Run",
		StartDate: time.Date(2026, 7, 8, 10, 0, 0, 0, time.UTC),
	})
	if err != nil {
		t.Fatal(err)
	}
	w2, err := repo.Create("alice", &workouts.Workout{
		Name: "Evening", SportType: "Run",
		StartDate: time.Date(2026, 7, 9, 18, 0, 0, 0, time.UTC),
	})
	if err != nil {
		t.Fatal(err)
	}

	got, err := repo.Get("alice", w1.ID)
	if err != nil || got.Name != "Morning" {
		t.Fatalf("Get: %#v err=%v", got, err)
	}

	page, hasMore, err := repo.ListPage("alice", nil, 1)
	if err != nil {
		t.Fatal(err)
	}
	if len(page) != 1 || page[0].ID != w2.ID || !hasMore {
		t.Fatalf("page=%#v hasMore=%v", page, hasMore)
	}
	page2, hasMore2, err := repo.ListPage("alice", &workouts.Cursor{StartDate: page[0].StartDate, ID: page[0].ID}, 10)
	if err != nil {
		t.Fatal(err)
	}
	if len(page2) != 1 || page2[0].ID != w1.ID || hasMore2 {
		t.Fatalf("page2=%#v hasMore=%v", page2, hasMore2)
	}

	if err := repo.Delete("alice", w1.ID); err != nil {
		t.Fatal(err)
	}
	if _, err := repo.Get("alice", w1.ID); !errors.Is(err, workouts.ErrWorkoutNotFound) {
		t.Fatalf("expected not found, got %v", err)
	}
}

func TestWorkoutsStoreListPageDoesNotDuplicateNewest(t *testing.T) {
	b := openTestBackend(t)
	repo := b.WorkoutsRepo()

	start := time.Date(2026, 7, 8, 15, 0, 0, 0, time.UTC)
	aliceIDs := make([]string, 0, 3)
	for i := 0; i < 3; i++ {
		created, err := repo.Create("alice", &workouts.Workout{
			Name: "Alice", SportType: "Run",
			StartDate: start.Add(-time.Duration(i) * time.Hour),
		})
		if err != nil {
			t.Fatal(err)
		}
		aliceIDs = append(aliceIDs, created.ID)
	}
	// incrementPrefix("alice/") == "alice0"; a neighbor nick must not leak into alice's page.
	alice0w, err := repo.Create("alice0", &workouts.Workout{
		Name: "Alice0", SportType: "Run",
		StartDate: start.Add(30 * time.Minute),
	})
	if err != nil {
		t.Fatal(err)
	}
	bobNewest, err := repo.Create("bob", &workouts.Workout{
		Name: "Bob", SportType: "Ride",
		StartDate: start.Add(time.Hour),
	})
	if err != nil {
		t.Fatal(err)
	}

	assertUniqueNewestFirst := func(t *testing.T, page []workouts.Workout, wantIDs []string, hasMore bool, gotMore bool) {
		t.Helper()
		if gotMore != hasMore {
			t.Fatalf("hasMore=%v want %v page=%#v", gotMore, hasMore, page)
		}
		if len(page) != len(wantIDs) {
			t.Fatalf("len=%d want %d page=%#v", len(page), len(wantIDs), page)
		}
		seen := make(map[string]struct{}, len(page))
		for i, w := range page {
			if w.ID != wantIDs[i] {
				t.Fatalf("page[%d]=%q want %q", i, w.ID, wantIDs[i])
			}
			if _, dup := seen[w.ID]; dup {
				t.Fatalf("duplicate id %q in page=%#v", w.ID, page)
			}
			seen[w.ID] = struct{}{}
		}
	}

	// Newer-nickname keys used to leave the cursor past alice's prefix, so the
	// newest alice workout was appended twice on the first page (limit > 1).
	page, more, err := repo.ListPage("alice", nil, 10)
	if err != nil {
		t.Fatal(err)
	}
	assertUniqueNewestFirst(t, page, aliceIDs, false, more)

	page, more, err = repo.ListPage("alice", nil, 2)
	if err != nil {
		t.Fatal(err)
	}
	assertUniqueNewestFirst(t, page, aliceIDs[:2], true, more)

	cursor := &workouts.Cursor{StartDate: page[1].StartDate, ID: page[1].ID}
	page2, more2, err := repo.ListPage("alice", cursor, 10)
	if err != nil {
		t.Fatal(err)
	}
	assertUniqueNewestFirst(t, page2, aliceIDs[2:], false, more2)

	alice0Page, alice0More, err := repo.ListPage("alice0", nil, 10)
	if err != nil {
		t.Fatal(err)
	}
	assertUniqueNewestFirst(t, alice0Page, []string{alice0w.ID}, false, alice0More)

	// Last nickname alphabetically: Seek(nextPrefix) hits end-of-bucket (nil), then Last().
	bobPage, bobMore, err := repo.ListPage("bob", nil, 10)
	if err != nil {
		t.Fatal(err)
	}
	assertUniqueNewestFirst(t, bobPage, []string{bobNewest.ID}, false, bobMore)

	empty, emptyMore, err := repo.ListPage("carol", nil, 10)
	if err != nil {
		t.Fatal(err)
	}
	if len(empty) != 0 || emptyMore {
		t.Fatalf("empty user: page=%#v hasMore=%v", empty, emptyMore)
	}
}

func TestWorkoutsUpdateMigratesChartsOnStartDateChange(t *testing.T) {
	b := openTestBackend(t)
	repo := b.WorkoutsRepo()
	svc := b.Workouts()

	start := time.Date(2026, 7, 8, 10, 0, 47, 0, time.UTC)
	created, err := repo.Create("alice", &workouts.Workout{
		Name:      "Ride",
		SportType: "Ride",
		StartDate: start,
		Track:     "track.gpx",
	})
	if err != nil {
		t.Fatal(err)
	}

	oldDirName, err := repo.WorkoutDirName("alice", created.ID)
	if err != nil {
		t.Fatal(err)
	}

	ctx := context.Background()
	speedStore := storebbolt.NewSpeedChartStore(b.DB())
	hrStore := storebbolt.NewHeartRateChartStore(b.DB())
	speedSamples := []workouts.SpeedSample{
		{Time: start, SpeedKmh: 22.5, DistanceM: 0},
		{Time: start.Add(time.Minute), SpeedKmh: 25, DistanceM: 400},
	}
	dist := 100.0
	hrSamples := []workouts.HeartRateSample{
		{Time: start, BPM: 120, DistanceM: &dist},
		{Time: start.Add(time.Minute), BPM: 140, DistanceM: &dist},
	}
	if err := speedStore.WriteLocal(ctx, "alice", oldDirName, speedSamples); err != nil {
		t.Fatalf("write speed chart: %v", err)
	}
	if err := hrStore.WriteLocal(ctx, "alice", oldDirName, hrSamples); err != nil {
		t.Fatalf("write heart rate chart: %v", err)
	}

	// Simulate edit form truncating seconds (old bug trigger).
	patched := *created
	patched.StartDate = time.Date(2026, 7, 8, 10, 0, 0, 0, time.UTC)
	patched.Equipment = []workouts.WorkoutEquipment{{ID: "eq-1", Name: "Shoes", Type: "shoes"}}
	if _, err := repo.Update("alice", &patched); err != nil {
		t.Fatalf("Update: %v", err)
	}

	newDirName, err := repo.WorkoutDirName("alice", created.ID)
	if err != nil {
		t.Fatal(err)
	}
	if newDirName == oldDirName {
		t.Fatalf("expected dir rename, still %q", oldDirName)
	}

	_, gotSpeed, err := svc.GetSpeedChart("alice", created.ID)
	if err != nil {
		t.Fatalf("GetSpeedChart: %v", err)
	}
	if len(gotSpeed) != 2 || gotSpeed[0].SpeedKmh != 22.5 {
		t.Fatalf("speed chart after rename = %#v", gotSpeed)
	}

	_, gotHR, err := svc.GetHeartRateChart("alice", created.ID)
	if err != nil {
		t.Fatalf("GetHeartRateChart: %v", err)
	}
	if len(gotHR) != 2 || gotHR[1].BPM != 140 {
		t.Fatalf("heart rate chart after rename = %#v", gotHR)
	}

	oldSpeed, err := speedStore.ReadLocal(ctx, "alice", oldDirName)
	if err != nil {
		t.Fatal(err)
	}
	if len(oldSpeed) != 0 {
		t.Fatalf("expected old speed chart key cleared, got %#v", oldSpeed)
	}
}

func TestWorkoutsBeginCreateWriteMetadata(t *testing.T) {
	b := openTestBackend(t)
	repo := b.WorkoutsRepo()
	w := &workouts.Workout{
		Name: "Track", SportType: "Ride",
		StartDate: time.Date(2026, 7, 10, 12, 0, 0, 0, time.UTC),
	}
	created, dirName, cleanup, err := repo.BeginCreate("alice", w)
	if err != nil {
		t.Fatal(err)
	}
	if created.ID == "" || dirName == "" {
		t.Fatalf("unexpected begin: %#v %q", created, dirName)
	}
	if err := repo.WriteMetadata("alice", created); err != nil {
		cleanup()
		t.Fatal(err)
	}
	got, err := repo.Get("alice", created.ID)
	if err != nil {
		t.Fatal(err)
	}
	if got.Name != "Track" {
		t.Fatalf("name=%q", got.Name)
	}
}

func TestFederationFollowersAndInboxMeta(t *testing.T) {
	b := openTestBackend(t)
	followers := b.Federation().Followers()
	if err := followers.Add("alice", federation.InboundFollower{
		ActorURI: "https://remote/users/bob",
		Inbox:    "https://remote/inbox",
		Handle:   "bob@remote",
	}); err != nil {
		t.Fatal(err)
	}
	list, err := followers.List("alice")
	if err != nil || len(list) != 1 {
		t.Fatalf("list=%v err=%v", list, err)
	}

	inbox := b.Federation().Inbox()
	w := &workouts.Workout{
		ID: "87654321", Name: "Remote ride", SportType: "Ride",
		StartDate: time.Date(2026, 7, 1, 9, 0, 0, 0, time.UTC),
	}
	if err := inbox.Save("alice", "bob@remote", w, nil, nil, map[string]any{"name": "Bob"}); err != nil {
		t.Fatal(err)
	}
	got, err := inbox.Get("alice", "bob", "87654321")
	if err != nil {
		t.Fatal(err)
	}
	if got.Name != "Remote ride" || got.Author.Name != "Bob" {
		t.Fatalf("got=%#v", got)
	}
}

func TestWorkoutsHasExternalID(t *testing.T) {
	b := openTestBackend(t)
	repo := b.WorkoutsRepo()

	created, err := repo.Create("alice", &workouts.Workout{
		Name:      "Ride",
		SportType: "Ride",
		StartDate: time.Date(2026, 7, 5, 14, 30, 0, 0, time.UTC),
		ExternalID: &workouts.ExternalID{
			Name: "strava",
			ID:   "strava-42",
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if created.ExternalID == nil || created.ExternalID.ID != "strava-42" {
		t.Fatalf("external_id not persisted: %#v", created.ExternalID)
	}

	ok, err := repo.HasExternalID("alice", "strava", "strava-42")
	if err != nil || !ok {
		t.Fatalf("HasExternalID = %v err=%v", ok, err)
	}
	ok, err = repo.HasExternalID("alice", "strava", "")
	if err != nil || ok {
		t.Fatalf("empty external id should be false: %v %v", ok, err)
	}
	ok, err = repo.HasExternalID("alice", "", "strava-42")
	if err != nil || ok {
		t.Fatalf("empty external name should be false: %v %v", ok, err)
	}
	ok, err = repo.HasExternalID("alice", "strava", "missing")
	if err != nil || ok {
		t.Fatalf("missing id should be false: %v %v", ok, err)
	}
}
