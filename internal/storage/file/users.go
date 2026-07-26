package file

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
	"unicode"

	"github.com/google/uuid"
	"github.com/solargate/grom/internal/auth"
	"github.com/solargate/grom/internal/users"
	"github.com/solargate/grom/internal/data"
	"gopkg.in/yaml.v3"
)

const usersFileName = "users.yaml"
const maxNicknameLen = 64

type usersFile struct {
	Users []users.User `yaml:"users"`
}

type UsersStore struct {
	dataDir string
	path    string
	mu      sync.Mutex
	users   []users.User
}

func NewUsersStore(dataDir string) (*UsersStore, error) {
	s := &UsersStore{
		dataDir: dataDir,
		path:    filepath.Join(dataDir, usersFileName),
	}
	if err := s.load(); err != nil {
		return nil, err
	}
	if err := s.ensureUserDirs(); err != nil {
		return nil, err
	}
	return s, nil
}

func (s *UsersStore) ensureUserDirs() error {
	for _, user := range s.users {
		if err := ensureUserDir(s.dataDir, user.Nickname); err != nil {
			return fmt.Errorf("ensure user dir for %q: %w", user.Nickname, err)
		}
	}
	return nil
}

func ensureUserDir(dataDir, nickname string) error {
	userDir := data.UserDir(dataDir, nickname)
	if err := os.MkdirAll(userDir, 0700); err != nil {
		return err
	}
	return nil
}

func validateNickname(nickname string) error {
	if nickname == "" {
		return users.ErrInvalidNickname
	}
	if len(nickname) > maxNicknameLen {
		return users.ErrInvalidNickname
	}
	if strings.Contains(nickname, "..") {
		return users.ErrInvalidNickname
	}
	for _, r := range nickname {
		if r == '/' || r == '\\' || r == 0 || unicode.IsControl(r) {
			return users.ErrInvalidNickname
		}
	}
	return nil
}

func (s *UsersStore) load() error {
	data, err := os.ReadFile(s.path)
	if err != nil {
		if os.IsNotExist(err) {
			s.users = nil
			return nil
		}
		return err
	}

	var file usersFile
	if err := yaml.Unmarshal(data, &file); err != nil {
		return fmt.Errorf("parse users file: %w", err)
	}
	s.users = file.Users
	return nil
}

func (s *UsersStore) save() error {
	file := usersFile{Users: s.users}
	data, err := yaml.Marshal(&file)
	if err != nil {
		return err
	}

	tmp := s.path + ".tmp"
	if err := os.WriteFile(tmp, data, 0600); err != nil {
		return err
	}
	return os.Rename(tmp, s.path)
}

func (s *UsersStore) FindByEmail(email string) (*users.User, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	normalized := strings.ToLower(strings.TrimSpace(email))
	for i := range s.users {
		if strings.ToLower(s.users[i].Email) == normalized {
			user := s.users[i]
			return &user, nil
		}
	}
	return nil, users.ErrUserNotFound
}

func (s *UsersStore) FindByID(id string) (*users.User, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for i := range s.users {
		if s.users[i].ID == id {
			user := s.users[i]
			return &user, nil
		}
	}
	return nil, users.ErrUserNotFound
}

func (s *UsersStore) FindByNickname(nickname string) (*users.User, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for i := range s.users {
		if strings.EqualFold(s.users[i].Nickname, nickname) {
			user := s.users[i]
			return &user, nil
		}
	}
	return nil, users.ErrUserNotFound
}

func (s *UsersStore) Search(query, excludeUserID string, limit int) ([]users.User, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if limit <= 0 {
		limit = 20
	}
	query = strings.ToLower(strings.TrimSpace(query))
	if query == "" {
		return nil, nil
	}

	result := make([]users.User, 0, limit)
	for i := range s.users {
		if s.users[i].ID == excludeUserID {
			continue
		}
		nick := strings.ToLower(s.users[i].Nickname)
		name := strings.ToLower(s.users[i].Name)
		if strings.HasPrefix(nick, query) || strings.Contains(nick, query) || strings.Contains(name, query) {
			result = append(result, s.users[i])
			if len(result) >= limit {
				break
			}
		}
	}
	return result, nil
}

