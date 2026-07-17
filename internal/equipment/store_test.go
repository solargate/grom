package equipment_test

import (
	"testing"

	"github.com/solargate/grom/internal/equipment"
	"github.com/solargate/grom/internal/storage/file"
)

func newTestStore(t *testing.T) *file.EquipmentStore {
	t.Helper()
	return file.NewEquipmentStore(t.TempDir())
}

func TestStoreCreateListUpdateDelete(t *testing.T) {
	store := newTestStore(t)

	weight := 9.5
	created, err := store.Create("athlete", &equipment.Equipment{
		Type:     equipment.TypeBike,
		Name:     "Gravel bike",
		BikeType: "gravel",
		Brand:    "Canyon",
		WeightKg: &weight,
	})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	if created.ID == "" {
		t.Fatal("expected generated id")
	}

	items, err := store.List("athlete")
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(items))
	}

	created.Name = "Updated bike"
	updated, err := store.Update("athlete", created)
	if err != nil {
		t.Fatalf("Update() error = %v", err)
	}
	if updated.Name != "Updated bike" {
		t.Fatalf("expected updated name, got %q", updated.Name)
	}

	if err := store.Delete("athlete", created.ID); err != nil {
		t.Fatalf("Delete() error = %v", err)
	}
	items, err = store.List("athlete")
	if err != nil {
		t.Fatalf("List() after delete error = %v", err)
	}
	if len(items) != 0 {
		t.Fatalf("expected 0 items after delete, got %d", len(items))
	}
}

func TestStoreRejectsInvalidType(t *testing.T) {
	store := newTestStore(t)
	_, err := store.Create("athlete", &equipment.Equipment{
		Type: "invalid",
		Name: "Test",
	})
	if err == nil {
		t.Fatal("expected error for invalid type")
	}
}

func TestStoreFindByIDsReturnsMatchingItems(t *testing.T) {
	store := newTestStore(t)

	first, err := store.Create("athlete", &equipment.Equipment{Type: equipment.TypeShoes, Name: "Shoes"})
	if err != nil {
		t.Fatalf("Create first: %v", err)
	}
	second, err := store.Create("athlete", &equipment.Equipment{Type: equipment.TypeBike, Name: "Bike"})
	if err != nil {
		t.Fatalf("Create second: %v", err)
	}

	// Request reverse of creation order; implementation returns storage order.
	found, err := store.FindByIDs("athlete", []string{second.ID, first.ID})
	if err != nil {
		t.Fatalf("FindByIDs() error = %v", err)
	}
	if len(found) != 2 {
		t.Fatalf("expected 2 items, got %d", len(found))
	}
	got := map[string]bool{found[0].ID: true, found[1].ID: true}
	if !got[first.ID] || !got[second.ID] {
		t.Fatalf("FindByIDs() = %v, want ids %q and %q", []string{found[0].ID, found[1].ID}, first.ID, second.ID)
	}
	// Storage order: first created, then second.
	if found[0].ID != first.ID || found[1].ID != second.ID {
		t.Fatalf("expected storage order [%q, %q], got [%q, %q]", first.ID, second.ID, found[0].ID, found[1].ID)
	}
}
