package tracks

import (
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
// Explicit SpeedMps is preferred; otherwise speed is derived only when there is
// real motion data: a device DistanceM delta or a valid GPS segment between
// consecutive timed points. Timed samples without GPS/distance/speed do not
// invent a zero-speed series (e.g. heart-rate-only indoor tracks).
// An all-zero series (explicit zeros or zero distance deltas) is discarded so
// blank speed charts are not stored for strength/indoor activities.
// Inclusion of zero speeds follows SpeedChartZeroPolicy (NaN/Inf always dropped).
//
// DistanceM is meters from the start of the sample list: FIT DistanceM is used
// when present, otherwise cumulative haversine over points with valid GPS.
func SpeedSeriesKmh(points []SamplePoint) []SpeedPoint {
	return speedSeriesKmh(points, SpeedChartZeroPolicy)
}

func speedSeriesKmh(points []SamplePoint, policy ChartZeroPolicy) []SpeedPoint {
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
			if acceptNonNegativeFinite(*cur.SpeedMps, policy) {
				kmh = roundFloat(mpsToKmh(*cur.SpeedMps))
				have = true
			}
		} else if prevTimed != nil {
			dt := cur.Time.Sub(prevTimed.Time).Seconds()
			if dt > 0 {
				if mps, ok := derivedMotionSpeedMps(*prevTimed, *cur, dt); ok {
					kmh = roundFloat(mpsToKmh(mps))
					have = true
				}
			}
		}

		if have && AcceptSpeedKmh(kmh, policy) {
			out = append(out, SpeedPoint{
				Time:      cur.Time.UTC(),
				Kmh:       kmh,
				DistanceM: roundFloat(cum),
			})
		}
		prevTimed = cur
	}
	return meaningfulSpeedSeries(out)
}

// meaningfulSpeedSeries drops an all-zero series (no sample with Kmh > 0).
// In-series zeros are kept when the activity has real motion elsewhere.
func meaningfulSpeedSeries(out []SpeedPoint) []SpeedPoint {
	if len(out) == 0 {
		return nil
	}
	for _, p := range out {
		if p.Kmh > 0 {
			return out
		}
	}
	return nil
}

// derivedMotionSpeedMps returns speed from device distance delta or valid GPS
// path. It does not treat missing/invalid coordinates as a stationary segment.
func derivedMotionSpeedMps(prev, cur SamplePoint, dtSeconds float64) (float64, bool) {
	if dtSeconds <= 0 {
		return 0, false
	}
	if prev.DistanceM != nil && cur.DistanceM != nil &&
		validFloat(*prev.DistanceM) && validFloat(*cur.DistanceM) {
		delta := *cur.DistanceM - *prev.DistanceM
		if delta < 0 {
			return 0, false
		}
		return delta / dtSeconds, true
	}
	if validCoord(prev.Lat, prev.Lng) && validCoord(cur.Lat, cur.Lng) {
		dist := haversine(
			LatLng{Lat: prev.Lat, Lng: prev.Lng},
			LatLng{Lat: cur.Lat, Lng: cur.Lng},
			earthRadiusMeters,
		)
		return dist / dtSeconds, true
	}
	return 0, false
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
