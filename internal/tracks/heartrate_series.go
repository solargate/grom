package tracks

import (
	"time"
)

// HeartRatePoint is a single heart-rate sample bound to an absolute UTC timestamp
// and optional cumulative distance from the start of the track (meters).
// DistanceM is only populated when the track has GPS (hasGPS=true).
type HeartRatePoint struct {
	Time        time.Time
	BPM         float64
	DistanceM   float64
	HasDistance bool
}

// HeartRateSeries builds a per-sample heart-rate series from track samples.
// Only timed points with a HeartRate reading are considered.
// Inclusion of zero BPM follows HeartRateChartZeroPolicy (NaN/Inf always dropped).
// When hasGPS is false, distance is omitted (HasDistance=false).
func HeartRateSeries(points []SamplePoint, hasGPS bool) []HeartRatePoint {
	return heartRateSeries(points, hasGPS, HeartRateChartZeroPolicy)
}

func heartRateSeries(points []SamplePoint, hasGPS bool, policy ChartZeroPolicy) []HeartRatePoint {
	if len(points) == 0 {
		return nil
	}

	out := make([]HeartRatePoint, 0, len(points))
	var prevPath *SamplePoint
	var cum float64

	for i := range points {
		cur := &points[i]
		if hasGPS {
			cum = advancePathDistance(cum, prevPath, cur)
			if validCoord(cur.Lat, cur.Lng) {
				prevPath = cur
			}
		}

		if !cur.HasTime || cur.HeartRate == nil {
			continue
		}
		bpm := *cur.HeartRate
		if !AcceptHeartRateBPM(bpm, policy) {
			continue
		}

		pt := HeartRatePoint{
			Time: cur.Time.UTC(),
			BPM:  roundFloat(bpm),
		}
		if hasGPS {
			pt.DistanceM = roundFloat(cum)
			pt.HasDistance = true
		}
		out = append(out, pt)
	}
	if len(out) == 0 {
		return nil
	}
	return out
}
