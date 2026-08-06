package bbolt

import (
	"fmt"
	"os"
	"strings"
	"time"
	"unicode"

	"github.com/google/uuid"
	bolt "go.etcd.io/bbolt"

	"github.com/solargate/grom/internal/auth"
	"github.com/solargate/grom/internal/data"
	"github.com/solargate/grom/internal/users"
)

const maxNicknameLen = 64

type UsersStore struct {
	db      *bolt.DB
	dataDir string
}

func NewUsersStore(db *bolt.DB, dataDir string) *UsersStore {
	return &UsersStore{db: db, dataDir: dataDir}
}

func ensureUserDir(dataDir, nickname string) error {
	return os.MkdirAll(data.UserDir(dataDir, nickname), 0700)
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

func (s *UsersStore) getByID(tx *bolt.Tx, id string) (*users.User, error) {
	b := tx.Bucket(bucketUsers)
	raw := b.Get([]byte(id))
	if raw == nil {
		return nil, users.ErrUserNotFound
	}
	var rec userRecord
	if err := unmarshalJSON(raw, &rec); err != nil {
		return nil, fmt.Errorf("parse user: %w", err)
	}
	u := recordToUser(rec)
	return &u, nil
}

func (s *UsersStore) putUser(tx *bolt.Tx, u users.User) error {
	raw, err := marshalJSON(userToRecord(u))
	if err != nil {
		return err
	}
	b := tx.Bucket(bucketUsers)
	if err := b.Put([]byte(u.ID), raw); err != nil {
		return err
	}
	emailKey := []byte(strings.ToLower(strings.TrimSpace(u.Email)))
	nickKey := []byte(strings.ToLower(u.Nickname))
	if err := tx.Bucket(bucketIdxUsersEmail).Put(emailKey, []byte(u.ID)); err != nil {
		return err
	}
	return tx.Bucket(bucketIdxUsersNick).Put(nickKey, []byte(u.ID))
}

func (s *UsersStore) FindByEmail(email string) (*users.User, error) {
	var result *users.User
	err := s.db.View(func(tx *bolt.Tx) error {
		id := tx.Bucket(bucketIdxUsersEmail).Get([]byte(strings.ToLower(strings.TrimSpace(email))))
		if id == nil {
			return users.ErrUserNotFound
		}
		u, err := s.getByID(tx, string(id))
		if err != nil {
			return err
		}
		result = u
		return nil
	})
	return result, err
}

func (s *UsersStore) FindByID(id string) (*users.User, error) {
	var result *users.User
	err := s.db.View(func(tx *bolt.Tx) error {
		u, err := s.getByID(tx, id)
		if err != nil {
			return err
		}
		result = u
		return nil
	})
	return result, err
}

func (s *UsersStore) FindByNickname(nickname string) (*users.User, error) {
	var result *users.User
	err := s.db.View(func(tx *bolt.Tx) error {
		id := tx.Bucket(bucketIdxUsersNick).Get([]byte(strings.ToLower(nickname)))
		if id == nil {
			return users.ErrUserNotFound
		}
		u, err := s.getByID(tx, string(id))
		if err != nil {
			return err
		}
		result = u
		return nil
	})
	return result, err
}

func (s *UsersStore) Search(query, excludeUserID string, limit int) ([]users.User, error) {
	if limit <= 0 {
		limit = 20
	}
	query = strings.ToLower(strings.TrimSpace(query))
	if query == "" {
		return nil, nil
	}

	var result []users.User
	err := s.db.View(func(tx *bolt.Tx) error {
		c := tx.Bucket(bucketUsers).Cursor()
		for k, v := c.First(); k != nil; k, v = c.Next() {
			var rec userRecord
			if err := unmarshalJSON(v, &rec); err != nil {
				return err
			}
			if rec.ID == excludeUserID {
				continue
			}
			nick := strings.ToLower(rec.Nickname)
			name := strings.ToLower(rec.Name)
			if strings.HasPrefix(nick, query) || strings.Contains(nick, query) || strings.Contains(name, query) {
				result = append(result, recordToUser(rec))
				if len(result) >= limit {
					break
				}
			}
		}
		return nil
	})
	return result, err
}

func (s *UsersStore) ListAll() ([]users.User, error) {
	var result []users.User
	err := s.db.View(func(tx *bolt.Tx) error {
		return tx.Bucket(bucketUsers).ForEach(func(_, v []byte) error {
			var rec userRecord
			if err := unmarshalJSON(v, &rec); err != nil {
				return err
			}
			result = append(result, recordToUser(rec))
			return nil
		})
	})
	return result, err
}

func (s *UsersStore) Create(nickname, name, email, password string) (*users.User, error) {
	nickname = strings.TrimSpace(nickname)
	name = strings.TrimSpace(name)
	email = strings.TrimSpace(email)

	if err := validateNickname(nickname); err != nil {
		return nil, err
	}

	hash, err := auth.HashPassword(password)
	if err != nil {
		return nil, err
	}
	if err := ensureUserDir(s.dataDir, nickname); err != nil {
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

	err = s.db.Update(func(tx *bolt.Tx) error {
		normalizedEmail := strings.ToLower(email)
		if tx.Bucket(bucketIdxUsersEmail).Get([]byte(normalizedEmail)) != nil {
			return users.ErrEmailTaken
		}
		if tx.Bucket(bucketIdxUsersNick).Get([]byte(strings.ToLower(nickname))) != nil {
			return users.ErrNicknameTaken
		}
		return s.putUser(tx, user)
	})
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (s *UsersStore) UpdateProfile(userID, name string) (*users.User, error) {
	name = strings.TrimSpace(name)
	var result *users.User
	err := s.db.Update(func(tx *bolt.Tx) error {
		u, err := s.getByID(tx, userID)
		if err != nil {
			return err
		}
		u.Name = name
		if err := s.putUser(tx, *u); err != nil {
			return err
		}
		result = u
		return nil
	})
	return result, err
}

// PutExisting writes a user record without hashing (used by migration).
func (s *UsersStore) PutExisting(u users.User) error {
	if err := ensureUserDir(s.dataDir, u.Nickname); err != nil {
		return err
	}
	return s.db.Update(func(tx *bolt.Tx) error {
		return s.putUser(tx, u)
	})
}

var _ users.Repository = (*UsersStore)(nil)
