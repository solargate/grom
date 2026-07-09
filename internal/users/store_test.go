package users

import (
	"testing"
)

func TestSetLastEquipmentForSport(t *testing.T) {
	dir := t.TempDir()
	store, err := NewStore(dir)
	if err != nil {
		t.Fatalf("NewStore() error = %v", err)
	}

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
	dir := t.TempDir()
	store, err := NewStore(dir)
	if err != nil {
		t.Fatalf("NewStore() error = %v", err)
	}

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
