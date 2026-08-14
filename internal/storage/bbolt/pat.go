package bbolt

import (
	"encoding/json"
	"fmt"
	"time"

	bolt "go.etcd.io/bbolt"

	"github.com/solargate/grom/internal/auth/pat"
)

// PATStore persists personal access tokens in bbolt.
type PATStore struct {
	db *bolt.DB
}

func NewPATStore(db *bolt.DB) *PATStore {
	return &PATStore{db: db}
}

func (s *PATStore) Create(record pat.TokenRecord) error {
	payload, err := json.Marshal(record)
	if err != nil {
		return err
	}
	return s.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(bucketPersonalAccessTokens)
		if b.Get([]byte(record.TokenHash)) != nil {
			return fmt.Errorf("pat hash collision")
		}
		return b.Put([]byte(record.TokenHash), payload)
	})
}

func (s *PATStore) ListByUser(userID string) ([]pat.TokenRecord, error) {
	var out []pat.TokenRecord
	err := s.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(bucketPersonalAccessTokens)
		return b.ForEach(func(k, v []byte) error {
			var rec pat.TokenRecord
			if err := json.Unmarshal(v, &rec); err != nil {
				return err
			}
			if rec.TokenHash == "" {
				rec.TokenHash = string(k)
			}
			if rec.UserID == userID {
				out = append(out, rec)
			}
			return nil
		})
	})
	return out, err
}

func (s *PATStore) CountByUser(userID string) (int, error) {
	count := 0
	err := s.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(bucketPersonalAccessTokens)
		return b.ForEach(func(_, v []byte) error {
			var rec pat.TokenRecord
			if err := json.Unmarshal(v, &rec); err != nil {
				return err
			}
			if rec.UserID == userID {
				count++
			}
			return nil
		})
	})
	return count, err
}

func (s *PATStore) GetByHash(hash string) (*pat.TokenRecord, error) {
	var found *pat.TokenRecord
	err := s.db.View(func(tx *bolt.Tx) error {
		raw := tx.Bucket(bucketPersonalAccessTokens).Get([]byte(hash))
		if raw == nil {
			return pat.ErrInvalidToken
		}
		var rec pat.TokenRecord
		if err := json.Unmarshal(raw, &rec); err != nil {
			return err
		}
		if rec.TokenHash == "" {
			rec.TokenHash = hash
		}
		found = &rec
		return nil
	})
	if err != nil {
		return nil, err
	}
	return found, nil
}

func (s *PATStore) DeleteByUserAndID(userID, id string) error {
	return s.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(bucketPersonalAccessTokens)
		var hashKey []byte
		c := b.Cursor()
		for k, v := c.First(); k != nil; k, v = c.Next() {
			var rec pat.TokenRecord
			if err := json.Unmarshal(v, &rec); err != nil {
				return err
			}
			if rec.UserID == userID && rec.ID == id {
				hashKey = append([]byte(nil), k...)
				break
			}
		}
		if hashKey == nil {
			return pat.ErrNotFound
		}
		return b.Delete(hashKey)
	})
}

func (s *PATStore) UpdateLastUsed(id string, at time.Time) error {
	return s.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(bucketPersonalAccessTokens)
		c := b.Cursor()
		for k, v := c.First(); k != nil; k, v = c.Next() {
			var rec pat.TokenRecord
			if err := json.Unmarshal(v, &rec); err != nil {
				return err
			}
			if rec.ID != id {
				continue
			}
			rec.LastUsedAt = &at
			payload, err := json.Marshal(rec)
			if err != nil {
				return err
			}
			return b.Put(k, payload)
		}
		return pat.ErrNotFound
	})
}

// ListAll returns every personal access token (migration).
func (s *PATStore) ListAll() ([]pat.TokenRecord, error) {
	var out []pat.TokenRecord
	err := s.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(bucketPersonalAccessTokens)
		return b.ForEach(func(k, v []byte) error {
			var rec pat.TokenRecord
			if err := json.Unmarshal(v, &rec); err != nil {
				return err
			}
			// TokenHash is omitted from JSON (json:"-"); restore from bucket key.
			if rec.TokenHash == "" {
				rec.TokenHash = string(k)
			}
			out = append(out, rec)
			return nil
		})
	})
	return out, err
}

// PutExisting writes a PAT record as-is (used by storage migration).
func (s *PATStore) PutExisting(record pat.TokenRecord) error {
	payload, err := json.Marshal(record)
	if err != nil {
		return err
	}
	return s.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(bucketPersonalAccessTokens)
		c := b.Cursor()
		for k, v := c.First(); k != nil; k, v = c.Next() {
			var existing pat.TokenRecord
			if err := json.Unmarshal(v, &existing); err != nil {
				return err
			}
			if existing.ID == record.ID && string(k) != record.TokenHash {
				if err := b.Delete(k); err != nil {
					return err
				}
				break
			}
		}
		return b.Put([]byte(record.TokenHash), payload)
	})
}

var _ pat.Repository = (*PATStore)(nil)
