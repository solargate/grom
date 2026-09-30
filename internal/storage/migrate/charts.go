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

type chartStoreSet struct {
	speed   workouts.SpeedChartStore
	hr      workouts.HeartRateChartStore
	cadence workouts.CadenceChartStore
}

func chartStores(backend storage.Backend) (chartStoreSet, error) {
	switch b := backend.(type) {
	case *file.Backend:
		blobs := b.Blobs()
		return chartStoreSet{
			speed:   workouts.NewBlobSpeedChartStore(blobs),
			hr:      workouts.NewBlobHeartRateChartStore(blobs),
			cadence: workouts.NewBlobCadenceChartStore(blobs),
		}, nil
	case *storebbolt.Backend:
		return chartStoreSet{
			speed:   storebbolt.NewSpeedChartStore(b.DB()),
			hr:      storebbolt.NewHeartRateChartStore(b.DB()),
			cadence: storebbolt.NewCadenceChartStore(b.DB()),
		}, nil
	default:
		return chartStoreSet{}, fmt.Errorf("unsupported backend type %T", backend)
	}
}

func copyLocalCharts(src, dst storage.Backend, nickname string, w *workouts.Workout) (speedCopied, hrCopied, cadenceCopied int, err error) {
	if w == nil || w.Track == "" {
		return 0, 0, 0, nil
	}
	srcStores, err := chartStores(src)
	if err != nil {
		return 0, 0, 0, err
	}
	dstStores, err := chartStores(dst)
	if err != nil {
		return 0, 0, 0, err
	}
	ctx := context.Background()
	dirName := keys.WorkoutDirName(w.StartDate, w.ID)

	speedSamples, err := srcStores.speed.ReadLocal(ctx, nickname, dirName)
	if err != nil {
		return 0, 0, 0, fmt.Errorf("read local speed chart: %w", err)
	}
	if len(speedSamples) > 0 {
		if err := dstStores.speed.WriteLocal(ctx, nickname, dirName, speedSamples); err != nil {
			return 0, 0, 0, fmt.Errorf("write local speed chart: %w", err)
		}
		speedCopied = 1
	}

	hrSamples, err := srcStores.hr.ReadLocal(ctx, nickname, dirName)
	if err != nil {
		return 0, 0, 0, fmt.Errorf("read local heart rate chart: %w", err)
	}
	if len(hrSamples) > 0 {
		if err := dstStores.hr.WriteLocal(ctx, nickname, dirName, hrSamples); err != nil {
			return 0, 0, 0, fmt.Errorf("write local heart rate chart: %w", err)
		}
		hrCopied = 1
	}

	cadenceSamples, err := srcStores.cadence.ReadLocal(ctx, nickname, dirName)
	if err != nil {
		return 0, 0, 0, fmt.Errorf("read local cadence chart: %w", err)
	}
	if len(cadenceSamples) > 0 {
		if err := dstStores.cadence.WriteLocal(ctx, nickname, dirName, cadenceSamples); err != nil {
			return 0, 0, 0, fmt.Errorf("write local cadence chart: %w", err)
		}
		cadenceCopied = 1
	}
	return speedCopied, hrCopied, cadenceCopied, nil
}

func copyFederatedCharts(src, dst storage.Backend, viewer, ownerKey string, w *workouts.Workout) (speedCopied, hrCopied, cadenceCopied int, err error) {
	if w == nil || w.Track == "" {
		return 0, 0, 0, nil
	}
	srcStores, err := chartStores(src)
	if err != nil {
		return 0, 0, 0, err
	}
	dstStores, err := chartStores(dst)
	if err != nil {
		return 0, 0, 0, err
	}
	ctx := context.Background()

	speedSamples, err := srcStores.speed.ReadFederated(ctx, viewer, ownerKey, w.ID)
	if err != nil {
		return 0, 0, 0, fmt.Errorf("read federated speed chart: %w", err)
	}
	if len(speedSamples) > 0 {
		if err := dstStores.speed.WriteFederated(ctx, viewer, ownerKey, w.ID, speedSamples); err != nil {
			return 0, 0, 0, fmt.Errorf("write federated speed chart: %w", err)
		}
		speedCopied = 1
	}

	hrSamples, err := srcStores.hr.ReadFederated(ctx, viewer, ownerKey, w.ID)
	if err != nil {
		return 0, 0, 0, fmt.Errorf("read federated heart rate chart: %w", err)
	}
	if len(hrSamples) > 0 {
		if err := dstStores.hr.WriteFederated(ctx, viewer, ownerKey, w.ID, hrSamples); err != nil {
			return 0, 0, 0, fmt.Errorf("write federated heart rate chart: %w", err)
		}
		hrCopied = 1
	}

	cadenceSamples, err := srcStores.cadence.ReadFederated(ctx, viewer, ownerKey, w.ID)
	if err != nil {
		return 0, 0, 0, fmt.Errorf("read federated cadence chart: %w", err)
	}
	if len(cadenceSamples) > 0 {
		if err := dstStores.cadence.WriteFederated(ctx, viewer, ownerKey, w.ID, cadenceSamples); err != nil {
			return 0, 0, 0, fmt.Errorf("write federated cadence chart: %w", err)
		}
		cadenceCopied = 1
	}

	return speedCopied, hrCopied, cadenceCopied, nil
}

func countCharts(backend storage.Backend, location string, result *Result) error {
	stores, err := chartStores(backend)
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
			speedSamples, err := stores.speed.ReadLocal(ctx, u.Nickname, dirName)
			if err != nil {
				return err
			}
			if len(speedSamples) > 0 {
				result.LocalSpeedCharts++
			}
			hrSamples, err := stores.hr.ReadLocal(ctx, u.Nickname, dirName)
			if err != nil {
				return err
			}
			if len(hrSamples) > 0 {
				result.LocalHeartRateCharts++
			}
			cadenceSamples, err := stores.cadence.ReadLocal(ctx, u.Nickname, dirName)
			if err != nil {
				return err
			}
			if len(cadenceSamples) > 0 {
				result.LocalCadenceCharts++
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
				speedSamples, err := stores.speed.ReadFederated(ctx, viewer, ownerKey, w.ID)
				if err != nil {
					return err
				}
				if len(speedSamples) > 0 {
					result.FedSpeedCharts++
				}
				hrSamples, err := stores.hr.ReadFederated(ctx, viewer, ownerKey, w.ID)
				if err != nil {
					return err
				}
				if len(hrSamples) > 0 {
					result.FedHeartRateCharts++
				}
				cadenceSamples, err := stores.cadence.ReadFederated(ctx, viewer, ownerKey, w.ID)
				if err != nil {
					return err
				}
				if len(cadenceSamples) > 0 {
					result.FedCadenceCharts++
				}
			}
		}
	}
	return nil
}
