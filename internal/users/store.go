package users

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/solargate/travka/internal/auth"
	"gopkg.in/yaml.v3"
)

var (
	ErrEmailTaken    = errors.New("email already registered")
	ErrNicknameTaken = errors.New("nickname already taken")
	ErrUserNotFound  = errors.New("user not found")
)

type Store struct {
	path  string
	mu    sync.Mutex
	users []User
}

func NewStore(path string) (*Store, error) {
	s := &Store{path: path}
	if err := s.load(); err != nil {
		return nil, err
	}
	return s, nil
}

func (s *Store) load() error {
	data, err := os.ReadFile(s.path)
	if err != nil {
		if os.IsNotExist(err) {
			s.users = nil
			return nil
		}
		return err
	}

	var file userFile
	if err := yaml.Unmarshal(data, &file); err != nil {
		return fmt.Errorf("parse users file: %w", err)
	}
	s.users = file.Users
	return nil
}

func (s *Store) save() error {
	file := userFile{Users: s.users}
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

func (s *Store) FindByEmail(email string) (*User, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	normalized := strings.ToLower(strings.TrimSpace(email))
	for i := range s.users {
		if strings.ToLower(s.users[i].Email) == normalized {
			user := s.users[i]
			return &user, nil
		}
	}
	return nil, ErrUserNotFound
}

func (s *Store) FindByID(id string) (*User, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for i := range s.users {
		if s.users[i].ID == id {
			user := s.users[i]
			return &user, nil
		}
	}
	return nil, ErrUserNotFound
}

func (s *Store) Create(nickname, name, email, password string) (*User, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	nickname = strings.TrimSpace(nickname)
	name = strings.TrimSpace(name)
	email = strings.TrimSpace(email)

	normalizedEmail := strings.ToLower(email)
	for i := range s.users {
		if strings.ToLower(s.users[i].Email) == normalizedEmail {
			return nil, ErrEmailTaken
		}
		if strings.EqualFold(s.users[i].Nickname, nickname) {
			return nil, ErrNicknameTaken
		}
	}

	hash, err := auth.HashPassword(password)
	if err != nil {
		return nil, err
	}

	user := User{
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
