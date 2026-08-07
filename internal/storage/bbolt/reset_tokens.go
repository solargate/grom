package bbolt

import (
	"encoding/json"
	"fmt"
	"time"

	bolt "go.etcd.io/bbolt"

	"github.com/solargate/grom/internal/auth/reset"
)

// ResetTokenStore persists password-reset tokens in bbolt.
type ResetTokenStore struct {
	db *bolt.DB
}

func NewResetTokenStore(db *bolt.DB) *ResetTokenStore {
	return &ResetTokenStore{db: db}
}

func (s *ResetTokenStore) ReplaceForUser(userID string, record reset.TokenRecord) error {
	return s.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(bucketResetTokens)
		c := b.Cursor()
		for k, v := c.First(); k != nil; k, v = c.Next() {
			var existing reset.TokenRecord
			if err := json.Unmarshal(v, &existing); err != nil {
				return err
			}
			if existing.UserID == userID {
				if err := b.Delete(k); err != nil {
					return err
				}
			}
		}
		payload, err := json.Marshal(record)
		if err != nil {
			return err
		}
		return b.Put([]byte(record.TokenHash), payload)
	})
}

func (s *ResetTokenStore) GetByHash(hash string) (*reset.TokenRecord, error) {
	var found *reset.TokenRecord
	err := s.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(bucketResetTokens)
		raw := b.Get([]byte(hash))
		if raw == nil {
			return reset.ErrInvalidToken
		}
		var rec reset.TokenRecord
		if err := json.Unmarshal(raw, &rec); err != nil {
			return err
		}
		if !rec.ExpiresAt.After(time.Now().UTC()) {
			_ = b.Delete([]byte(hash))
			return reset.ErrInvalidToken
		}
		found = &rec
		return nil
	})
	if err != nil {
		return nil, err
	}
	return found, nil
}

func (s *ResetTokenStore) DeleteByHash(hash string) error {
	return s.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(bucketResetTokens)
		if err := b.Delete([]byte(hash)); err != nil {
			return fmt.Errorf("delete reset token: %w", err)
		}
		return nil
	})
}

var _ reset.TokenStore = (*ResetTokenStore)(nil)
