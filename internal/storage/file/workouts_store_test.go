package file

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/solargate/grom/internal/workouts"
	"gopkg.in/yaml.v3"
)

func TestWorkoutsStoreRejectsDuplicateIDWithDifferentDate(t *testing.T) {
	dir := t.TempDir()
	store := NewWorkoutsStore(dir)

	startDate1 := time.Date(2026, 7, 4, 10, 0, 0, 0, time.UTC)
	startDate2 := time.Date(2026, 7, 6, 11, 0, 0, 0, time.UTC)

	created, err := store.saveNewWorkout("athlete", &workouts.Workout{
		ID: "12345678", Name: "First", SportType: "Run", StartDate: startDate1,
	})
	if err != nil {
		t.Fatalf("saveNewWorkout: %v", err)
	}

	_, err = store.saveNewWorkout("athlete", &workouts.Workout{
		ID: created.ID, Name: "Duplicate ID", SportType: "Run", StartDate: startDate2,
	})
	if err != workouts.ErrWorkoutExists {
		t.Fatalf("expected ErrWorkoutExists, got %v", err)
	}
}

func TestWorkoutsStoreRejectsDuplicateIDAcrossUsers(t *testing.T) {
	store := NewWorkoutsStore(t.TempDir())
	startDate := time.Date(2026, 7, 4, 10, 0, 0, 0, time.UTC)

	created, err := store.saveNewWorkout("alice", &workouts.Workout{
		ID: "12345678", Name: "Alice run", SportType: "Run", StartDate: startDate,
	})
	if err != nil {
		t.Fatalf("saveNewWorkout alice: %v", err)
	}

	_, err = store.saveNewWorkout("bob", &workouts.Workout{
		ID: created.ID, Name: "Bob run", SportType: "Run", StartDate: startDate,
	})
	if err != workouts.ErrWorkoutExists {
		t.Fatalf("expected ErrWorkoutExists across users, got %v", err)
	}
}

func TestWorkoutsStoreCreateFailsIfWorkoutDirExists(t *testing.T) {
	dir := t.TempDir()
	store := NewWorkoutsStore(dir)

	startDate := time.Date(2026, 7, 5, 14, 30, 0, 0, time.UTC)
	created, err := store.saveNewWorkout("solarwind", &workouts.Workout{
		ID: "87654321", Name: "First", SportType: "Run", StartDate: startDate,
	})
	if err != nil {
		t.Fatalf("saveNewWorkout: %v", err)
	}

	_, err = store.saveNewWorkout("solarwind", &workouts.Workout{
		ID: created.ID, Name: "Duplicate", SportType: "Run", StartDate: startDate,
	})
	if err != workouts.ErrWorkoutExists {
		t.Fatalf("expected ErrWorkoutExists, got %v", err)
	}
}

func TestWorkoutsStoreCreateGetListDelete(t *testing.T) {
	store := NewWorkoutsStore(t.TempDir())

	created, err := store.Create("alice", &workouts.Workout{
		Name:      "Morning",
		SportType: "Run",
		StartDate: mustTime(t, "2026-07-08T10:00:00Z"),
	})
	if err != nil {
		t.Fatal(err)
	}
	if created.ID == "" || created.Device != workouts.DeviceGrom {
		t.Fatalf("unexpected create: %#v", created)
	}

	got, err := store.Get("alice", created.ID)
	if err != nil {
		t.Fatal(err)
	}
	if got.Name != "Morning" {
		t.Fatalf("get name = %q", got.Name)
	}

	dirName, err := store.WorkoutDirName("alice", created.ID)
	if err != nil {
		t.Fatal(err)
	}
	if dirName == "" || !strings.Contains(dirName, created.ID) {
		t.Fatalf("dirName = %q", dirName)
	}

	later, err := store.Create("alice", &workouts.Workout{
		Name:      "Evening",
		SportType: "Run",
		StartDate: mustTime(t, "2026-07-09T18:00:00Z"),
	})
	if err != nil {
		t.Fatal(err)
	}
	list, err := store.List("alice")
	if err != nil {
		t.Fatal(err)
	}
	if len(list) != 2 || list[0].ID != later.ID {
		t.Fatalf("list sort: %#v", list)
	}

	if err := store.Delete("alice", created.ID); err != nil {
		t.Fatal(err)
	}
	if _, err := store.Get("alice", created.ID); !errors.Is(err, workouts.ErrWorkoutNotFound) {
		t.Fatalf("get after delete: %v", err)
	}
	if err := store.Delete("alice", created.ID); !errors.Is(err, workouts.ErrWorkoutNotFound) {
		t.Fatalf("second delete: %v", err)
	}
}

