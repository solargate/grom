package strava

import (
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/solargate/grom/internal/equipment"
	"github.com/solargate/grom/internal/workouts"
)

type gearRecord struct {
	Name  string
	Brand string
	Model string
	Type  string
}

type EquipmentResolver struct {
	store    equipment.Repository
	nickname string
	byName   map[string]equipment.Equipment
	gear     map[string]gearRecord
}

func NewEquipmentResolver(store equipment.Repository, nickname string, archive *Archive) (*EquipmentResolver, error) {
	items, err := store.List(nickname)
	if err != nil {
		return nil, err
	}
	byName := make(map[string]equipment.Equipment, len(items))
	for _, item := range items {
		byName[normalizeName(item.Name)] = item
	}

	gear := make(map[string]gearRecord)
	for _, source := range []struct {
		file     string
		gearType string
	}{
		{"bikes.csv", equipment.TypeBike},
		{"shoes.csv", equipment.TypeShoes},
	} {
		records, err := readGearCSVFromArchive(archive, source.file, source.gearType)
		if err != nil {
			return nil, err
		}
		for _, rec := range records {
			for _, alias := range gearAliases(rec) {
				if _, exists := gear[alias]; !exists {
					gear[alias] = rec
				}
			}
		}
	}

	return &EquipmentResolver{
		store:    store,
		nickname: nickname,
		byName:   byName,
		gear:     gear,
	}, nil
}

func readGearCSVFromArchive(archive *Archive, name, gearType string) ([]gearRecord, error) {
	data, err := archive.ReadFile(name)
	if err != nil {
		return nil, nil
	}
	return readGearCSVReader(strings.NewReader(string(data)), gearType)
}

func readGearCSV(path, gearType string) ([]gearRecord, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	return readGearCSVReader(file, gearType)
}

func readGearCSVReader(r io.Reader, gearType string) ([]gearRecord, error) {
	reader := csv.NewReader(r)
	reader.FieldsPerRecord = -1
	if _, err := reader.Read(); err != nil {
		return nil, fmt.Errorf("read gear header: %w", err)
	}

	var result []gearRecord
	for {
		record, err := reader.Read()
		if err != nil {
			break
		}
		if len(record) == 0 {
			continue
		}
		rec := gearRecord{Name: strings.TrimSpace(record[0]), Type: gearType}
		if len(record) > 1 {
			rec.Brand = strings.TrimSpace(record[1])
		}
		if len(record) > 2 {
			rec.Model = strings.TrimSpace(record[2])
		}
		// Strava leaves the nickname column empty when the gear has no custom name,
		// so a record is usable as long as brand or model is present.
		if rec.Name == "" && rec.Brand == "" && rec.Model == "" {
			continue
		}
		result = append(result, rec)
	}
	return result, nil
}

// gearAliases lists the names an activity row may use for the gear record.
// Strava writes the display name into activities.csv: brand, model and nickname
// joined together, while bikes.csv/shoes.csv keep those parts in separate columns.
func gearAliases(rec gearRecord) []string {
	candidates := []string{
		rec.Name,
		joinFields(rec.Brand, rec.Model),
		joinFields(rec.Brand, rec.Model, rec.Name),
		joinFields(rec.Brand, rec.Name),
		joinFields(rec.Model, rec.Name),
	}

	aliases := make([]string, 0, len(candidates))
	seen := make(map[string]struct{}, len(candidates))
	for _, candidate := range candidates {
		alias := normalizeName(candidate)
		if alias == "" {
			continue
		}
		if _, ok := seen[alias]; ok {
			continue
		}
		seen[alias] = struct{}{}
		aliases = append(aliases, alias)
	}
	return aliases
}

func joinFields(parts ...string) string {
	kept := make([]string, 0, len(parts))
	for _, part := range parts {
		if strings.TrimSpace(part) != "" {
			kept = append(kept, strings.TrimSpace(part))
		}
	}
	return strings.Join(kept, " ")
}

// gearDisplayName picks the name for the equipment created from a gear record,
// falling back to the raw activity value when the record carries no usable name.
func gearDisplayName(rec gearRecord, raw string) string {
	if name := joinFields(rec.Name); name != "" {
		return name
	}
	if name := joinFields(rec.Brand, rec.Model); name != "" {
		return name
	}
	return strings.TrimSpace(raw)
}

func normalizeName(name string) string {
	return strings.Join(strings.Fields(strings.ToLower(name)), " ")
}

func (r *EquipmentResolver) Resolve(name string) ([]equipment.Equipment, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return nil, nil
	}

	key := normalizeName(name)
	if item, ok := r.byName[key]; ok {
		return []equipment.Equipment{item}, nil
	}

	if rec, ok := r.gear[key]; ok {
		aliases := gearAliases(rec)
		for _, alias := range aliases {
			if item, ok := r.byName[alias]; ok {
				r.byName[key] = item
				return []equipment.Equipment{item}, nil
			}
		}
		created, err := r.store.Create(r.nickname, &equipment.Equipment{
			Type:  rec.Type,
			Name:  gearDisplayName(rec, name),
			Brand: rec.Brand,
			Model: rec.Model,
		})
		if err != nil {
			return nil, err
		}
		r.remember(*created, append(aliases, key)...)
		return []equipment.Equipment{*created}, nil
	}

	created, err := r.store.Create(r.nickname, &equipment.Equipment{
		Type: equipment.TypeOther,
		Name: name,
	})
	if err != nil {
		return nil, err
	}
	r.remember(*created, key)
	return []equipment.Equipment{*created}, nil
}

func (r *EquipmentResolver) remember(item equipment.Equipment, keys ...string) {
	r.byName[normalizeName(item.Name)] = item
	for _, key := range keys {
		if key == "" {
			continue
		}
		r.byName[key] = item
	}
}

func toWorkoutEquipment(items []equipment.Equipment) []workouts.WorkoutEquipment {
	result := make([]workouts.WorkoutEquipment, 0, len(items))
	for _, item := range items {
		result = append(result, workouts.WorkoutEquipment{
			ID:   item.ID,
			Name: item.Name,
			Type: item.Type,
		})
	}
	return result
}
