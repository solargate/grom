package bbolt

import (
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
	bolt "go.etcd.io/bbolt"

	"github.com/solargate/grom/internal/social"
)

type SocialStore struct {
	db *bolt.DB
}

func NewSocialStore(db *bolt.DB) *SocialStore {
	return &SocialStore{db: db}
}

func (s *SocialStore) getFollow(tx *bolt.Tx, id string) (*social.Follow, error) {
	raw := tx.Bucket(bucketFollows).Get([]byte(id))
	if raw == nil {
		return nil, social.ErrFollowNotFound
	}
	var f social.Follow
	if err := unmarshalJSON(raw, &f); err != nil {
		return nil, fmt.Errorf("parse follow: %w", err)
	}
	return &f, nil
}

func (s *SocialStore) putFollow(tx *bolt.Tx, f social.Follow) error {
	raw, err := marshalJSON(f)
	if err != nil {
		return err
	}
	if err := tx.Bucket(bucketFollows).Put([]byte(f.ID), raw); err != nil {
		return err
	}
	followerKey := []byte(f.FollowerID + "/" + f.ID)
	if err := tx.Bucket(bucketIdxFollowsFollower).Put(followerKey, []byte(f.ID)); err != nil {
		return err
	}
	targetKey := []byte(f.TargetHandle + "/" + f.ID)
	if err := tx.Bucket(bucketIdxFollowsTarget).Put(targetKey, []byte(f.ID)); err != nil {
		return err
	}
	if f.FollowActivityID != "" {
		if err := tx.Bucket(bucketIdxFollowsActivity).Put([]byte(f.FollowActivityID), []byte(f.ID)); err != nil {
			return err
		}
	}
	return nil
}

func (s *SocialStore) deleteFollowIndexes(tx *bolt.Tx, f social.Follow) error {
	_ = tx.Bucket(bucketIdxFollowsFollower).Delete([]byte(f.FollowerID + "/" + f.ID))
	_ = tx.Bucket(bucketIdxFollowsTarget).Delete([]byte(f.TargetHandle + "/" + f.ID))
	if f.FollowActivityID != "" {
		_ = tx.Bucket(bucketIdxFollowsActivity).Delete([]byte(f.FollowActivityID))
	}
	return nil
}

func (s *SocialStore) FindByID(id string) (*social.Follow, error) {
	var result *social.Follow
	err := s.db.View(func(tx *bolt.Tx) error {
		f, err := s.getFollow(tx, id)
		if err != nil {
			return err
		}
		result = f
		return nil
	})
	return result, err
}

func (s *SocialStore) ListByFollower(followerID string) ([]social.Follow, error) {
	prefix := []byte(followerID + "/")
	var result []social.Follow
	err := s.db.View(func(tx *bolt.Tx) error {
		c := tx.Bucket(bucketIdxFollowsFollower).Cursor()
		for k, id := c.Seek(prefix); k != nil && strings.HasPrefix(string(k), string(prefix)); k, id = c.Next() {
			f, err := s.getFollow(tx, string(id))
			if err != nil {
				return err
			}
			result = append(result, *f)
		}
		return nil
	})
	return result, err
}

func (s *SocialStore) ListActiveFollowing(followerID string) ([]social.Follow, error) {
	all, err := s.ListByFollower(followerID)
	if err != nil {
		return nil, err
	}
	result := make([]social.Follow, 0, len(all))
	for i := range all {
		if all[i].Status == social.StatusActive {
			result = append(result, all[i])
		}
	}
	return result, nil
}

func (s *SocialStore) ListActiveByTarget(targetHandle string) ([]social.Follow, error) {
	prefix := []byte(targetHandle + "/")
	var result []social.Follow
	err := s.db.View(func(tx *bolt.Tx) error {
		c := tx.Bucket(bucketIdxFollowsTarget).Cursor()
		for k, id := c.Seek(prefix); k != nil && strings.HasPrefix(string(k), string(prefix)); k, id = c.Next() {
			f, err := s.getFollow(tx, string(id))
			if err != nil {
				return err
			}
			if f.Status == social.StatusActive {
				result = append(result, *f)
			}
		}
		return nil
	})
	return result, err
}

func (s *SocialStore) FindExisting(followerID, targetHandle string) (*social.Follow, error) {
	all, err := s.ListByFollower(followerID)
	if err != nil {
		return nil, err
	}
	for i := range all {
		f := all[i]
		if f.TargetHandle == targetHandle && f.Status != social.StatusRejected {
			return &f, nil
		}
	}
	return nil, social.ErrFollowNotFound
}

