package bbolt

import (
	"context"
	"fmt"

	bolt "go.etcd.io/bbolt"

	"github.com/solargate/grom/internal/workouts"
)

// HeartRateChartStore stores heart-rate charts in bbolt buckets (bbolt storage driver).
type HeartRateChartStore struct {
	db *bolt.DB
}

func NewHeartRateChartStore(db *bolt.DB) *HeartRateChartStore {
	return &HeartRateChartStore{db: db}
}

func localHeartRateChartKey(nickname, workoutDirName string) []byte {
	return []byte(workouts.LocalHeartRateChartKey(nickname, workoutDirName))
}

func federatedHeartRateChartKey(viewer, ownerKey, workoutID string) []byte {
	return []byte(workouts.FederatedHeartRateChartKey(viewer, ownerKey, workoutID))
}

func (s *HeartRateChartStore) ReadLocal(_ context.Context, nickname, workoutDirName string) ([]workouts.HeartRateSample, error) {
	var raw []byte
	err := s.db.View(func(tx *bolt.Tx) error {
		raw = append([]byte(nil), tx.Bucket(bucketHeartRateCharts).Get(localHeartRateChartKey(nickname, workoutDirName))...)
		return nil
	})
	if err != nil {
		return nil, err
	}
	return workouts.UnmarshalHeartRateChart(raw)
}

func (s *HeartRateChartStore) WriteLocal(_ context.Context, nickname, workoutDirName string, samples []workouts.HeartRateSample) error {
	key := localHeartRateChartKey(nickname, workoutDirName)
	return s.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(bucketHeartRateCharts)
		if len(samples) == 0 {
			return b.Delete(key)
		}
		data, err := workouts.MarshalHeartRateChart(samples)
		if err != nil {
			return err
		}
		if len(data) == 0 {
			return b.Delete(key)
		}
		return b.Put(key, data)
	})
}

func (s *HeartRateChartStore) DeleteLocal(_ context.Context, nickname, workoutDirName string) error {
	return s.db.Update(func(tx *bolt.Tx) error {
		return tx.Bucket(bucketHeartRateCharts).Delete(localHeartRateChartKey(nickname, workoutDirName))
	})
}

func (s *HeartRateChartStore) ReadFederated(_ context.Context, viewer, ownerKey, workoutID string) ([]workouts.HeartRateSample, error) {
	var raw []byte
	err := s.db.View(func(tx *bolt.Tx) error {
		raw = append([]byte(nil), tx.Bucket(bucketFedHeartRateCharts).Get(federatedHeartRateChartKey(viewer, ownerKey, workoutID))...)
		return nil
	})
	if err != nil {
		return nil, err
	}
	return workouts.UnmarshalHeartRateChart(raw)
}

func (s *HeartRateChartStore) WriteFederated(_ context.Context, viewer, ownerKey, workoutID string, samples []workouts.HeartRateSample) error {
	key := federatedHeartRateChartKey(viewer, ownerKey, workoutID)
	return s.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(bucketFedHeartRateCharts)
		if len(samples) == 0 {
			return b.Delete(key)
		}
		data, err := workouts.MarshalHeartRateChart(samples)
		if err != nil {
			return err
		}
		if len(data) == 0 {
			return b.Delete(key)
		}
		return b.Put(key, data)
	})
}

func (s *HeartRateChartStore) DeleteFederated(_ context.Context, viewer, ownerKey, workoutID string) error {
	return s.db.Update(func(tx *bolt.Tx) error {
		return tx.Bucket(bucketFedHeartRateCharts).Delete(federatedHeartRateChartKey(viewer, ownerKey, workoutID))
	})
}

// DeleteLocalHeartRateChartInTx removes a local heart-rate chart within an existing transaction.
func DeleteLocalHeartRateChartInTx(tx *bolt.Tx, nickname, workoutDirName string) error {
	if tx.Bucket(bucketHeartRateCharts) == nil {
		return nil
	}
	return tx.Bucket(bucketHeartRateCharts).Delete(localHeartRateChartKey(nickname, workoutDirName))
}

// DeleteFederatedHeartRateChartInTx removes a federated heart-rate chart within an existing transaction.
func DeleteFederatedHeartRateChartInTx(tx *bolt.Tx, viewer, ownerKey, workoutID string) error {
	if tx.Bucket(bucketFedHeartRateCharts) == nil {
		return nil
	}
	return tx.Bucket(bucketFedHeartRateCharts).Delete(federatedHeartRateChartKey(viewer, ownerKey, workoutID))
}

// MigrateLocalHeartRateChartInTx moves a local heart-rate chart payload when a workout dir name changes.
func MigrateLocalHeartRateChartInTx(tx *bolt.Tx, nickname, oldDirName, newDirName string) error {
	if oldDirName == "" || newDirName == "" || oldDirName == newDirName {
		return nil
	}
	if err := moveBucketValueInTx(
		tx,
		bucketHeartRateCharts,
		localHeartRateChartKey(nickname, oldDirName),
		localHeartRateChartKey(nickname, newDirName),
	); err != nil {
		return fmt.Errorf("migrate heart rate chart: %w", err)
	}
	return nil
}

var _ workouts.HeartRateChartStore = (*HeartRateChartStore)(nil)
