package workouts

import "math"

// DownsampleCadenceSamples evenly spaces samples keeping first and last.
// No-op when len(samples) <= maxPoints or maxPoints < 2.
func DownsampleCadenceSamples(samples []CadenceSample, maxPoints int) []CadenceSample {
	if len(samples) <= maxPoints || maxPoints < 2 {
		return samples
	}
	out := make([]CadenceSample, 0, maxPoints)
	last := len(samples) - 1
	prev := -1
	for i := 0; i < maxPoints; i++ {
		idx := int(math.Round(float64(i*last) / float64(maxPoints-1)))
		if idx == prev {
			continue
		}
		out = append(out, samples[idx])
		prev = idx
	}
	return out
}