func (s *SocialStore) Create(follow social.Follow) (*social.Follow, error) {
	if follow.ID == "" {
		follow.ID = uuid.NewString()
	}
	err := s.db.Update(func(tx *bolt.Tx) error {
		prefix := []byte(follow.FollowerID + "/")
		c := tx.Bucket(bucketIdxFollowsFollower).Cursor()
		for k, id := c.Seek(prefix); k != nil && strings.HasPrefix(string(k), string(prefix)); k, id = c.Next() {
			f, err := s.getFollow(tx, string(id))
			if err != nil {
				return err
			}
			if f.TargetHandle == follow.TargetHandle && f.Status != social.StatusRejected {
				return social.ErrAlreadyFollowing
			}
		}
		return s.putFollow(tx, follow)
	})
	if err != nil {
		return nil, err
	}
	result := follow
	return &result, nil
}

func (s *SocialStore) UpdateStatus(id, status string) (*social.Follow, error) {
	var result *social.Follow
	err := s.db.Update(func(tx *bolt.Tx) error {
		f, err := s.getFollow(tx, id)
		if err != nil {
			return err
		}
		_ = s.deleteFollowIndexes(tx, *f)
		f.Status = status
		if err := s.putFollow(tx, *f); err != nil {
			return err
		}
		result = f
		return nil
	})
	return result, err
}

func (s *SocialStore) FindByFollowActivityID(activityID string) (*social.Follow, error) {
	var result *social.Follow
	err := s.db.View(func(tx *bolt.Tx) error {
		id := tx.Bucket(bucketIdxFollowsActivity).Get([]byte(activityID))
		if id == nil {
			return social.ErrFollowNotFound
		}
		f, err := s.getFollow(tx, string(id))
		if err != nil {
			return err
		}
		result = f
		return nil
	})
	return result, err
}

func (s *SocialStore) UpdateActivityID(id, activityID string) (*social.Follow, error) {
	var result *social.Follow
	err := s.db.Update(func(tx *bolt.Tx) error {
		f, err := s.getFollow(tx, id)
		if err != nil {
			return err
		}
		_ = s.deleteFollowIndexes(tx, *f)
		f.FollowActivityID = activityID
		if err := s.putFollow(tx, *f); err != nil {
			return err
		}
		result = f
		return nil
	})
	return result, err
}

func (s *SocialStore) Delete(id string) error {
	return s.db.Update(func(tx *bolt.Tx) error {
		f, err := s.getFollow(tx, id)
		if err != nil {
			return err
		}
		_ = s.deleteFollowIndexes(tx, *f)
		return tx.Bucket(bucketFollows).Delete([]byte(id))
	})
}

func (s *SocialStore) DeleteInvolving(followerID, localHandle string) error {
	return s.db.Update(func(tx *bolt.Tx) error {
		var toDelete []string
		c := tx.Bucket(bucketFollows).Cursor()
		for k, v := c.First(); k != nil; k, v = c.Next() {
			var f social.Follow
			if err := unmarshalJSON(v, &f); err != nil {
				return err
			}
			if f.FollowerID == followerID {
				toDelete = append(toDelete, f.ID)
				continue
			}
			if localHandle != "" && strings.EqualFold(f.TargetHandle, localHandle) {
				toDelete = append(toDelete, f.ID)
			}
		}
		for _, id := range toDelete {
			f, err := s.getFollow(tx, id)
			if err != nil {
				if errors.Is(err, social.ErrFollowNotFound) {
					continue
				}
				return err
			}
			_ = s.deleteFollowIndexes(tx, *f)
			if err := tx.Bucket(bucketFollows).Delete([]byte(id)); err != nil {
				return err
			}
		}
		return nil
	})
}

// PutExisting writes a follow as-is (migration).
func (s *SocialStore) PutExisting(follow social.Follow) error {
	return s.db.Update(func(tx *bolt.Tx) error {
		return s.putFollow(tx, follow)
	})
}

func (s *SocialStore) ListAll() ([]social.Follow, error) {
	var result []social.Follow
	err := s.db.View(func(tx *bolt.Tx) error {
		return tx.Bucket(bucketFollows).ForEach(func(_, v []byte) error {
			var f social.Follow
			if err := unmarshalJSON(v, &f); err != nil {
				return err
			}
			result = append(result, f)
			return nil
		})
	})
	return result, err
}

var _ social.Repository = (*SocialStore)(nil)
