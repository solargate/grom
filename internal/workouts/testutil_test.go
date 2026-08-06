package workouts_test

import (
	"path/filepath"
	"testing"

	storebbolt "github.com/solargate/grom/internal/storage/bbolt"
	"github.com/solargate/grom/internal/storage/file"
	blobfs "github.com/solargate/grom/internal/storage/blob/fs"
	"github.com/solargate/grom/internal/workouts"
)

func newTestService(dir string) *workouts.Service {
	blobs := blobfs.NewStore(dir)
	speedCharts := workouts.NewBlobSpeedChartStore(blobs)
	hrCharts := workouts.NewBlobHeartRateChartStore(blobs)
	return workouts.NewService(file.NewWorkoutsStore(dir), blobs, speedCharts, hrCharts)
}

func newTestServiceWithEquipment(dir string) *workouts.Service {
	svc := newTestService(dir)
	svc.SetEquipmentCatalog(file.NewEquipmentStore(dir))
	return svc
}

// newTestServiceForDriver returns a workouts.Service backed by file or bbolt metadata.
func newTestServiceForDriver(t *testing.T, driver string) *workouts.Service {
	t.Helper()
	dir := t.TempDir()
	switch driver {
	case "file":
		return newTestService(dir)
	case "bbolt":
		backend, err := storebbolt.Open(filepath.Join(dir, "grom.db"), dir)
		if err != nil {
			t.Fatalf("open bbolt: %v", err)
		}
		t.Cleanup(func() { _ = backend.Close() })
		return backend.Workouts()
	default:
		t.Fatalf("unknown storage driver %q", driver)
		return nil
	}
}
