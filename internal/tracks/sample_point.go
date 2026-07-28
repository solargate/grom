package tracks

import (
	"math"
	"time"
)

const (
	movingSpeedThresholdMps = 0.5
	gapThresholdSeconds     = 120
	elevationThresholdMeters = 1.0
)

type SamplePoint struct {
	Lat         float64
	Lng         float64
	Time        time.Time
	HasTime     bool
	Elevation   *float64
	SpeedMps    *float64
	DistanceM   *float64 // cumulative from activity/track start when known (e.g. FIT)
	HeartRate   *float64
	Cadence     *float64
	Power       *float64
	Temperature *float64
}

func calculateStatsFromSamples(points []SamplePoint) Stats {
	var stats Stats
	if len(points) == 0 {
		return stats
	}

	stats.DurationTotalSeconds.setCalculated(calcElapsedSeconds(points))
	stats.DurationSeconds.setCalculated(calcActiveSeconds(points))

	dist := samplePathDistanceMeters(points)
	if dist > 0 {
		moving := 0
		if stats.DurationSeconds.Value != nil {
			moving = *stats.DurationSeconds.Value
		}
		if moving > 0 {
			stats.SpeedAvgKmh.setCalculated(roundFloat(mpsToKmh(dist / float64(moving))))
		}
	}

	calcElevationStats(points, &stats)
	calcGradeStats(points, &stats)
	calcSpeedStatsFromSamples(points, &stats)
	calcSensorStats(points, &stats)

	return stats
}

func calcElapsedSeconds(points []SamplePoint) int {
	first, last := firstLastTimed(points)
	if first == nil || last == nil || !last.Time.After(first.Time) {
		return 0
	}
	return int(last.Time.Sub(first.Time).Seconds())
}

func calcMovingSeconds(points []SamplePoint) int {
	var total float64
	for i := 1; i < len(points); i++ {
		prev, cur := points[i-1], points[i]
		if !prev.HasTime || !cur.HasTime {
			continue
		}
		dt := cur.Time.Sub(prev.Time).Seconds()
		if dt <= 0 || dt > gapThresholdSeconds {
			continue
		}
		speed := pointSpeedMps(prev, cur, dt)
		if speed >= movingSpeedThresholdMps {
			total += dt
		}
	}
	return int(mathRound(total))
}

func calcActiveSeconds(points []SamplePoint) int {
	var total float64
	for i := 1; i < len(points); i++ {
		prev, cur := points[i-1], points[i]
		if !prev.HasTime || !cur.HasTime {
			continue
		}
		dt := cur.Time.Sub(prev.Time).Seconds()
		if dt <= 0 || dt > gapThresholdSeconds {
			continue
		}
		total += dt
	}
	if total > 0 {
		return int(mathRound(total))
	}
	return calcMovingSeconds(points)
}

func firstLastTimed(points []SamplePoint) (*SamplePoint, *SamplePoint) {
	var first, last *SamplePoint
	for i := range points {
		if !points[i].HasTime {
			continue
		}
		if first == nil {
			first = &points[i]
		}
		last = &points[i]
	}
	return first, last
}

func pointSpeedMps(prev, cur SamplePoint, dtSeconds float64) float64 {
	if cur.SpeedMps != nil && *cur.SpeedMps >= 0 {
		return *cur.SpeedMps
	}
	if dtSeconds <= 0 {
		return 0
	}
	dist := haversine(LatLng{Lat: prev.Lat, Lng: prev.Lng}, LatLng{Lat: cur.Lat, Lng: cur.Lng}, earthRadiusMeters)
	return dist / dtSeconds
}

func samplePathDistanceMeters(points []SamplePoint) float64 {
	if len(points) < 2 {
		return 0
	}
	var total float64
	for i := 1; i < len(points); i++ {
		total += haversine(
			LatLng{Lat: points[i-1].Lat, Lng: points[i-1].Lng},
			LatLng{Lat: points[i].Lat, Lng: points[i].Lng},
			earthRadiusMeters,
		)
	}
	return total
}

func calcElevationStats(points []SamplePoint, stats *Stats) {
	var low, high *float64
	var gain, loss float64
	var prevEle *float64

	for i := range points {
		ele := points[i].Elevation
		if ele == nil {
			continue
		}
		if low == nil || *ele < *low {
			low = ele
		}
		if high == nil || *ele > *high {
			high = ele
		}
		if prevEle != nil {
			delta := *ele - *prevEle
			if delta >= elevationThresholdMeters {
				gain += delta
			} else if delta <= -elevationThresholdMeters {
				loss += -delta
			}
		}
		prevEle = ele
	}

	if low != nil {
		stats.ElevationLow.setCalculated(roundFloat(*low))
	}
	if high != nil {
		stats.ElevationHigh.setCalculated(roundFloat(*high))
	}
	if gain > 0 {
		stats.ElevationGain.setCalculated(roundFloat(gain))
	}
	if loss > 0 {
		stats.ElevationLoss.setCalculated(roundFloat(loss))
	}
}

