package tracks

import (
	"time"
)

// CadencePoint is a single cadence sample bound to an absolute UTC timestamp
// and optional cumulative distance from the start of the track (meters).
// DistanceM is only populated when the track has GPS (hasGPS=true).
// Values are raw device/track units (FIT/Strava cadence); UI may scale for foot sports.
type CadencePoint struct {
	Time        time.Time
	Cadence     float64
	DistanceM   float64
	HasDistance bool
}

// CadenceSeries builds a per-sample cadence series from track samples.
// Only timed points with a Cadence reading are considered.
// Inclusion of zero cadence follows CadenceChartZeroPolicy (NaN/Inf always dropped).
// When hasGPS is false, distance is omitted (HasDistance=false).
func CadenceSeries(points []SamplePoint, hasGPS bool) []CadencePoint {
	return cadenceSeries(points, hasGPS, CadenceChartZeroPolicy)
}

func cadenceSeries(points []SamplePoint, hasGPS bool, policy ChartZeroPolicy) []CadencePoint {
	if len(points) == 0 {
		return nil
	}

	out := make([]CadencePoint, 0, len(points))
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

		if !cur.HasTime || cur.Cadence == nil {
			continue
		}
		cad := *cur.Cadence
		if !AcceptCadence(cad, policy) {
			continue
		}

		pt := CadencePoint{
			Time:    cur.Time.UTC(),
			Cadence: roundFloat(cad),
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
