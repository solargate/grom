package file

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/google/uuid"
	"github.com/solargate/grom/internal/data"
	"github.com/solargate/grom/internal/equipment"
	"gopkg.in/yaml.v3"
)

const equipmentFileName = "equipment.yaml"

type equipmentFile struct {
	Equipment []equipment.Equipment `yaml:"equipment"`
}

type EquipmentStore struct {
	dataDir string
	mu      sync.Mutex
}

func NewEquipmentStore(dataDir string) *EquipmentStore {
	return &EquipmentStore{dataDir: dataDir}
}

func (s *EquipmentStore) userDir(nickname string) string {
	return data.UserDir(s.dataDir, nickname)
}

func (s *EquipmentStore) filePath(nickname string) string {
	return filepath.Join(s.userDir(nickname), equipmentFileName)
}

func (s *EquipmentStore) load(nickname string) ([]equipment.Equipment, error) {
	path := s.filePath(nickname)
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	var file equipmentFile
	if err := yaml.Unmarshal(data, &file); err != nil {
		return nil, fmt.Errorf("parse equipment file: %w", err)
	}
	return file.Equipment, nil
}

func (s *EquipmentStore) save(nickname string, items []equipment.Equipment) error {
	if err := os.MkdirAll(s.userDir(nickname), 0700); err != nil {
		return fmt.Errorf("create user dir: %w", err)
	}

	file := equipmentFile{Equipment: items}
	data, err := yaml.Marshal(&file)
	if err != nil {
		return err
	}

	path := s.filePath(nickname)
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, data, 0600); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}

func (s *EquipmentStore) List(nickname string) ([]equipment.Equipment, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.load(nickname)
}

func (s *EquipmentStore) FindByID(nickname, id string) (*equipment.Equipment, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	items, err := s.load(nickname)
	if err != nil {
		return nil, err
	}
	for i := range items {
		if items[i].ID == id {
			item := items[i]
			return &item, nil
		}
	}
	return nil, equipment.ErrEquipmentNotFound
}

func (s *EquipmentStore) FindByIDs(nickname string, ids []string) ([]equipment.Equipment, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	items, err := s.load(nickname)
	if err != nil {
		return nil, err
	}
	if len(ids) == 0 {
		return nil, nil
	}

	idSet := make(map[string]struct{}, len(ids))
	for _, id := range ids {
		idSet[id] = struct{}{}
	}

	result := make([]equipment.Equipment, 0, len(ids))
	for i := range items {
		if _, ok := idSet[items[i].ID]; ok {
			result = append(result, items[i])
		}
	}
	return result, nil
}

func (s *EquipmentStore) validate(item *equipment.Equipment) error {
	if item == nil {
		return equipment.ErrInvalidEquipment
	}
	if !equipment.IsValidType(item.Type) {
		return fmt.Errorf("%w: invalid type", equipment.ErrInvalidEquipment)
	}
	if strings.TrimSpace(item.Name) == "" {
		return fmt.Errorf("%w: name is required", equipment.ErrInvalidEquipment)
	}
	if item.Type == equipment.TypeBike && !equipment.IsValidBikeType(item.BikeType) {
		return fmt.Errorf("%w: invalid bike type", equipment.ErrInvalidEquipment)
	}
	if item.Type == equipment.TypeWater && !equipment.IsValidWaterType(item.WaterType) {
		return fmt.Errorf("%w: invalid water type", equipment.ErrInvalidEquipment)
	}
	if item.Type != equipment.TypeBike {
		item.BikeType = ""
	}
	if item.Type != equipment.TypeWater {
		item.WaterType = ""
	}
	if item.WeightKg != nil && *item.WeightKg < 0 {
		return fmt.Errorf("%w: weight must be non-negative", equipment.ErrInvalidEquipment)
	}
	return nil
}

func normalize(item *equipment.Equipment) {
	item.Name = strings.TrimSpace(item.Name)
	item.Brand = strings.TrimSpace(item.Brand)
	item.Model = strings.TrimSpace(item.Model)
	item.Notes = strings.TrimSpace(item.Notes)
}

func (s *EquipmentStore) Create(nickname string, item *equipment.Equipment) (*equipment.Equipment, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if err := s.validate(item); err != nil {
		return nil, err
	}
	normalize(item)

	items, err := s.load(nickname)
	if err != nil {
		return nil, err
	}

	item.ID = uuid.NewString()
	items = append(items, *item)
	if err := s.save(nickname, items); err != nil {
		return nil, err
	}

	result := *item
	return &result, nil
}

func (s *EquipmentStore) Update(nickname string, item *equipment.Equipment) (*equipment.Equipment, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if item == nil || strings.TrimSpace(item.ID) == "" {
		return nil, equipment.ErrInvalidEquipment
	}
	if err := s.validate(item); err != nil {
		return nil, err
	}
	normalize(item)

	items, err := s.load(nickname)
	if err != nil {
		return nil, err
	}

	found := false
	for i := range items {
		if items[i].ID == item.ID {
			items[i] = *item
			found = true
			break
		}
	}
	if !found {
		return nil, equipment.ErrEquipmentNotFound
	}

	if err := s.save(nickname, items); err != nil {
		return nil, err
	}

	result := *item
	return &result, nil
}

func (s *EquipmentStore) Delete(nickname, id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	items, err := s.load(nickname)
	if err != nil {
		return err
	}

	found := false
	updated := make([]equipment.Equipment, 0, len(items))
	for i := range items {
		if items[i].ID == id {
			found = true
			continue
		}
		updated = append(updated, items[i])
	}
	if !found {
		return equipment.ErrEquipmentNotFound
	}

	return s.save(nickname, updated)
}

// Import writes an equipment item as-is (used by storage migration).
func (s *EquipmentStore) Import(nickname string, item equipment.Equipment) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	items, err := s.load(nickname)
	if err != nil {
		return err
	}
	for i := range items {
		if items[i].ID == item.ID {
			items[i] = item
			return s.save(nickname, items)
		}
	}
	items = append(items, item)
	return s.save(nickname, items)
}

var _ equipment.Repository = (*EquipmentStore)(nil)
