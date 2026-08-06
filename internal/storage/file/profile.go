package file

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/solargate/grom/internal/data"
	"github.com/solargate/grom/internal/users"
	"gopkg.in/yaml.v3"
)

const profileFileName = "profile.yaml"

func (s *UsersStore) profilePath(nickname string) string {
	return filepath.Join(data.UserDir(s.dataDir, nickname), profileFileName)
}

func (s *UsersStore) nicknameByIDLocked(userID string) (string, error) {
	for i := range s.users {
		if s.users[i].ID == userID {
			return s.users[i].Nickname, nil
		}
	}
	return "", users.ErrUserNotFound
}

func (s *UsersStore) loadProfile(nickname string) (users.Profile, error) {
	raw, err := os.ReadFile(s.profilePath(nickname))
	if err != nil {
		if os.IsNotExist(err) {
			return users.Profile{}, nil
		}
		return users.Profile{}, err
	}
	var profile users.Profile
	if err := yaml.Unmarshal(raw, &profile); err != nil {
		return users.Profile{}, fmt.Errorf("parse profile file: %w", err)
	}
	return profile, nil
}

func (s *UsersStore) saveProfile(nickname string, profile users.Profile) error {
	if err := ensureUserDir(s.dataDir, nickname); err != nil {
		return err
	}
	raw, err := yaml.Marshal(&profile)
	if err != nil {
		return err
	}
	path := s.profilePath(nickname)
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, raw, 0600); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}

func (s *UsersStore) GetProfile(userID string) (*users.Profile, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	nickname, err := s.nicknameByIDLocked(userID)
	if err != nil {
		return nil, err
	}
	profile, err := s.loadProfile(nickname)
	if err != nil {
		return nil, err
	}
	return &profile, nil
}

func (s *UsersStore) PutProfile(userID string, profile users.Profile) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	nickname, err := s.nicknameByIDLocked(userID)
	if err != nil {
		return err
	}
	return s.saveProfile(nickname, cloneProfile(profile))
}

func (s *UsersStore) SetLastSportType(userID, sportType string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	nickname, err := s.nicknameByIDLocked(userID)
	if err != nil {
		return err
	}
	profile, err := s.loadProfile(nickname)
	if err != nil {
		return err
	}
	profile.LastSportType = sportType
	return s.saveProfile(nickname, profile)
}

func (s *UsersStore) SetLastEquipmentForSport(userID, sportType string, equipmentIDs []string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	nickname, err := s.nicknameByIDLocked(userID)
	if err != nil {
		return err
	}
	profile, err := s.loadProfile(nickname)
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
	return s.saveProfile(nickname, profile)
}

func (s *UsersStore) RemoveEquipmentFromLastSets(userID, equipmentID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	nickname, err := s.nicknameByIDLocked(userID)
	if err != nil {
		return err
	}
	profile, err := s.loadProfile(nickname)
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
	return s.saveProfile(nickname, profile)
}

func cloneProfile(p users.Profile) users.Profile {
	out := users.Profile{LastSportType: p.LastSportType}
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
