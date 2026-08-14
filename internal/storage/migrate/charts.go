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

func copyLocalCharts(src, dst storage.Backend, nickname string, w *workouts.Workout) (speedCopied, hrCopied int, err error) {
	if w == nil || w.Track == "" {
		return 0, 0, nil
	}
	srcSpeed, srcHR, err := chartStores(src)
	if err != nil {
		return 0, 0, err
	}
	dstSpeed, dstHR, err := chartStores(dst)
	if err != nil {
		return 0, 0, err
	}
	ctx := context.Background()
	dirName := keys.WorkoutDirName(w.StartDate, w.ID)

	speedSamples, err := srcSpeed.ReadLocal(ctx, nickname, dirName)
	if err != nil {
		return 0, 0, fmt.Errorf("read local speed chart: %w", err)
	}
	if len(speedSamples) > 0 {
		if err := dstSpeed.WriteLocal(ctx, nickname, dirName, speedSamples); err != nil {
			return 0, 0, fmt.Errorf("write local speed chart: %w", err)
		}
		speedCopied = 1
	}

	hrSamples, err := srcHR.ReadLocal(ctx, nickname, dirName)
	if err != nil {
		return 0, 0, fmt.Errorf("read local heart rate chart: %w", err)
	}
	if len(hrSamples) > 0 {
		if err := dstHR.WriteLocal(ctx, nickname, dirName, hrSamples); err != nil {
			return 0, 0, fmt.Errorf("write local heart rate chart: %w", err)
		}
		hrCopied = 1
	}
	return speedCopied, hrCopied, nil
}

func copyFederatedCharts(src, dst storage.Backend, viewer, ownerKey string, w *workouts.Workout) (speedCopied, hrCopied int, err error) {
	if w == nil || w.Track == "" {
		return 0, 0, nil
	}
	srcSpeed, srcHR, err := chartStores(src)
	if err != nil {
		return 0, 0, err
	}
	dstSpeed, dstHR, err := chartStores(dst)
	if err != nil {
		return 0, 0, err
	}
	ctx := context.Background()

	speedSamples, err := srcSpeed.ReadFederated(ctx, viewer, ownerKey, w.ID)
	if err != nil {
		return 0, 0, fmt.Errorf("read federated speed chart: %w", err)
	}
	if len(speedSamples) > 0 {
		if err := dstSpeed.WriteFederated(ctx, viewer, ownerKey, w.ID, speedSamples); err != nil {
			return 0, 0, fmt.Errorf("write federated speed chart: %w", err)
		}
		speedCopied = 1
	}

	hrSamples, err := srcHR.ReadFederated(ctx, viewer, ownerKey, w.ID)
	if err != nil {
		return 0, 0, fmt.Errorf("read federated heart rate chart: %w", err)
	}
	if len(hrSamples) > 0 {
		if err := dstHR.WriteFederated(ctx, viewer, ownerKey, w.ID, hrSamples); err != nil {
			return 0, 0, fmt.Errorf("write federated heart rate chart: %w", err)
		}
		hrCopied = 1
	}

	return speedCopied, hrCopied, nil
}

func countCharts(backend storage.Backend, location string, result *Result) error {
	speedStore, hrStore, err := chartStores(backend)
	if err != nil {
		return err
	}
	ctx := context.Background()

	usersList, err := backend.Users().ListAll()
	if err != nil {
		return err
	}
	for _, u := range usersList {
		ws, err := backend.Workouts().List(u.Nickname)
		if err != nil {
			return err
		}
		for i := range ws {
			w := ws[i]
			if w.Track == "" {
				continue
			}
			dirName := keys.WorkoutDirName(w.StartDate, w.ID)
			speedSamples, err := speedStore.ReadLocal(ctx, u.Nickname, dirName)
			if err != nil {
				return err
			}
			if len(speedSamples) > 0 {
				result.LocalSpeedCharts++
			}
			hrSamples, err := hrStore.ReadLocal(ctx, u.Nickname, dirName)
			if err != nil {
				return err
			}
			if len(hrSamples) > 0 {
				result.LocalHeartRateCharts++
			}
		}
	}

	_, inbox, err := loadFederationInbox(backend, location)
	if err != nil {
		return err
	}
	for viewer, byOwner := range inbox {
		for ownerKey, list := range byOwner {
			for i := range list {
				w := list[i]
				if w.Track == "" {
					continue
				}
				speedSamples, err := speedStore.ReadFederated(ctx, viewer, ownerKey, w.ID)
				if err != nil {
					return err
				}
				if len(speedSamples) > 0 {
					result.FedSpeedCharts++
				}
				hrSamples, err := hrStore.ReadFederated(ctx, viewer, ownerKey, w.ID)
				if err != nil {
					return err
				}
				if len(hrSamples) > 0 {
					result.FedHeartRateCharts++
				}
			}
		}
	}
	return nil
}
