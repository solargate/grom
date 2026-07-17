package workouts

import (
	"context"

	"github.com/solargate/grom/internal/storage/blob"
	"github.com/solargate/grom/internal/storage/keys"
)

// Service provides workout business operations over a metadata repository and blob store.
type Service struct {
	repo  Repository
	blobs blob.Store
}

func NewService(repo Repository, blobs blob.Store) *Service {
	return &Service{repo: repo, blobs: blobs}
}

func (s *Service) List(nickname string) ([]Workout, error) {
	items, err := s.repo.List(nickname)
	if err != nil {
		return nil, err
	}
	for i := range items {
		s.enrichWorkout(nickname, &items[i])
	}
	return items, nil
}

func (s *Service) Delete(nickname, workoutID string) error {
	return s.repo.Delete(nickname, workoutID)
}

func (s *Service) RemoveEquipmentFromAll(nickname, equipmentID string) error {
	return s.repo.RemoveEquipmentFromAll(nickname, equipmentID)
}

func (s *Service) HasStravaActivityID(nickname, stravaActivityID string) (bool, error) {
	return s.repo.HasStravaActivityID(nickname, stravaActivityID)
}

func (s *Service) Create(nickname string, workout *Workout) (*Workout, error) {
	if err := validateWorkout(workout); err != nil {
		return nil, err
	}
	return s.repo.Create(nickname, workout)
}

func (s *Service) enrichWorkout(nickname string, w *Workout) {
	dirName, err := s.repo.WorkoutDirName(nickname, w.ID)
	if err != nil {
		return
	}
	ctx := context.Background()
	if ok, _ := s.blobs.Exists(ctx, keys.WorkoutMapPreview(nickname, dirName)); ok {
		w.HasMapPreview = true
	}
	w.HasMedia = len(w.MediaFiles) > 0
}
