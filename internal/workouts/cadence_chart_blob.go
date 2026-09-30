package workouts

import (
	"context"
	"fmt"
	"os"

	"github.com/solargate/grom/internal/storage/blob"
	"github.com/solargate/grom/internal/storage/keys"
)

// BlobCadenceChartStore stores cadence charts as blob files (file storage driver).
type BlobCadenceChartStore struct {
	blobs blob.Store
}

func NewBlobCadenceChartStore(blobs blob.Store) *BlobCadenceChartStore {
	return &BlobCadenceChartStore{blobs: blobs}
}

func (s *BlobCadenceChartStore) localKey(nickname, workoutDirName string) string {
	return keys.WorkoutSpeed(nickname, workoutDirName, keys.CadenceChartFileJSON)
}

func (s *BlobCadenceChartStore) federatedKey(viewer, ownerKey, workoutID string) string {
	return keys.FederatedInboxSpeed(viewer, ownerKey, workoutID, keys.CadenceChartFileJSON)
}

func (s *BlobCadenceChartStore) ReadLocal(ctx context.Context, nickname, workoutDirName string) ([]CadenceSample, error) {
	return s.read(ctx, s.localKey(nickname, workoutDirName))
}

func (s *BlobCadenceChartStore) WriteLocal(ctx context.Context, nickname, workoutDirName string, samples []CadenceSample) error {
	return s.write(ctx, s.localKey(nickname, workoutDirName), samples)
}

func (s *BlobCadenceChartStore) DeleteLocal(ctx context.Context, nickname, workoutDirName string) error {
	return s.delete(ctx, s.localKey(nickname, workoutDirName))
}

func (s *BlobCadenceChartStore) ReadFederated(ctx context.Context, viewer, ownerKey, workoutID string) ([]CadenceSample, error) {
	return s.read(ctx, s.federatedKey(viewer, ownerKey, workoutID))
}

func (s *BlobCadenceChartStore) WriteFederated(ctx context.Context, viewer, ownerKey, workoutID string, samples []CadenceSample) error {
	return s.write(ctx, s.federatedKey(viewer, ownerKey, workoutID), samples)
}

func (s *BlobCadenceChartStore) DeleteFederated(ctx context.Context, viewer, ownerKey, workoutID string) error {
	return s.delete(ctx, s.federatedKey(viewer, ownerKey, workoutID))
}

func (s *BlobCadenceChartStore) read(ctx context.Context, key string) ([]CadenceSample, error) {
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
	return UnmarshalCadenceChart(data)
}

func (s *BlobCadenceChartStore) write(ctx context.Context, key string, samples []CadenceSample) error {
	if s.blobs == nil {
		return fmt.Errorf("blob store is nil")
	}
	if len(samples) == 0 {
		return s.blobs.Delete(ctx, key)
	}
	data, err := MarshalCadenceChart(samples)
	if err != nil {
		return err
	}
	if len(data) == 0 {
		return s.blobs.Delete(ctx, key)
	}
	_, err = blob.PutBytes(ctx, s.blobs, key, data, blob.PutOptions{ContentType: "application/json"})
	return err
}

func (s *BlobCadenceChartStore) delete(ctx context.Context, key string) error {
	if s.blobs == nil {
		return nil
	}
	return s.blobs.Delete(ctx, key)
}

var _ CadenceChartStore = (*BlobCadenceChartStore)(nil)
