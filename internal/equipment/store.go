package equipment

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/google/uuid"
	"github.com/solargate/travka/internal/data"
	"gopkg.in/yaml.v3"
)

const equipmentFileName = "equipment.yaml"

type Store struct {
	dataDir string
	mu      sync.Mutex
}

func NewStore(dataDir string) *Store {
	return &Store{dataDir: dataDir}
}

func (s *Store) userDir(nickname string) string {
	return data.UserDir(s.dataDir, nickname)
}

func (s *Store) filePath(nickname string) string {
	return filepath.Join(s.userDir(nickname), equipmentFileName)
}

func (s *Store) load(nickname string) ([]Equipment, error) {
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

func (s *Store) save(nickname string, items []Equipment) error {
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

func (s *Store) List(nickname string) ([]Equipment, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.load(nickname)
}

func (s *Store) FindByID(nickname, id string) (*Equipment, error) {
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
	return nil, ErrEquipmentNotFound
}

func (s *Store) FindByIDs(nickname string, ids []string) ([]Equipment, error) {
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

	result := make([]Equipment, 0, len(ids))
	for i := range items {
		if _, ok := idSet[items[i].ID]; ok {
			result = append(result, items[i])
		}
	}
	return result, nil
}

func (s *Store) validate(item *Equipment) error {
	if item == nil {
		return ErrInvalidEquipment
	}
	if !IsValidType(item.Type) {
		return fmt.Errorf("%w: invalid type", ErrInvalidEquipment)
	}
	if strings.TrimSpace(item.Name) == "" {
		return fmt.Errorf("%w: name is required", ErrInvalidEquipment)
	}
	if item.Type == TypeBike && !IsValidBikeType(item.BikeType) {
		return fmt.Errorf("%w: invalid bike type", ErrInvalidEquipment)
	}
	if item.Type == TypeWater && !IsValidWaterType(item.WaterType) {
		return fmt.Errorf("%w: invalid water type", ErrInvalidEquipment)
	}
	if item.Type != TypeBike {
		item.BikeType = ""
	}
	if item.Type != TypeWater {
		item.WaterType = ""
	}
	if item.WeightKg != nil && *item.WeightKg < 0 {
		return fmt.Errorf("%w: weight must be non-negative", ErrInvalidEquipment)
	}
	return nil
}

func normalize(item *Equipment) {
	item.Name = strings.TrimSpace(item.Name)
	item.Brand = strings.TrimSpace(item.Brand)
	item.Model = strings.TrimSpace(item.Model)
	item.Notes = strings.TrimSpace(item.Notes)
}

func (s *Store) Create(nickname string, item *Equipment) (*Equipment, error) {
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

func (s *Store) Update(nickname string, item *Equipment) (*Equipment, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if item == nil || strings.TrimSpace(item.ID) == "" {
		return nil, ErrInvalidEquipment
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
		return nil, ErrEquipmentNotFound
	}

	if err := s.save(nickname, items); err != nil {
		return nil, err
	}

	result := *item
	return &result, nil
}

func (s *Store) Delete(nickname, id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	items, err := s.load(nickname)
	if err != nil {
		return err
	}

	found := false
	updated := make([]Equipment, 0, len(items))
	for i := range items {
		if items[i].ID == id {
			found = true
			continue
		}
		updated = append(updated, items[i])
	}
	if !found {
		return ErrEquipmentNotFound
	}

	return s.save(nickname, updated)
}
