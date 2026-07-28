package workouts

import (
	"context"
	"fmt"
	"os"

	"github.com/solargate/grom/internal/storage/blob"
	"github.com/solargate/grom/internal/tracks"
)

// WriteSpeedBlob persists a speed sidecar. Empty samples deletes any existing file.
func WriteSpeedBlob(ctx context.Context, store blob.Store, key string, format SpeedSidecarFormat, samples []SpeedSample) error {
	if store == nil {
		return fmt.Errorf("blob store is nil")
	}
	if len(samples) == 0 {
		return store.Delete(ctx, key)
	}
	data, err := MarshalSpeedSidecar(format, samples)
	if err != nil {
		return err
	}
	if len(data) == 0 {
		return store.Delete(ctx, key)
	}
	_, err = blob.PutBytes(ctx, store, key, data, blob.PutOptions{})
	return err
}

// ReadSpeedBlob loads a speed sidecar. Missing file returns nil samples.
func ReadSpeedBlob(ctx context.Context, store blob.Store, key string, format SpeedSidecarFormat) ([]SpeedSample, error) {
	if store == nil {
		return nil, nil
	}
	exists, err := store.Exists(ctx, key)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, nil
	}
	data, err := blob.ReadAll(ctx, store, key)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	return UnmarshalSpeedSidecar(format, data)
}

// DeleteSpeedBlob removes a speed sidecar if present.
func DeleteSpeedBlob(ctx context.Context, store blob.Store, key string) error {
	if store == nil {
		return nil
	}
	return store.Delete(ctx, key)
}

// SpeedSamplesFromParsed returns workout speed samples from parsed track data.
func SpeedSamplesFromParsed(parsed *tracks.Data) []SpeedSample {
	if parsed == nil {
		return nil
	}
	return SpeedSamplesFromTrack(parsed.SpeedSeries)
}
