package bbolt

import (
	"fmt"
	"strings"

	bolt "go.etcd.io/bbolt"

	"github.com/solargate/grom/internal/federation"
	"github.com/solargate/grom/internal/storage/keys"
	"github.com/solargate/grom/internal/workouts"
)

type WorkoutLikesStore struct {
	db       *bolt.DB
	workouts *WorkoutsStore
}

// FederatedLikeEntry is a federated likes cache record (migration).
type FederatedLikeEntry struct {
	ViewerNickname string
	OwnerHandle    string
	WorkoutID      string
	Likes          workouts.WorkoutLikes
}

// LikeActivityEntry maps an outbound Like activity id (migration).
type LikeActivityEntry struct {
	ActorNickname string
	ObjectID      string
	ActivityID    string
}

func NewWorkoutLikesStore(db *bolt.DB, workoutStore *WorkoutsStore) *WorkoutLikesStore {
	return &WorkoutLikesStore{db: db, workouts: workoutStore}
}

func workoutLikesKey(ownerNickname, dirName string) []byte {
	return []byte(ownerNickname + "/" + dirName)
}

func fedWorkoutLikesKey(viewerNickname, ownerHandle, workoutID string) []byte {
	return []byte(viewerNickname + "/" + federation.OwnerKeyFromHandle(ownerHandle) + "/" + workoutID)
}

func likeActivityKey(actorNickname, objectID string) []byte {
	return []byte(actorNickname + "|" + objectID)
}

func getLikesFromBucket(tx *bolt.Tx, bucketName, key []byte) (*workouts.WorkoutLikes, error) {
	raw := tx.Bucket(bucketName).Get(key)
	if raw == nil {
		empty := workouts.NormalizeWorkoutLikes(nil)
		return &empty, nil
	}
	var likes workouts.WorkoutLikes
	if err := unmarshalJSON(raw, &likes); err != nil {
		return nil, fmt.Errorf("parse likes: %w", err)
	}
	norm := workouts.NormalizeWorkoutLikes(&likes)
	return &norm, nil
}

func putLikesToBucket(tx *bolt.Tx, bucketName, key []byte, likes *workouts.WorkoutLikes) error {
	norm := workouts.NormalizeWorkoutLikes(likes)
	raw, err := marshalJSON(&norm)
	if err != nil {
		return err
	}
	return tx.Bucket(bucketName).Put(key, raw)
}

func (s *WorkoutLikesStore) GetLocal(ownerNickname, workoutID string) (*workouts.WorkoutLikes, error) {
	var likes *workouts.WorkoutLikes
	err := s.db.View(func(tx *bolt.Tx) error {
		owner, key, err := s.workouts.lookupPrimaryKey(tx, workoutID)
		if err != nil {
			return err
		}
		if owner != ownerNickname {
			return workouts.ErrWorkoutNotFound
		}
		workout, err := s.workouts.getByPrimaryKey(tx, key)
		if err != nil {
			return err
		}
		dirName := keys.WorkoutDirName(workout.StartDate, workout.ID)
		likes, err = getLikesFromBucket(tx, bucketWorkoutLikes, workoutLikesKey(ownerNickname, dirName))
		return err
	})
	return likes, err
}

func (s *WorkoutLikesStore) PutLocal(ownerNickname, workoutID string, likes *workouts.WorkoutLikes) error {
	return s.db.Update(func(tx *bolt.Tx) error {
		owner, key, err := s.workouts.lookupPrimaryKey(tx, workoutID)
		if err != nil {
			return err
		}
		if owner != ownerNickname {
			return workouts.ErrWorkoutNotFound
		}
		workout, err := s.workouts.getByPrimaryKey(tx, key)
		if err != nil {
			return err
		}
		dirName := keys.WorkoutDirName(workout.StartDate, workout.ID)
		return putLikesToBucket(tx, bucketWorkoutLikes, workoutLikesKey(ownerNickname, dirName), likes)
	})
}

func (s *WorkoutLikesStore) DeleteLocal(ownerNickname, workoutID string) error {
	return s.db.Update(func(tx *bolt.Tx) error {
		owner, key, err := s.workouts.lookupPrimaryKey(tx, workoutID)
		if err != nil {
			if err == workouts.ErrWorkoutNotFound {
				return nil
			}
			return err
		}
		if owner != ownerNickname {
			return nil
		}
		workout, err := s.workouts.getByPrimaryKey(tx, key)
		if err != nil {
			return err
		}
		dirName := keys.WorkoutDirName(workout.StartDate, workout.ID)
		return tx.Bucket(bucketWorkoutLikes).Delete(workoutLikesKey(ownerNickname, dirName))
	})
}

