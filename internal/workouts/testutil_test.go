package workouts_test

import (
	"github.com/solargate/grom/internal/storage/file"
	blobfs "github.com/solargate/grom/internal/storage/blob/fs"
	"github.com/solargate/grom/internal/workouts"
)

func newTestService(dir string) *workouts.Service {
	return workouts.NewService(file.NewWorkoutsStore(dir), blobfs.NewStore(dir))
}

func newTestServiceWithEquipment(dir string) *workouts.Service {
	svc := workouts.NewService(file.NewWorkoutsStore(dir), blobfs.NewStore(dir))
	svc.SetEquipmentCatalog(file.NewEquipmentStore(dir))
	return svc
}
