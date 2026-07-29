package bbolt

import (
	"context"
	"fmt"

	bolt "go.etcd.io/bbolt"

	"github.com/solargate/grom/internal/workouts"
)

// SpeedChartStore stores speed charts in bbolt buckets as packed binary (bbolt storage driver).
type SpeedChartStore struct {
	db *bolt.DB
}

func NewSpeedChartStore(db *bolt.DB) *SpeedChartStore {
	return &SpeedChartStore{db: db}
}

func localSpeedChartKey(nickname, workoutDirName string) []byte {
	return []byte(workouts.LocalSpeedChartKey(nickname, workoutDirName))
}

func federatedSpeedChartKey(viewer, ownerKey, workoutID string) []byte {
	return []byte(workouts.FederatedSpeedChartKey(viewer, ownerKey, workoutID))
}

func (s *SpeedChartStore) ReadLocal(_ context.Context, nickname, workoutDirName string) ([]workouts.SpeedSample, error) {
	var raw []byte
	err := s.db.View(func(tx *bolt.Tx) error {
		raw = append([]byte(nil), tx.Bucket(bucketSpeedCharts).Get(localSpeedChartKey(nickname, workoutDirName))...)
		return nil
	})
	if err != nil {
		return nil, err
	}
	return workouts.UnmarshalSpeedChartBinary(raw)
}

func (s *SpeedChartStore) WriteLocal(_ context.Context, nickname, workoutDirName string, samples []workouts.SpeedSample) error {
	key := localSpeedChartKey(nickname, workoutDirName)
	return s.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(bucketSpeedCharts)
		if len(samples) == 0 {
			return b.Delete(key)
		}
		data, err := workouts.MarshalSpeedChartBinary(samples)
		if err != nil {
			return err
		}
		if len(data) == 0 {
			return b.Delete(key)
		}
		return b.Put(key, data)
	})
}

func (s *SpeedChartStore) DeleteLocal(_ context.Context, nickname, workoutDirName string) error {
	return s.db.Update(func(tx *bolt.Tx) error {
		return tx.Bucket(bucketSpeedCharts).Delete(localSpeedChartKey(nickname, workoutDirName))
	})
}

func (s *SpeedChartStore) ReadFederated(_ context.Context, viewer, ownerKey, workoutID string) ([]workouts.SpeedSample, error) {
	var raw []byte
	err := s.db.View(func(tx *bolt.Tx) error {
		raw = append([]byte(nil), tx.Bucket(bucketFedSpeedCharts).Get(federatedSpeedChartKey(viewer, ownerKey, workoutID))...)
		return nil
	})
	if err != nil {
		return nil, err
	}
	return workouts.UnmarshalSpeedChartBinary(raw)
}

func (s *SpeedChartStore) WriteFederated(_ context.Context, viewer, ownerKey, workoutID string, samples []workouts.SpeedSample) error {
	key := federatedSpeedChartKey(viewer, ownerKey, workoutID)
	return s.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(bucketFedSpeedCharts)
		if len(samples) == 0 {
			return b.Delete(key)
		}
		data, err := workouts.MarshalSpeedChartBinary(samples)
		if err != nil {
			return err
		}
		if len(data) == 0 {
			return b.Delete(key)
		}
		return b.Put(key, data)
	})
}

func (s *SpeedChartStore) DeleteFederated(_ context.Context, viewer, ownerKey, workoutID string) error {
	return s.db.Update(func(tx *bolt.Tx) error {
		return tx.Bucket(bucketFedSpeedCharts).Delete(federatedSpeedChartKey(viewer, ownerKey, workoutID))
	})
}

// DeleteLocalInTx removes a local speed chart within an existing transaction.
func DeleteLocalSpeedChartInTx(tx *bolt.Tx, nickname, workoutDirName string) error {
	if tx == nil {
		return fmt.Errorf("transaction is nil")
	}
	return tx.Bucket(bucketSpeedCharts).Delete(localSpeedChartKey(nickname, workoutDirName))
}

// DeleteFederatedInTx removes a federated speed chart within an existing transaction.
func DeleteFederatedSpeedChartInTx(tx *bolt.Tx, viewer, ownerKey, workoutID string) error {
	if tx == nil {
		return fmt.Errorf("transaction is nil")
	}
	return tx.Bucket(bucketFedSpeedCharts).Delete(federatedSpeedChartKey(viewer, ownerKey, workoutID))
}

// MigrateLocalSpeedChartInTx moves a local speed chart payload when a workout dir name changes.
func MigrateLocalSpeedChartInTx(tx *bolt.Tx, nickname, oldDirName, newDirName string) error {
	if oldDirName == "" || newDirName == "" || oldDirName == newDirName {
		return nil
	}
	if err := moveBucketValueInTx(
		tx,
		bucketSpeedCharts,
		localSpeedChartKey(nickname, oldDirName),
		localSpeedChartKey(nickname, newDirName),
	); err != nil {
		return fmt.Errorf("migrate speed chart: %w", err)
	}
	return nil
}

// moveBucketValueInTx copies a value to a new key and deletes the old key (no-op if missing).
func moveBucketValueInTx(tx *bolt.Tx, bucketName, oldKey, newKey []byte) error {
	if tx == nil {
		return fmt.Errorf("transaction is nil")
	}
	if string(oldKey) == string(newKey) {
		return nil
	}
	b := tx.Bucket(bucketName)
	if b == nil {
		return nil
	}
	raw := b.Get(oldKey)
	if raw == nil {
		return nil
	}
	copied := append([]byte(nil), raw...)
	if err := b.Put(newKey, copied); err != nil {
		return err
	}
	return b.Delete(oldKey)
}

var _ workouts.SpeedChartStore = (*SpeedChartStore)(nil)
