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
	iso := startDate.UTC().Format("2006-01-02T15:04:05Z")
	iso = strings.ReplaceAll(iso, ":", "")
	return iso + "-" + id
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

	id, err := s.allocateWorkoutID(wd)
	if err != nil {
		return nil, err
	}
	workout.ID = id
	workout.Name = workouts.TrimWorkoutName(workout.Name)
	workout.Description = workouts.TrimWorkoutDescription(workout.Description)
	workout.Device = workouts.DeviceGrom

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

	id, err := s.allocateWorkoutID(wd)
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

	if s.workoutIDExists(wd, workout.ID) {
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

		dirName := entry.Name()
		dirPath := filepath.Join(wd, dirName)
		filePath := filepath.Join(dirPath, dirName+".yaml")
		data, err := os.ReadFile(filePath)
		if err != nil {
			return nil, fmt.Errorf("read workout %q: %w", dirName, err)
		}

		var workout workouts.Workout
		if err := yaml.Unmarshal(data, &workout); err != nil {
			return nil, fmt.Errorf("parse workout %q: %w", dirName, err)
		}
		if workout.ID == "" {
			if idx := strings.Index(dirName, "Z-"); idx >= 0 && idx+2 < len(dirName) {
				workout.ID = dirName[idx+2:]
			}
		}
		result = append(result, workout)
	}

	sort.Slice(result, func(i, j int) bool {
		return result[i].StartDate.After(result[j].StartDate)
	})

	return result, nil
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

func (s *WorkoutsStore) HasStravaActivityID(nickname, stravaActivityID string) (bool, error) {
	stravaActivityID = strings.TrimSpace(stravaActivityID)
	if stravaActivityID == "" {
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
		if workout.StravaActivityID == stravaActivityID {
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

func (s *WorkoutsStore) workoutIDExists(workoutsRoot, id string) bool {
	suffix := "-" + id
	entries, err := os.ReadDir(workoutsRoot)
	if err != nil {
		return false
	}
	for _, entry := range entries {
		if entry.IsDir() && strings.HasSuffix(entry.Name(), suffix) {
			return true
		}
	}
	return false
}

func (s *WorkoutsStore) allocateWorkoutID(workoutsRoot string) (string, error) {
	const maxAttempts = 10
	for range maxAttempts {
		id, err := newWorkoutID()
		if err != nil {
			return "", err
		}
		if !s.workoutIDExists(workoutsRoot, id) {
			return id, nil
		}
	}
	return "", fmt.Errorf("failed to allocate unique workout id after %d attempts", maxAttempts)
}

var _ workouts.Repository = (*WorkoutsStore)(nil)
