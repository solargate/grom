package workouts

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/solargate/grom/internal/tracks"
)

// HeartRateChartMaxPoints is the maximum number of heart-rate samples stored for the detail chart.
const HeartRateChartMaxPoints = 500

// HeartRateChartStore persists pre-downsampled heart-rate chart payloads.
type HeartRateChartStore interface {
	ReadLocal(ctx context.Context, nickname, workoutDirName string) ([]HeartRateSample, error)
	WriteLocal(ctx context.Context, nickname, workoutDirName string, samples []HeartRateSample) error
	DeleteLocal(ctx context.Context, nickname, workoutDirName string) error

	ReadFederated(ctx context.Context, viewer, ownerKey, workoutID string) ([]HeartRateSample, error)
	WriteFederated(ctx context.Context, viewer, ownerKey, workoutID string, samples []HeartRateSample) error
	DeleteFederated(ctx context.Context, viewer, ownerKey, workoutID string) error
}

type heartRateChartSampleJSON struct {
	T            string   `json:"t"`
	HeartRateBPM float64  `json:"heart_rate_bpm"`
	DistanceM    *float64 `json:"distance_m,omitempty"`
}

type heartRateChartJSON struct {
	Samples []heartRateChartSampleJSON `json:"samples"`
}

// LocalHeartRateChartKey returns the storage key for a local workout heart-rate chart.
func LocalHeartRateChartKey(nickname, workoutDirName string) string {
	return nickname + "/" + workoutDirName
}

// FederatedHeartRateChartKey returns the storage key for a federated inbox heart-rate chart.
func FederatedHeartRateChartKey(viewer, ownerKey, workoutID string) string {
	return viewer + "/" + ownerKey + "/" + workoutID
}

// HeartRateSamplesFromTrack converts parsed track HR points into workout samples.
func HeartRateSamplesFromTrack(series []tracks.HeartRatePoint) []HeartRateSample {
	if len(series) == 0 {
		return nil
	}
	out := make([]HeartRateSample, len(series))
	for i := range series {
		out[i] = HeartRateSample{
			Time: series[i].Time.UTC(),
			BPM:  series[i].BPM,
		}
		if series[i].HasDistance {
			d := series[i].DistanceM
			out[i].DistanceM = &d
		}
	}
	return out
}

// HeartRateSamplesFromParsed returns workout HR samples from parsed track data.
func HeartRateSamplesFromParsed(parsed *tracks.Data) []HeartRateSample {
	if parsed == nil {
		return nil
	}
	return HeartRateSamplesFromTrack(parsed.HeartRateSeries)
}

// BuildHeartRateChartSamples builds the stored chart series (downsampled).
// Zero-BPM inclusion follows tracks.HeartRateChartZeroPolicy.
func BuildHeartRateChartSamples(parsed *tracks.Data) []HeartRateSample {
	full := HeartRateSamplesFromParsed(parsed)
	if len(full) == 0 {
		return nil
	}
	filtered := make([]HeartRateSample, 0, len(full))
	for _, s := range full {
		if !tracks.AcceptHeartRateBPM(s.BPM, tracks.HeartRateChartZeroPolicy) {
			continue
		}
		filtered = append(filtered, s)
	}
	return DownsampleHeartRateSamples(filtered, HeartRateChartMaxPoints)
}

// MarshalHeartRateChart encodes chart samples as compact JSON.
func MarshalHeartRateChart(samples []HeartRateSample) ([]byte, error) {
	if len(samples) == 0 {
		return nil, nil
	}
	payload := heartRateChartJSON{Samples: make([]heartRateChartSampleJSON, 0, len(samples))}
	for _, s := range samples {
		payload.Samples = append(payload.Samples, heartRateChartSampleJSON{
			T:            s.Time.UTC().Format(time.RFC3339),
			HeartRateBPM: s.BPM,
			DistanceM:    s.DistanceM,
		})
	}
	data, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("marshal heart rate chart: %w", err)
	}
	return data, nil
}

// UnmarshalHeartRateChart decodes chart JSON into samples.
func UnmarshalHeartRateChart(data []byte) ([]HeartRateSample, error) {
	if len(data) == 0 {
		return nil, nil
	}
	var payload heartRateChartJSON
	if err := json.Unmarshal(data, &payload); err != nil {
		return nil, fmt.Errorf("parse heart rate chart: %w", err)
	}
	if len(payload.Samples) == 0 {
		return nil, nil
	}
	out := make([]HeartRateSample, 0, len(payload.Samples))
	for _, s := range payload.Samples {
		ts, err := time.Parse(time.RFC3339, s.T)
		if err != nil {
			return nil, fmt.Errorf("parse heart rate chart timestamp %q: %w", s.T, err)
		}
		out = append(out, HeartRateSample{
			Time:      ts.UTC(),
			BPM:       s.HeartRateBPM,
			DistanceM: s.DistanceM,
		})
	}
	return out, nil
}
