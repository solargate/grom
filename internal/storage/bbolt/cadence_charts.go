package bbolt

import (
	"context"
	"fmt"

	bolt "go.etcd.io/bbolt"

	"github.com/solargate/grom/internal/workouts"
)

// CadenceChartStore stores cadence charts in bbolt buckets as packed binary (bbolt storage driver).
type CadenceChartStore struct {
	db *bolt.DB
}

func NewCadenceChartStore(db *bolt.DB) *CadenceChartStore {
	return &CadenceChartStore{db: db}
}

func localCadenceChartKey(nickname, workoutDirName string) []byte {
	return []byte(workouts.LocalCadenceChartKey(nickname, workoutDirName))
}

func federatedCadenceChartKey(viewer, ownerKey, workoutID string) []byte {
	return []byte(workouts.FederatedCadenceChartKey(viewer, ownerKey, workoutID))
}

func (s *CadenceChartStore) ReadLocal(_ context.Context, nickname, workoutDirName string) ([]workouts.CadenceSample, error) {
	var raw []byte
	err := s.db.View(func(tx *bolt.Tx) error {
		raw = append([]byte(nil), tx.Bucket(bucketCadenceCharts).Get(localCadenceChartKey(nickname, workoutDirName))...)
		return nil
	})
	if err != nil {
		return nil, err
	}
	return workouts.UnmarshalCadenceChartBinary(raw)
}

func (s *CadenceChartStore) WriteLocal(_ context.Context, nickname, workoutDirName string, samples []workouts.CadenceSample) error {
	key := localCadenceChartKey(nickname, workoutDirName)
	return s.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(bucketCadenceCharts)
		if len(samples) == 0 {
			return b.Delete(key)
		}
		data, err := workouts.MarshalCadenceChartBinary(samples)
		if err != nil {
			return err
		}
		if len(data) == 0 {
			return b.Delete(key)
		}
		return b.Put(key, data)
	})
}

func (s *CadenceChartStore) DeleteLocal(_ context.Context, nickname, workoutDirName string) error {
	return s.db.Update(func(tx *bolt.Tx) error {
		return tx.Bucket(bucketCadenceCharts).Delete(localCadenceChartKey(nickname, workoutDirName))
	})
}

func (s *CadenceChartStore) ReadFederated(_ context.Context, viewer, ownerKey, workoutID string) ([]workouts.CadenceSample, error) {
	var raw []byte
	err := s.db.View(func(tx *bolt.Tx) error {
		raw = append([]byte(nil), tx.Bucket(bucketFedCadenceCharts).Get(federatedCadenceChartKey(viewer, ownerKey, workoutID))...)
		return nil
	})
	if err != nil {
		return nil, err
	}
	return workouts.UnmarshalCadenceChartBinary(raw)
}

func (s *CadenceChartStore) WriteFederated(_ context.Context, viewer, ownerKey, workoutID string, samples []workouts.CadenceSample) error {
	key := federatedCadenceChartKey(viewer, ownerKey, workoutID)
	return s.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(bucketFedCadenceCharts)
		if len(samples) == 0 {
			return b.Delete(key)
		}
		data, err := workouts.MarshalCadenceChartBinary(samples)
		if err != nil {
			return err
		}
		if len(data) == 0 {
			return b.Delete(key)
		}
		return b.Put(key, data)
	})
}

func (s *CadenceChartStore) DeleteFederated(_ context.Context, viewer, ownerKey, workoutID string) error {
	return s.db.Update(func(tx *bolt.Tx) error {
		return tx.Bucket(bucketFedCadenceCharts).Delete(federatedCadenceChartKey(viewer, ownerKey, workoutID))
	})
}

// DeleteLocalCadenceChartInTx removes a local cadence chart within an existing transaction.
func DeleteLocalCadenceChartInTx(tx *bolt.Tx, nickname, workoutDirName string) error {
	if tx.Bucket(bucketCadenceCharts) == nil {
		return nil
	}
	return tx.Bucket(bucketCadenceCharts).Delete(localCadenceChartKey(nickname, workoutDirName))
}

// DeleteFederatedCadenceChartInTx removes a federated cadence chart within an existing transaction.
func DeleteFederatedCadenceChartInTx(tx *bolt.Tx, viewer, ownerKey, workoutID string) error {
	if tx.Bucket(bucketFedCadenceCharts) == nil {
		return nil
	}
	return tx.Bucket(bucketFedCadenceCharts).Delete(federatedCadenceChartKey(viewer, ownerKey, workoutID))
}

// MigrateLocalCadenceChartInTx moves a local cadence chart payload when a workout dir name changes.
func MigrateLocalCadenceChartInTx(tx *bolt.Tx, nickname, oldDirName, newDirName string) error {
	if oldDirName == "" || newDirName == "" || oldDirName == newDirName {
		return nil
	}
	if err := moveBucketValueInTx(
		tx,
		bucketCadenceCharts,
		localCadenceChartKey(nickname, oldDirName),
		localCadenceChartKey(nickname, newDirName),
	); err != nil {
		return fmt.Errorf("migrate cadence chart: %w", err)
	}
	return nil
}

var _ workouts.CadenceChartStore = (*CadenceChartStore)(nil)
