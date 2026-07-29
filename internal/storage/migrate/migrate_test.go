package migrate_test

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/solargate/grom/internal/config"
	"github.com/solargate/grom/internal/equipment"
	"github.com/solargate/grom/internal/social"
	"github.com/solargate/grom/internal/storage"
	"github.com/solargate/grom/internal/storage/migrate"
	"github.com/solargate/grom/internal/workouts"
)

func TestMigrateFileToBBoltAndBack(t *testing.T) {
	dir := t.TempDir()
	cfg := config.StorageConfig{
		Driver:            config.StorageDriverFile,
		ResolvedLocation:  dir,
		ResolvedBBoltPath: filepath.Join(dir, "grom.db"),
	}

	fileBackend, err := storage.Open(cfg)
	if err != nil {
		t.Fatal(err)
	}

	user, err := fileBackend.Users().Create("alice", "Alice", "alice@example.com", "password123")
	if err != nil {
		t.Fatal(err)
	}
	bike := &equipment.Equipment{Type: equipment.TypeBike, Name: "Road", BikeType: "road"}
	if _, err := fileBackend.Equipment().Create("alice", bike); err != nil {
		t.Fatal(err)
	}
	if _, err := fileBackend.Workouts().Create("alice", &workouts.Workout{
		Name: "Run", SportType: "Run",
		StartDate: time.Date(2026, 7, 8, 10, 0, 0, 0, time.UTC),
	}); err != nil {
		t.Fatal(err)
	}
	if _, err := fileBackend.Social().Create(social.Follow{
		FollowerID:     user.ID,
		TargetHandle:   "bob@local",
		TargetNickname: "bob",
		TargetIsLocal:  true,
		Status:         social.StatusActive,
		CreatedAt:      time.Now().UTC(),
	}); err != nil {
		t.Fatal(err)
	}
	_ = fileBackend.Close()

	result, err := migrate.Run(migrate.Options{
		From:   config.StorageDriverFile,
		To:     config.StorageDriverBBolt,
		Config: cfg,
		Verify: true,
	})
	if err != nil {
		t.Fatalf("file→bbolt: %v", err)
	}
	if result.Users != 1 || result.Workouts != 1 || result.Equipment != 1 || result.Follows != 1 {
		t.Fatalf("unexpected result: %+v", result)
	}

	bboltCfg := cfg
	bboltCfg.Driver = config.StorageDriverBBolt
	boltBackend, err := storage.Open(bboltCfg)
	if err != nil {
		t.Fatal(err)
	}
	got, err := boltBackend.Users().FindByNickname("alice")
	if err != nil {
		t.Fatal(err)
	}
	if got.PasswordHash == "" {
		t.Fatal("password hash missing after migrate")
	}
	ws, err := boltBackend.Workouts().List("alice")
	if err != nil || len(ws) != 1 {
		t.Fatalf("workouts=%v err=%v", ws, err)
	}
	_ = boltBackend.Close()

	// Round-trip back to file into the same location (upsert).
	result2, err := migrate.Run(migrate.Options{
		From:   config.StorageDriverBBolt,
		To:     config.StorageDriverFile,
		Config: cfg,
		Verify: true,
	})
	if err != nil {
		t.Fatalf("bbolt→file: %v", err)
	}
	if result2.Users != 1 || result2.Workouts != 1 {
		t.Fatalf("round-trip result: %+v", result2)
	}
}

