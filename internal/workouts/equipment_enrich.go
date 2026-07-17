package workouts

import "github.com/solargate/grom/internal/equipment"

// EquipmentCatalog resolves equipment metadata for workout read enrichment.
type EquipmentCatalog interface {
	FindByIDs(nickname string, ids []string) ([]equipment.Equipment, error)
}

// ApplyEquipmentCatalog replaces Name/Type on workout equipment items from the
// catalog when the id is present. Missing catalog entries keep the stored snapshot.
func ApplyEquipmentCatalog(items []WorkoutEquipment, byID map[string]equipment.Equipment) {
	if len(items) == 0 || len(byID) == 0 {
		return
	}
	for i := range items {
		id := items[i].ID
		if id == "" {
			continue
		}
		cat, ok := byID[id]
		if !ok {
			continue
		}
		items[i].Name = cat.Name
		items[i].Type = cat.Type
	}
}

func collectEquipmentIDs(workouts []Workout) []string {
	seen := make(map[string]struct{})
	ids := make([]string, 0)
	for i := range workouts {
		for _, item := range workouts[i].Equipment {
			if item.ID == "" {
				continue
			}
			if _, ok := seen[item.ID]; ok {
				continue
			}
			seen[item.ID] = struct{}{}
			ids = append(ids, item.ID)
		}
	}
	return ids
}

func equipmentByID(items []equipment.Equipment) map[string]equipment.Equipment {
	byID := make(map[string]equipment.Equipment, len(items))
	for i := range items {
		byID[items[i].ID] = items[i]
	}
	return byID
}
