package bbolt

import (
	"fmt"

	bolt "go.etcd.io/bbolt"

	"github.com/solargate/grom/internal/users"
)

type profileRecord struct {
	LastSportType        string              `json:"last_sport_type,omitempty"`
	LastEquipmentBySport map[string][]string `json:"last_equipment_by_sport,omitempty"`
	UsedSportTypes       []string            `json:"used_sport_types,omitempty"`
}

func profileToRecord(p users.Profile) profileRecord {
	return profileRecord{
		LastSportType:        p.LastSportType,
		LastEquipmentBySport: p.LastEquipmentBySport,
		UsedSportTypes:       p.UsedSportTypes,
	}
}

func recordToProfile(r profileRecord) users.Profile {
	return users.Profile{
		LastSportType:        r.LastSportType,
		LastEquipmentBySport: r.LastEquipmentBySport,
		UsedSportTypes:       r.UsedSportTypes,
	}
}

func (s *UsersStore) getProfile(tx *bolt.Tx, userID string) (users.Profile, error) {
	raw := tx.Bucket(bucketUserProfiles).Get([]byte(userID))
	if raw == nil {
		return users.Profile{}, nil
	}
	var rec profileRecord
	if err := unmarshalJSON(raw, &rec); err != nil {
		return users.Profile{}, fmt.Errorf("parse profile: %w", err)
	}
	return recordToProfile(rec), nil
}

func (s *UsersStore) putProfile(tx *bolt.Tx, userID string, profile users.Profile) error {
	raw, err := marshalJSON(profileToRecord(profile))
	if err != nil {
		return err
	}
	return tx.Bucket(bucketUserProfiles).Put([]byte(userID), raw)
}

func (s *UsersStore) GetProfile(userID string) (*users.Profile, error) {
	var profile users.Profile
	err := s.db.View(func(tx *bolt.Tx) error {
		if _, err := s.getByID(tx, userID); err != nil {
			return err
		}
		p, err := s.getProfile(tx, userID)
		if err != nil {
			return err
		}
		profile = p
		return nil
	})
	if err != nil {
		return nil, err
	}
	return &profile, nil
}

func (s *UsersStore) PutProfile(userID string, profile users.Profile) error {
	return s.db.Update(func(tx *bolt.Tx) error {
		if _, err := s.getByID(tx, userID); err != nil {
			return err
		}
		return s.putProfile(tx, userID, cloneProfile(profile))
	})
}

func (s *UsersStore) SetLastSportType(userID, sportType string) error {
	return s.db.Update(func(tx *bolt.Tx) error {
		if _, err := s.getByID(tx, userID); err != nil {
			return err
		}
		profile, err := s.getProfile(tx, userID)
		if err != nil {
			return err
		}
		profile.LastSportType = sportType
		return s.putProfile(tx, userID, profile)
	})
}

func (s *UsersStore) TouchUsedSportType(userID, sportType string) error {
	if sportType == "" {
		return nil
	}
	return s.db.Update(func(tx *bolt.Tx) error {
		if _, err := s.getByID(tx, userID); err != nil {
			return err
		}
		profile, err := s.getProfile(tx, userID)
		if err != nil {
			return err
		}
		updated := users.MoveSportToFront(profile.UsedSportTypes, sportType)
		if slicesEqual(profile.UsedSportTypes, updated) {
			return nil
		}
		profile.UsedSportTypes = updated
		return s.putProfile(tx, userID, profile)
	})
}

func (s *UsersStore) PruneUsedSportTypes(userID string, remaining map[string]struct{}) error {
	return s.db.Update(func(tx *bolt.Tx) error {
		if _, err := s.getByID(tx, userID); err != nil {
			return err
		}
		profile, err := s.getProfile(tx, userID)
		if err != nil {
			return err
		}
		pruned := users.PruneUsedSports(profile.UsedSportTypes, remaining)
		if slicesEqual(profile.UsedSportTypes, pruned) {
			return nil
		}
		profile.UsedSportTypes = pruned
		return s.putProfile(tx, userID, profile)
	})
}

func (s *UsersStore) SetLastEquipmentForSport(userID, sportType string, equipmentIDs []string) error {
	return s.db.Update(func(tx *bolt.Tx) error {
		if _, err := s.getByID(tx, userID); err != nil {
			return err
		}
		profile, err := s.getProfile(tx, userID)
		if err != nil {
			return err
		}
		if profile.LastEquipmentBySport == nil {
			profile.LastEquipmentBySport = make(map[string][]string)
		}
		if len(equipmentIDs) == 0 {
			delete(profile.LastEquipmentBySport, sportType)
		} else {
			copied := make([]string, len(equipmentIDs))
			copy(copied, equipmentIDs)
			profile.LastEquipmentBySport[sportType] = copied
		}
		return s.putProfile(tx, userID, profile)
	})
}

func (s *UsersStore) RemoveEquipmentFromLastSets(userID, equipmentID string) error {
	return s.db.Update(func(tx *bolt.Tx) error {
		if _, err := s.getByID(tx, userID); err != nil {
			return err
		}
		profile, err := s.getProfile(tx, userID)
		if err != nil {
			return err
		}
		if len(profile.LastEquipmentBySport) == 0 {
			return nil
		}
		changed := false
		for sportType, ids := range profile.LastEquipmentBySport {
			filtered := make([]string, 0, len(ids))
			for _, id := range ids {
				if id != equipmentID {
					filtered = append(filtered, id)
				} else {
					changed = true
				}
			}
			if len(filtered) == 0 {
				delete(profile.LastEquipmentBySport, sportType)
			} else if changed {
				profile.LastEquipmentBySport[sportType] = filtered
			}
		}
		if !changed {
			return nil
		}
		return s.putProfile(tx, userID, profile)
	})
}

func cloneProfile(p users.Profile) users.Profile {
	out := users.Profile{
		LastSportType:  p.LastSportType,
		UsedSportTypes: cloneStrings(p.UsedSportTypes),
	}
	if len(p.LastEquipmentBySport) == 0 {
		return out
	}
	out.LastEquipmentBySport = make(map[string][]string, len(p.LastEquipmentBySport))
	for sport, ids := range p.LastEquipmentBySport {
		copied := make([]string, len(ids))
		copy(copied, ids)
		out.LastEquipmentBySport[sport] = copied
	}
	return out
}

func cloneStrings(in []string) []string {
	if len(in) == 0 {
		return nil
	}
	out := make([]string, len(in))
	copy(out, in)
	return out
}

func slicesEqual(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
