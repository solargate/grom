package workouts

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/solargate/grom/internal/tracks"
)

// CadenceChartMaxPoints is the maximum number of cadence samples stored for the detail chart.
const CadenceChartMaxPoints = 500

// CadenceChartStore persists pre-downsampled cadence chart payloads.
type CadenceChartStore interface {
	ReadLocal(ctx context.Context, nickname, workoutDirName string) ([]CadenceSample, error)
	WriteLocal(ctx context.Context, nickname, workoutDirName string, samples []CadenceSample) error
	DeleteLocal(ctx context.Context, nickname, workoutDirName string) error

	ReadFederated(ctx context.Context, viewer, ownerKey, workoutID string) ([]CadenceSample, error)
	WriteFederated(ctx context.Context, viewer, ownerKey, workoutID string, samples []CadenceSample) error
	DeleteFederated(ctx context.Context, viewer, ownerKey, workoutID string) error
}

type cadenceChartSampleJSON struct {
	T         string   `json:"t"`
	Cadence   float64  `json:"cadence"`
	DistanceM *float64 `json:"distance_m,omitempty"`
}

type cadenceChartJSON struct {
	Samples []cadenceChartSampleJSON `json:"samples"`
}

// LocalCadenceChartKey returns the storage key for a local workout cadence chart.
func LocalCadenceChartKey(nickname, workoutDirName string) string {
	return nickname + "/" + workoutDirName
}

// FederatedCadenceChartKey returns the storage key for a federated inbox cadence chart.
func FederatedCadenceChartKey(viewer, ownerKey, workoutID string) string {
	return viewer + "/" + ownerKey + "/" + workoutID
}

// CadenceSamplesFromTrack converts parsed track cadence points into workout samples.
func CadenceSamplesFromTrack(series []tracks.CadencePoint) []CadenceSample {
	if len(series) == 0 {
		return nil
	}
	out := make([]CadenceSample, len(series))
	for i := range series {
		out[i] = CadenceSample{
			Time:    series[i].Time.UTC(),
			Cadence: series[i].Cadence,
		}
		if series[i].HasDistance {
			d := series[i].DistanceM
			out[i].DistanceM = &d
		}
	}
	return out
}

// CadenceSamplesFromParsed returns workout cadence samples from parsed track data.
func CadenceSamplesFromParsed(parsed *tracks.Data) []CadenceSample {
	if parsed == nil {
		return nil
	}
	return CadenceSamplesFromTrack(parsed.CadenceSeries)
}

// BuildCadenceChartSamples builds the stored chart series (downsampled).
// Zero-cadence inclusion follows tracks.CadenceChartZeroPolicy.
func BuildCadenceChartSamples(parsed *tracks.Data) []CadenceSample {
	full := CadenceSamplesFromParsed(parsed)
	if len(full) == 0 {
		return nil
	}
	filtered := make([]CadenceSample, 0, len(full))
	for _, s := range full {
		if !tracks.AcceptCadence(s.Cadence, tracks.CadenceChartZeroPolicy) {
			continue
		}
		filtered = append(filtered, s)
	}
	return DownsampleCadenceSamples(filtered, CadenceChartMaxPoints)
}

// MarshalCadenceChart encodes chart samples as compact JSON.
func MarshalCadenceChart(samples []CadenceSample) ([]byte, error) {
	if len(samples) == 0 {
		return nil, nil
	}
	payload := cadenceChartJSON{Samples: make([]cadenceChartSampleJSON, 0, len(samples))}
	for _, s := range samples {
		payload.Samples = append(payload.Samples, cadenceChartSampleJSON{
			T:         s.Time.UTC().Format(time.RFC3339),
			Cadence:   s.Cadence,
			DistanceM: s.DistanceM,
		})
	}
	data, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("marshal cadence chart: %w", err)
	}
	return data, nil
}

// UnmarshalCadenceChart decodes chart JSON into samples.
func UnmarshalCadenceChart(data []byte) ([]CadenceSample, error) {
	if len(data) == 0 {
		return nil, nil
	}
	var payload cadenceChartJSON
	if err := json.Unmarshal(data, &payload); err != nil {
		return nil, fmt.Errorf("parse cadence chart: %w", err)
	}
	if len(payload.Samples) == 0 {
		return nil, nil
	}
	out := make([]CadenceSample, 0, len(payload.Samples))
	for _, s := range payload.Samples {
		ts, err := time.Parse(time.RFC3339, s.T)
		if err != nil {
			return nil, fmt.Errorf("parse cadence chart timestamp %q: %w", s.T, err)
		}
		out = append(out, CadenceSample{
			Time:      ts.UTC(),
			Cadence:   s.Cadence,
			DistanceM: s.DistanceM,
		})
	}
	return out, nil
}
