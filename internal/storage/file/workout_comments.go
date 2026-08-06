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

const commentsFileName = "comments.yaml"

type WorkoutCommentsStore struct {
	dataDir string
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

type commentActivityFile struct {
	NoteID     string `yaml:"note_id"`
	ActivityID string `yaml:"activity_id"`
}

func NewWorkoutCommentsStore(dataDir string) *WorkoutCommentsStore {
	return &WorkoutCommentsStore{dataDir: dataDir}
}

func (s *WorkoutCommentsStore) GetLocal(ownerNickname, workoutID string) (*workouts.WorkoutComments, error) {
	store := NewWorkoutsStore(s.dataDir)
	dir, err := store.findWorkoutDir(ownerNickname, workoutID)
	if err != nil {
		return nil, err
	}
	return readCommentsYAML(filepath.Join(dir, commentsFileName))
}

func (s *WorkoutCommentsStore) PutLocal(ownerNickname, workoutID string, comments *workouts.WorkoutComments) error {
	store := NewWorkoutsStore(s.dataDir)
	dir, err := store.findWorkoutDir(ownerNickname, workoutID)
	if err != nil {
		return err
	}
	return writeCommentsYAML(filepath.Join(dir, commentsFileName), comments)
}

func (s *WorkoutCommentsStore) DeleteLocal(ownerNickname, workoutID string) error {
	store := NewWorkoutsStore(s.dataDir)
	dir, err := store.findWorkoutDir(ownerNickname, workoutID)
	if err != nil {
		if err == workouts.ErrWorkoutNotFound {
			return nil
		}
		return err
	}
	if err := os.Remove(filepath.Join(dir, commentsFileName)); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("delete comments yaml: %w", err)
	}
	return nil
}

func (s *WorkoutCommentsStore) GetFederated(viewerNickname, ownerHandle, workoutID string) (*workouts.WorkoutComments, error) {
	return readCommentsYAML(s.federatedCommentsPath(viewerNickname, ownerHandle, workoutID))
}

func (s *WorkoutCommentsStore) PutFederated(viewerNickname, ownerHandle, workoutID string, comments *workouts.WorkoutComments) error {
	path := s.federatedCommentsPath(viewerNickname, ownerHandle, workoutID)
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return fmt.Errorf("create federated comments dir: %w", err)
	}
	return writeCommentsYAML(path, comments)
}

func (s *WorkoutCommentsStore) DeleteFederated(viewerNickname, ownerHandle, workoutID string) error {
	path := s.federatedCommentsPath(viewerNickname, ownerHandle, workoutID)
	if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("delete federated comments yaml: %w", err)
	}
	return nil
}

func (s *WorkoutCommentsStore) GetCommentActivityID(actorNickname, noteID string) (string, error) {
	path := s.commentActivityPath(actorNickname, noteID)
	raw, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return "", nil
		}
		return "", fmt.Errorf("read comment activity id: %w", err)
	}
	return parseCommentActivityFile(raw), nil
}

func (s *WorkoutCommentsStore) PutCommentActivityID(actorNickname, noteID, activityID string) error {
	path := s.commentActivityPath(actorNickname, noteID)
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return fmt.Errorf("create comment activity dir: %w", err)
	}
	payload, err := yaml.Marshal(&commentActivityFile{
		NoteID:     noteID,
		ActivityID: activityID,
	})
	if err != nil {
		return fmt.Errorf("marshal comment activity id: %w", err)
	}
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, payload, 0600); err != nil {
		return fmt.Errorf("write comment activity id: %w", err)
	}
	if err := os.Rename(tmp, path); err != nil {
		return fmt.Errorf("rename comment activity id: %w", err)
	}
	return nil
}

func (s *WorkoutCommentsStore) DeleteCommentActivityID(actorNickname, noteID string) error {
	path := s.commentActivityPath(actorNickname, noteID)
	if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("delete comment activity id: %w", err)
	}
	return nil
}

// ListAllFederated returns every federated comments cache record (migration).
func (s *WorkoutCommentsStore) ListAllFederated() ([]FederatedCommentEntry, error) {
	var out []FederatedCommentEntry
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
			commentsDir := filepath.Join(ownerDir, "comments")
			files, err := os.ReadDir(commentsDir)
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
				comments, err := readCommentsYAML(filepath.Join(commentsDir, name))
				if err != nil {
					return nil, err
				}
				if comments.CommentsNum == 0 {
					continue
				}
				out = append(out, FederatedCommentEntry{
					ViewerNickname: viewer,
					OwnerHandle:    ownerHandle,
					WorkoutID:      strings.TrimSuffix(name, ".yaml"),
					Comments:       *comments,
				})
			}
		}
	}
	return out, nil
}

// ListAllCommentActivities returns outbound Create Note activity ids (migration).
func (s *WorkoutCommentsStore) ListAllCommentActivities() ([]CommentActivityEntry, error) {
	var out []CommentActivityEntry
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
		dir := filepath.Join(usersRoot, actor, "federation", "outbox", "comments")
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
			var parsed commentActivityFile
			if err := yaml.Unmarshal(raw, &parsed); err != nil || parsed.NoteID == "" || parsed.ActivityID == "" {
				continue
			}
			out = append(out, CommentActivityEntry{
				ActorNickname: actor,
				NoteID:        parsed.NoteID,
				ActivityID:    parsed.ActivityID,
			})
		}
	}
	return out, nil
}

func parseCommentActivityFile(raw []byte) string {
	var parsed commentActivityFile
	if err := yaml.Unmarshal(raw, &parsed); err == nil && parsed.ActivityID != "" {
		return parsed.ActivityID
	}
	return string(raw)
}

func (s *WorkoutCommentsStore) federatedCommentsPath(viewerNickname, ownerHandle, workoutID string) string {
	ownerKey := federation.OwnerKeyFromHandle(ownerHandle)
	return filepath.Join(
		data.UserDir(s.dataDir, viewerNickname),
		"federation",
		"inbox",
		"workouts",
		ownerKey,
		"comments",
		workoutID+".yaml",
	)
}

func (s *WorkoutCommentsStore) commentActivityPath(actorNickname, noteID string) string {
	sum := sha256.Sum256([]byte(noteID))
	filename := hex.EncodeToString(sum[:]) + ".txt"
	return filepath.Join(
		data.UserDir(s.dataDir, actorNickname),
		"federation",
		"outbox",
		"comments",
		filename,
	)
}

func readCommentsYAML(path string) (*workouts.WorkoutComments, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			empty := workouts.NormalizeWorkoutComments(nil)
			return &empty, nil
		}
		return nil, fmt.Errorf("read comments yaml: %w", err)
	}
	var comments workouts.WorkoutComments
	if err := yaml.Unmarshal(data, &comments); err != nil {
		return nil, fmt.Errorf("parse comments yaml: %w", err)
	}
	norm := workouts.NormalizeWorkoutComments(&comments)
	return &norm, nil
}

func writeCommentsYAML(path string, comments *workouts.WorkoutComments) error {
	norm := workouts.NormalizeWorkoutComments(comments)
	data, err := yaml.Marshal(&norm)
	if err != nil {
		return fmt.Errorf("marshal comments yaml: %w", err)
	}
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, data, 0600); err != nil {
		return fmt.Errorf("write comments yaml: %w", err)
	}
	if err := os.Rename(tmp, path); err != nil {
		return fmt.Errorf("rename comments yaml: %w", err)
	}
	return nil
}

var _ workouts.CommentsRepository = (*WorkoutCommentsStore)(nil)
