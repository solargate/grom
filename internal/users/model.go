package users

import "time"

type User struct {
	ID           string    `yaml:"id" json:"id"`
	Nickname     string    `yaml:"nickname" json:"nickname"`
	Name         string    `yaml:"name,omitempty" json:"name,omitempty"`
	Email        string    `yaml:"email" json:"email"`
	PasswordHash string    `yaml:"password_hash" json:"-"`
	CreatedAt    time.Time `yaml:"created_at" json:"created_at"`
}

// Profile holds per-user UI/service preferences (not identity fields).
type Profile struct {
	LastSportType        string              `yaml:"last_sport_type,omitempty" json:"last_sport_type,omitempty"`
	LastEquipmentBySport map[string][]string `yaml:"last_equipment_by_sport,omitempty" json:"last_equipment_by_sport,omitempty"`
}
