package bbolt

import (
	"encoding/json"
	"time"

	"github.com/solargate/grom/internal/users"
)

type userRecord struct {
	ID                   string              `json:"id"`
	Nickname             string              `json:"nickname"`
	Name                 string              `json:"name,omitempty"`
	Email                string              `json:"email"`
	PasswordHash         string              `json:"password_hash"`
	CreatedAt            time.Time           `json:"created_at"`
	LastEquipmentBySport map[string][]string `json:"last_equipment_by_sport,omitempty"`
}

func userToRecord(u users.User) userRecord {
	return userRecord{
		ID:                   u.ID,
		Nickname:             u.Nickname,
		Name:                 u.Name,
		Email:                u.Email,
		PasswordHash:         u.PasswordHash,
		CreatedAt:            u.CreatedAt,
		LastEquipmentBySport: u.LastEquipmentBySport,
	}
}

func recordToUser(r userRecord) users.User {
	return users.User{
		ID:                   r.ID,
		Nickname:             r.Nickname,
		Name:                 r.Name,
		Email:                r.Email,
		PasswordHash:         r.PasswordHash,
		CreatedAt:            r.CreatedAt,
		LastEquipmentBySport: r.LastEquipmentBySport,
	}
}

func marshalJSON(v any) ([]byte, error) {
	return json.Marshal(v)
}

func unmarshalJSON(data []byte, v any) error {
	return json.Unmarshal(data, v)
}
