package workouts

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"

	"github.com/solargate/grom/internal/maprender"
	"github.com/solargate/grom/internal/tracks"
)

const MapPreviewFileName = "map-preview.webp"

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

	trackName, parsed, err := prepareTrackInput(track)
	if err != nil {
		return nil, err
	}

	startDate := workout.StartDate
	durationSeconds := workout.DurationSeconds
	distanceMeters := workout.Distance
	parsed.ApplyToWorkout(&startDate, &durationSeconds, &distanceMeters)

	workout.StartDate = startDate
	workout.DurationSeconds = durationSeconds
	workout.Distance = distanceMeters
	workout.Track = trackName
	workout.Device = deviceForTrack(trackName, parsed)

	if err := s.validateWorkout(workout); err != nil {
		return nil, err
	}

	created, workoutDirPath, cleanup, err := s.prepareNewWorkoutDir(nickname, workout)
	if err != nil {
		return nil, err
	}

	if err := writeTrackArtifacts(workoutDirPath, track.Data, parsed, workout); err != nil {
		cleanup()
		return nil, err
	}

	filePath := workoutFilePath(s.userDir(nickname), workout.StartDate, workout.ID)
	if err := writeWorkoutYAML(filePath, workout); err != nil {
		cleanup()
		return nil, err
	}

	result := *created
	return &result, nil
}

// AttachTrack saves a track file and map preview for an existing workout without
// overwriting start_date, duration_seconds, or distance from the track.
func (s *Store) AttachTrack(nickname string, workout *Workout, track *TrackInput) (*Workout, error) {
	if workout == nil || workout.ID == "" {
		return nil, ErrInvalidWorkout
	}
	if track == nil {
		return workout, nil
	}

	trackName, parsed, err := prepareTrackInput(track)
	if err != nil {
		return nil, err
	}

	workoutDirPath, err := s.findWorkoutDir(nickname, workout.ID)
	if err != nil {
		return nil, err
	}

	workout.Track = trackName
	if device := deviceForTrack(trackName, parsed); device != DeviceGrom {
		workout.Device = device
	}

	if err := writeTrackArtifacts(workoutDirPath, track.Data, parsed, workout); err != nil {
		return nil, err
	}

	dirName := filepath.Base(workoutDirPath)
	filePath := filepath.Join(workoutDirPath, dirName+".yaml")
	if err := writeWorkoutYAML(filePath, workout); err != nil {
		return nil, err
	}

	result := *workout
	return &result, nil
}

func prepareTrackInput(track *TrackInput) (string, *tracks.Data, error) {
	trackName, err := tracks.TrackFileName(track.Filename)
	if err != nil {
		return "", nil, err
	}

	parsed := track.Parsed
	if parsed == nil {
		parsed, err = tracks.Parse(track.Data, track.Filename)
		if err != nil {
			return "", nil, err
		}
	}

	return trackName, parsed, nil
}

func (s *Store) prepareNewWorkoutDir(nickname string, workout *Workout) (*Workout, string, func(), error) {
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
	workout.Name = trimWorkoutName(workout.Name)
	workout.Description = trimWorkoutDescription(workout.Description)

	workoutDirPath := workoutDir(userDir, workout.StartDate, workout.ID)
	if _, err := os.Stat(workoutDirPath); err == nil {
		return nil, "", nil, ErrWorkoutExists
	}
	if err := os.MkdirAll(workoutDirPath, 0700); err != nil {
		return nil, "", nil, fmt.Errorf("create workout dir: %w", err)
	}

	cleanup := func() {
		_ = os.RemoveAll(workoutDirPath)
	}

	result := *workout
	return &result, workoutDirPath, cleanup, nil
}

func writeTrackArtifacts(workoutDirPath string, trackData []byte, parsed *tracks.Data, workout *Workout) error {
	trackPath := filepath.Join(workoutDirPath, workout.Track)
	if err := os.WriteFile(trackPath, trackData, 0600); err != nil {
		return fmt.Errorf("write track: %w", err)
	}

	if parsed != nil && parsed.HasGPS() {
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

	return nil
}

func deviceForTrack(trackName string, parsed *tracks.Data) string {
	if trackName == tracks.TrackFileFIT && parsed != nil && parsed.Device != nil {
		if device := strings.TrimSpace(*parsed.Device); device != "" {
			return device
		}
	}
	return DeviceGrom
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
	wd := workoutsDir(s.userDir(nickname))
	entries, err := os.ReadDir(wd)
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
			return filepath.Join(wd, entry.Name()), nil
		}
	}
	return "", ErrWorkoutNotFound
}

func (s *Store) HasStravaActivityID(nickname, stravaActivityID string) (bool, error) {
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

func workoutHasMapPreview(dir string) bool {
	info, err := os.Stat(filepath.Join(dir, MapPreviewFileName))
	return err == nil && !info.IsDir()
}
