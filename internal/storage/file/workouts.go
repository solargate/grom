package file

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/solargate/grom/internal/data"
	"github.com/solargate/grom/internal/storage/keys"
	"github.com/solargate/grom/internal/workouts"
	"gopkg.in/yaml.v3"
)

const (
	workoutsSubdir    = "workouts"
	workoutIDLength   = workouts.WorkoutIDLength
	workoutIDAlphabet = "0123456789"
)

type WorkoutsStore struct {
	dataDir string
}

func NewWorkoutsStore(dataDir string) *WorkoutsStore {
	return &WorkoutsStore{dataDir: dataDir}
}

func (s *WorkoutsStore) userDir(nickname string) string {
	return data.UserDir(s.dataDir, nickname)
}

func workoutBaseName(startDate time.Time, id string) string {
	return keys.WorkoutDirName(startDate, id)
}

func workoutFileName(startDate time.Time, id string) string {
	return workoutBaseName(startDate, id) + ".yaml"
}

func workoutsDir(userDir string) string {
	return filepath.Join(userDir, workoutsSubdir)
}

func workoutDir(userDir string, startDate time.Time, id string) string {
	return filepath.Join(workoutsDir(userDir), workoutBaseName(startDate, id))
}

func workoutFilePath(userDir string, startDate time.Time, id string) string {
	return filepath.Join(workoutDir(userDir, startDate, id), workoutFileName(startDate, id))
}

func (s *WorkoutsStore) Create(nickname string, workout *workouts.Workout) (*workouts.Workout, error) {
	userDir := s.userDir(nickname)
	if err := os.MkdirAll(userDir, 0700); err != nil {
		return nil, fmt.Errorf("create user dir: %w", err)
	}
	wd := workoutsDir(userDir)
	if err := os.MkdirAll(wd, 0700); err != nil {
		return nil, fmt.Errorf("create workouts dir: %w", err)
	}

	id, err := s.allocateWorkoutID()
	if err != nil {
		return nil, err
	}
	workout.ID = id
	workout.Name = workouts.TrimWorkoutName(workout.Name)
	workout.Description = workouts.TrimWorkoutDescription(workout.Description)
	workout.Device = workouts.NormalizeDevice(workout.Device)

	return s.saveNewWorkout(nickname, workout)
}

func (s *WorkoutsStore) BeginCreate(nickname string, workout *workouts.Workout) (*workouts.Workout, string, func(), error) {
	userDir := s.userDir(nickname)
	if err := os.MkdirAll(userDir, 0700); err != nil {
		return nil, "", nil, fmt.Errorf("create user dir: %w", err)
	}
	wd := workoutsDir(userDir)
	if err := os.MkdirAll(wd, 0700); err != nil {
		return nil, "", nil, fmt.Errorf("create workouts dir: %w", err)
	}

	id, err := s.allocateWorkoutID()
	if err != nil {
		return nil, "", nil, err
	}
	workout.ID = id
	workout.Name = workouts.TrimWorkoutName(workout.Name)
	workout.Description = workouts.TrimWorkoutDescription(workout.Description)

	workoutDirPath := workoutDir(userDir, workout.StartDate, workout.ID)
	if _, err := os.Stat(workoutDirPath); err == nil {
		return nil, "", nil, workouts.ErrWorkoutExists
	}
	if err := os.MkdirAll(workoutDirPath, 0700); err != nil {
		return nil, "", nil, fmt.Errorf("create workout dir: %w", err)
	}

	cleanup := func() {
		_ = os.RemoveAll(workoutDirPath)
	}

	result := *workout
	dirName := filepath.Base(workoutDirPath)
	return &result, dirName, cleanup, nil
}

func (s *WorkoutsStore) WriteMetadata(nickname string, workout *workouts.Workout) error {
	userDir := s.userDir(nickname)
	filePath := workoutFilePath(userDir, workout.StartDate, workout.ID)
	return writeWorkoutYAML(filePath, workout)
}

