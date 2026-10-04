package bbolt

import (
	"encoding/json"
	"fmt"
	"time"

	bolt "go.etcd.io/bbolt"

	"github.com/solargate/grom/internal/notifications"
)

// PushSubscriptionStore persists Web Push subscriptions in bbolt.
type PushSubscriptionStore struct {
	db *bolt.DB
}

func NewPushSubscriptionStore(db *bolt.DB) *PushSubscriptionStore {
	return &PushSubscriptionStore{db: db}
}

func pushSubKey(userID, installationID string) []byte {
	return []byte(userID + "\x00" + installationID)
}

func (s *PushSubscriptionStore) Upsert(sub notifications.PushSubscription) error {
	payload, err := json.Marshal(sub)
	if err != nil {
		return err
	}
	return s.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(bucketPushSubscriptions)
		return b.Put(pushSubKey(sub.UserID, sub.InstallationID), payload)
	})
}

func (s *PushSubscriptionStore) ListByUser(userID string) ([]notifications.PushSubscription, error) {
	var out []notifications.PushSubscription
	err := s.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(bucketPushSubscriptions)
		return b.ForEach(func(_, v []byte) error {
			var sub notifications.PushSubscription
			if err := json.Unmarshal(v, &sub); err != nil {
				return err
			}
			if sub.UserID == userID {
				out = append(out, sub)
			}
			return nil
		})
	})
	return out, err
}

func (s *PushSubscriptionStore) DeleteByUserAndInstallation(userID, installationID string) error {
	return s.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(bucketPushSubscriptions)
		key := pushSubKey(userID, installationID)
		if b.Get(key) == nil {
			return notifications.ErrNotFound
		}
		return b.Delete(key)
	})
}

func (s *PushSubscriptionStore) DeleteAllForUser(userID string) error {
	return s.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(bucketPushSubscriptions)
		var toDelete [][]byte
		c := b.Cursor()
		for k, v := c.First(); k != nil; k, v = c.Next() {
			var sub notifications.PushSubscription
			if err := json.Unmarshal(v, &sub); err != nil {
				return err
			}
			if sub.UserID == userID {
				toDelete = append(toDelete, append([]byte(nil), k...))
			}
		}
		for _, k := range toDelete {
			if err := b.Delete(k); err != nil {
				return err
			}
		}
		return nil
	})
}

func (s *PushSubscriptionStore) DeleteByEndpoint(endpoint string) error {
	return s.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(bucketPushSubscriptions)
		var toDelete [][]byte
		c := b.Cursor()
		for k, v := c.First(); k != nil; k, v = c.Next() {
			var sub notifications.PushSubscription
			if err := json.Unmarshal(v, &sub); err != nil {
				return err
			}
			if sub.Endpoint == endpoint {
				toDelete = append(toDelete, append([]byte(nil), k...))
			}
		}
		for _, k := range toDelete {
			if err := b.Delete(k); err != nil {
				return err
			}
		}
		return nil
	})
}

// ListAll returns every push subscription (migration).
func (s *PushSubscriptionStore) ListAll() ([]notifications.PushSubscription, error) {
	var out []notifications.PushSubscription
	err := s.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(bucketPushSubscriptions)
		return b.ForEach(func(_, v []byte) error {
			var sub notifications.PushSubscription
			if err := json.Unmarshal(v, &sub); err != nil {
				return err
			}
			out = append(out, sub)
			return nil
		})
	})
	return out, err
}

// PutExisting writes a subscription as-is (storage migration).
func (s *PushSubscriptionStore) PutExisting(sub notifications.PushSubscription) error {
	if sub.UserID == "" || sub.InstallationID == "" {
		return fmt.Errorf("user id and installation id are required")
	}
	if sub.UpdatedAt.IsZero() {
		sub.UpdatedAt = time.Now().UTC()
	}
	return s.Upsert(sub)
}

var _ notifications.Repository = (*PushSubscriptionStore)(nil)
