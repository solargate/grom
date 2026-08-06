package file

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/solargate/grom/internal/data"
	"github.com/solargate/grom/internal/federation"
	"github.com/solargate/grom/internal/workouts"
	"gopkg.in/yaml.v3"
)

const likesFileName = "likes.yaml"

type WorkoutLikesStore struct {
	dataDir string
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

type likeActivityFile struct {
	ObjectID   string `yaml:"object_id"`
	ActivityID string `yaml:"activity_id"`
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
	raw, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return "", nil
		}
		return "", fmt.Errorf("read like activity id: %w", err)
	}
	return parseLikeActivityFile(raw), nil
}

func (s *WorkoutLikesStore) PutLikeActivityID(actorNickname, objectID, activityID string) error {
	path := s.likeActivityPath(actorNickname, objectID)
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return fmt.Errorf("create like activity dir: %w", err)
	}
	payload, err := yaml.Marshal(&likeActivityFile{
		ObjectID:   objectID,
		ActivityID: activityID,
	})
	if err != nil {
		return fmt.Errorf("marshal like activity id: %w", err)
	}
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, payload, 0600); err != nil {
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

// ListAllFederated returns every federated likes cache record (migration).
func (s *WorkoutLikesStore) ListAllFederated() ([]FederatedLikeEntry, error) {
	var out []FederatedLikeEntry
	usersRoot := filepath.Join(s.dataDir, data.UsersSubdir)
	userEntries, err := os.ReadDir(usersRoot)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	for _, userEntry := range userEntries {
		if !userEntry.IsDir() {
			continue
		}
		viewer := userEntry.Name()
		root := filepath.Join(usersRoot, viewer, "federation", "inbox", "workouts")
		ownerEntries, err := os.ReadDir(root)
		if err != nil {
			if os.IsNotExist(err) {
				continue
			}
			return nil, err
		}
		for _, ownerEntry := range ownerEntries {
			if !ownerEntry.IsDir() {
				continue
			}
			ownerKey := ownerEntry.Name()
			ownerDir := filepath.Join(root, ownerKey)
			ownerHandle := federation.OwnerHandleFromKey(ownerKey)
			if raw, err := os.ReadFile(filepath.Join(ownerDir, "author.yaml")); err == nil {
				var meta federation.AuthorMeta
				if err := yaml.Unmarshal(raw, &meta); err != nil {
					return nil, fmt.Errorf("parse author meta %s: %w", ownerDir, err)
				}
				if meta.Handle != "" {
					ownerHandle = meta.Handle
				}
			} else if !os.IsNotExist(err) {
				return nil, err
			}
			likesDir := filepath.Join(ownerDir, "likes")
			files, err := os.ReadDir(likesDir)
			if err != nil {
				if os.IsNotExist(err) {
					continue
				}
				return nil, err
			}
			for _, f := range files {
				name := f.Name()
				if f.IsDir() || !strings.HasSuffix(name, ".yaml") {
					continue
				}
				likes, err := readLikesYAML(filepath.Join(likesDir, name))
				if err != nil {
					return nil, err
				}
				if likes.Likes == 0 {
					continue
				}
				out = append(out, FederatedLikeEntry{
					ViewerNickname: viewer,
					OwnerHandle:    ownerHandle,
					WorkoutID:      strings.TrimSuffix(name, ".yaml"),
					Likes:          *likes,
				})
			}
		}
	}
	return out, nil
}

// ListAllLikeActivities returns outbound Like activity ids that include object_id (migration).
// Legacy plain-text activity files without object_id are skipped; migrate reconstructs those via inbox.
func (s *WorkoutLikesStore) ListAllLikeActivities() ([]LikeActivityEntry, error) {
	var out []LikeActivityEntry
	usersRoot := filepath.Join(s.dataDir, data.UsersSubdir)
	userEntries, err := os.ReadDir(usersRoot)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	for _, userEntry := range userEntries {
		if !userEntry.IsDir() {
			continue
		}
		actor := userEntry.Name()
		dir := filepath.Join(usersRoot, actor, "federation", "outbox", "likes")
		files, err := os.ReadDir(dir)
		if err != nil {
			if os.IsNotExist(err) {
				continue
			}
			return nil, err
		}
		for _, f := range files {
			if f.IsDir() || !strings.HasSuffix(f.Name(), ".txt") {
				continue
			}
			raw, err := os.ReadFile(filepath.Join(dir, f.Name()))
			if err != nil {
				return nil, err
			}
			var parsed likeActivityFile
			if err := yaml.Unmarshal(raw, &parsed); err != nil || parsed.ObjectID == "" || parsed.ActivityID == "" {
				continue
			}
			out = append(out, LikeActivityEntry{
				ActorNickname: actor,
				ObjectID:      parsed.ObjectID,
				ActivityID:    parsed.ActivityID,
			})
		}
	}
	return out, nil
}

func parseLikeActivityFile(raw []byte) string {
	var parsed likeActivityFile
	if err := yaml.Unmarshal(raw, &parsed); err == nil && parsed.ActivityID != "" {
		return parsed.ActivityID
	}
	return string(raw)
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
