package federation

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/solargate/travka/internal/data"
	"github.com/solargate/travka/internal/maprender"
	"github.com/solargate/travka/internal/tracks"
	"github.com/solargate/travka/internal/workouts"
	"gopkg.in/yaml.v3"
)

type WorkoutInboxStore struct {
	dataDir string
}

func NewWorkoutInboxStore(dataDir string) *WorkoutInboxStore {
	return &WorkoutInboxStore{dataDir: dataDir}
}

func (s *WorkoutInboxStore) inboxDir(viewerNickname string) string {
	return filepath.Join(data.UserDir(s.dataDir, viewerNickname), "federation", "inbox", "workouts")
}

func ownerDirName(handle string) string {
	return strings.NewReplacer("@", "_", ":", "_", "/", "_").Replace(handle)
}

func federatedTrackPath(ownerDir, workoutID, trackName string) string {
	return filepath.Join(ownerDir, workoutID+"_"+trackName)
}

func federatedMapPreviewPath(ownerDir, workoutID string) string {
	return filepath.Join(ownerDir, workoutID+"_"+workouts.MapPreviewFileName)
}

func (s *WorkoutInboxStore) findOwnerDir(viewerNickname, ownerNickname string) (string, error) {
	root := s.inboxDir(viewerNickname)
	entries, err := os.ReadDir(root)
	if err != nil {
		if os.IsNotExist(err) {
			return "", workouts.ErrWorkoutNotFound
		}
		return "", err
	}
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		if strings.EqualFold(ownerNicknameFromDir(entry.Name()), ownerNickname) {
			return filepath.Join(root, entry.Name()), nil
		}
	}
	return "", workouts.ErrWorkoutNotFound
}

func ownerNicknameFromDir(dirName string) string {
	handle := strings.ReplaceAll(dirName, "_", "@")
	if idx := strings.Index(handle, "@"); idx > 0 {
		return handle[:idx]
	}
	return handle
}

func ownerHandleFromDir(dirName string) string {
	return strings.ReplaceAll(dirName, "_", "@")
}

func (s *WorkoutInboxStore) Save(viewerNickname, ownerHandle string, workout *workouts.Workout, trackData []byte) error {
	dir := filepath.Join(s.inboxDir(viewerNickname), ownerDirName(ownerHandle))
	if err := os.MkdirAll(dir, 0700); err != nil {
		return err
	}

	if workout.Track != "" && len(trackData) > 0 {
		trackPath := federatedTrackPath(dir, workout.ID, workout.Track)
		tmp := trackPath + ".tmp"
		if err := os.WriteFile(tmp, trackData, 0600); err != nil {
			return err
		}
		if err := os.Rename(tmp, trackPath); err != nil {
			return err
		}

		if parsed, err := tracks.Parse(trackData, workout.Track); err == nil && parsed.HasGPS() {
			if preview, err := maprender.RenderPreview(parsed.Points); err != nil {
				log.Printf("federated map preview render failed for workout %s: %v", workout.ID, err)
			} else if len(preview) > 0 {
				previewPath := federatedMapPreviewPath(dir, workout.ID)
				tmpPreview := previewPath + ".tmp"
				if err := os.WriteFile(tmpPreview, preview, 0600); err != nil {
					log.Printf("federated map preview write failed for workout %s: %v", workout.ID, err)
				} else if err := os.Rename(tmpPreview, previewPath); err != nil {
					log.Printf("federated map preview rename failed for workout %s: %v", workout.ID, err)
				} else {
					workout.HasMapPreview = true
				}
			}
		}
	}

	path := filepath.Join(dir, workout.ID+".yaml")
	data, err := yaml.Marshal(workout)
	if err != nil {
		return err
	}
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, data, 0600); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}

func (s *WorkoutInboxStore) readWorkout(ownerDir, workoutID string) (*workouts.Workout, error) {
	path := filepath.Join(ownerDir, workoutID+".yaml")
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, workouts.ErrWorkoutNotFound
		}
		return nil, err
	}
	var workout workouts.Workout
	if err := yaml.Unmarshal(data, &workout); err != nil {
		return nil, fmt.Errorf("parse federated workout: %w", err)
	}
	if _, err := os.Stat(federatedMapPreviewPath(ownerDir, workoutID)); err == nil {
		workout.HasMapPreview = true
	}
	return &workout, nil
}

func (s *WorkoutInboxStore) TrackFile(viewerNickname, ownerNickname, workoutID string) ([]byte, string, error) {
	ownerDir, err := s.findOwnerDir(viewerNickname, ownerNickname)
	if err != nil {
		return nil, "", err
	}
	workout, err := s.readWorkout(ownerDir, workoutID)
	if err != nil {
		return nil, "", err
	}
	if workout.Track == "" {
		return nil, "", workouts.ErrWorkoutNotFound
	}
	path := federatedTrackPath(ownerDir, workoutID, workout.Track)
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, "", workouts.ErrWorkoutNotFound
		}
		return nil, "", fmt.Errorf("read federated track: %w", err)
	}
	return data, workout.Track, nil
}

func (s *WorkoutInboxStore) MapPreviewPath(viewerNickname, ownerNickname, workoutID string) (string, error) {
	ownerDir, err := s.findOwnerDir(viewerNickname, ownerNickname)
	if err != nil {
		return "", err
	}
	if _, err := s.readWorkout(ownerDir, workoutID); err != nil {
		return "", err
	}
	path := federatedMapPreviewPath(ownerDir, workoutID)
	if _, err := os.Stat(path); err != nil {
		if os.IsNotExist(err) {
			return "", workouts.ErrWorkoutNotFound
		}
		return "", err
	}
	return path, nil
}

func (s *WorkoutInboxStore) List(viewerNickname string) ([]workouts.FeedWorkout, error) {
	root := s.inboxDir(viewerNickname)
	entries, err := os.ReadDir(root)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	items := make([]workouts.FeedWorkout, 0)
	for _, ownerEntry := range entries {
		if !ownerEntry.IsDir() {
			continue
		}
		ownerDir := filepath.Join(root, ownerEntry.Name())
		handle := ownerHandleFromDir(ownerEntry.Name())
		nickname := ownerNicknameFromDir(ownerEntry.Name())
		files, err := os.ReadDir(ownerDir)
		if err != nil {
			return nil, err
		}
		for _, fileEntry := range files {
			if fileEntry.IsDir() || !strings.HasSuffix(fileEntry.Name(), ".yaml") {
				continue
			}
			workoutID := strings.TrimSuffix(fileEntry.Name(), ".yaml")
			workout, err := s.readWorkout(ownerDir, workoutID)
			if err != nil {
				return nil, err
			}
			items = append(items, workouts.FeedWorkout{
				Workout: *workout,
				Owner:   nickname,
				Author: workouts.FeedAuthor{
					Nickname: nickname,
					Handle:   handle,
					IsLocal:  false,
				},
			})
		}
	}

	sort.Slice(items, func(i, j int) bool {
		return items[i].StartDate.After(items[j].StartDate)
	})
	return items, nil
}
