package tracks

import "math"

// ChartZeroPolicy controls whether zero values are kept on speed/heart-rate chart series.
type ChartZeroPolicy int

const (
	// ChartZeroOmit drops non-positive values (value > 0 required).
	ChartZeroOmit ChartZeroPolicy = iota
	// ChartZeroKeep keeps finite zeros (value >= 0); NaN, Inf, and negatives are still dropped.
	ChartZeroKeep
)

// SpeedChartZeroPolicy selects how zero speeds are handled when building the speed chart series.
const SpeedChartZeroPolicy = ChartZeroKeep

// HeartRateChartZeroPolicy selects how zero BPM values are handled when building the heart-rate chart series.
const HeartRateChartZeroPolicy = ChartZeroKeep

// AcceptSpeedKmh reports whether a km/h sample should be included under policy.
func AcceptSpeedKmh(kmh float64, policy ChartZeroPolicy) bool {
	return acceptNonNegativeFinite(kmh, policy)
}

// AcceptHeartRateBPM reports whether a BPM sample should be included under policy.
func AcceptHeartRateBPM(bpm float64, policy ChartZeroPolicy) bool {
	return acceptNonNegativeFinite(bpm, policy)
}

// AcceptSpeedMpsForSample reports whether an explicit m/s reading should be stored on a sample
// for later chart series construction under SpeedChartZeroPolicy.
func AcceptSpeedMpsForSample(mps float64) bool {
	return acceptNonNegativeFinite(mps, SpeedChartZeroPolicy)
}

// AcceptHeartRateForSample reports whether an explicit BPM reading should be stored on a sample
// for later chart series construction under HeartRateChartZeroPolicy.
func AcceptHeartRateForSample(bpm float64) bool {
	return acceptNonNegativeFinite(bpm, HeartRateChartZeroPolicy)
}

func acceptNonNegativeFinite(v float64, policy ChartZeroPolicy) bool {
	if math.IsNaN(v) || math.IsInf(v, 0) {
		return false
	}
	switch policy {
	case ChartZeroKeep:
		return v >= 0
	default:
		return v > 0
	}
}
