package bbolt

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	bolt "go.etcd.io/bbolt"

	"github.com/solargate/grom/internal/data"
	"github.com/solargate/grom/internal/storage/keys"
	"github.com/solargate/grom/internal/workouts"
)

const (
	workoutIDLength   = workouts.WorkoutIDLength
	workoutIDAlphabet = "0123456789"
)

type workoutIndexRef struct {
	Nickname string `json:"nickname"`
	Key      string `json:"key"`
}

type WorkoutsStore struct {
	db      *bolt.DB
	dataDir string
}

func NewWorkoutsStore(db *bolt.DB, dataDir string) *WorkoutsStore {
	return &WorkoutsStore{db: db, dataDir: dataDir}
}

func workoutPrimaryKey(nickname string, startDate time.Time, id string) string {
	return nickname + "/" + keys.WorkoutDirName(startDate, id)
}

func workoutDirPath(dataDir, nickname string, startDate time.Time, id string) string {
	return filepath.Join(data.UserDir(dataDir, nickname), "workouts", keys.WorkoutDirName(startDate, id))
}

func (s *WorkoutsStore) getByPrimaryKey(tx *bolt.Tx, primaryKey string) (*workouts.Workout, error) {
	raw := tx.Bucket(bucketWorkouts).Get([]byte(primaryKey))
	if raw == nil {
		return nil, workouts.ErrWorkoutNotFound
	}
	var w workouts.Workout
	if err := unmarshalJSON(raw, &w); err != nil {
		return nil, fmt.Errorf("parse workout: %w", err)
	}
	return &w, nil
}

func (s *WorkoutsStore) lookupPrimaryKey(tx *bolt.Tx, workoutID string) (string, string, error) {
	raw := tx.Bucket(bucketIdxWorkoutsID).Get([]byte(workoutID))
	if raw == nil {
		return "", "", workouts.ErrWorkoutNotFound
	}
	var ref workoutIndexRef
	if err := unmarshalJSON(raw, &ref); err != nil {
		return "", "", fmt.Errorf("parse workout index: %w", err)
	}
	return ref.Nickname, ref.Key, nil
}

func (s *WorkoutsStore) putWorkout(tx *bolt.Tx, nickname string, w *workouts.Workout) error {
	primaryKey := workoutPrimaryKey(nickname, w.StartDate, w.ID)
	raw, err := marshalJSON(w)
	if err != nil {
		return err
	}
	if err := tx.Bucket(bucketWorkouts).Put([]byte(primaryKey), raw); err != nil {
		return err
	}
	ref, err := marshalJSON(workoutIndexRef{Nickname: nickname, Key: primaryKey})
	if err != nil {
		return err
	}
	if err := tx.Bucket(bucketIdxWorkoutsID).Put([]byte(w.ID), ref); err != nil {
		return err
	}
	if sid := strings.TrimSpace(w.StravaActivityID); sid != "" {
		stravaKey := []byte(nickname + "/" + sid)
		if err := tx.Bucket(bucketIdxWorkoutsStrava).Put(stravaKey, []byte(w.ID)); err != nil {
			return err
		}
	}
	return nil
}

func (s *WorkoutsStore) deleteWorkoutMeta(tx *bolt.Tx, nickname string, w *workouts.Workout) error {
	primaryKey := workoutPrimaryKey(nickname, w.StartDate, w.ID)
	dirName := keys.WorkoutDirName(w.StartDate, w.ID)
	if err := DeleteLocalSpeedChartInTx(tx, nickname, dirName); err != nil {
		return err
	}
	_ = tx.Bucket(bucketWorkouts).Delete([]byte(primaryKey))
	_ = tx.Bucket(bucketIdxWorkoutsID).Delete([]byte(w.ID))
	if sid := strings.TrimSpace(w.StravaActivityID); sid != "" {
		_ = tx.Bucket(bucketIdxWorkoutsStrava).Delete([]byte(nickname + "/" + sid))
	}
	return nil
}

