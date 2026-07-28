package workouts

import (
	"context"
	"fmt"
	"os"

	"github.com/solargate/grom/internal/storage/blob"
	"github.com/solargate/grom/internal/storage/keys"
)

// BlobSpeedChartStore stores speed charts as blob files (file storage driver).
type BlobSpeedChartStore struct {
	blobs blob.Store
}

func NewBlobSpeedChartStore(blobs blob.Store) *BlobSpeedChartStore {
	return &BlobSpeedChartStore{blobs: blobs}
}

func (s *BlobSpeedChartStore) localKey(nickname, workoutDirName string) string {
	return keys.WorkoutSpeed(nickname, workoutDirName, keys.SpeedChartFileJSON)
}

func (s *BlobSpeedChartStore) federatedKey(viewer, ownerKey, workoutID string) string {
	return keys.FederatedInboxSpeed(viewer, ownerKey, workoutID, keys.SpeedChartFileJSON)
}

func (s *BlobSpeedChartStore) ReadLocal(ctx context.Context, nickname, workoutDirName string) ([]SpeedSample, error) {
	return s.read(ctx, s.localKey(nickname, workoutDirName))
}

func (s *BlobSpeedChartStore) WriteLocal(ctx context.Context, nickname, workoutDirName string, samples []SpeedSample) error {
	return s.write(ctx, s.localKey(nickname, workoutDirName), samples)
}

func (s *BlobSpeedChartStore) DeleteLocal(ctx context.Context, nickname, workoutDirName string) error {
	return s.delete(ctx, s.localKey(nickname, workoutDirName))
}

func (s *BlobSpeedChartStore) ReadFederated(ctx context.Context, viewer, ownerKey, workoutID string) ([]SpeedSample, error) {
	return s.read(ctx, s.federatedKey(viewer, ownerKey, workoutID))
}

func (s *BlobSpeedChartStore) WriteFederated(ctx context.Context, viewer, ownerKey, workoutID string, samples []SpeedSample) error {
	return s.write(ctx, s.federatedKey(viewer, ownerKey, workoutID), samples)
}

func (s *BlobSpeedChartStore) DeleteFederated(ctx context.Context, viewer, ownerKey, workoutID string) error {
	return s.delete(ctx, s.federatedKey(viewer, ownerKey, workoutID))
}

func (s *BlobSpeedChartStore) read(ctx context.Context, key string) ([]SpeedSample, error) {
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
	return UnmarshalSpeedChart(data)
}

func (s *BlobSpeedChartStore) write(ctx context.Context, key string, samples []SpeedSample) error {
	if s.blobs == nil {
		return fmt.Errorf("blob store is nil")
	}
	if len(samples) == 0 {
		return s.blobs.Delete(ctx, key)
	}
	data, err := MarshalSpeedChart(samples)
	if err != nil {
		return err
	}
	if len(data) == 0 {
		return s.blobs.Delete(ctx, key)
	}
	_, err = blob.PutBytes(ctx, s.blobs, key, data, blob.PutOptions{ContentType: "application/json"})
	return err
}

func (s *BlobSpeedChartStore) delete(ctx context.Context, key string) error {
	if s.blobs == nil {
		return nil
	}
	return s.blobs.Delete(ctx, key)
}

var _ SpeedChartStore = (*BlobSpeedChartStore)(nil)
