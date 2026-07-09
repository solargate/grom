package federation

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/solargate/travka/internal/config"
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
	return filepath.Join(s.dataDir, viewerNickname, "federation", "inbox", "workouts")
}

func ownerDirName(handle string) string {
	return strings.NewReplacer("@", "_", ":", "_", "/", "_").Replace(handle)
}

func (s *WorkoutInboxStore) Save(viewerNickname, ownerHandle string, workout *workouts.Workout) error {
	dir := filepath.Join(s.inboxDir(viewerNickname), ownerDirName(ownerHandle))
	if err := os.MkdirAll(dir, 0700); err != nil {
		return err
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

func (s *WorkoutInboxStore) List(viewerNickname string) ([]workouts.FeedWorkout, error) {
	root := s.inboxDir(viewerNickname)
	entries, err := os.ReadDir(root)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	domain := config.Cfg.Federation.Domain
	if domain == "" {
		domain = "localhost"
	}

	items := make([]workouts.FeedWorkout, 0)
	for _, ownerEntry := range entries {
		if !ownerEntry.IsDir() {
			continue
		}
		ownerDir := filepath.Join(root, ownerEntry.Name())
		files, err := os.ReadDir(ownerDir)
		if err != nil {
			return nil, err
		}
		for _, fileEntry := range files {
			if fileEntry.IsDir() || !strings.HasSuffix(fileEntry.Name(), ".yaml") {
				continue
			}
			data, err := os.ReadFile(filepath.Join(ownerDir, fileEntry.Name()))
			if err != nil {
				return nil, err
			}
			var workout workouts.Workout
			if err := yaml.Unmarshal(data, &workout); err != nil {
				return nil, fmt.Errorf("parse federated workout: %w", err)
			}
			handle := strings.ReplaceAll(ownerEntry.Name(), "_", "@")
			nickname := handle
			if idx := strings.Index(handle, "@"); idx > 0 {
				nickname = handle[:idx]
			}
			items = append(items, workouts.FeedWorkout{
				Workout: workout,
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
