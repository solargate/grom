package workouts_test

import (
	"github.com/solargate/grom/internal/storage/file"
	blobfs "github.com/solargate/grom/internal/storage/blob/fs"
	"github.com/solargate/grom/internal/workouts"
)

func newTestService(dir string) *workouts.Service {
	blobs := blobfs.NewStore(dir)
	charts := workouts.NewBlobSpeedChartStore(blobs)
	return workouts.NewService(file.NewWorkoutsStore(dir), blobs, charts)
}

func newTestServiceWithEquipment(dir string) *workouts.Service {
	svc := newTestService(dir)
	svc.SetEquipmentCatalog(file.NewEquipmentStore(dir))
	return svc
}
