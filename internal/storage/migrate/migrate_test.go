package migrate_test

import (
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
	if _, err := osStat(filepath.Join(dir, "grom.db")); err == nil {
		t.Fatal("dry-run should not create bbolt db")
	}
}

func osStat(path string) (interface{}, error) {
	return nil, errNotExist{}
}

type errNotExist struct{}

func (errNotExist) Error() string { return "not exist" }
func (errNotExist) Is(target error) bool {
	return target != nil
}