func newWorkoutID() (string, error) {
	b := make([]byte, workoutIDLength)
	alphabetSize := big.NewInt(int64(len(workoutIDAlphabet)))
	for i := range b {
		n, err := rand.Int(rand.Reader, alphabetSize)
		if err != nil {
			return "", fmt.Errorf("generate workout id: %w", err)
		}
		b[i] = workoutIDAlphabet[n.Int64()]
	}
	return string(b), nil
}

func (s *WorkoutsStore) workoutIDExists(tx *bolt.Tx, id string) bool {
	return tx.Bucket(bucketIdxWorkoutsID).Get([]byte(id)) != nil
}

func (s *WorkoutsStore) allocateWorkoutID(tx *bolt.Tx) (string, error) {
	const maxAttempts = 10
	for range maxAttempts {
		id, err := newWorkoutID()
		if err != nil {
			return "", err
		}
		if !s.workoutIDExists(tx, id) {
			return id, nil
		}
	}
	return "", fmt.Errorf("failed to allocate unique workout id after %d attempts", maxAttempts)
}

func (s *WorkoutsStore) Create(nickname string, workout *workouts.Workout) (*workouts.Workout, error) {
	workout.Name = workouts.TrimWorkoutName(workout.Name)
	workout.Description = workouts.TrimWorkoutDescription(workout.Description)
	workout.Device = workouts.DeviceGrom

	if err := os.MkdirAll(data.UserDir(s.dataDir, nickname), 0700); err != nil {
		return nil, fmt.Errorf("create user dir: %w", err)
	}

	err := s.db.Update(func(tx *bolt.Tx) error {
		id, err := s.allocateWorkoutID(tx)
		if err != nil {
			return err
		}
		workout.ID = id
		primaryKey := workoutPrimaryKey(nickname, workout.StartDate, workout.ID)
		if tx.Bucket(bucketWorkouts).Get([]byte(primaryKey)) != nil {
			return workouts.ErrWorkoutExists
		}
		return s.putWorkout(tx, nickname, workout)
	})
	if err != nil {
		return nil, err
	}
	dir := workoutDirPath(s.dataDir, nickname, workout.StartDate, workout.ID)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return nil, fmt.Errorf("create workout dir: %w", err)
	}
	result := *workout
	return &result, nil
}

func (s *WorkoutsStore) BeginCreate(nickname string, workout *workouts.Workout) (*workouts.Workout, string, func(), error) {
	workout.Name = workouts.TrimWorkoutName(workout.Name)
	workout.Description = workouts.TrimWorkoutDescription(workout.Description)

	if err := os.MkdirAll(data.UserDir(s.dataDir, nickname), 0700); err != nil {
		return nil, "", nil, fmt.Errorf("create user dir: %w", err)
	}

	var allocatedID string
	err := s.db.Update(func(tx *bolt.Tx) error {
		id, err := s.allocateWorkoutID(tx)
		if err != nil {
			return err
		}
		allocatedID = id
		return nil
	})
	if err != nil {
		return nil, "", nil, err
	}
	workout.ID = allocatedID

	dirPath := workoutDirPath(s.dataDir, nickname, workout.StartDate, workout.ID)
	if _, err := os.Stat(dirPath); err == nil {
		return nil, "", nil, workouts.ErrWorkoutExists
	}
	if err := os.MkdirAll(dirPath, 0700); err != nil {
		return nil, "", nil, fmt.Errorf("create workout dir: %w", err)
	}

	cleanup := func() {
		_ = os.RemoveAll(dirPath)
	}

	result := *workout
	dirName := keys.WorkoutDirName(workout.StartDate, workout.ID)
	return &result, dirName, cleanup, nil
}

func (s *WorkoutsStore) WriteMetadata(nickname string, workout *workouts.Workout) error {
	return s.db.Update(func(tx *bolt.Tx) error {
		if s.workoutIDExists(tx, workout.ID) {
			// Allow overwrite of same id only if same primary key (idempotent retry).
			_, key, err := s.lookupPrimaryKey(tx, workout.ID)
			if err != nil {
				return err
			}
			want := workoutPrimaryKey(nickname, workout.StartDate, workout.ID)
			if key != want {
				return workouts.ErrWorkoutExists
			}
		}
		return s.putWorkout(tx, nickname, workout)
	})
}

