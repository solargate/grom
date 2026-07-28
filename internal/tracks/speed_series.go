package tracks

import "time"

// SpeedPoint is a single speed sample bound to an absolute UTC timestamp.
type SpeedPoint struct {
	Time time.Time
	Kmh  float64
}

// SpeedSeriesKmh builds a per-sample speed series in km/h from track samples.
// Explicit SpeedMps is preferred; otherwise speed is derived from consecutive
// timed points (including across long gaps). Points without time are skipped.
func SpeedSeriesKmh(points []SamplePoint) []SpeedPoint {
	if len(points) == 0 {
		return nil
	}

	out := make([]SpeedPoint, 0, len(points))
	var prev *SamplePoint
	for i := range points {
		cur := &points[i]
		if !cur.HasTime {
			continue
		}

		var kmh float64
		have := false
		if cur.SpeedMps != nil && *cur.SpeedMps >= 0 {
			kmh = roundFloat(mpsToKmh(*cur.SpeedMps))
			have = true
		} else if prev != nil {
			dt := cur.Time.Sub(prev.Time).Seconds()
			if dt > 0 {
				kmh = roundFloat(mpsToKmh(pointSpeedMps(*prev, *cur, dt)))
				have = true
			}
		}

		if have {
			out = append(out, SpeedPoint{
				Time: cur.Time.UTC(),
				Kmh:  kmh,
			})
		}
		prev = cur
	}
	if len(out) == 0 {
		return nil
	}
	return out
}
