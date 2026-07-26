package bbolt

import (
	"fmt"
	"strings"

	"github.com/google/uuid"
	bolt "go.etcd.io/bbolt"

	"github.com/solargate/grom/internal/equipment"
)

type EquipmentStore struct {
	db *bolt.DB
}

func NewEquipmentStore(db *bolt.DB) *EquipmentStore {
	return &EquipmentStore{db: db}
}

func equipmentKey(nickname, id string) []byte {
	return []byte(nickname + "/" + id)
}

func equipmentPrefix(nickname string) []byte {
	return []byte(nickname + "/")
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

func normalizeEquipment(item *equipment.Equipment) {
	item.Name = strings.TrimSpace(item.Name)
	item.Brand = strings.TrimSpace(item.Brand)
	item.Model = strings.TrimSpace(item.Model)
	item.Notes = strings.TrimSpace(item.Notes)
}

func (s *EquipmentStore) List(nickname string) ([]equipment.Equipment, error) {
	prefix := equipmentPrefix(nickname)
	var result []equipment.Equipment
	err := s.db.View(func(tx *bolt.Tx) error {
		c := tx.Bucket(bucketEquipment).Cursor()
		for k, v := c.Seek(prefix); k != nil && strings.HasPrefix(string(k), string(prefix)); k, v = c.Next() {
			var item equipment.Equipment
			if err := unmarshalJSON(v, &item); err != nil {
				return err
			}
			result = append(result, item)
		}
		return nil
	})
	return result, err
}

func (s *EquipmentStore) FindByID(nickname, id string) (*equipment.Equipment, error) {
	var result *equipment.Equipment
	err := s.db.View(func(tx *bolt.Tx) error {
		raw := tx.Bucket(bucketEquipment).Get(equipmentKey(nickname, id))
		if raw == nil {
			return equipment.ErrEquipmentNotFound
		}
		var item equipment.Equipment
		if err := unmarshalJSON(raw, &item); err != nil {
			return err
		}
		result = &item
		return nil
	})
	return result, err
}

func (s *EquipmentStore) FindByIDs(nickname string, ids []string) ([]equipment.Equipment, error) {
	if len(ids) == 0 {
		return nil, nil
	}
	idSet := make(map[string]struct{}, len(ids))
	for _, id := range ids {
		idSet[id] = struct{}{}
	}
	items, err := s.List(nickname)
	if err != nil {
		return nil, err
	}
	result := make([]equipment.Equipment, 0, len(ids))
	for i := range items {
		if _, ok := idSet[items[i].ID]; ok {
			result = append(result, items[i])
		}
	}
	return result, nil
}

func (s *EquipmentStore) Create(nickname string, item *equipment.Equipment) (*equipment.Equipment, error) {
	if err := s.validate(item); err != nil {
		return nil, err
	}
	normalizeEquipment(item)
	item.ID = uuid.NewString()

	raw, err := marshalJSON(item)
	if err != nil {
		return nil, err
	}
	err = s.db.Update(func(tx *bolt.Tx) error {
		return tx.Bucket(bucketEquipment).Put(equipmentKey(nickname, item.ID), raw)
	})
	if err != nil {
		return nil, err
	}
	result := *item
	return &result, nil
}

func (s *EquipmentStore) Update(nickname string, item *equipment.Equipment) (*equipment.Equipment, error) {
	if item == nil || strings.TrimSpace(item.ID) == "" {
		return nil, equipment.ErrInvalidEquipment
	}
	if err := s.validate(item); err != nil {
		return nil, err
	}
	normalizeEquipment(item)

	raw, err := marshalJSON(item)
	if err != nil {
		return nil, err
	}
	err = s.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(bucketEquipment)
		if b.Get(equipmentKey(nickname, item.ID)) == nil {
			return equipment.ErrEquipmentNotFound
		}
		return b.Put(equipmentKey(nickname, item.ID), raw)
	})
	if err != nil {
		return nil, err
	}
	result := *item
	return &result, nil
}

func (s *EquipmentStore) Delete(nickname, id string) error {
	return s.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(bucketEquipment)
		key := equipmentKey(nickname, id)
		if b.Get(key) == nil {
			return equipment.ErrEquipmentNotFound
		}
		return b.Delete(key)
	})
}

// PutExisting writes equipment without generating a new ID (migration).
func (s *EquipmentStore) PutExisting(nickname string, item equipment.Equipment) error {
	raw, err := marshalJSON(item)
	if err != nil {
		return err
	}
	return s.db.Update(func(tx *bolt.Tx) error {
		return tx.Bucket(bucketEquipment).Put(equipmentKey(nickname, item.ID), raw)
	})
}

var _ equipment.Repository = (*EquipmentStore)(nil)