func (s *UsersStore) ListAll() ([]users.User, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	result := make([]users.User, len(s.users))
	copy(result, s.users)
	return result, nil
}

func (s *UsersStore) Create(nickname, name, email, password string) (*users.User, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	nickname = strings.TrimSpace(nickname)
	name = strings.TrimSpace(name)
	email = strings.TrimSpace(email)

	if err := validateNickname(nickname); err != nil {
		return nil, err
	}

	normalizedEmail := strings.ToLower(email)
	for i := range s.users {
		if strings.ToLower(s.users[i].Email) == normalizedEmail {
			return nil, users.ErrEmailTaken
		}
		if strings.EqualFold(s.users[i].Nickname, nickname) {
			return nil, users.ErrNicknameTaken
		}
	}

	if err := ensureUserDir(s.dataDir, nickname); err != nil {
		return nil, err
	}

	hash, err := auth.HashPassword(password)
	if err != nil {
		return nil, err
	}

	user := users.User{
		ID:           uuid.NewString(),
		Nickname:     nickname,
		Name:         name,
		Email:        email,
		PasswordHash: hash,
		CreatedAt:    time.Now().UTC(),
	}

	s.users = append(s.users, user)
	if err := s.save(); err != nil {
		s.users = s.users[:len(s.users)-1]
		return nil, err
	}

	return &user, nil
}

func (s *UsersStore) UpdateProfile(userID, name string) (*users.User, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	name = strings.TrimSpace(name)
	for i := range s.users {
		if s.users[i].ID != userID {
			continue
		}
		s.users[i].Name = name
		if err := s.save(); err != nil {
			return nil, err
		}
		user := s.users[i]
		return &user, nil
	}
	return nil, users.ErrUserNotFound
}

func (s *UsersStore) SetLastEquipmentForSport(userID, sportType string, equipmentIDs []string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for i := range s.users {
		if s.users[i].ID != userID {
			continue
		}
		if s.users[i].LastEquipmentBySport == nil {
			s.users[i].LastEquipmentBySport = make(map[string][]string)
		}
		if len(equipmentIDs) == 0 {
			delete(s.users[i].LastEquipmentBySport, sportType)
		} else {
			copied := make([]string, len(equipmentIDs))
			copy(copied, equipmentIDs)
			s.users[i].LastEquipmentBySport[sportType] = copied
		}
		return s.save()
	}
	return users.ErrUserNotFound
}

func (s *UsersStore) RemoveEquipmentFromLastSets(userID, equipmentID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for i := range s.users {
		if s.users[i].ID != userID {
			continue
		}
		if len(s.users[i].LastEquipmentBySport) == 0 {
			return nil
		}
		changed := false
		for sportType, ids := range s.users[i].LastEquipmentBySport {
			filtered := make([]string, 0, len(ids))
			for _, id := range ids {
				if id != equipmentID {
					filtered = append(filtered, id)
				} else {
					changed = true
				}
			}
			if len(filtered) == 0 {
				delete(s.users[i].LastEquipmentBySport, sportType)
			} else if changed {
				s.users[i].LastEquipmentBySport[sportType] = filtered
			}
		}
		if changed {
			return s.save()
		}
		return nil
	}
	return users.ErrUserNotFound
}

// Import writes a user record as-is (used by storage migration).
func (s *UsersStore) Import(user users.User) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if err := validateNickname(user.Nickname); err != nil {
		return err
	}
	if err := ensureUserDir(s.dataDir, user.Nickname); err != nil {
		return err
	}
	normalizedEmail := strings.ToLower(strings.TrimSpace(user.Email))
	for i := range s.users {
		if s.users[i].ID == user.ID {
			s.users[i] = user
			return s.save()
		}
		if strings.ToLower(s.users[i].Email) == normalizedEmail && s.users[i].ID != user.ID {
			return users.ErrEmailTaken
		}
		if strings.EqualFold(s.users[i].Nickname, user.Nickname) && s.users[i].ID != user.ID {
			return users.ErrNicknameTaken
		}
	}
	s.users = append(s.users, user)
	return s.save()
}

var _ users.Repository = (*UsersStore)(nil)
