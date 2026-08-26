package bbolt

import (
	"strings"

	bolt "go.etcd.io/bbolt"

	"github.com/solargate/grom/internal/federation"
)

type FederationFollowersStore struct {
	db *bolt.DB
}

func NewFederationFollowersStore(db *bolt.DB) *FederationFollowersStore {
	return &FederationFollowersStore{db: db}
}

func fedFollowerKey(nickname, actorURI string) []byte {
	return []byte(nickname + "/" + actorURI)
}

func fedFollowerPrefix(nickname string) []byte {
	return []byte(nickname + "/")
}

func (s *FederationFollowersStore) Add(nickname string, follower federation.InboundFollower) error {
	raw, err := marshalJSON(follower)
	if err != nil {
		return err
	}
	return s.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(bucketFedFollowers)
		key := fedFollowerKey(nickname, follower.ActorURI)
		return b.Put(key, raw)
	})
}

func (s *FederationFollowersStore) List(nickname string) ([]federation.InboundFollower, error) {
	prefix := fedFollowerPrefix(nickname)
	var result []federation.InboundFollower
	err := s.db.View(func(tx *bolt.Tx) error {
		c := tx.Bucket(bucketFedFollowers).Cursor()
		for k, v := c.Seek(prefix); k != nil && strings.HasPrefix(string(k), string(prefix)); k, v = c.Next() {
			var f federation.InboundFollower
			if err := unmarshalJSON(v, &f); err != nil {
				return err
			}
			result = append(result, f)
		}
		return nil
	})
	return result, err
}

func (s *FederationFollowersStore) ListInboxes(nickname string) ([]string, error) {
	followers, err := s.List(nickname)
	if err != nil {
		return nil, err
	}
	return federation.DeduplicateDeliveryInboxes(followers), nil
}

func (s *FederationFollowersStore) Remove(nickname, actorURI string) error {
	return s.db.Update(func(tx *bolt.Tx) error {
		return tx.Bucket(bucketFedFollowers).Delete(fedFollowerKey(nickname, actorURI))
	})
}

// ListAllNicknames returns distinct nicknames that have followers (migration helper).
func (s *FederationFollowersStore) ListAll() (map[string][]federation.InboundFollower, error) {
	out := make(map[string][]federation.InboundFollower)
	err := s.db.View(func(tx *bolt.Tx) error {
		return tx.Bucket(bucketFedFollowers).ForEach(func(k, v []byte) error {
			parts := strings.SplitN(string(k), "/", 2)
			if len(parts) != 2 {
				return nil
			}
			var f federation.InboundFollower
			if err := unmarshalJSON(v, &f); err != nil {
				return err
			}
			out[parts[0]] = append(out[parts[0]], f)
			return nil
		})
	})
	return out, err
}

var _ federation.FollowersRepository = (*FederationFollowersStore)(nil)