func (s *WorkoutsStore) Update(nickname string, workout *workouts.Workout) (*workouts.Workout, error) {
	if workout == nil || workout.ID == "" {
		return nil, workouts.ErrInvalidWorkout
	}
	workout.Name = workouts.TrimWorkoutName(workout.Name)
	workout.Description = workouts.TrimWorkoutDescription(workout.Description)

	err := s.db.Update(func(tx *bolt.Tx) error {
		owner, oldKey, err := s.lookupPrimaryKey(tx, workout.ID)
		if err != nil {
			return err
		}
		if owner != nickname {
			return workouts.ErrWorkoutNotFound
		}
		old, err := s.getByPrimaryKey(tx, oldKey)
		if err != nil {
			return err
		}

		newKey := workoutPrimaryKey(nickname, workout.StartDate, workout.ID)
		if oldKey != newKey {
			if tx.Bucket(bucketWorkouts).Get([]byte(newKey)) != nil {
				return workouts.ErrWorkoutExists
			}
			oldDir := workoutDirPath(s.dataDir, nickname, old.StartDate, old.ID)
			newDir := workoutDirPath(s.dataDir, nickname, workout.StartDate, workout.ID)
			if err := os.Rename(oldDir, newDir); err != nil && !os.IsNotExist(err) {
				return fmt.Errorf("rename workout dir: %w", err)
			}
			_ = s.deleteWorkoutMeta(tx, nickname, old)
		} else if sid := strings.TrimSpace(old.StravaActivityID); sid != "" && sid != strings.TrimSpace(workout.StravaActivityID) {
			_ = tx.Bucket(bucketIdxWorkoutsStrava).Delete([]byte(nickname + "/" + sid))
		}
		return s.putWorkout(tx, nickname, workout)
	})
	if err != nil {
		return nil, err
	}
	result := *workout
	return &result, nil
}

func (s *WorkoutsStore) Get(nickname, workoutID string) (*workouts.Workout, error) {
	var result *workouts.Workout
	err := s.db.View(func(tx *bolt.Tx) error {
		owner, key, err := s.lookupPrimaryKey(tx, workoutID)
		if err != nil {
			return err
		}
		if owner != nickname {
			return workouts.ErrWorkoutNotFound
		}
		w, err := s.getByPrimaryKey(tx, key)
		if err != nil {
			return err
		}
		result = w
		return nil
	})
	return result, err
}

func (s *WorkoutsStore) WorkoutDirName(nickname, workoutID string) (string, error) {
	w, err := s.Get(nickname, workoutID)
	if err != nil {
		return "", err
	}
	return keys.WorkoutDirName(w.StartDate, w.ID), nil
}