func (s *WorkoutsStore) Update(nickname string, workout *workouts.Workout) (*workouts.Workout, error) {
	if workout == nil || workout.ID == "" {
		return nil, workouts.ErrInvalidWorkout
	}

	oldDir, err := s.findWorkoutDir(nickname, workout.ID)
	if err != nil {
		return nil, err
	}

	workout.Name = workouts.TrimWorkoutName(workout.Name)
	workout.Description = workouts.TrimWorkoutDescription(workout.Description)

	userDir := s.userDir(nickname)
	newDir := workoutDir(userDir, workout.StartDate, workout.ID)
	oldBase := filepath.Base(oldDir)
	newBase := filepath.Base(newDir)

	if oldBase != newBase {
		if _, err := os.Stat(newDir); err == nil {
			return nil, workouts.ErrWorkoutExists
		}
		if err := os.Rename(oldDir, newDir); err != nil {
			return nil, fmt.Errorf("rename workout dir: %w", err)
		}
		oldYAML := filepath.Join(newDir, oldBase+".yaml")
		newYAML := filepath.Join(newDir, newBase+".yaml")
		if err := os.Rename(oldYAML, newYAML); err != nil {
			// Best-effort rollback of directory rename.
			_ = os.Rename(newDir, oldDir)
			return nil, fmt.Errorf("rename workout yaml: %w", err)
		}
	}

	filePath := workoutFilePath(userDir, workout.StartDate, workout.ID)
	if err := writeWorkoutYAML(filePath, workout); err != nil {
		return nil, err
	}
	result := *workout
	return &result, nil
}

func (s *WorkoutsStore) Get(nickname, workoutID string) (*workouts.Workout, error) {
	dir, err := s.findWorkoutDir(nickname, workoutID)
	if err != nil {
		return nil, err
	}
	return readWorkoutFromDir(dir)
}

func (s *WorkoutsStore) WorkoutDirName(nickname, workoutID string) (string, error) {
	dir, err := s.findWorkoutDir(nickname, workoutID)
	if err != nil {
		return "", err
	}
	return filepath.Base(dir), nil
}

func (s *WorkoutsStore) saveNewWorkout(nickname string, workout *workouts.Workout) (*workouts.Workout, error) {
	userDir := s.userDir(nickname)
	if err := os.MkdirAll(userDir, 0700); err != nil {
		return nil, fmt.Errorf("create user dir: %w", err)
	}
	wd := workoutsDir(userDir)
	if err := os.MkdirAll(wd, 0700); err != nil {
		return nil, fmt.Errorf("create workouts dir: %w", err)
	}

	exists, err := s.workoutIDExists(workout.ID)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, workouts.ErrWorkoutExists
	}

	workoutDirPath := workoutDir(userDir, workout.StartDate, workout.ID)
	if _, err := os.Stat(workoutDirPath); err == nil {
		return nil, workouts.ErrWorkoutExists
	}

	if err := os.MkdirAll(workoutDirPath, 0700); err != nil {
		return nil, fmt.Errorf("create workout dir: %w", err)
	}

	filePath := workoutFilePath(userDir, workout.StartDate, workout.ID)
	if err := writeWorkoutYAML(filePath, workout); err != nil {
		return nil, err
	}

	result := *workout
	return &result, nil
}

func writeWorkoutYAML(filePath string, workout *workouts.Workout) error {
	data, err := yaml.Marshal(workout)
	if err != nil {
		return fmt.Errorf("marshal workout: %w", err)
	}

	tmp := filePath + ".tmp"
	if err := os.WriteFile(tmp, data, 0600); err != nil {
		return fmt.Errorf("write workout: %w", err)
	}
	if err := os.Rename(tmp, filePath); err != nil {
		return fmt.Errorf("rename workout: %w", err)
	}
	return nil
}

