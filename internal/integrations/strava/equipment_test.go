package strava

import (
	"fmt"
	"strings"
	"testing"

	"github.com/solargate/grom/internal/equipment"
)

type fakeEquipmentStore struct {
	items []equipment.Equipment
}

func (s *fakeEquipmentStore) List(string) ([]equipment.Equipment, error) {
	return append([]equipment.Equipment(nil), s.items...), nil
}

func (s *fakeEquipmentStore) FindByID(string, string) (*equipment.Equipment, error) {
	return nil, equipment.ErrEquipmentNotFound
}

func (s *fakeEquipmentStore) FindByIDs(string, []string) ([]equipment.Equipment, error) {
	return nil, nil
}

func (s *fakeEquipmentStore) Create(_ string, item *equipment.Equipment) (*equipment.Equipment, error) {
	created := *item
	created.ID = fmt.Sprintf("eq-%d", len(s.items)+1)
	s.items = append(s.items, created)
	return &created, nil
}

func (s *fakeEquipmentStore) Update(_ string, item *equipment.Equipment) (*equipment.Equipment, error) {
	return item, nil
}

func (s *fakeEquipmentStore) Delete(string, string) error { return nil }

func (s *fakeEquipmentStore) SetDistance(string, string, float64) error { return nil }

// Column layout of a Russian Strava bulk export; the nickname column is empty
// whenever the gear has no custom name.
const (
	shoesCSV = "Название обуви,Марка обуви,Модель обуви,Модели обуви и их виды спорта по умолчанию\n" +
		"\"\",Nike,Lunar Forever 3,\"\"\n" +
		"Wave Inspire 16,Mizuno,WAVE INSPIRE 16,\"\"\n" +
		"Arishi ,New Balance,Arishi Trail v1 GTX,\"\"\n"
	bikesCSV = "Название велосипеда,Марка велосипеда,Модель велосипеда,Модели велосипедов и их виды спорта по умолчанию\n" +
		"GT Avalanche 29 Elite,GT,Avalanche 29 Elite,Заезд\n"
)

func newTestResolver(t *testing.T, store equipment.Repository) *EquipmentResolver {
	t.Helper()

	shoes, err := readGearCSVReader(strings.NewReader(shoesCSV), equipment.TypeShoes)
	if err != nil {
		t.Fatalf("read shoes: %v", err)
	}
	bikes, err := readGearCSVReader(strings.NewReader(bikesCSV), equipment.TypeBike)
	if err != nil {
		t.Fatalf("read bikes: %v", err)
	}

	gear := make(map[string]gearRecord)
	for _, rec := range append(bikes, shoes...) {
		for _, alias := range gearAliases(rec) {
			if _, exists := gear[alias]; !exists {
				gear[alias] = rec
			}
		}
	}

	return &EquipmentResolver{
		store:    store,
		nickname: "athlete",
		byName:   map[string]equipment.Equipment{},
		gear:     gear,
	}
}

func TestResolveMatchesGearDisplayNames(t *testing.T) {
	t.Parallel()

	tests := []struct {
		activityValue string
		wantType      string
		wantName      string
		wantBrand     string
		wantModel     string
	}{
		{"Nike Lunar Forever 3", equipment.TypeShoes, "Nike Lunar Forever 3", "Nike", "Lunar Forever 3"},
		{"Mizuno WAVE INSPIRE 16 Wave Inspire 16", equipment.TypeShoes, "Wave Inspire 16", "Mizuno", "WAVE INSPIRE 16"},
		{"New Balance Arishi Trail v1 GTX Arishi ", equipment.TypeShoes, "Arishi", "New Balance", "Arishi Trail v1 GTX"},
		{"GT Avalanche 29 Elite", equipment.TypeBike, "GT Avalanche 29 Elite", "GT", "Avalanche 29 Elite"},
	}

	for _, tt := range tests {
		t.Run(tt.activityValue, func(t *testing.T) {
			t.Parallel()

			resolver := newTestResolver(t, &fakeEquipmentStore{})
			items, err := resolver.Resolve(tt.activityValue)
			if err != nil {
				t.Fatalf("Resolve: %v", err)
			}
			if len(items) != 1 {
				t.Fatalf("got %d items, want 1", len(items))
			}
			got := items[0]
			if got.Type != tt.wantType {
				t.Errorf("Type = %q, want %q", got.Type, tt.wantType)
			}
			if got.Name != tt.wantName {
				t.Errorf("Name = %q, want %q", got.Name, tt.wantName)
			}
			if got.Brand != tt.wantBrand {
				t.Errorf("Brand = %q, want %q", got.Brand, tt.wantBrand)
			}
			if got.Model != tt.wantModel {
				t.Errorf("Model = %q, want %q", got.Model, tt.wantModel)
			}
		})
	}
}

func TestResolveCreatesGearOnce(t *testing.T) {
	t.Parallel()

	store := &fakeEquipmentStore{}
	resolver := newTestResolver(t, store)

	first, err := resolver.Resolve("Mizuno WAVE INSPIRE 16 Wave Inspire 16")
	if err != nil {
		t.Fatalf("Resolve: %v", err)
	}
	second, err := resolver.Resolve("Mizuno WAVE INSPIRE 16 Wave Inspire 16")
	if err != nil {
		t.Fatalf("Resolve again: %v", err)
	}
	if first[0].ID != second[0].ID {
		t.Fatalf("created duplicates: %q and %q", first[0].ID, second[0].ID)
	}
	if len(store.items) != 1 {
		t.Fatalf("store has %d items, want 1", len(store.items))
	}
}

func TestResolveReusesExistingEquipmentByGearAlias(t *testing.T) {
	t.Parallel()

	// A previous import stored the gear under its nickname only.
	store := &fakeEquipmentStore{items: []equipment.Equipment{{
		ID:    "eq-existing",
		Type:  equipment.TypeShoes,
		Name:  "Wave Inspire 16",
		Brand: "Mizuno",
		Model: "WAVE INSPIRE 16",
	}}}
	resolver := newTestResolver(t, store)
	resolver.byName[normalizeName("Wave Inspire 16")] = store.items[0]

	items, err := resolver.Resolve("Mizuno WAVE INSPIRE 16 Wave Inspire 16")
	if err != nil {
		t.Fatalf("Resolve: %v", err)
	}
	if items[0].ID != "eq-existing" {
		t.Fatalf("ID = %q, want eq-existing", items[0].ID)
	}
	if len(store.items) != 1 {
		t.Fatalf("store has %d items, want 1", len(store.items))
	}
}

func TestResolveFallsBackToOther(t *testing.T) {
	t.Parallel()

	resolver := newTestResolver(t, &fakeEquipmentStore{})
	items, err := resolver.Resolve("Unknown gear")
	if err != nil {
		t.Fatalf("Resolve: %v", err)
	}
	if items[0].Type != equipment.TypeOther {
		t.Fatalf("Type = %q, want %q", items[0].Type, equipment.TypeOther)
	}
}
