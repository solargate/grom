package workouts

import (
	"testing"

	"github.com/solargate/grom/internal/equipment"
)

func TestApplyEquipmentCatalogUpdatesNameAndType(t *testing.T) {
	items := []WorkoutEquipment{
		{ID: "eq-1", Name: "Old shoes", Type: equipment.TypeShoes},
		{ID: "eq-2", Name: "Old bike", Type: equipment.TypeBike},
	}
	byID := map[string]equipment.Equipment{
		"eq-1": {ID: "eq-1", Name: "New shoes", Type: equipment.TypeShoes},
		"eq-2": {ID: "eq-2", Name: "New bike", Type: equipment.TypeOther},
	}

	ApplyEquipmentCatalog(items, byID)

	if items[0].Name != "New shoes" {
		t.Fatalf("eq-1 name = %q, want New shoes", items[0].Name)
	}
	if items[1].Name != "New bike" {
		t.Fatalf("eq-2 name = %q, want New bike", items[1].Name)
	}
	if items[1].Type != equipment.TypeOther {
		t.Fatalf("eq-2 type = %q, want %q", items[1].Type, equipment.TypeOther)
	}
}

func TestApplyEquipmentCatalogKeepsSnapshotWhenMissing(t *testing.T) {
	items := []WorkoutEquipment{
		{ID: "eq-gone", Name: "Deleted gear", Type: equipment.TypeOther},
		{ID: "", Name: "No id", Type: equipment.TypeShoes},
	}

	ApplyEquipmentCatalog(items, map[string]equipment.Equipment{
		"eq-other": {ID: "eq-other", Name: "Other", Type: equipment.TypeBike},
	})

	if items[0].Name != "Deleted gear" || items[0].Type != equipment.TypeOther {
		t.Fatalf("missing id should keep snapshot, got %+v", items[0])
	}
	if items[1].Name != "No id" {
		t.Fatalf("empty id should keep snapshot, got %+v", items[1])
	}
}

func TestApplyEquipmentCatalogNoopOnEmpty(t *testing.T) {
	ApplyEquipmentCatalog(nil, map[string]equipment.Equipment{"eq-1": {ID: "eq-1", Name: "X"}})
	ApplyEquipmentCatalog([]WorkoutEquipment{{ID: "eq-1", Name: "Old"}}, nil)
}

func TestCollectEquipmentIDsDedupes(t *testing.T) {
	ids := collectEquipmentIDs([]Workout{
		{Equipment: []WorkoutEquipment{{ID: "a"}, {ID: "b"}, {ID: "a"}, {ID: ""}}},
		{Equipment: []WorkoutEquipment{{ID: "b"}, {ID: "c"}}},
	})
	if len(ids) != 3 {
		t.Fatalf("expected 3 unique ids, got %#v", ids)
	}
}