func readWorkoutFromDir(dir string) (*workouts.Workout, error) {
	dirName := filepath.Base(dir)
	filePath := filepath.Join(dir, dirName+".yaml")
	data, err := os.ReadFile(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, workouts.ErrWorkoutNotFound
		}
		return nil, fmt.Errorf("read workout: %w", err)
	}

	var workout workouts.Workout
	if err := yaml.Unmarshal(data, &workout); err != nil {
		return nil, fmt.Errorf("parse workout: %w", err)
	}
	if workout.ID == "" {
		if idx := strings.Index(dirName, "Z-"); idx >= 0 && idx+2 < len(dirName) {
			workout.ID = dirName[idx+2:]
		}
	}
	return &workout, nil
}

func (s *WorkoutsStore) List(nickname string) ([]workouts.Workout, error) {
	return s.listAll(nickname)
}

func (s *WorkoutsStore) listAll(nickname string) ([]workouts.Workout, error) {
	wd := workoutsDir(s.userDir(nickname))
	entries, err := os.ReadDir(wd)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("read workouts dir: %w", err)
	}

	result := make([]workouts.Workout, 0)
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		workout, err := readWorkoutFromDir(filepath.Join(wd, entry.Name()))
		if err != nil {
			return nil, fmt.Errorf("read workout %q: %w", entry.Name(), err)
		}
		result = append(result, *workout)
	}

	sort.Slice(result, func(i, j int) bool {
		return workouts.FeedNewer(result[i].StartDate, result[i].ID, result[j].StartDate, result[j].ID)
	})
	return result, nil
}

func (s *WorkoutsStore) ListPage(nickname string, cursor *workouts.Cursor, limit int, sportTypes map[string]struct{}) ([]workouts.Workout, bool, error) {
	if limit <= 0 {
		limit = workouts.DefaultPageLimit
	}

	wd := workoutsDir(s.userDir(nickname))
	entries, err := os.ReadDir(wd)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, false, nil
		}
		return nil, false, fmt.Errorf("read workouts dir: %w", err)
	}

	dirNames := make([]string, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() {
			dirNames = append(dirNames, entry.Name())
		}
	}
	sort.Slice(dirNames, func(i, j int) bool {
		return dirNames[i] > dirNames[j]
	})

	var cursorKey string
	if cursor != nil {
		cursorKey = keys.WorkoutDirName(cursor.StartDate, cursor.ID)
	}

	result := make([]workouts.Workout, 0, limit)
	hasMore := false
	for _, dirName := range dirNames {
		if cursorKey != "" && dirName >= cursorKey {
			continue
		}
		workout, err := readWorkoutFromDir(filepath.Join(wd, dirName))
		if err != nil {
			return nil, false, fmt.Errorf("read workout %q: %w", dirName, err)
		}
		if cursor != nil && !workouts.AfterCursor(workout.StartDate, workout.ID, cursor) {
			continue
		}
		if !workouts.MatchSportType(workout.SportType, sportTypes) {
			continue
		}
		if len(result) >= limit {
			hasMore = true
			break
		}
		result = append(result, *workout)
	}
	return result, hasMore, nil
}

func (s *WorkoutsStore) RemoveEquipmentFromAll(nickname, equipmentID string) error {
	wd := workoutsDir(s.userDir(nickname))
	entries, err := os.ReadDir(wd)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("read workouts dir: %w", err)
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		dirName := entry.Name()
		dirPath := filepath.Join(wd, dirName)
		filePath := filepath.Join(dirPath, dirName+".yaml")
		data, err := os.ReadFile(filePath)
		if err != nil {
			return fmt.Errorf("read workout %q: %w", dirName, err)
		}

		var workout workouts.Workout
		if err := yaml.Unmarshal(data, &workout); err != nil {
			return fmt.Errorf("parse workout %q: %w", dirName, err)
		}

		changed := false
		filtered := make([]workouts.WorkoutEquipment, 0, len(workout.Equipment))
		for _, item := range workout.Equipment {
			if item.ID == equipmentID {
				changed = true
				continue
			}
			filtered = append(filtered, item)
		}
		if !changed {
			continue
		}

		workout.Equipment = filtered
		if err := writeWorkoutYAML(filePath, &workout); err != nil {
			return err
		}
	}

	return nil
}

