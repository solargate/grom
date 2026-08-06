package bbolt

import (
	"fmt"
	"strings"

	bolt "go.etcd.io/bbolt"

	"github.com/solargate/grom/internal/federation"
	"github.com/solargate/grom/internal/storage/keys"
	"github.com/solargate/grom/internal/workouts"
)

type WorkoutCommentsStore struct {
	db       *bolt.DB
	workouts *WorkoutsStore
}

// FederatedCommentEntry is a federated comments cache record (migration).
type FederatedCommentEntry struct {
	ViewerNickname string
	OwnerHandle    string
	WorkoutID      string
	Comments       workouts.WorkoutComments
}

// CommentActivityEntry maps an outbound Create Note activity id (migration).
type CommentActivityEntry struct {
	ActorNickname string
	NoteID        string
	ActivityID    string
}

func NewWorkoutCommentsStore(db *bolt.DB, workoutStore *WorkoutsStore) *WorkoutCommentsStore {
	return &WorkoutCommentsStore{db: db, workouts: workoutStore}
}

func workoutCommentsKey(ownerNickname, dirName string) []byte {
	return []byte(ownerNickname + "/" + dirName)
}

func fedWorkoutCommentsKey(viewerNickname, ownerHandle, workoutID string) []byte {
	return []byte(viewerNickname + "/" + federation.OwnerKeyFromHandle(ownerHandle) + "/" + workoutID)
}

func commentActivityKey(actorNickname, noteID string) []byte {
	return []byte(actorNickname + "|" + noteID)
}

func getCommentsFromBucket(tx *bolt.Tx, bucketName, key []byte) (*workouts.WorkoutComments, error) {
	raw := tx.Bucket(bucketName).Get(key)
	if raw == nil {
		empty := workouts.NormalizeWorkoutComments(nil)
		return &empty, nil
	}
	var comments workouts.WorkoutComments
	if err := unmarshalJSON(raw, &comments); err != nil {
		return nil, fmt.Errorf("parse comments: %w", err)
	}
	norm := workouts.NormalizeWorkoutComments(&comments)
	return &norm, nil
}

func putCommentsToBucket(tx *bolt.Tx, bucketName, key []byte, comments *workouts.WorkoutComments) error {
	norm := workouts.NormalizeWorkoutComments(comments)
	raw, err := marshalJSON(&norm)
	if err != nil {
		return err
	}
	return tx.Bucket(bucketName).Put(key, raw)
}

func (s *WorkoutCommentsStore) GetLocal(ownerNickname, workoutID string) (*workouts.WorkoutComments, error) {
	var comments *workouts.WorkoutComments
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
		comments, err = getCommentsFromBucket(tx, bucketWorkoutComments, workoutCommentsKey(ownerNickname, dirName))
		return err
	})
	return comments, err
}

func (s *WorkoutCommentsStore) PutLocal(ownerNickname, workoutID string, comments *workouts.WorkoutComments) error {
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
		return putCommentsToBucket(tx, bucketWorkoutComments, workoutCommentsKey(ownerNickname, dirName), comments)
	})
}

func (s *WorkoutCommentsStore) DeleteLocal(ownerNickname, workoutID string) error {
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
		return tx.Bucket(bucketWorkoutComments).Delete(workoutCommentsKey(ownerNickname, dirName))
	})
}

func (s *WorkoutCommentsStore) GetFederated(viewerNickname, ownerHandle, workoutID string) (*workouts.WorkoutComments, error) {
	var comments *workouts.WorkoutComments
	err := s.db.View(func(tx *bolt.Tx) error {
		var err error
		comments, err = getCommentsFromBucket(tx, bucketFedWorkoutComments, fedWorkoutCommentsKey(viewerNickname, ownerHandle, workoutID))
		return err
	})
	return comments, err
}

func (s *WorkoutCommentsStore) PutFederated(viewerNickname, ownerHandle, workoutID string, comments *workouts.WorkoutComments) error {
	return s.db.Update(func(tx *bolt.Tx) error {
		return putCommentsToBucket(tx, bucketFedWorkoutComments, fedWorkoutCommentsKey(viewerNickname, ownerHandle, workoutID), comments)
	})
}

func (s *WorkoutCommentsStore) DeleteFederated(viewerNickname, ownerHandle, workoutID string) error {
	return s.db.Update(func(tx *bolt.Tx) error {
		return tx.Bucket(bucketFedWorkoutComments).Delete(fedWorkoutCommentsKey(viewerNickname, ownerHandle, workoutID))
	})
}

func (s *WorkoutCommentsStore) GetCommentActivityID(actorNickname, noteID string) (string, error) {
	var activityID string
	err := s.db.View(func(tx *bolt.Tx) error {
		raw := tx.Bucket(bucketCommentActivities).Get(commentActivityKey(actorNickname, noteID))
		if raw != nil {
			activityID = string(raw)
		}
		return nil
	})
	return activityID, err
}

func (s *WorkoutCommentsStore) PutCommentActivityID(actorNickname, noteID, activityID string) error {
	return s.db.Update(func(tx *bolt.Tx) error {
		return tx.Bucket(bucketCommentActivities).Put(commentActivityKey(actorNickname, noteID), []byte(activityID))
	})
}

func (s *WorkoutCommentsStore) DeleteCommentActivityID(actorNickname, noteID string) error {
	return s.db.Update(func(tx *bolt.Tx) error {
		return tx.Bucket(bucketCommentActivities).Delete(commentActivityKey(actorNickname, noteID))
	})
}

// ListAllFederated returns every federated comments cache record (migration).
func (s *WorkoutCommentsStore) ListAllFederated() ([]FederatedCommentEntry, error) {
	var out []FederatedCommentEntry
	err := s.db.View(func(tx *bolt.Tx) error {
		return tx.Bucket(bucketFedWorkoutComments).ForEach(func(k, v []byte) error {
			parts := strings.SplitN(string(k), "/", 3)
			if len(parts) != 3 {
				return nil
			}
			var comments workouts.WorkoutComments
			if err := unmarshalJSON(v, &comments); err != nil {
				return fmt.Errorf("parse federated comments: %w", err)
			}
			norm := workouts.NormalizeWorkoutComments(&comments)
			if norm.CommentsNum == 0 {
				return nil
			}
			out = append(out, FederatedCommentEntry{
				ViewerNickname: parts[0],
				OwnerHandle:    federation.OwnerHandleFromKey(parts[1]),
				WorkoutID:      parts[2],
				Comments:       norm,
			})
			return nil
		})
	})
	return out, err
}

// ListAllCommentActivities returns every outbound Create Note activity id (migration).
func (s *WorkoutCommentsStore) ListAllCommentActivities() ([]CommentActivityEntry, error) {
	var out []CommentActivityEntry
	err := s.db.View(func(tx *bolt.Tx) error {
		return tx.Bucket(bucketCommentActivities).ForEach(func(k, v []byte) error {
			parts := strings.SplitN(string(k), "|", 2)
			if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
				return nil
			}
			out = append(out, CommentActivityEntry{
				ActorNickname: parts[0],
				NoteID:        parts[1],
				ActivityID:    string(v),
			})
			return nil
		})
	})
	return out, err
}

var _ workouts.CommentsRepository = (*WorkoutCommentsStore)(nil)