func calcGradeStats(points []SamplePoint, stats *Stats) {
	var sumGradeDist float64
	var totalDist float64
	var maxGrade float64

	for i := 1; i < len(points); i++ {
		prev, cur := points[i-1], points[i]
		if prev.Elevation == nil || cur.Elevation == nil {
			continue
		}
		horiz := haversine(
			LatLng{Lat: prev.Lat, Lng: prev.Lng},
			LatLng{Lat: cur.Lat, Lng: cur.Lng},
			earthRadiusMeters,
		)
		if horiz <= 0 {
			continue
		}
		grade := ((*cur.Elevation - *prev.Elevation) / horiz) * 100
		absGrade := mathAbs(grade)
		if absGrade > maxGrade {
			maxGrade = absGrade
		}
		sumGradeDist += grade * horiz
		totalDist += horiz
	}

	if maxGrade > 0 {
		stats.GradeMax.setCalculated(roundFloat(maxGrade))
	}
	if totalDist > 0 {
		stats.GradeAvg.setCalculated(roundFloat(sumGradeDist / totalDist))
	}
}

func calcSpeedStatsFromSamples(points []SamplePoint, stats *Stats) {
	var maxKmh float64
	for i := range points {
		if points[i].SpeedMps == nil || *points[i].SpeedMps < 0 {
			continue
		}
		kmh := mpsToKmh(*points[i].SpeedMps)
		if kmh > maxKmh {
			maxKmh = kmh
		}
	}
	if maxKmh > 0 {
		stats.SpeedMaxKmh.setCalculated(roundFloat(maxKmh))
	}

	for i := 1; i < len(points); i++ {
		prev, cur := points[i-1], points[i]
		if !prev.HasTime || !cur.HasTime {
			continue
		}
		dt := cur.Time.Sub(prev.Time).Seconds()
		if dt <= 0 {
			continue
		}
		speed := pointSpeedMps(prev, cur, dt)
		if speed <= 0 {
			continue
		}
		kmh := mpsToKmh(speed)
		if kmh > maxKmh {
			maxKmh = kmh
		}
	}
	if maxKmh > 0 && stats.SpeedMaxKmh.Source == SourceNone {
		stats.SpeedMaxKmh.setCalculated(roundFloat(maxKmh))
	}
}

func calcSensorStats(points []SamplePoint, stats *Stats) {
	var hrSum, hrCount float64
	var hrMax float64
	var cadSum, cadCount float64
	var cadMax float64
	var pwrSum, pwrCount float64
	var pwrMax float64
	var tempSum, tempCount float64
	var tempMax float64

	for i := range points {
		if points[i].HeartRate != nil && *points[i].HeartRate > 0 {
			v := *points[i].HeartRate
			hrSum += v
			hrCount++
			if v > hrMax {
				hrMax = v
			}
		}
		if points[i].Cadence != nil && *points[i].Cadence > 0 {
			v := *points[i].Cadence
			cadSum += v
			cadCount++
			if v > cadMax {
				cadMax = v
			}
		}
		if points[i].Power != nil && *points[i].Power > 0 {
			v := *points[i].Power
			pwrSum += v
			pwrCount++
			if v > pwrMax {
				pwrMax = v
			}
		}
		if points[i].Temperature != nil {
			v := *points[i].Temperature
			tempSum += v
			tempCount++
			if tempCount == 1 || v > tempMax {
				tempMax = v
			}
		}
	}

	if hrMax > 0 {
		stats.HeartRateMax.setCalculated(roundFloat(hrMax))
	}
	if hrCount > 0 {
		stats.HeartRateAvg.setCalculated(roundFloat(hrSum / hrCount))
	}
	if cadMax > 0 {
		stats.CadenceMax.setCalculated(roundFloat(cadMax))
	}
	if cadCount > 0 {
		stats.CadenceAvg.setCalculated(roundFloat(cadSum / cadCount))
	}
	if pwrMax > 0 {
		stats.WattsMax.setCalculated(roundFloat(pwrMax))
	}
	if pwrCount > 0 {
		stats.WattsAvg.setCalculated(roundFloat(pwrSum / pwrCount))
	}
	if tempCount > 0 {
		stats.TemperatureMax.setCalculated(roundFloat(tempMax))
		stats.TemperatureAvg.setCalculated(roundFloat(tempSum / tempCount))
	}
}

const earthRadiusMeters = 6371000.0

func mathRound(v float64) float64 {
	return math.Round(v)
}

func mathAbs(v float64) float64 {
	return math.Abs(v)
}
