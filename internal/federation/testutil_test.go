package federation

import (
	blobfs "github.com/solargate/grom/internal/storage/blob/fs"
	"github.com/solargate/grom/internal/workouts"
)

func newTestInboxStore(dir string) *WorkoutInboxStore {
	blobs := blobfs.NewStore(dir)
	speedCharts := workouts.NewBlobSpeedChartStore(blobs)
	hrCharts := workouts.NewBlobHeartRateChartStore(blobs)
	cadenceCharts := workouts.NewBlobCadenceChartStore(blobs)
	return NewWorkoutInboxStore(dir, blobs, speedCharts, hrCharts, cadenceCharts)
}
