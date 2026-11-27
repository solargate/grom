package migrate

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/solargate/grom/internal/config"
	"github.com/solargate/grom/internal/storage"
	"github.com/solargate/grom/internal/users"
	"github.com/solargate/grom/internal/workouts"
)

func TestVerifyDetectsDifferentRecordsWithEqualCounts(t *testing.T) {
	open := func(driver config.StorageDriver) config.StorageConfig {
		dir := t.TempDir()
		return config.StorageConfig{
			Driver: driver, ResolvedLocation: dir, ResolvedBBoltPath: filepath.Join(dir, "grom.db"),
		}
	}
	fileCfg, boltCfg := open(config.StorageDriverFile), open(config.StorageDriverBBolt)
	file, err := openDriver(fileCfg.Driver, fileCfg)
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	bolt, err := openDriver(boltCfg.Driver, boltCfg)
	if err != nil {
		t.Fatal(err)
	}
	defer bolt.Close()

	user, err := file.Users().Create("alice", "Alice", "alice@example.test", "password123")
	if err != nil {
		t.Fatal(err)
	}
	if err := importUser(bolt, *user); err != nil {
		t.Fatal(err)
	}
	item := workouts.Workout{ID: "42", Name: "First", SportType: "Run", StartDate: time.Now().UTC()}
	if err := importWorkout(file, "alice", &item); err != nil {
		t.Fatal(err)
	}
	item.Name = "Different"
	if err := importWorkout(bolt, "alice", &item); err != nil {
		t.Fatal(err)
	}
	for _, backend := range []storage.Backend{file, bolt} {
		if _, err := backend.GetProfileWithWorkoutTotal(user.ID); err != nil {
			t.Fatal(err)
		}
	}
	for _, err := range []error{
		verifyContent(file, bolt, fileCfg.ResolvedLocation, ""),
		verifyContent(bolt, file, fileCfg.ResolvedLocation, ""),
	} {
		if err == nil || !strings.Contains(err.Error(), "different workout record") {
			t.Fatalf("expected workout verification mismatch; got %v", err)
		}
	}
	profile, err := bolt.Users().GetProfile(user.ID)
	if err != nil {
		t.Fatal(err)
	}
	(*profile.Workouts)[item.StartDate.Format(time.DateOnly)][0].DurationSeconds = 123
	if err := bolt.Users().PutProfile(user.ID, *profile); err != nil {
		t.Fatal(err)
	}
	if err := verifyContent(file, bolt, fileCfg.ResolvedLocation, ""); err == nil ||
		!strings.Contains(err.Error(), "workout history differs") {
		t.Fatalf("expected workout history verification mismatch; got %v", err)
	}
	wrongTotal := 0
	if err := bolt.Users().PutProfile(user.ID, users.Profile{TotalWorkouts: &wrongTotal}); err != nil {
		t.Fatal(err)
	}
	if err := verifyContent(file, bolt, fileCfg.ResolvedLocation, ""); err == nil ||
		!strings.Contains(err.Error(), "workout total differs") {
		t.Fatalf("expected workout total verification mismatch; got %v", err)
	}
}

func TestFailedForceMigrationPreservesExistingDatabase(t *testing.T) {
	dir := t.TempDir()
	cfg := config.StorageConfig{ResolvedLocation: dir, ResolvedBBoltPath: filepath.Join(dir, "grom.db")}
	original := []byte("previous database")
	if err := os.WriteFile(cfg.ResolvedBBoltPath, original, 0600); err != nil {
		t.Fatal(err)
	}
	source, err := openDriver(config.StorageDriverFile, cfg)
	if err != nil {
		t.Fatal(err)
	}
	for _, nickname := range []string{"alice", "bob"} {
		if _, err := source.Users().Create(nickname, nickname, nickname+"@example.test", "password123"); err != nil {
			t.Fatal(err)
		}
	}
	// The second workout has the same ID as the first. bbolt rejects this
	// when importing into its instance-wide workout ID index.
	for _, nickname := range []string{"alice", "bob"} {
		item := workouts.Workout{ID: "42", Name: "Run", SportType: "Run", StartDate: time.Now().UTC()}
		if err := importWorkout(source, nickname, &item); err != nil {
			t.Fatal(err)
		}
	}
	_ = source.Close()

	if _, err := Run(Options{
		From: config.StorageDriverFile, To: config.StorageDriverBBolt, Config: cfg, Force: true, Verify: true,
	}); err == nil {
		t.Fatal("migration unexpectedly succeeded")
	}
	got, err := os.ReadFile(cfg.ResolvedBBoltPath)
	if err != nil || !bytes.Equal(got, original) {
		t.Fatalf("existing database changed, err=%v", err)
	}
	staged, err := filepath.Glob(filepath.Join(dir, ".grom-migrate-*.db"))
	if err != nil || len(staged) != 0 {
		t.Fatalf("staged database still exists: %v, err=%v", staged, err)
	}
}
