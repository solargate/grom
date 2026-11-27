package bbolt

import (
	"strings"

	"github.com/solargate/grom/internal/users"
	"github.com/solargate/grom/internal/workouts"
	bolt "go.etcd.io/bbolt"
)

func (b *Backend) GetProfileWithWorkoutTotal(userID string) (*users.Profile, error) {
	profile, err := b.users.GetProfile(userID)
	if err != nil || profile.HasWorkoutStats() {
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
		if profile.HasWorkoutStats() {
			return nil
		}
		if err := rebuildWorkoutStats(tx, user.Nickname, profile); err != nil {
			return err
		}
		return b.users.putProfile(tx, userID, *profile)
	})
	return profile, err
}

// updateWorkoutStats runs after the metadata change in the same transaction.
func (s *WorkoutsStore) updateWorkoutStats(tx *bolt.Tx, nickname string, before, after *workouts.Workout) error {
	userID := tx.Bucket(bucketIdxUsersNick).Get([]byte(strings.ToLower(nickname)))
	if userID == nil {
		return nil
	}
	store := UsersStore{db: s.db, dataDir: s.dataDir}
	profile, err := store.getProfile(tx, string(userID))
	if err != nil {
		return err
	}
	if profile.HasWorkoutStats() && users.SameWorkoutSummary(before, after) {
		return nil
	}
	if !profile.ApplyWorkoutChange(before, after) {
		if err := rebuildWorkoutStats(tx, nickname, &profile); err != nil {
			return err
		}
	}
	return store.putProfile(tx, string(userID), profile)
}

func rebuildWorkoutStats(tx *bolt.Tx, nickname string, profile *users.Profile) error {
	items, err := listWorkouts(tx, nickname)
	if err != nil {
		return err
	}
	profile.RebuildWorkoutStats(items)
	return nil
}
