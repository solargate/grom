package distance_test

import (
	"testing"

	"github.com/solargate/grom/internal/equipment"
	"github.com/solargate/grom/internal/equipment/distance"
	"github.com/solargate/grom/internal/storage/file"
	"github.com/solargate/grom/internal/workouts"
)

type stubWorkoutLister struct {
	items []workouts.Workout
	err   error
}

func (s stubWorkoutLister) List(nickname string) ([]workouts.Workout, error) {
	if s.err != nil {
		return nil, s.err
	}
	return s.items, nil
}

func TestRecalculateForIDsSumsWorkoutDistances(t *testing.T) {
	dir := t.TempDir()
	store := file.NewEquipmentStore(dir)
	bike, err := store.Create("athlete", &equipment.Equipment{Type: equipment.TypeBike, Name: "Road"})
	if err != nil {
		t.Fatalf("Create bike: %v", err)
	}
	shoes, err := store.Create("athlete", &equipment.Equipment{Type: equipment.TypeShoes, Name: "Run"})
	if err != nil {
		t.Fatalf("Create shoes: %v", err)
	}

	svc := distance.NewService(store, stubWorkoutLister{
		items: []workouts.Workout{
			{
				Distance: 5000,
				Equipment: []workouts.WorkoutEquipment{
					{ID: bike.ID},
					{ID: shoes.ID},
				},
			},
			{
				Distance: 3200,
				Equipment: []workouts.WorkoutEquipment{
					{ID: bike.ID},
				},
			},
			{
				Distance: 0,
				Equipment: []workouts.WorkoutEquipment{
					{ID: shoes.ID},
				},
			},
		},
	})

	if err := svc.RecalculateForIDs("athlete", []string{bike.ID, shoes.ID}); err != nil {
		t.Fatalf("RecalculateForIDs() error = %v", err)
	}

	gotBike, err := store.FindByID("athlete", bike.ID)
	if err != nil {
		t.Fatalf("FindByID bike: %v", err)
	}
	if gotBike.Distance != 8200 {
		t.Fatalf("bike distance = %v, want 8200", gotBike.Distance)
	}

	gotShoes, err := store.FindByID("athlete", shoes.ID)
	if err != nil {
		t.Fatalf("FindByID shoes: %v", err)
	}
	if gotShoes.Distance != 5000 {
		t.Fatalf("shoes distance = %v, want 5000", gotShoes.Distance)
	}
}

func TestRecalculateForIDsZerosMissingUsage(t *testing.T) {
	dir := t.TempDir()
	store := file.NewEquipmentStore(dir)
	item, err := store.Create("athlete", &equipment.Equipment{Type: equipment.TypeOther, Name: "Pack"})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if err := store.SetDistance("athlete", item.ID, 1500); err != nil {
		t.Fatalf("SetDistance: %v", err)
	}

	svc := distance.NewService(store, stubWorkoutLister{items: nil})
	if err := svc.RecalculateForIDs("athlete", []string{item.ID}); err != nil {
		t.Fatalf("RecalculateForIDs() error = %v", err)
	}

	got, err := store.FindByID("athlete", item.ID)
	if err != nil {
		t.Fatalf("FindByID: %v", err)
	}
	if got.Distance != 0 {
		t.Fatalf("distance = %v, want 0", got.Distance)
	}
}

func TestCollectWorkoutEquipmentIDsDedupes(t *testing.T) {
	ids := distance.CollectWorkoutEquipmentIDs([]workouts.WorkoutEquipment{
		{ID: "a"},
		{ID: "b"},
		{ID: "a"},
		{ID: ""},
	})
	if len(ids) != 2 || ids[0] != "a" || ids[1] != "b" {
		t.Fatalf("ids = %v, want [a b]", ids)
	}
}

func TestScheduleRecalculateWaitDrainsWorkers(t *testing.T) {
	dir := t.TempDir()
	store := file.NewEquipmentStore(dir)
	item, err := store.Create("athlete", &equipment.Equipment{Type: equipment.TypeShoes, Name: "Road"})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	svc := distance.NewService(store, stubWorkoutLister{
		items: []workouts.Workout{{
			Distance:  4200,
			Equipment: []workouts.WorkoutEquipment{{ID: item.ID}},
		}},
	})
	svc.ScheduleRecalculateForIDs("athlete", []string{item.ID})
	svc.Wait()

	got, err := store.FindByID("athlete", item.ID)
	if err != nil {
		t.Fatalf("FindByID: %v", err)
	}
	if got.Distance != 4200 {
		t.Fatalf("distance = %v, want 4200", got.Distance)
	}
}