func (s *WorkoutLikesStore) GetFederated(viewerNickname, ownerHandle, workoutID string) (*workouts.WorkoutLikes, error) {
	var likes *workouts.WorkoutLikes
	err := s.db.View(func(tx *bolt.Tx) error {
		var err error
		likes, err = getLikesFromBucket(tx, bucketFedWorkoutLikes, fedWorkoutLikesKey(viewerNickname, ownerHandle, workoutID))
		return err
	})
	return likes, err
}

func (s *WorkoutLikesStore) PutFederated(viewerNickname, ownerHandle, workoutID string, likes *workouts.WorkoutLikes) error {
	return s.db.Update(func(tx *bolt.Tx) error {
		return putLikesToBucket(tx, bucketFedWorkoutLikes, fedWorkoutLikesKey(viewerNickname, ownerHandle, workoutID), likes)
	})
}

func (s *WorkoutLikesStore) DeleteFederated(viewerNickname, ownerHandle, workoutID string) error {
	return s.db.Update(func(tx *bolt.Tx) error {
		return tx.Bucket(bucketFedWorkoutLikes).Delete(fedWorkoutLikesKey(viewerNickname, ownerHandle, workoutID))
	})
}

func (s *WorkoutLikesStore) GetLikeActivityID(actorNickname, objectID string) (string, error) {
	var activityID string
	err := s.db.View(func(tx *bolt.Tx) error {
		raw := tx.Bucket(bucketLikeActivities).Get(likeActivityKey(actorNickname, objectID))
		if raw != nil {
			activityID = string(raw)
		}
		return nil
	})
	return activityID, err
}

func (s *WorkoutLikesStore) PutLikeActivityID(actorNickname, objectID, activityID string) error {
	return s.db.Update(func(tx *bolt.Tx) error {
		return tx.Bucket(bucketLikeActivities).Put(likeActivityKey(actorNickname, objectID), []byte(activityID))
	})
}

func (s *WorkoutLikesStore) DeleteLikeActivityID(actorNickname, objectID string) error {
	return s.db.Update(func(tx *bolt.Tx) error {
		return tx.Bucket(bucketLikeActivities).Delete(likeActivityKey(actorNickname, objectID))
	})
}

// ListAllFederated returns every federated likes cache record (migration).
func (s *WorkoutLikesStore) ListAllFederated() ([]FederatedLikeEntry, error) {
	var out []FederatedLikeEntry
	err := s.db.View(func(tx *bolt.Tx) error {
		return tx.Bucket(bucketFedWorkoutLikes).ForEach(func(k, v []byte) error {
			parts := strings.SplitN(string(k), "/", 3)
			if len(parts) != 3 {
				return nil
			}
			var likes workouts.WorkoutLikes
			if err := unmarshalJSON(v, &likes); err != nil {
				return fmt.Errorf("parse federated likes: %w", err)
			}
			norm := workouts.NormalizeWorkoutLikes(&likes)
			if norm.Likes == 0 {
				return nil
			}
			out = append(out, FederatedLikeEntry{
				ViewerNickname: parts[0],
				OwnerHandle:    federation.OwnerHandleFromKey(parts[1]),
				WorkoutID:      parts[2],
				Likes:          norm,
			})
			return nil
		})
	})
	return out, err
}

// ListAllLikeActivities returns every outbound Like activity id (migration).
func (s *WorkoutLikesStore) ListAllLikeActivities() ([]LikeActivityEntry, error) {
	var out []LikeActivityEntry
	err := s.db.View(func(tx *bolt.Tx) error {
		return tx.Bucket(bucketLikeActivities).ForEach(func(k, v []byte) error {
			parts := strings.SplitN(string(k), "|", 2)
			if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
				return nil
			}
			out = append(out, LikeActivityEntry{
				ActorNickname: parts[0],
				ObjectID:      parts[1],
				ActivityID:    string(v),
			})
			return nil
		})
	})
	return out, err
}

var _ workouts.LikesRepository = (*WorkoutLikesStore)(nil)