func (s *WorkoutsStore) List(nickname string) ([]workouts.Workout, error) {
	prefix := []byte(nickname + "/")
	var result []workouts.Workout
	err := s.db.View(func(tx *bolt.Tx) error {
		c := tx.Bucket(bucketWorkouts).Cursor()
		for k, v := c.Seek(prefix); k != nil && strings.HasPrefix(string(k), string(prefix)); k, v = c.Next() {
			var w workouts.Workout
			if err := unmarshalJSON(v, &w); err != nil {
				return err
			}
			result = append(result, w)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	sort.Slice(result, func(i, j int) bool {
		return workouts.FeedNewer(result[i].StartDate, result[i].ID, result[j].StartDate, result[j].ID)
	})
	return result, nil
}

func (s *WorkoutsStore) ListPage(nickname string, cursor *workouts.Cursor, limit int) ([]workouts.Workout, bool, error) {
	if limit <= 0 {
		limit = workouts.DefaultPageLimit
	}
	prefix := []byte(nickname + "/")
	var cursorKey string
	if cursor != nil {
		cursorKey = workoutPrimaryKey(nickname, cursor.StartDate, cursor.ID)
	}

	result := make([]workouts.Workout, 0, limit)
	hasMore := false
	err := s.db.View(func(tx *bolt.Tx) error {
		c := tx.Bucket(bucketWorkouts).Cursor()
		// Keys sort ascending; we need newest first → walk backwards.
		var k, v []byte
		if cursorKey != "" {
			k, v = c.Seek([]byte(cursorKey))
			if k != nil && string(k) == cursorKey {
				k, v = c.Prev()
			} else if k == nil {
				k, v = c.Last()
			} else {
				k, v = c.Prev()
			}
		} else {
			// Seek to last key with this nickname prefix.
			k, v = c.Seek(prefix)
			if k == nil || !strings.HasPrefix(string(k), string(prefix)) {
				return nil
			}
			for {
				nk, nv := c.Next()
				if nk == nil || !strings.HasPrefix(string(nk), string(prefix)) {
					break
				}
				k, v = nk, nv
			}
		}

		for k != nil && strings.HasPrefix(string(k), string(prefix)) {
			var w workouts.Workout
			if err := unmarshalJSON(v, &w); err != nil {
				return err
			}
			if cursor != nil && !workouts.AfterCursor(w.StartDate, w.ID, cursor) {
				k, v = c.Prev()
				continue
			}
			if len(result) >= limit {
				hasMore = true
				break
			}
			result = append(result, w)
			k, v = c.Prev()
		}
		return nil
	})
	return result, hasMore, err
}

func (s *WorkoutsStore) RemoveEquipmentFromAll(nickname, equipmentID string) error {
	prefix := []byte(nickname + "/")
	return s.db.Update(func(tx *bolt.Tx) error {
		c := tx.Bucket(bucketWorkouts).Cursor()
		for k, v := c.Seek(prefix); k != nil && strings.HasPrefix(string(k), string(prefix)); k, v = c.Next() {
			var w workouts.Workout
			if err := unmarshalJSON(v, &w); err != nil {
				return err
			}
			changed := false
			filtered := make([]workouts.WorkoutEquipment, 0, len(w.Equipment))
			for _, item := range w.Equipment {
				if item.ID == equipmentID {
					changed = true
					continue
				}
				filtered = append(filtered, item)
			}
			if !changed {
				continue
			}
			w.Equipment = filtered
			if err := s.putWorkout(tx, nickname, &w); err != nil {
				return err
			}
		}
		return nil
	})
}

func (s *WorkoutsStore) Delete(nickname, workoutID string) error {
	var dir string
	err := s.db.Update(func(tx *bolt.Tx) error {
		owner, key, err := s.lookupPrimaryKey(tx, workoutID)
		if err != nil {
			return err
		}
		if owner != nickname {
			return workouts.ErrWorkoutNotFound
		}
		w, err := s.getByPrimaryKey(tx, key)
		if err != nil {
			return err
		}
		dir = workoutDirPath(s.dataDir, nickname, w.StartDate, w.ID)
		return s.deleteWorkoutMeta(tx, nickname, w)
	})
	if err != nil {
		return err
	}
	if err := os.RemoveAll(dir); err != nil {
		return fmt.Errorf("delete workout dir: %w", err)
	}
	return nil
}

func (s *WorkoutsStore) HasStravaActivityID(nickname, stravaActivityID string) (bool, error) {
	stravaActivityID = strings.TrimSpace(stravaActivityID)
	if stravaActivityID == "" {
		return false, nil
	}
	var found bool
	err := s.db.View(func(tx *bolt.Tx) error {
		found = tx.Bucket(bucketIdxWorkoutsStrava).Get([]byte(nickname+"/"+stravaActivityID)) != nil
		return nil
	})
	return found, err
}

// PutExisting writes workout metadata without allocating a new ID (migration).
func (s *WorkoutsStore) PutExisting(nickname string, workout *workouts.Workout) error {
	return s.db.Update(func(tx *bolt.Tx) error {
		return s.putWorkout(tx, nickname, workout)
	})
}

func (s *WorkoutsStore) ListAllForUser(nickname string) ([]workouts.Workout, error) {
	return s.List(nickname)
}

var _ workouts.Repository = (*WorkoutsStore)(nil)
