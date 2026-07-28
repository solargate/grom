package workouts

import (
	"encoding/json"
	"fmt"
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
			Time:     series[i].Time.UTC(),
			SpeedKmh: series[i].Kmh,
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
	var raw map[string]float64
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

func speedSamplesToMap(samples []SpeedSample) map[string]float64 {
	if len(samples) == 0 {
		return nil
	}
	m := make(map[string]float64, len(samples))
	for _, s := range samples {
		m[s.Time.UTC().Format(time.RFC3339)] = s.SpeedKmh
	}
	return m
}

func speedMapToSamples(raw map[string]float64) ([]SpeedSample, error) {
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
		out = append(out, SpeedSample{
			Time:     ts.UTC(),
			SpeedKmh: raw[k],
		})
	}
	return out, nil
}
