package workouts

import (
	"context"

	"github.com/solargate/grom/internal/equipment"
	"github.com/solargate/grom/internal/storage/blob"
	"github.com/solargate/grom/internal/storage/keys"
)

// Service provides workout business operations over a metadata repository and blob store.
type Service struct {
	repo      Repository
	blobs     blob.Store
	equipment EquipmentCatalog
}

func NewService(repo Repository, blobs blob.Store) *Service {
	return &Service{repo: repo, blobs: blobs}
}

// SetEquipmentCatalog enables read-time equipment name/type enrichment in List.
func (s *Service) SetEquipmentCatalog(catalog EquipmentCatalog) {
	s.equipment = catalog
}

func (s *Service) List(nickname string) ([]Workout, error) {
	items, err := s.repo.List(nickname)
	if err != nil {
		return nil, err
	}
	byID := s.loadEquipmentByID(nickname, items)
	for i := range items {
		s.enrichWorkout(nickname, &items[i])
		ApplyEquipmentCatalog(items[i].Equipment, byID)
	}
	return items, nil
}

func (s *Service) loadEquipmentByID(nickname string, items []Workout) map[string]equipment.Equipment {
	if s.equipment == nil {
		return nil
	}
	ids := collectEquipmentIDs(items)
	if len(ids) == 0 {
		return nil
	}
	catalogItems, err := s.equipment.FindByIDs(nickname, ids)
	if err != nil {
		return nil
	}
	return equipmentByID(catalogItems)
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

// Update applies editable metadata fields onto an existing workout.
// Track, media, device, strava id, and track-derived stats not present on patch are preserved.
func (s *Service) Update(nickname string, workoutID string, patch *Workout) (*Workout, error) {
	if patch == nil {
		return nil, ErrInvalidWorkout
	}
	existing, err := s.repo.Get(nickname, workoutID)
	if err != nil {
		return nil, err
	}

	existing.Name = patch.Name
	existing.Description = patch.Description
	existing.SportType = patch.SportType
	existing.StartDate = patch.StartDate
	existing.DurationSeconds = patch.DurationSeconds
	existing.DurationTotalSeconds = patch.DurationTotalSeconds
	existing.Distance = patch.Distance
	existing.SpeedMaxKmh = patch.SpeedMaxKmh
	existing.SpeedAvgKmh = patch.SpeedAvgKmh
	existing.Equipment = patch.Equipment

	if err := validateWorkout(existing); err != nil {
		return nil, err
	}

	updated, err := s.repo.Update(nickname, existing)
	if err != nil {
		return nil, err
	}
	s.enrichWorkout(nickname, updated)
	byID := s.loadEquipmentByID(nickname, []Workout{*updated})
	ApplyEquipmentCatalog(updated.Equipment, byID)
	return updated, nil
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
