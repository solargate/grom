package migrate

import (
	"context"
	"fmt"

	"github.com/solargate/grom/internal/storage"
	storebbolt "github.com/solargate/grom/internal/storage/bbolt"
	"github.com/solargate/grom/internal/storage/file"
	"github.com/solargate/grom/internal/storage/keys"
	"github.com/solargate/grom/internal/workouts"
)

func chartStores(backend storage.Backend) (workouts.SpeedChartStore, workouts.HeartRateChartStore, error) {
	switch b := backend.(type) {
	case *file.Backend:
		blobs := b.Blobs()
		return workouts.NewBlobSpeedChartStore(blobs), workouts.NewBlobHeartRateChartStore(blobs), nil
	case *storebbolt.Backend:
		return storebbolt.NewSpeedChartStore(b.DB()), storebbolt.NewHeartRateChartStore(b.DB()), nil
	default:
		return nil, nil, fmt.Errorf("unsupported backend type %T", backend)
	}
}

func copyLocalCharts(src, dst storage.Backend, nickname string, w *workouts.Workout) error {
	if w == nil || w.Track == "" {
		return nil
	}
	srcSpeed, srcHR, err := chartStores(src)
	if err != nil {
		return err
	}
	dstSpeed, dstHR, err := chartStores(dst)
	if err != nil {
		return err
	}
	ctx := context.Background()
	dirName := keys.WorkoutDirName(w.StartDate, w.ID)

	speedSamples, err := srcSpeed.ReadLocal(ctx, nickname, dirName)
	if err != nil {
		return fmt.Errorf("read local speed chart: %w", err)
	}
	if len(speedSamples) > 0 {
		if err := dstSpeed.WriteLocal(ctx, nickname, dirName, speedSamples); err != nil {
			return fmt.Errorf("write local speed chart: %w", err)
		}
	}

	hrSamples, err := srcHR.ReadLocal(ctx, nickname, dirName)
	if err != nil {
		return fmt.Errorf("read local heart rate chart: %w", err)
	}
	if len(hrSamples) > 0 {
		if err := dstHR.WriteLocal(ctx, nickname, dirName, hrSamples); err != nil {
			return fmt.Errorf("write local heart rate chart: %w", err)
		}
	}
	return nil
}

func copyFederatedCharts(src, dst storage.Backend, viewer, ownerKey string, w *workouts.Workout) error {
	if w == nil || w.Track == "" {
		return nil
	}
	srcSpeed, srcHR, err := chartStores(src)
	if err != nil {
		return err
	}
	dstSpeed, dstHR, err := chartStores(dst)
	if err != nil {
		return err
	}
	ctx := context.Background()

	speedSamples, err := srcSpeed.ReadFederated(ctx, viewer, ownerKey, w.ID)
	if err != nil {
		return fmt.Errorf("read federated speed chart: %w", err)
	}
	if len(speedSamples) > 0 {
		if err := dstSpeed.WriteFederated(ctx, viewer, ownerKey, w.ID, speedSamples); err != nil {
			return fmt.Errorf("write federated speed chart: %w", err)
		}
	}

	hrSamples, err := srcHR.ReadFederated(ctx, viewer, ownerKey, w.ID)
	if err != nil {
		return fmt.Errorf("read federated heart rate chart: %w", err)
	}
	if len(hrSamples) > 0 {
		if err := dstHR.WriteFederated(ctx, viewer, ownerKey, w.ID, hrSamples); err != nil {
			return fmt.Errorf("write federated heart rate chart: %w", err)
		}
	}

	return nil
}