func TestMigratePreservesLocalCharts(t *testing.T) {
	dir := t.TempDir()
	cfg := config.StorageConfig{
		Driver:            config.StorageDriverFile,
		ResolvedLocation:  dir,
		ResolvedBBoltPath: filepath.Join(dir, "grom.db"),
	}

	fileBackend, err := storage.Open(cfg)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := fileBackend.Users().Create("alice", "Alice", "alice@example.com", "password123"); err != nil {
		t.Fatal(err)
	}

	gpx, err := os.ReadFile(filepath.Join("..", "..", "..", "testdata", "tracks", "1-sample.gpx"))
	if err != nil {
		t.Fatal(err)
	}
	fit, err := os.ReadFile(filepath.Join("..", "..", "..", "testdata", "tracks", "1-ride.fit"))
	if err != nil {
		t.Fatal(err)
	}

	gpxWorkout, err := fileBackend.Workouts().CreateWithTrack("alice", &workouts.Workout{
		Name: "GPX run", SportType: "Run",
		StartDate: time.Date(2026, 7, 8, 10, 0, 0, 0, time.UTC),
	}, &workouts.TrackInput{Filename: "1-sample.gpx", Data: gpx})
	if err != nil {
		t.Fatalf("CreateWithTrack GPX: %v", err)
	}
	fitWorkout, err := fileBackend.Workouts().CreateWithTrack("alice", &workouts.Workout{
		Name: "FIT ride", SportType: "Ride",
		StartDate: time.Date(2026, 7, 9, 10, 0, 0, 0, time.UTC),
	}, &workouts.TrackInput{Filename: "1-ride.fit", Data: fit})
	if err != nil {
		t.Fatalf("CreateWithTrack FIT: %v", err)
	}

	_, wantSpeed, err := fileBackend.Workouts().GetSpeedChart("alice", gpxWorkout.ID)
	if err != nil || len(wantSpeed) < 1 {
		t.Fatalf("file speed chart: len=%d err=%v", len(wantSpeed), err)
	}
	_, wantHR, err := fileBackend.Workouts().GetHeartRateChart("alice", fitWorkout.ID)
	if err != nil || len(wantHR) < 1 {
		t.Fatalf("file HR chart: len=%d err=%v", len(wantHR), err)
	}
	_ = fileBackend.Close()

	if _, err := migrate.Run(migrate.Options{
		From: config.StorageDriverFile, To: config.StorageDriverBBolt, Config: cfg, Force: true,
	}); err != nil {
		t.Fatalf("file→bbolt: %v", err)
	}

	bboltCfg := cfg
	bboltCfg.Driver = config.StorageDriverBBolt
	boltBackend, err := storage.Open(bboltCfg)
	if err != nil {
		t.Fatal(err)
	}
	_, gotSpeed, err := boltBackend.Workouts().GetSpeedChart("alice", gpxWorkout.ID)
	if err != nil {
		t.Fatalf("bbolt GetSpeedChart: %v", err)
	}
	if len(gotSpeed) != len(wantSpeed) {
		t.Fatalf("bbolt speed len=%d want %d", len(gotSpeed), len(wantSpeed))
	}
	_, gotHR, err := boltBackend.Workouts().GetHeartRateChart("alice", fitWorkout.ID)
	if err != nil {
		t.Fatalf("bbolt GetHeartRateChart: %v", err)
	}
	if len(gotHR) != len(wantHR) {
		t.Fatalf("bbolt HR len=%d want %d", len(gotHR), len(wantHR))
	}
	_ = boltBackend.Close()

	// Remove JSON chart blobs so bbolt→file must rewrite them from binary buckets.
	gpxDir := filepath.Join(dir, "users", "alice", "workouts")
	entries, err := os.ReadDir(gpxDir)
	if err != nil {
		t.Fatal(err)
	}
	for _, e := range entries {
		_ = os.Remove(filepath.Join(gpxDir, e.Name(), "speed-chart.json"))
		_ = os.Remove(filepath.Join(gpxDir, e.Name(), "heartrate-chart.json"))
	}

	if _, err := migrate.Run(migrate.Options{
		From: config.StorageDriverBBolt, To: config.StorageDriverFile, Config: cfg,
	}); err != nil {
		t.Fatalf("bbolt→file: %v", err)
	}

	fileCfg := cfg
	fileCfg.Driver = config.StorageDriverFile
	fileBackend2, err := storage.Open(fileCfg)
	if err != nil {
		t.Fatal(err)
	}
	defer fileBackend2.Close()
	_, roundSpeed, err := fileBackend2.Workouts().GetSpeedChart("alice", gpxWorkout.ID)
	if err != nil || len(roundSpeed) != len(wantSpeed) {
		t.Fatalf("round-trip speed len=%d want %d err=%v", len(roundSpeed), len(wantSpeed), err)
	}
	_, roundHR, err := fileBackend2.Workouts().GetHeartRateChart("alice", fitWorkout.ID)
	if err != nil || len(roundHR) != len(wantHR) {
		t.Fatalf("round-trip HR len=%d want %d err=%v", len(roundHR), len(wantHR), err)
	}
}

func TestMigrateDryRun(t *testing.T) {
	dir := t.TempDir()
	cfg := config.StorageConfig{
		Driver:            config.StorageDriverFile,
		ResolvedLocation:  dir,
		ResolvedBBoltPath: filepath.Join(dir, "grom.db"),
	}
	backend, err := storage.Open(cfg)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := backend.Users().Create("alice", "Alice", "alice@example.com", "password123"); err != nil {
		t.Fatal(err)
	}
	_ = backend.Close()

	result, err := migrate.Run(migrate.Options{
		From:   config.StorageDriverFile,
		To:     config.StorageDriverBBolt,
		Config: cfg,
		DryRun: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	if result.Users != 1 {
		t.Fatalf("dry-run users=%d", result.Users)
	}
	if _, err := os.Stat(filepath.Join(dir, "grom.db")); err == nil {
		t.Fatal("dry-run should not create bbolt db")
	}
}
