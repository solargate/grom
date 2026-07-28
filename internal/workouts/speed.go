package workouts

import (
	"encoding/json"
	"fmt"
	"math"
	"sort"
	"time"

	"github.com/solargate/grom/internal/storage/keys"
	"github.com/solargate/grom/internal/tracks"
	"gopkg.in/yaml.v3"
)

// SpeedSidecarFormat selects how the speed series is stored on disk.
type SpeedSidecarFormat int

const (
	// SpeedSidecarYAML stores speed as speed.yaml (file storage driver).
	SpeedSidecarYAML SpeedSidecarFormat = iota
	// SpeedSidecarJSON stores speed as speed.json (bbolt storage driver).
	SpeedSidecarJSON
)

// speedSidecarPoint is the per-timestamp value stored in speed.yaml / speed.json.
type speedSidecarPoint struct {
	SpeedKmh  float64 `json:"speed_kmh" yaml:"speed_kmh"`
	DistanceM float64 `json:"distance_m" yaml:"distance_m"`
}

// SpeedFileName returns the sidecar basename for the given format.
func SpeedFileName(format SpeedSidecarFormat) string {
	if format == SpeedSidecarJSON {
		return keys.SpeedFileJSON
	}
	return keys.SpeedFileYAML
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

// MarshalSpeedSidecar encodes a speed series for the given sidecar format.
func MarshalSpeedSidecar(format SpeedSidecarFormat, samples []SpeedSample) ([]byte, error) {
	m := speedSamplesToMap(samples)
	if len(m) == 0 {
		return nil, nil
	}
	switch format {
	case SpeedSidecarJSON:
		data, err := json.MarshalIndent(m, "", "  ")
		if err != nil {
			return nil, fmt.Errorf("marshal speed json: %w", err)
		}
		return append(data, '\n'), nil
	default:
		data, err := yaml.Marshal(m)
		if err != nil {
			return nil, fmt.Errorf("marshal speed yaml: %w", err)
		}
		return data, nil
	}
}

// UnmarshalSpeedSidecar decodes a speed sidecar payload.
func UnmarshalSpeedSidecar(format SpeedSidecarFormat, data []byte) ([]SpeedSample, error) {
	if len(data) == 0 {
		return nil, nil
	}
	var raw map[string]speedSidecarPoint
	var err error
	switch format {
	case SpeedSidecarJSON:
		err = json.Unmarshal(data, &raw)
	default:
		err = yaml.Unmarshal(data, &raw)
	}
	if err != nil {
		return nil, fmt.Errorf("parse speed sidecar: %w", err)
	}
	return speedMapToSamples(raw)
}

func speedSamplesToMap(samples []SpeedSample) map[string]speedSidecarPoint {
	if len(samples) == 0 {
		return nil
	}
	m := make(map[string]speedSidecarPoint, len(samples))
	for _, s := range samples {
		if !positiveSpeedKmh(s.SpeedKmh) {
			continue
		}
		m[s.Time.UTC().Format(time.RFC3339)] = speedSidecarPoint{
			SpeedKmh:  s.SpeedKmh,
			DistanceM: s.DistanceM,
		}
	}
	if len(m) == 0 {
		return nil
	}
	return m
}

func positiveSpeedKmh(kmh float64) bool {
	return !math.IsNaN(kmh) && !math.IsInf(kmh, 0) && kmh > 0
}

func speedMapToSamples(raw map[string]speedSidecarPoint) ([]SpeedSample, error) {
	if len(raw) == 0 {
		return nil, nil
	}
	keysSorted := make([]string, 0, len(raw))
	for k := range raw {
		keysSorted = append(keysSorted, k)
	}
	sort.Strings(keysSorted)
	out := make([]SpeedSample, 0, len(keysSorted))
	for _, k := range keysSorted {
		ts, err := time.Parse(time.RFC3339, k)
		if err != nil {
			return nil, fmt.Errorf("parse speed timestamp %q: %w", k, err)
		}
		pt := raw[k]
		out = append(out, SpeedSample{
			Time:      ts.UTC(),
			SpeedKmh:  pt.SpeedKmh,
			DistanceM: pt.DistanceM,
		})
	}
	return out, nil
}
