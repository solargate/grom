package bbolt

import (
	"fmt"

	bolt "go.etcd.io/bbolt"

	"github.com/solargate/grom/internal/federation"
	"github.com/solargate/grom/internal/storage/keys"
	"github.com/solargate/grom/internal/workouts"
)

type WorkoutLikesStore struct {
	db       *bolt.DB
	workouts *WorkoutsStore
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

var _ workouts.LikesRepository = (*WorkoutLikesStore)(nil)
