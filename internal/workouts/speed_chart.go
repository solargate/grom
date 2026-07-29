package workouts

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/solargate/grom/internal/tracks"
)

// SpeedChartMaxPoints is the maximum number of speed samples stored for the detail chart.
const SpeedChartMaxPoints = 500

// SpeedChartStore persists pre-downsampled speed chart payloads.
type SpeedChartStore interface {
	ReadLocal(ctx context.Context, nickname, workoutDirName string) ([]SpeedSample, error)
	WriteLocal(ctx context.Context, nickname, workoutDirName string, samples []SpeedSample) error
	DeleteLocal(ctx context.Context, nickname, workoutDirName string) error

	ReadFederated(ctx context.Context, viewer, ownerKey, workoutID string) ([]SpeedSample, error)
	WriteFederated(ctx context.Context, viewer, ownerKey, workoutID string, samples []SpeedSample) error
	DeleteFederated(ctx context.Context, viewer, ownerKey, workoutID string) error
}

type speedChartSampleJSON struct {
	T         string  `json:"t"`
	SpeedKmh  float64 `json:"speed_kmh"`
	DistanceM float64 `json:"distance_m"`
}

type speedChartJSON struct {
	Samples []speedChartSampleJSON `json:"samples"`
}

// LocalSpeedChartKey returns the storage key for a local workout speed chart.
func LocalSpeedChartKey(nickname, workoutDirName string) string {
	return nickname + "/" + workoutDirName
}

// FederatedSpeedChartKey returns the storage key for a federated inbox speed chart.
func FederatedSpeedChartKey(viewer, ownerKey, workoutID string) string {
	return viewer + "/" + ownerKey + "/" + workoutID
}

// SpeedSamplesFromTrack converts parsed track speed points into workout samples.
func SpeedSamplesFromTrack(series []tracks.SpeedPoint) []SpeedSample {
	if len(series) == 0 {
		return nil
	}
	out := make([]SpeedSample, len(series))
	for i := range series {
		out[i] = SpeedSample{
			Time:      series[i].Time.UTC(),
			SpeedKmh:  series[i].Kmh,
			DistanceM: series[i].DistanceM,
		}
	}
	return out
}

// SpeedSamplesFromParsed returns workout speed samples from parsed track data.
func SpeedSamplesFromParsed(parsed *tracks.Data) []SpeedSample {
	if parsed == nil {
		return nil
	}
	return SpeedSamplesFromTrack(parsed.SpeedSeries)
}

// BuildSpeedChartSamples builds the stored chart series (downsampled).
// Zero-speed inclusion follows tracks.SpeedChartZeroPolicy.
func BuildSpeedChartSamples(parsed *tracks.Data) []SpeedSample {
	full := SpeedSamplesFromParsed(parsed)
	if len(full) == 0 {
		return nil
	}
	filtered := make([]SpeedSample, 0, len(full))
	for _, s := range full {
		if !tracks.AcceptSpeedKmh(s.SpeedKmh, tracks.SpeedChartZeroPolicy) {
			continue
		}
		filtered = append(filtered, s)
	}
	return DownsampleSpeedSamples(filtered, SpeedChartMaxPoints)
}

// MarshalSpeedChart encodes chart samples as compact JSON.
func MarshalSpeedChart(samples []SpeedSample) ([]byte, error) {
	if len(samples) == 0 {
		return nil, nil
	}
	payload := speedChartJSON{Samples: make([]speedChartSampleJSON, 0, len(samples))}
	for _, s := range samples {
		payload.Samples = append(payload.Samples, speedChartSampleJSON{
			T:         s.Time.UTC().Format(time.RFC3339),
			SpeedKmh:  s.SpeedKmh,
			DistanceM: s.DistanceM,
		})
	}
	data, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("marshal speed chart: %w", err)
	}
	return data, nil
}

// UnmarshalSpeedChart decodes chart JSON into samples.
func UnmarshalSpeedChart(data []byte) ([]SpeedSample, error) {
	if len(data) == 0 {
		return nil, nil
	}
	var payload speedChartJSON
	if err := json.Unmarshal(data, &payload); err != nil {
		return nil, fmt.Errorf("parse speed chart: %w", err)
	}
	if len(payload.Samples) == 0 {
		return nil, nil
	}
	out := make([]SpeedSample, 0, len(payload.Samples))
	for _, s := range payload.Samples {
		ts, err := time.Parse(time.RFC3339, s.T)
		if err != nil {
			return nil, fmt.Errorf("parse speed chart timestamp %q: %w", s.T, err)
		}
		out = append(out, SpeedSample{
			Time:      ts.UTC(),
			SpeedKmh:  s.SpeedKmh,
			DistanceM: s.DistanceM,
		})
	}
	return out, nil
}
