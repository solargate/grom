package workouts

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/solargate/grom/internal/data"
	"gopkg.in/yaml.v3"
)

var (
	ErrInvalidSportType = errors.New("invalid sport type")
	ErrInvalidWorkout   = errors.New("invalid workout")
	ErrWorkoutExists    = errors.New("workout already exists")
	ErrWorkoutNotFound  = errors.New("workout not found")
)

type Store struct {
	dataDir string
}

func NewStore(dataDir string) *Store {
	return &Store{dataDir: dataDir}
}

func (s *Store) DataDir() string {
	return s.dataDir
}

const workoutsSubdir = "workouts"

func (s *Store) userDir(nickname string) string {
	return data.UserDir(s.dataDir, nickname)
}

func workoutsDir(userDir string) string {
	return filepath.Join(userDir, workoutsSubdir)
}

func workoutBaseName(startDate time.Time, id string) string {
	iso := startDate.UTC().Format("2006-01-02T15:04:05Z")
	iso = strings.ReplaceAll(iso, ":", "")
	return iso + "-" + id
}

func workoutFileName(startDate time.Time, id string) string {
	return workoutBaseName(startDate, id) + ".yaml"
}

func workoutDir(userDir string, startDate time.Time, id string) string {
	return filepath.Join(workoutsDir(userDir), workoutBaseName(startDate, id))
}

func workoutFilePath(userDir string, startDate time.Time, id string) string {
	return filepath.Join(workoutDir(userDir, startDate, id), workoutFileName(startDate, id))
}

func (s *Store) Create(nickname string, workout *Workout) (*Workout, error) {
	if err := s.validateWorkout(workout); err != nil {
		return nil, err
	}

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
	workout.Name = trimWorkoutName(workout.Name)
	workout.Description = trimWorkoutDescription(workout.Description)
	workout.Device = DeviceGrom

	return s.saveWorkout(nickname, workout)
}

func trimWorkoutName(name string) string {
	return strings.TrimSpace(name)
}

func trimWorkoutDescription(description string) string {
	return strings.TrimSpace(description)
}

func (s *Store) validateWorkout(workout *Workout) error {
	if workout == nil {
		return ErrInvalidWorkout
	}
	if strings.TrimSpace(workout.Name) == "" {
		return fmt.Errorf("%w: name is required", ErrInvalidWorkout)
	}
	if !IsValidSportType(workout.SportType) {
		return ErrInvalidSportType
	}
	if workout.DurationSeconds < 0 {
		return fmt.Errorf("%w: duration must be non-negative", ErrInvalidWorkout)
	}
	if workout.Distance < 0 {
		return fmt.Errorf("%w: distance must be non-negative", ErrInvalidWorkout)
	}
	if workout.StartDate.IsZero() {
		return fmt.Errorf("%w: start_date is required", ErrInvalidWorkout)
	}
	return nil
}

func (s *Store) saveWorkout(nickname string, workout *Workout) (*Workout, error) {
	userDir := s.userDir(nickname)
	if err := os.MkdirAll(userDir, 0700); err != nil {
		return nil, fmt.Errorf("create user dir: %w", err)
	}
	wd := workoutsDir(userDir)
	if err := os.MkdirAll(wd, 0700); err != nil {
		return nil, fmt.Errorf("create workouts dir: %w", err)
	}

	if s.workoutIDExists(wd, workout.ID) {
		return nil, ErrWorkoutExists
	}

	workoutDirPath := workoutDir(userDir, workout.StartDate, workout.ID)
	if _, err := os.Stat(workoutDirPath); err == nil {
		return nil, ErrWorkoutExists
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

func writeWorkoutYAML(filePath string, workout *Workout) error {
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

func (s *Store) List(nickname string) ([]Workout, error) {
	wd := workoutsDir(s.userDir(nickname))
	entries, err := os.ReadDir(wd)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("read workouts dir: %w", err)
	}

	workouts := make([]Workout, 0)
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

		var workout Workout
		if err := yaml.Unmarshal(data, &workout); err != nil {
			return nil, fmt.Errorf("parse workout %q: %w", dirName, err)
		}
		if workout.ID == "" {
			if idx := strings.Index(dirName, "Z-"); idx >= 0 && idx+2 < len(dirName) {
				workout.ID = dirName[idx+2:]
			}
		}
		workout.HasMapPreview = workoutHasMapPreview(dirPath)
		populateWorkoutMedia(&workout, dirPath)
		workouts = append(workouts, workout)
	}

	sort.Slice(workouts, func(i, j int) bool {
		return workouts[i].StartDate.After(workouts[j].StartDate)
	})

	return workouts, nil
}

func (s *Store) RemoveEquipmentFromAll(nickname, equipmentID string) error {
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

		var workout Workout
		if err := yaml.Unmarshal(data, &workout); err != nil {
			return fmt.Errorf("parse workout %q: %w", dirName, err)
		}

		changed := false
		filtered := make([]WorkoutEquipment, 0, len(workout.Equipment))
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
