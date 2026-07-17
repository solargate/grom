package users_test

import (
	"testing"

	"github.com/solargate/grom/internal/storage/file"
)

func newTestStore(t *testing.T) *file.UsersStore {
	t.Helper()
	store, err := file.NewUsersStore(t.TempDir())
	if err != nil {
		t.Fatalf("NewUsersStore() error = %v", err)
	}
	return store
}

func TestSetLastEquipmentForSport(t *testing.T) {
	store := newTestStore(t)

	user, err := store.Create("athlete", "Athlete", "athlete@example.com", "password123")
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	ids := []string{"eq-1", "eq-2"}
	if err := store.SetLastEquipmentForSport(user.ID, "Run", ids); err != nil {
		t.Fatalf("SetLastEquipmentForSport() error = %v", err)
	}

	loaded, err := store.FindByID(user.ID)
	if err != nil {
		t.Fatalf("FindByID() error = %v", err)
	}
	got := loaded.LastEquipmentBySport["Run"]
	if len(got) != 2 || got[0] != "eq-1" || got[1] != "eq-2" {
		t.Fatalf("unexpected last equipment: %#v", got)
	}
}

func TestRemoveEquipmentFromLastSets(t *testing.T) {
	store := newTestStore(t)

	user, err := store.Create("athlete", "Athlete", "athlete@example.com", "password123")
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	if err := store.SetLastEquipmentForSport(user.ID, "Run", []string{"eq-1", "eq-2"}); err != nil {
		t.Fatalf("SetLastEquipmentForSport() error = %v", err)
	}
	if err := store.SetLastEquipmentForSport(user.ID, "Ride", []string{"eq-2", "eq-3"}); err != nil {
		t.Fatalf("SetLastEquipmentForSport() error = %v", err)
	}

	if err := store.RemoveEquipmentFromLastSets(user.ID, "eq-2"); err != nil {
		t.Fatalf("RemoveEquipmentFromLastSets() error = %v", err)
	}

	loaded, err := store.FindByID(user.ID)
	if err != nil {
		t.Fatalf("FindByID() error = %v", err)
	}
	if got := loaded.LastEquipmentBySport["Run"]; len(got) != 1 || got[0] != "eq-1" {
		t.Fatalf("unexpected Run equipment: %#v", got)
	}
	if got := loaded.LastEquipmentBySport["Ride"]; len(got) != 1 || got[0] != "eq-3" {
		t.Fatalf("unexpected Ride equipment: %#v", got)
	}
}

func TestSetLastEquipmentForSportClearsEmpty(t *testing.T) {
	store := newTestStore(t)
	user, err := store.Create("athlete", "Athlete", "athlete@example.com", "password123")
	if err != nil {
		t.Fatal(err)
	}
	if err := store.SetLastEquipmentForSport(user.ID, "Run", []string{"eq-1"}); err != nil {
		t.Fatal(err)
	}
	if err := store.SetLastEquipmentForSport(user.ID, "Run", nil); err != nil {
		t.Fatal(err)
	}
	loaded, err := store.FindByID(user.ID)
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := loaded.LastEquipmentBySport["Run"]; ok {
		t.Fatalf("expected Run key removed, got %#v", loaded.LastEquipmentBySport)
	}
}
