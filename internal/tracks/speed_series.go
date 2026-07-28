package tracks

import (
	"math"
	"time"
)

// SpeedPoint is a single speed sample bound to an absolute UTC timestamp
// and cumulative distance from the start of the track (meters).
type SpeedPoint struct {
	Time      time.Time
	Kmh       float64
	DistanceM float64
}

// SpeedSeriesKmh builds a per-sample speed series in km/h from track samples.
// Explicit SpeedMps is preferred; otherwise speed is derived from consecutive
// timed points (including across long gaps). Points without time are skipped
// for the series, but still contribute to path distance.
// Only positive finite speeds are included (zero, NaN, and Inf are dropped).
//
// DistanceM is meters from the start of the sample list: FIT DistanceM is used
// when present, otherwise cumulative haversine over points with valid GPS.
func SpeedSeriesKmh(points []SamplePoint) []SpeedPoint {
	if len(points) == 0 {
		return nil
	}

	out := make([]SpeedPoint, 0, len(points))
	var prevTimed *SamplePoint
	var prevPath *SamplePoint
	var cum float64

	for i := range points {
		cur := &points[i]
		cum = advancePathDistance(cum, prevPath, cur)
		if validCoord(cur.Lat, cur.Lng) {
			prevPath = cur
		}

		if !cur.HasTime {
			continue
		}

		var kmh float64
		have := false
		if cur.SpeedMps != nil {
			if validSpeedMps(*cur.SpeedMps) {
				kmh = roundFloat(mpsToKmh(*cur.SpeedMps))
				have = true
			}
		} else if prevTimed != nil {
			dt := cur.Time.Sub(prevTimed.Time).Seconds()
			if dt > 0 {
				kmh = roundFloat(mpsToKmh(pointSpeedMps(*prevTimed, *cur, dt)))
				have = true
			}
		}

		if have && positiveSpeedKmh(kmh) {
			out = append(out, SpeedPoint{
				Time:      cur.Time.UTC(),
				Kmh:       kmh,
				DistanceM: roundFloat(cum),
			})
		}
		prevTimed = cur
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

func positiveSpeedKmh(kmh float64) bool {
	return !math.IsNaN(kmh) && !math.IsInf(kmh, 0) && kmh > 0
}

func advancePathDistance(cum float64, prev, cur *SamplePoint) float64 {
	if cur.DistanceM != nil && validFloat(*cur.DistanceM) {
		return *cur.DistanceM
	}
	if prev == nil || !validCoord(prev.Lat, prev.Lng) || !validCoord(cur.Lat, cur.Lng) {
		return cum
	}
	return cum + haversine(
		LatLng{Lat: prev.Lat, Lng: prev.Lng},
		LatLng{Lat: cur.Lat, Lng: cur.Lng},
		earthRadiusMeters,
	)
}
