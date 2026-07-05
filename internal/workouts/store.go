package workouts

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"
	"gopkg.in/yaml.v3"
)

var (
	ErrInvalidSportType = errors.New("invalid sport type")
	ErrInvalidWorkout   = errors.New("invalid workout")
	ErrWorkoutExists    = errors.New("workout already exists")
)

type Store struct {
	dataDir string
}

func NewStore(dataDir string) *Store {
	return &Store{dataDir: dataDir}
}

func (s *Store) userDir(nickname string) string {
	return filepath.Join(s.dataDir, nickname)
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
	return filepath.Join(userDir, workoutBaseName(startDate, id))
}

func workoutFilePath(userDir string, startDate time.Time, id string) string {
	return filepath.Join(workoutDir(userDir, startDate, id), workoutFileName(startDate, id))
}

func (s *Store) Create(nickname string, workout *Workout) (*Workout, error) {
	if err := s.validateWorkout(workout); err != nil {
		return nil, err
	}

	workout.ID = uuid.NewString()
	workout.Name = strings.TrimSpace(workout.Name)
	workout.Description = strings.TrimSpace(workout.Description)

	return s.saveWorkout(nickname, workout)
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

	workoutDirPath := workoutDir(userDir, workout.StartDate, workout.ID)
	if _, err := os.Stat(workoutDirPath); err == nil {
		return nil, ErrWorkoutExists
	}

	if err := os.MkdirAll(workoutDirPath, 0700); err != nil {
		return nil, fmt.Errorf("create workout dir: %w", err)
	}

	filePath := workoutFilePath(userDir, workout.StartDate, workout.ID)

	data, err := yaml.Marshal(workout)
	if err != nil {
		return nil, fmt.Errorf("marshal workout: %w", err)
	}

	tmp := filePath + ".tmp"
	if err := os.WriteFile(tmp, data, 0600); err != nil {
		return nil, fmt.Errorf("write workout: %w", err)
	}
	if err := os.Rename(tmp, filePath); err != nil {
		return nil, fmt.Errorf("rename workout: %w", err)
	}

	result := *workout
	return &result, nil
}

func (s *Store) List(nickname string) ([]Workout, error) {
	userDir := s.userDir(nickname)
	entries, err := os.ReadDir(userDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("read user dir: %w", err)
	}

	workouts := make([]Workout, 0)
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		dirName := entry.Name()
		filePath := filepath.Join(userDir, dirName, dirName+".yaml")
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
		workouts = append(workouts, workout)
	}

	sort.Slice(workouts, func(i, j int) bool {
		return workouts[i].StartDate.After(workouts[j].StartDate)
	})

	return workouts, nil
}
