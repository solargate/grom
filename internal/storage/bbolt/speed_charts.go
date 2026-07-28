package bbolt

import (
	"context"
	"fmt"

	bolt "go.etcd.io/bbolt"

	"github.com/solargate/grom/internal/workouts"
)

// SpeedChartStore stores speed charts in bbolt buckets (bbolt storage driver).
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
	return workouts.UnmarshalSpeedChart(raw)
}

func (s *SpeedChartStore) WriteLocal(_ context.Context, nickname, workoutDirName string, samples []workouts.SpeedSample) error {
	key := localSpeedChartKey(nickname, workoutDirName)
	return s.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(bucketSpeedCharts)
		if len(samples) == 0 {
			return b.Delete(key)
		}
		data, err := workouts.MarshalSpeedChart(samples)
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
	return workouts.UnmarshalSpeedChart(raw)
}

func (s *SpeedChartStore) WriteFederated(_ context.Context, viewer, ownerKey, workoutID string, samples []workouts.SpeedSample) error {
	key := federatedSpeedChartKey(viewer, ownerKey, workoutID)
	return s.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(bucketFedSpeedCharts)
		if len(samples) == 0 {
			return b.Delete(key)
		}
		data, err := workouts.MarshalSpeedChart(samples)
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

var _ workouts.SpeedChartStore = (*SpeedChartStore)(nil)
