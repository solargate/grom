package bbolt

import (
	"bytes"
	"strings"

	"github.com/solargate/grom/internal/users"
	bolt "go.etcd.io/bbolt"
)

func (b *Backend) GetProfileWithWorkoutTotal(userID string) (*users.Profile, error) {
	profile, err := b.users.GetProfile(userID)
	if err != nil || profile.TotalWorkouts != nil {
		return profile, err
	}
	err = b.db.Update(func(tx *bolt.Tx) error {
		user, err := b.users.getByID(tx, userID)
		if err != nil {
			return err
		}
		current, err := b.users.getProfile(tx, userID)
		if err != nil {
			return err
		}
		profile = &current
		if profile.TotalWorkouts != nil {
			return nil
		}
		total := countWorkouts(tx, user.Nickname)
		profile.TotalWorkouts = &total
		return b.users.putProfile(tx, userID, *profile)
	})
	return profile, err
}

// updateWorkoutTotal runs after the metadata change in the same transaction.
func (s *WorkoutsStore) updateWorkoutTotal(tx *bolt.Tx, nickname string, delta int) error {
	userID := tx.Bucket(bucketIdxUsersNick).Get([]byte(strings.ToLower(nickname)))
	if userID == nil {
		return nil
	}
	store := UsersStore{db: s.db, dataDir: s.dataDir}
	profile, err := store.getProfile(tx, string(userID))
	if err != nil {
		return err
	}
	var total int
	if profile.TotalWorkouts != nil && *profile.TotalWorkouts+delta >= 0 {
		total = *profile.TotalWorkouts + delta
	} else {
		total = countWorkouts(tx, nickname)
	}
	profile.TotalWorkouts = &total
	return store.putProfile(tx, string(userID), profile)
}

func countWorkouts(tx *bolt.Tx, nickname string) int {
	prefix := []byte(nickname + "/")
	cursor := tx.Bucket(bucketWorkouts).Cursor()
	total := 0
	for key, _ := cursor.Seek(prefix); bytes.HasPrefix(key, prefix); key, _ = cursor.Next() {
		total++
	}
	return total
}
