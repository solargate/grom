package workouts

import (
	"context"
	"fmt"
	"os"

	"github.com/solargate/grom/internal/storage/blob"
	"github.com/solargate/grom/internal/storage/keys"
)

// BlobHeartRateChartStore stores heart-rate charts as blob files (file storage driver).
type BlobHeartRateChartStore struct {
	blobs blob.Store
}

func NewBlobHeartRateChartStore(blobs blob.Store) *BlobHeartRateChartStore {
	return &BlobHeartRateChartStore{blobs: blobs}
}

func (s *BlobHeartRateChartStore) localKey(nickname, workoutDirName string) string {
	return keys.WorkoutSpeed(nickname, workoutDirName, keys.HeartRateChartFileJSON)
}

func (s *BlobHeartRateChartStore) federatedKey(viewer, ownerKey, workoutID string) string {
	return keys.FederatedInboxSpeed(viewer, ownerKey, workoutID, keys.HeartRateChartFileJSON)
}

func (s *BlobHeartRateChartStore) ReadLocal(ctx context.Context, nickname, workoutDirName string) ([]HeartRateSample, error) {
	return s.read(ctx, s.localKey(nickname, workoutDirName))
}

func (s *BlobHeartRateChartStore) WriteLocal(ctx context.Context, nickname, workoutDirName string, samples []HeartRateSample) error {
	return s.write(ctx, s.localKey(nickname, workoutDirName), samples)
}

func (s *BlobHeartRateChartStore) DeleteLocal(ctx context.Context, nickname, workoutDirName string) error {
	return s.delete(ctx, s.localKey(nickname, workoutDirName))
}

func (s *BlobHeartRateChartStore) ReadFederated(ctx context.Context, viewer, ownerKey, workoutID string) ([]HeartRateSample, error) {
	return s.read(ctx, s.federatedKey(viewer, ownerKey, workoutID))
}

func (s *BlobHeartRateChartStore) WriteFederated(ctx context.Context, viewer, ownerKey, workoutID string, samples []HeartRateSample) error {
	return s.write(ctx, s.federatedKey(viewer, ownerKey, workoutID), samples)
}

func (s *BlobHeartRateChartStore) DeleteFederated(ctx context.Context, viewer, ownerKey, workoutID string) error {
	return s.delete(ctx, s.federatedKey(viewer, ownerKey, workoutID))
}

func (s *BlobHeartRateChartStore) read(ctx context.Context, key string) ([]HeartRateSample, error) {
	if s.blobs == nil {
		return nil, nil
	}
	exists, err := s.blobs.Exists(ctx, key)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, nil
	}
	data, err := blob.ReadAll(ctx, s.blobs, key)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	return UnmarshalHeartRateChart(data)
}

func (s *BlobHeartRateChartStore) write(ctx context.Context, key string, samples []HeartRateSample) error {
	if s.blobs == nil {
		return fmt.Errorf("blob store is nil")
	}
	if len(samples) == 0 {
		return s.blobs.Delete(ctx, key)
	}
	data, err := MarshalHeartRateChart(samples)
	if err != nil {
		return err
	}
	if len(data) == 0 {
		return s.blobs.Delete(ctx, key)
	}
	_, err = blob.PutBytes(ctx, s.blobs, key, data, blob.PutOptions{ContentType: "application/json"})
	return err
}

func (s *BlobHeartRateChartStore) delete(ctx context.Context, key string) error {
	if s.blobs == nil {
		return nil
	}
	return s.blobs.Delete(ctx, key)
}

var _ HeartRateChartStore = (*BlobHeartRateChartStore)(nil)