func TestWorkoutsStoreUpdateRenamesDirOnStartDateChange(t *testing.T) {
	dir := t.TempDir()
	store := NewWorkoutsStore(dir)

	created, err := store.Create("alice", &workouts.Workout{
		Name:      "Morning",
		SportType: "Run",
		StartDate: mustTime(t, "2026-07-08T10:00:00Z"),
		Distance:  5000,
	})
	if err != nil {
		t.Fatal(err)
	}
	oldDirName, err := store.WorkoutDirName("alice", created.ID)
	if err != nil {
		t.Fatal(err)
	}

	marker := filepath.Join(dir, "users", "alice", "workouts", oldDirName, "track.gpx")
	if err := os.WriteFile(marker, []byte("gpx"), 0600); err != nil {
		t.Fatal(err)
	}

	created.Name = "Evening"
	created.StartDate = mustTime(t, "2026-07-09T18:00:00Z")
	updated, err := store.Update("alice", created)
	if err != nil {
		t.Fatalf("Update: %v", err)
	}
	if updated.Name != "Evening" {
		t.Fatalf("name = %q", updated.Name)
	}

	newDirName, err := store.WorkoutDirName("alice", created.ID)
	if err != nil {
		t.Fatal(err)
	}
	if newDirName == oldDirName {
		t.Fatalf("expected renamed dir, still %q", newDirName)
	}
	if _, err := os.Stat(filepath.Join(dir, "users", "alice", "workouts", newDirName, "track.gpx")); err != nil {
		t.Fatalf("track should move with dir: %v", err)
	}
	if _, err := os.Stat(filepath.Join(dir, "users", "alice", "workouts", oldDirName)); !os.IsNotExist(err) {
		t.Fatalf("old dir should be gone, err=%v", err)
	}

	got, err := store.Get("alice", created.ID)
	if err != nil {
		t.Fatal(err)
	}
	if got.Name != "Evening" || !got.StartDate.Equal(mustTime(t, "2026-07-09T18:00:00Z")) {
		t.Fatalf("get after update: %#v", got)
	}
}

func TestWorkoutsStoreBeginCreateCleanupAndWriteMetadata(t *testing.T) {
	dir := t.TempDir()
	store := NewWorkoutsStore(dir)

	created, dirName, cleanup, err := store.BeginCreate("alice", &workouts.Workout{
		Name:      "Draft",
		SportType: "Run",
		StartDate: mustTime(t, "2026-07-08T10:00:00Z"),
	})
	if err != nil {
		t.Fatal(err)
	}
	workoutPath := filepath.Join(dir, "users", "alice", "workouts", dirName)
	if _, err := os.Stat(workoutPath); err != nil {
		t.Fatalf("expected workout dir: %v", err)
	}
	cleanup()
	if _, err := os.Stat(workoutPath); !os.IsNotExist(err) {
		t.Fatalf("cleanup should remove dir, err=%v", err)
	}

	created, dirName, cleanup, err = store.BeginCreate("alice", &workouts.Workout{
		Name:      "Tracked",
		SportType: "Run",
		StartDate: mustTime(t, "2026-07-08T11:00:00Z"),
	})
	if err != nil {
		t.Fatal(err)
	}
	defer cleanup()

	created.Track = "track.gpx"
	created.StravaActivityID = "strava-42"
	if err := store.WriteMetadata("alice", created); err != nil {
		t.Fatal(err)
	}
	got, err := store.Get("alice", created.ID)
	if err != nil {
		t.Fatal(err)
	}
	if got.Track != "track.gpx" {
		t.Fatalf("track = %q", got.Track)
	}

	ok, err := store.HasStravaActivityID("alice", "strava-42")
	if err != nil || !ok {
		t.Fatalf("HasStravaActivityID = %v err=%v", ok, err)
	}
	ok, err = store.HasStravaActivityID("alice", "")
	if err != nil || ok {
		t.Fatalf("empty strava id should be false: %v %v", ok, err)
	}
	_ = dirName
}

func TestWorkoutsStoreRecoversIDFromDirName(t *testing.T) {
	dir := t.TempDir()
	store := NewWorkoutsStore(dir)
	start := mustTime(t, "2026-07-08T10:00:00Z")
	id := "11223344"
	base := workoutBaseName(start, id)
	workoutDirPath := filepath.Join(dir, "users", "alice", "workouts", base)
	if err := os.MkdirAll(workoutDirPath, 0700); err != nil {
		t.Fatal(err)
	}
	payload := workouts.Workout{
		Name:      "Legacy",
		SportType: "Run",
		StartDate: start,
	}
	data, err := yaml.Marshal(&payload)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(workoutDirPath, base+".yaml"), data, 0600); err != nil {
		t.Fatal(err)
	}

	got, err := store.Get("alice", id)
	if err != nil {
		t.Fatal(err)
	}
	if got.ID != id {
		t.Fatalf("recovered id = %q, want %q", got.ID, id)
	}
}

func TestWorkoutsStoreListMissingDir(t *testing.T) {
	store := NewWorkoutsStore(t.TempDir())
	list, err := store.List("nobody")
	if err != nil || list != nil {
		t.Fatalf("List missing = %#v err=%v", list, err)
	}
}
