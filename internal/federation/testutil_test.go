package federation

import (
	blobfs "github.com/solargate/grom/internal/storage/blob/fs"
)

func newTestInboxStore(dir string) *WorkoutInboxStore {
	return NewWorkoutInboxStore(dir, blobfs.NewStore(dir))
}