func (s *WorkoutsStore) Delete(nickname, workoutID string) error {
	dir, err := s.findWorkoutDir(nickname, workoutID)
	if err != nil {
		return err
	}
	if err := os.RemoveAll(dir); err != nil {
		return fmt.Errorf("delete workout dir: %w", err)
	}
	return nil
}

func (s *WorkoutsStore) findWorkoutDir(nickname, workoutID string) (string, error) {
	wd := workoutsDir(s.userDir(nickname))
	entries, err := os.ReadDir(wd)
	if err != nil {
		if os.IsNotExist(err) {
			return "", workouts.ErrWorkoutNotFound
		}
		return "", err
	}

	suffix := "-" + workoutID
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		if len(entry.Name()) > len(suffix) && entry.Name()[len(entry.Name())-len(suffix):] == suffix {
			return filepath.Join(wd, entry.Name()), nil
		}
	}
	return "", workouts.ErrWorkoutNotFound
}

func (s *WorkoutsStore) HasExternalID(nickname, name, id string) (bool, error) {
	name = strings.TrimSpace(name)
	id = strings.TrimSpace(id)
	if name == "" || id == "" {
		return false, nil
	}

	wd := workoutsDir(s.userDir(nickname))
	entries, err := os.ReadDir(wd)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		workout, err := readWorkoutFromDir(filepath.Join(wd, entry.Name()))
		if err != nil {
			continue
		}
		if workout.ExternalID == nil {
			continue
		}
		if strings.TrimSpace(workout.ExternalID.Name) == name && strings.TrimSpace(workout.ExternalID.ID) == id {
			return true, nil
		}
	}
	return false, nil
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

func (s *WorkoutsStore) workoutIDExists(id string) (bool, error) {
	usersRoot := filepath.Join(s.dataDir, data.UsersSubdir)
	users, err := os.ReadDir(usersRoot)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}

	suffix := "-" + id
	for _, userEntry := range users {
		if !userEntry.IsDir() {
			continue
		}
		entries, err := os.ReadDir(workoutsDir(filepath.Join(usersRoot, userEntry.Name())))
		if err != nil {
			if os.IsNotExist(err) {
				continue
			}
			return false, err
		}
		for _, entry := range entries {
			if entry.IsDir() && strings.HasSuffix(entry.Name(), suffix) {
				return true, nil
			}
		}
	}
	return false, nil
}

func (s *WorkoutsStore) allocateWorkoutID() (string, error) {
	const maxAttempts = 10
	for range maxAttempts {
		id, err := newWorkoutID()
		if err != nil {
			return "", err
		}
		exists, err := s.workoutIDExists(id)
		if err != nil {
			return "", fmt.Errorf("check workout id: %w", err)
		}
		if !exists {
			return id, nil
		}
	}
	return "", fmt.Errorf("failed to allocate unique workout id after %d attempts", maxAttempts)
}

// Import writes workout metadata as-is without allocating a new ID (migration).
func (s *WorkoutsStore) Import(nickname string, workout *workouts.Workout) error {
	if workout == nil || workout.ID == "" {
		return workouts.ErrInvalidWorkout
	}
	userDir := s.userDir(nickname)
	if err := os.MkdirAll(workoutsDir(userDir), 0700); err != nil {
		return fmt.Errorf("create workouts dir: %w", err)
	}
	workoutDirPath := workoutDir(userDir, workout.StartDate, workout.ID)
	if err := os.MkdirAll(workoutDirPath, 0700); err != nil {
		return fmt.Errorf("create workout dir: %w", err)
	}
	return writeWorkoutYAML(workoutFilePath(userDir, workout.StartDate, workout.ID), workout)
}

var _ workouts.Repository = (*WorkoutsStore)(nil)
