package workouts

import (
	"context"
	"fmt"

	"github.com/solargate/grom/internal/equipment"
	"github.com/solargate/grom/internal/storage/blob"
	"github.com/solargate/grom/internal/storage/keys"
	"github.com/solargate/grom/internal/tracks"
)

// Service provides workout business operations over a metadata repository and blob store.
type Service struct {
	repo        Repository
	blobs       blob.Store
	speedCharts SpeedChartStore
	equipment   EquipmentCatalog
}

func NewService(repo Repository, blobs blob.Store, speedCharts SpeedChartStore) *Service {
	return &Service{repo: repo, blobs: blobs, speedCharts: speedCharts}
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
	s.enrichList(nickname, items)
	return items, nil
}

// ListPage returns a cursor page of workouts for nickname (enriched).
func (s *Service) ListPage(nickname string, cursor *Cursor, limit int) ([]Workout, bool, error) {
	limit = ClampLimit(limit)
	items, hasMore, err := s.repo.ListPage(nickname, cursor, limit)
	if err != nil {
		return nil, false, err
	}
	s.enrichList(nickname, items)
	return items, hasMore, nil
}

func (s *Service) enrichList(nickname string, items []Workout) {
	byID := s.loadEquipmentByID(nickname, items)
	for i := range items {
		s.enrichWorkout(nickname, &items[i])
		ApplyEquipmentCatalog(items[i].Equipment, byID)
	}
}

func (s *Service) Get(nickname, workoutID string) (*Workout, error) {
	workout, err := s.repo.Get(nickname, workoutID)
	if err != nil {
		return nil, err
	}
	s.enrichWorkout(nickname, workout)
	byID := s.loadEquipmentByID(nickname, []Workout{*workout})
	ApplyEquipmentCatalog(workout.Equipment, byID)
	return workout, nil
}

// GetSpeedChart returns workout metadata and precomputed chart samples for /speed.
func (s *Service) GetSpeedChart(nickname, workoutID string) (*Workout, []SpeedSample, error) {
	workout, err := s.repo.Get(nickname, workoutID)
	if err != nil {
		return nil, nil, err
	}
	if workout.Track == "" || s.speedCharts == nil {
		return workout, nil, nil
	}
	dirName := keys.WorkoutDirName(workout.StartDate, workout.ID)
	samples, err := s.speedCharts.ReadLocal(context.Background(), nickname, dirName)
	if err != nil {
		return nil, nil, fmt.Errorf("read speed chart: %w", err)
	}
	return workout, samples, nil
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
	dirName := keys.WorkoutDirName(w.StartDate, w.ID)
	ctx := context.Background()
	if ok, _ := s.blobs.Exists(ctx, keys.WorkoutMapPreview(nickname, dirName)); ok {
		w.HasMapPreview = true
	}
	w.HasMedia = len(w.MediaFiles) > 0
}

func (s *Service) writeSpeedChart(nickname, dirName string, parsed *tracks.Data) error {
	if s.speedCharts == nil {
		return fmt.Errorf("speed chart store is nil")
	}
	return s.speedCharts.WriteLocal(context.Background(), nickname, dirName, BuildSpeedChartSamples(parsed))
}
