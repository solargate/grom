package file

import (
	"errors"
	"testing"

	"github.com/solargate/grom/internal/equipment"
)

func TestEquipmentStoreIsolationAndPersistence(t *testing.T) {
	dir := t.TempDir()
	store := NewEquipmentStore(dir)

	created, err := store.Create("alice", &equipment.Equipment{
		Type: equipment.TypeShoes,
		Name: "Trail",
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if _, err := store.Create("bob", &equipment.Equipment{
		Type: equipment.TypeBike,
		Name: "Road",
		BikeType: "road",
	}); err != nil {
		t.Fatal(err)
	}

	aliceItems, err := store.List("alice")
	if err != nil {
		t.Fatal(err)
	}
	if len(aliceItems) != 1 || aliceItems[0].ID != created.ID {
		t.Fatalf("alice items: %#v", aliceItems)
	}
	bobItems, err := store.List("bob")
	if err != nil {
		t.Fatal(err)
	}
	if len(bobItems) != 1 || bobItems[0].Name != "Road" {
		t.Fatalf("bob items: %#v", bobItems)
	}

	reloaded := NewEquipmentStore(dir)
	got, err := reloaded.FindByID("alice", created.ID)
	if err != nil {
		t.Fatal(err)
	}
	if got.Name != "Trail" {
		t.Fatalf("reloaded name = %q", got.Name)
	}

	if _, err := store.FindByID("alice", "missing"); !errors.Is(err, equipment.ErrEquipmentNotFound) {
		t.Fatalf("missing find: %v", err)
	}
	if err := store.Delete("alice", "missing"); !errors.Is(err, equipment.ErrEquipmentNotFound) {
		t.Fatalf("missing delete: %v", err)
	}
	if _, err := store.Update("alice", &equipment.Equipment{ID: "missing", Type: equipment.TypeOther, Name: "X"}); !errors.Is(err, equipment.ErrEquipmentNotFound) {
		t.Fatalf("missing update: %v", err)
	}
}

func TestEquipmentStoreUpdatePreservesDistance(t *testing.T) {
	dir := t.TempDir()
	store := NewEquipmentStore(dir)

	created, err := store.Create("alice", &equipment.Equipment{
		Type: equipment.TypeBike,
		Name: "Road",
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if err := store.SetDistance("alice", created.ID, 4200); err != nil {
		t.Fatalf("SetDistance: %v", err)
	}

	updated, err := store.Update("alice", &equipment.Equipment{
		ID:   created.ID,
		Type: equipment.TypeBike,
		Name: "Gravel",
	})
	if err != nil {
		t.Fatalf("Update: %v", err)
	}
	if updated.Distance != 4200 {
		t.Fatalf("distance = %v, want 4200", updated.Distance)
	}
}
