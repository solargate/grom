package federation

import (
	blobfs "github.com/solargate/grom/internal/storage/blob/fs"
	"github.com/solargate/grom/internal/workouts"
)

func newTestInboxStore(dir string) *WorkoutInboxStore {
	blobs := blobfs.NewStore(dir)
	charts := workouts.NewBlobSpeedChartStore(blobs)
	return NewWorkoutInboxStore(dir, blobs, charts)
}
