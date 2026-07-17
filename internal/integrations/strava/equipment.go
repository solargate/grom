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
}

type EquipmentResolver struct {
	store    equipment.Repository
	nickname string
	byName   map[string]equipment.Equipment
	bikes    map[string]gearRecord
	shoes    map[string]gearRecord
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

	bikes, err := readGearCSVFromArchive(archive, "bikes.csv")
	if err != nil {
		return nil, err
	}
	shoes, err := readGearCSVFromArchive(archive, "shoes.csv")
	if err != nil {
		return nil, err
	}

	return &EquipmentResolver{
		store:    store,
		nickname: nickname,
		byName:   byName,
		bikes:    bikes,
		shoes:    shoes,
	}, nil
}

func readGearCSVFromArchive(archive *Archive, name string) (map[string]gearRecord, error) {
	data, err := archive.ReadFile(name)
	if err != nil {
		return map[string]gearRecord{}, nil
	}
	return readGearCSVReader(strings.NewReader(string(data)))
}

func readGearCSV(path string) (map[string]gearRecord, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	return readGearCSVReader(file)
}

func readGearCSVReader(r io.Reader) (map[string]gearRecord, error) {
	reader := csv.NewReader(r)
	reader.FieldsPerRecord = -1
	if _, err := reader.Read(); err != nil {
		return nil, fmt.Errorf("read gear header: %w", err)
	}

	result := make(map[string]gearRecord)
	for {
		record, err := reader.Read()
		if err != nil {
			break
		}
		if len(record) == 0 {
			continue
		}
		name := strings.TrimSpace(record[0])
		if name == "" {
			continue
		}
		rec := gearRecord{Name: name}
		if len(record) > 1 {
			rec.Brand = strings.TrimSpace(record[1])
		}
		if len(record) > 2 {
			rec.Model = strings.TrimSpace(record[2])
		}
		result[normalizeName(name)] = rec
	}
	return result, nil
}

func normalizeName(name string) string {
	return strings.ToLower(strings.TrimSpace(name))
}

func (r *EquipmentResolver) Resolve(name string) ([]equipment.Equipment, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return nil, nil
	}

	if item, ok := r.byName[normalizeName(name)]; ok {
		return []equipment.Equipment{item}, nil
	}

	if bike, ok := r.bikes[normalizeName(name)]; ok {
		created, err := r.store.Create(r.nickname, &equipment.Equipment{
			Type:  equipment.TypeBike,
			Name:  bike.Name,
			Brand: bike.Brand,
			Model: bike.Model,
		})
		if err != nil {
			return nil, err
		}
		r.byName[normalizeName(created.Name)] = *created
		return []equipment.Equipment{*created}, nil
	}

	if shoe, ok := r.shoes[normalizeName(name)]; ok {
		created, err := r.store.Create(r.nickname, &equipment.Equipment{
			Type:  equipment.TypeShoes,
			Name:  shoe.Name,
			Brand: shoe.Brand,
			Model: shoe.Model,
		})
		if err != nil {
			return nil, err
		}
		r.byName[normalizeName(created.Name)] = *created
		return []equipment.Equipment{*created}, nil
	}

	created, err := r.store.Create(r.nickname, &equipment.Equipment{
		Type: equipment.TypeOther,
		Name: name,
	})
	if err != nil {
		return nil, err
	}
	r.byName[normalizeName(created.Name)] = *created
	return []equipment.Equipment{*created}, nil
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
