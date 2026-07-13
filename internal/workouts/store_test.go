package workouts

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestStoreCreateAndList(t *testing.T) {
	dir := t.TempDir()
	store := NewStore(dir)

	startDate := time.Date(2026, 7, 5, 14, 30, 0, 0, time.UTC)
	created, err := store.Create("solarwind", &Workout{
		Name:            "Morning run",
		Description:     "Easy session",
		SportType:       "Run",
		StartDate:       startDate,
		DurationSeconds: 3600,
		Distance:        5200,
	})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	if created.ID == "" {
		t.Fatal("expected generated id")
	}
	if len(created.ID) != workoutIDLength {
		t.Fatalf("expected id length %d, got %d (%q)", workoutIDLength, len(created.ID), created.ID)
	}
	if strings.Trim(created.ID, "0123456789") != "" {
		t.Fatalf("expected numeric id, got %q", created.ID)
	}

	expectedBase := "2026-07-05T143000Z-" + created.ID
	expectedDir := filepath.Join(dir, "users", "solarwind", "workouts", expectedBase)
	if info, err := os.Stat(expectedDir); err != nil || !info.IsDir() {
		t.Fatalf("expected workout dir %q: %v", expectedDir, err)
	}

	expectedFile := filepath.Join(expectedDir, expectedBase+".yaml")
	if _, err := os.Stat(expectedFile); err != nil {
		t.Fatalf("expected workout file %q: %v", expectedFile, err)
	}

	workouts, err := store.List("solarwind")
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
	if len(workouts) != 1 {
		t.Fatalf("expected 1 workout, got %d", len(workouts))
	}
	if workouts[0].Name != "Morning run" {
		t.Fatalf("expected name Morning run, got %q", workouts[0].Name)
	}
	if workouts[0].Distance != 5200 {
		t.Fatalf("expected distance 5200, got %v", workouts[0].Distance)
	}
	if workouts[0].Device != DeviceGrom {
		t.Fatalf("expected device %q, got %q", DeviceGrom, workouts[0].Device)
	}
}

func TestStoreRejectsInvalidSportType(t *testing.T) {
	store := NewStore(t.TempDir())
	_, err := store.Create("solarwind", &Workout{
		Name:      "Test",
		SportType: "InvalidType",
		StartDate: time.Now().UTC(),
	})
	if err != ErrInvalidSportType {
		t.Fatalf("expected ErrInvalidSportType, got %v", err)
	}
}

func TestStoreListSortedDescending(t *testing.T) {
	dir := t.TempDir()
	store := NewStore(dir)

	older := time.Date(2026, 7, 4, 10, 0, 0, 0, time.UTC)
	newer := time.Date(2026, 7, 5, 10, 0, 0, 0, time.UTC)

	if _, err := store.Create("athlete", &Workout{
		Name: "Older", SportType: "Run", StartDate: older,
	}); err != nil {
		t.Fatalf("Create older: %v", err)
	}
	if _, err := store.Create("athlete", &Workout{
		Name: "Newer", SportType: "Ride", StartDate: newer,
	}); err != nil {
		t.Fatalf("Create newer: %v", err)
	}

	workouts, err := store.List("athlete")
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
	if len(workouts) != 2 {
		t.Fatalf("expected 2 workouts, got %d", len(workouts))
	}
	if workouts[0].Name != "Newer" {
		t.Fatalf("expected newer workout first, got %q", workouts[0].Name)
	}
}

func TestStoreCreateFailsIfWorkoutDirExists(t *testing.T) {
	dir := t.TempDir()
	store := NewStore(dir)

	startDate := time.Date(2026, 7, 5, 14, 30, 0, 0, time.UTC)
	created, err := store.Create("solarwind", &Workout{
		Name: "First", SportType: "Run", StartDate: startDate,
	})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	_, err = store.saveWorkout("solarwind", &Workout{
		ID:        created.ID,
		Name:      "Duplicate",
		SportType: "Run",
		StartDate: startDate,
	})
	if err != ErrWorkoutExists {
		t.Fatalf("expected ErrWorkoutExists, got %v", err)
	}
}

func TestStoreRemoveEquipmentFromAll(t *testing.T) {
	dir := t.TempDir()
	store := NewStore(dir)

	startDate := time.Date(2026, 7, 5, 14, 30, 0, 0, time.UTC)
	created, err := store.Create("athlete", &Workout{
		Name:      "Morning run",
		SportType: "Run",
		StartDate: startDate,
		Equipment: []WorkoutEquipment{
			{ID: "eq-1", Name: "Shoes"},
			{ID: "eq-2", Name: "Watch"},
		},
	})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	if err := store.RemoveEquipmentFromAll("athlete", "eq-1"); err != nil {
		t.Fatalf("RemoveEquipmentFromAll() error = %v", err)
	}

	workouts, err := store.List("athlete")
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
	if len(workouts) != 1 {
		t.Fatalf("expected 1 workout, got %d", len(workouts))
	}
	if len(workouts[0].Equipment) != 1 {
		t.Fatalf("expected 1 equipment item, got %d", len(workouts[0].Equipment))
	}
	if workouts[0].Equipment[0].ID != "eq-2" {
		t.Fatalf("expected eq-2 to remain, got %q", workouts[0].Equipment[0].ID)
	}
	if workouts[0].ID != created.ID {
		t.Fatalf("expected same workout id")
	}
}

func TestStoreDelete(t *testing.T) {
	dir := t.TempDir()
	store := NewStore(dir)

	startDate := time.Date(2026, 7, 5, 14, 30, 0, 0, time.UTC)
	created, err := store.Create("solarwind", &Workout{
		Name: "Morning run", SportType: "Run", StartDate: startDate,
	})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	expectedBase := "2026-07-05T143000Z-" + created.ID
	expectedDir := filepath.Join(dir, "users", "solarwind", "workouts", expectedBase)
	if err := os.WriteFile(filepath.Join(expectedDir, "extra.txt"), []byte("x"), 0600); err != nil {
		t.Fatal(err)
	}

	if err := store.Delete("solarwind", created.ID); err != nil {
		t.Fatalf("Delete() error = %v", err)
	}
	if _, err := os.Stat(expectedDir); !os.IsNotExist(err) {
		t.Fatalf("expected workout dir removed, stat err = %v", err)
	}

	workouts, err := store.List("solarwind")
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
	if len(workouts) != 0 {
		t.Fatalf("expected 0 workouts, got %d", len(workouts))
	}

	if err := store.Delete("solarwind", created.ID); err != ErrWorkoutNotFound {
		t.Fatalf("expected ErrWorkoutNotFound on second delete, got %v", err)
	}
}
