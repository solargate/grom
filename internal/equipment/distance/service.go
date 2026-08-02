package distance

import (
	"log/slog"
	"sync"

	"github.com/solargate/grom/internal/equipment"
	"github.com/solargate/grom/internal/workouts"
)

// WorkoutLister lists workouts for mileage aggregation.
type WorkoutLister interface {
	List(nickname string) ([]workouts.Workout, error)
}

// Service recalculates cached equipment mileage from workout distances.
type Service struct {
	equipment equipment.Repository
	workouts  WorkoutLister
	locks     sync.Map
	wg        sync.WaitGroup
}

func NewService(equipment equipment.Repository, workouts WorkoutLister) *Service {
	return &Service{
		equipment: equipment,
		workouts:  workouts,
	}
}

func (s *Service) lockFor(nickname string) *sync.Mutex {
	mu, _ := s.locks.LoadOrStore(nickname, &sync.Mutex{})
	return mu.(*sync.Mutex)
}

// Wait blocks until all scheduled recalculations finish.
func (s *Service) Wait() {
	s.wg.Wait()
}

// RecalculateForIDs sums workout distances per equipment ID and stores totals in meters.
func (s *Service) RecalculateForIDs(nickname string, ids []string) error {
	unique := dedupeIDs(ids)
	if len(unique) == 0 {
		return nil
	}

	workoutItems, err := s.workouts.List(nickname)
	if err != nil {
		return err
	}

	totals := sumDistanceByEquipmentID(workoutItems)
	for _, id := range unique {
		if err := s.equipment.SetDistance(nickname, id, totals[id]); err != nil {
			return err
		}
	}
	return nil
}

// RecalculateAll updates cached mileage for every equipment item owned by nickname.
func (s *Service) RecalculateAll(nickname string) error {
	items, err := s.equipment.List(nickname)
	if err != nil {
		return err
	}
	ids := make([]string, 0, len(items))
	for i := range items {
		if items[i].ID != "" {
			ids = append(ids, items[i].ID)
		}
	}
	return s.RecalculateForIDs(nickname, ids)
}

// ScheduleRecalculateForIDs runs RecalculateForIDs asynchronously for nickname.
func (s *Service) ScheduleRecalculateForIDs(nickname string, ids []string) {
	if len(ids) == 0 {
		return
	}
	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		s.runLocked(nickname, func() error {
			return s.RecalculateForIDs(nickname, ids)
		}, "for_ids")
	}()
}

// ScheduleRecalculateAll runs RecalculateAll asynchronously for nickname.
func (s *Service) ScheduleRecalculateAll(nickname string) {
	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		s.runLocked(nickname, func() error {
			return s.RecalculateAll(nickname)
		}, "all")
	}()
}

func (s *Service) runLocked(nickname string, fn func() error, scope string) {
	mu := s.lockFor(nickname)
	mu.Lock()
	defer mu.Unlock()

	if err := fn(); err != nil {
		slog.Error("equipment distance recalculation failed",
			"user", nickname,
			"scope", scope,
			"err", err,
		)
	}
}

func dedupeIDs(ids []string) []string {
	if len(ids) == 0 {
		return nil
	}
	seen := make(map[string]struct{}, len(ids))
	unique := make([]string, 0, len(ids))
	for _, id := range ids {
		if id == "" {
			continue
		}
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		unique = append(unique, id)
	}
	return unique
}

func sumDistanceByEquipmentID(workoutItems []workouts.Workout) map[string]float64 {
	totals := make(map[string]float64)
	for i := range workoutItems {
		distance := workoutItems[i].Distance
		for _, eq := range workoutItems[i].Equipment {
			if eq.ID == "" {
				continue
			}
			totals[eq.ID] += distance
		}
	}
	return totals
}

// CollectWorkoutEquipmentIDs returns deduplicated equipment IDs from workout items.
func CollectWorkoutEquipmentIDs(items []workouts.WorkoutEquipment) []string {
	if len(items) == 0 {
		return nil
	}
	ids := make([]string, 0, len(items))
	for i := range items {
		ids = append(ids, items[i].ID)
	}
	return dedupeIDs(ids)
}
