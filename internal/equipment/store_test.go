package equipment

import (
	"testing"
)

func TestStoreCreateListUpdateDelete(t *testing.T) {
	dir := t.TempDir()
	store := NewStore(dir)

	weight := 9.5
	created, err := store.Create("athlete", &Equipment{
		Type:     TypeBike,
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
	store := NewStore(t.TempDir())
	_, err := store.Create("athlete", &Equipment{
		Type: "invalid",
		Name: "Test",
	})
	if err == nil {
		t.Fatal("expected error for invalid type")
	}
}

func TestStoreFindByIDsPreservesOrder(t *testing.T) {
	dir := t.TempDir()
	store := NewStore(dir)

	first, err := store.Create("athlete", &Equipment{Type: TypeShoes, Name: "Shoes"})
	if err != nil {
		t.Fatalf("Create first: %v", err)
	}
	second, err := store.Create("athlete", &Equipment{Type: TypeBike, Name: "Bike"})
	if err != nil {
		t.Fatalf("Create second: %v", err)
	}

	found, err := store.FindByIDs("athlete", []string{second.ID, first.ID})
	if err != nil {
		t.Fatalf("FindByIDs() error = %v", err)
	}
	if len(found) != 2 {
		t.Fatalf("expected 2 items, got %d", len(found))
	}
}
