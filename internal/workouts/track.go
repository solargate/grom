package workouts

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"

	"github.com/solargate/travka/internal/maprender"
	"github.com/solargate/travka/internal/tracks"
)

const MapPreviewFileName = "map-preview.png"

type TrackInput struct {
	Filename string
	Data     []byte
	Parsed   *tracks.Data
}

func (s *Store) CreateWithTrack(nickname string, workout *Workout, track *TrackInput) (*Workout, error) {
	if track == nil {
		return s.Create(nickname, workout)
	}

	if err := s.validateWorkout(workout); err != nil {
		return nil, err
	}

	trackName, err := tracks.TrackFileName(track.Filename)
	if err != nil {
		return nil, err
	}

	parsed := track.Parsed
	if parsed == nil {
		parsed, err = tracks.Parse(track.Data, track.Filename)
		if err != nil {
			return nil, err
		}
	}

	startDate := workout.StartDate
	durationSeconds := workout.DurationSeconds
	distanceMeters := workout.Distance
	parsed.ApplyToWorkout(&startDate, &durationSeconds, &distanceMeters)

	workout.StartDate = startDate
	workout.DurationSeconds = durationSeconds
	workout.Distance = distanceMeters
	workout.Track = trackName

	if err := s.validateWorkout(workout); err != nil {
		return nil, err
	}

	userDir := s.userDir(nickname)
	if err := os.MkdirAll(userDir, 0700); err != nil {
		return nil, fmt.Errorf("create user dir: %w", err)
	}

	id, err := s.allocateWorkoutID(userDir)
	if err != nil {
		return nil, err
	}
	workout.ID = id
	workout.Name = trimWorkoutName(workout.Name)
	workout.Description = trimWorkoutDescription(workout.Description)

	workoutDirPath := workoutDir(userDir, workout.StartDate, workout.ID)
	if _, err := os.Stat(workoutDirPath); err == nil {
		return nil, ErrWorkoutExists
	}
	if err := os.MkdirAll(workoutDirPath, 0700); err != nil {
		return nil, fmt.Errorf("create workout dir: %w", err)
	}

	cleanup := func() {
		_ = os.RemoveAll(workoutDirPath)
	}

	trackPath := filepath.Join(workoutDirPath, trackName)
	if err := os.WriteFile(trackPath, track.Data, 0600); err != nil {
		cleanup()
		return nil, fmt.Errorf("write track: %w", err)
	}

	filePath := workoutFilePath(userDir, workout.StartDate, workout.ID)
	if err := writeWorkoutYAML(filePath, workout); err != nil {
		cleanup()
		return nil, err
	}

	if parsed.HasGPS() {
		if preview, err := maprender.RenderPreview(parsed.Points); err != nil {
			log.Printf("map preview render failed for workout %s: %v", workout.ID, err)
		} else if len(preview) > 0 {
			previewPath := filepath.Join(workoutDirPath, MapPreviewFileName)
			if err := os.WriteFile(previewPath, preview, 0600); err != nil {
				log.Printf("map preview write failed for workout %s: %v", workout.ID, err)
			} else {
				workout.HasMapPreview = true
			}
		}
	}

	result := *workout
	return &result, nil
}

func (s *Store) TrackFile(nickname, workoutID string) ([]byte, string, error) {
	dir, err := s.findWorkoutDir(nickname, workoutID)
	if err != nil {
		return nil, "", err
	}

	workout, err := readWorkoutFromDir(dir)
	if err != nil {
		return nil, "", err
	}
	if workout.Track == "" {
		return nil, "", ErrWorkoutNotFound
	}

	path := filepath.Join(dir, workout.Track)
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, "", ErrWorkoutNotFound
		}
		return nil, "", fmt.Errorf("read track: %w", err)
	}
	return data, workout.Track, nil
}

func readWorkoutFromDir(dir string) (*Workout, error) {
	dirName := filepath.Base(dir)
	filePath := filepath.Join(dir, dirName+".yaml")
	data, err := os.ReadFile(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, ErrWorkoutNotFound
		}
		return nil, fmt.Errorf("read workout: %w", err)
	}

	var workout Workout
	if err := yaml.Unmarshal(data, &workout); err != nil {
		return nil, fmt.Errorf("parse workout: %w", err)
	}
	return &workout, nil
}

func (s *Store) MapPreviewPath(nickname, workoutID string) (string, error) {
	dir, err := s.findWorkoutDir(nickname, workoutID)
	if err != nil {
		return "", err
	}
	path := filepath.Join(dir, MapPreviewFileName)
	if _, err := os.Stat(path); err != nil {
		if os.IsNotExist(err) {
			return "", ErrWorkoutNotFound
		}
		return "", err
	}
	return path, nil
}

func (s *Store) findWorkoutDir(nickname, workoutID string) (string, error) {
	userDir := s.userDir(nickname)
	entries, err := os.ReadDir(userDir)
	if err != nil {
		if os.IsNotExist(err) {
			return "", ErrWorkoutNotFound
		}
		return "", err
	}

	suffix := "-" + workoutID
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		if len(entry.Name()) > len(suffix) && entry.Name()[len(entry.Name())-len(suffix):] == suffix {
			return filepath.Join(userDir, entry.Name()), nil
		}
	}
	return "", ErrWorkoutNotFound
}

func workoutHasMapPreview(dir string) bool {
	info, err := os.Stat(filepath.Join(dir, MapPreviewFileName))
	return err == nil && !info.IsDir()
}
