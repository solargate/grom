package file

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"

	"github.com/solargate/grom/internal/data"
	"github.com/solargate/grom/internal/federation"
	"github.com/solargate/grom/internal/workouts"
	"gopkg.in/yaml.v3"
)

const likesFileName = "likes.yaml"

type WorkoutLikesStore struct {
	dataDir string
}

func NewWorkoutLikesStore(dataDir string) *WorkoutLikesStore {
	return &WorkoutLikesStore{dataDir: dataDir}
}

func (s *WorkoutLikesStore) GetLocal(ownerNickname, workoutID string) (*workouts.WorkoutLikes, error) {
	store := NewWorkoutsStore(s.dataDir)
	dir, err := store.findWorkoutDir(ownerNickname, workoutID)
	if err != nil {
		return nil, err
	}
	return readLikesYAML(filepath.Join(dir, likesFileName))
}

func (s *WorkoutLikesStore) PutLocal(ownerNickname, workoutID string, likes *workouts.WorkoutLikes) error {
	store := NewWorkoutsStore(s.dataDir)
	dir, err := store.findWorkoutDir(ownerNickname, workoutID)
	if err != nil {
		return err
	}
	return writeLikesYAML(filepath.Join(dir, likesFileName), likes)
}

func (s *WorkoutLikesStore) DeleteLocal(ownerNickname, workoutID string) error {
	store := NewWorkoutsStore(s.dataDir)
	dir, err := store.findWorkoutDir(ownerNickname, workoutID)
	if err != nil {
		if err == workouts.ErrWorkoutNotFound {
			return nil
		}
		return err
	}
	if err := os.Remove(filepath.Join(dir, likesFileName)); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("delete likes yaml: %w", err)
	}
	return nil
}

func (s *WorkoutLikesStore) GetFederated(viewerNickname, ownerHandle, workoutID string) (*workouts.WorkoutLikes, error) {
	return readLikesYAML(s.federatedLikesPath(viewerNickname, ownerHandle, workoutID))
}

func (s *WorkoutLikesStore) PutFederated(viewerNickname, ownerHandle, workoutID string, likes *workouts.WorkoutLikes) error {
	path := s.federatedLikesPath(viewerNickname, ownerHandle, workoutID)
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return fmt.Errorf("create federated likes dir: %w", err)
	}
	return writeLikesYAML(path, likes)
}

func (s *WorkoutLikesStore) DeleteFederated(viewerNickname, ownerHandle, workoutID string) error {
	path := s.federatedLikesPath(viewerNickname, ownerHandle, workoutID)
	if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("delete federated likes yaml: %w", err)
	}
	return nil
}

func (s *WorkoutLikesStore) GetLikeActivityID(actorNickname, objectID string) (string, error) {
	path := s.likeActivityPath(actorNickname, objectID)
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return "", nil
		}
		return "", fmt.Errorf("read like activity id: %w", err)
	}
	return string(data), nil
}

func (s *WorkoutLikesStore) PutLikeActivityID(actorNickname, objectID, activityID string) error {
	path := s.likeActivityPath(actorNickname, objectID)
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return fmt.Errorf("create like activity dir: %w", err)
	}
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, []byte(activityID), 0600); err != nil {
		return fmt.Errorf("write like activity id: %w", err)
	}
	if err := os.Rename(tmp, path); err != nil {
		return fmt.Errorf("rename like activity id: %w", err)
	}
	return nil
}

func (s *WorkoutLikesStore) DeleteLikeActivityID(actorNickname, objectID string) error {
	path := s.likeActivityPath(actorNickname, objectID)
	if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("delete like activity id: %w", err)
	}
	return nil
}

func (s *WorkoutLikesStore) federatedLikesPath(viewerNickname, ownerHandle, workoutID string) string {
	ownerKey := federation.OwnerKeyFromHandle(ownerHandle)
	return filepath.Join(
		data.UserDir(s.dataDir, viewerNickname),
		"federation",
		"inbox",
		"workouts",
		ownerKey,
		"likes",
		workoutID+".yaml",
	)
}

func (s *WorkoutLikesStore) likeActivityPath(actorNickname, objectID string) string {
	sum := sha256.Sum256([]byte(objectID))
	filename := hex.EncodeToString(sum[:]) + ".txt"
	return filepath.Join(
		data.UserDir(s.dataDir, actorNickname),
		"federation",
		"outbox",
		"likes",
		filename,
	)
}

func readLikesYAML(path string) (*workouts.WorkoutLikes, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			empty := workouts.NormalizeWorkoutLikes(nil)
			return &empty, nil
		}
		return nil, fmt.Errorf("read likes yaml: %w", err)
	}
	var likes workouts.WorkoutLikes
	if err := yaml.Unmarshal(data, &likes); err != nil {
		return nil, fmt.Errorf("parse likes yaml: %w", err)
	}
	norm := workouts.NormalizeWorkoutLikes(&likes)
	return &norm, nil
}

func writeLikesYAML(path string, likes *workouts.WorkoutLikes) error {
	norm := workouts.NormalizeWorkoutLikes(likes)
	data, err := yaml.Marshal(&norm)
	if err != nil {
		return fmt.Errorf("marshal likes yaml: %w", err)
	}
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, data, 0600); err != nil {
		return fmt.Errorf("write likes yaml: %w", err)
	}
	if err := os.Rename(tmp, path); err != nil {
		return fmt.Errorf("rename likes yaml: %w", err)
	}
	return nil
}

var _ workouts.LikesRepository = (*WorkoutLikesStore)(nil)
