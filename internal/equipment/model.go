package equipment

import "errors"

const (
	TypeBike  = "bike"
	TypeShoes = "shoes"
	TypeWater = "water"
	TypeOther = "other"
)

var (
	ErrInvalidEquipment = errors.New("invalid equipment")
	ErrEquipmentExists  = errors.New("equipment already exists")
	ErrEquipmentNotFound = errors.New("equipment not found")
)

type Equipment struct {
	ID        string   `yaml:"id" json:"id"`
	Type      string   `yaml:"type" json:"type"`
	Name      string   `yaml:"name" json:"name"`
	BikeType  string   `yaml:"bike_type,omitempty" json:"bike_type,omitempty"`
	WaterType string   `yaml:"water_type,omitempty" json:"water_type,omitempty"`
	Brand     string   `yaml:"brand,omitempty" json:"brand,omitempty"`
	Model     string   `yaml:"model,omitempty" json:"model,omitempty"`
	WeightKg  *float64 `yaml:"weight_kg,omitempty" json:"weight_kg,omitempty"`
	Notes     string   `yaml:"notes,omitempty" json:"notes,omitempty"`
}

type equipmentFile struct {
	Equipment []Equipment `yaml:"equipment"`
}

func IsValidType(t string) bool {
	switch t {
	case TypeBike, TypeShoes, TypeWater, TypeOther:
		return true
	default:
		return false
	}
}

func IsValidBikeType(t string) bool {
	if t == "" {
		return true
	}
	switch t {
	case "mountain", "gravel", "road", "touring", "triathlon", "cyclocross", "fixie", "folding", "bmx":
		return true
	default:
		return false
	}
}

func IsValidWaterType(t string) bool {
	if t == "" {
		return true
	}
	switch t {
	case "sup", "kayak", "canoe", "canoe_double", "packraft", "surf":
		return true
	default:
		return false
	}
}
