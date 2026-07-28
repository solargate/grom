package tracks

import (
	"math"

	"github.com/muktihari/fit/profile/basetype"
	"github.com/muktihari/fit/profile/filedef"
	"github.com/muktihari/fit/profile/mesgdef"
	"github.com/muktihari/fit/profile/typedef"
	"github.com/muktihari/fit/profile/untyped/mesgnum"
)

func extractFITStats(activity *filedef.Activity) (Stats, []SpeedPoint) {
	var stats Stats
	if activity == nil {
		return stats, nil
	}

	samples := fitSamplePoints(activity.Records)
	series := SpeedSeriesKmh(samples)

	if len(activity.Sessions) == 0 {
		calc := calculateStatsFromSamples(samples)
		return calc, series
	}

	session := activity.Sessions[0]
	extractFITSessionStats(session, &stats)
	extractFITSetStats(activity, &stats)

	calc := calculateStatsFromSamples(samples)
	mergeCalculatedStats(&stats, &calc)
	applyCalculatedCadence(&stats, &calc)

	return stats, series
}

func extractFITSessionStats(session *mesgdef.Session, stats *Stats) {
	if elapsed := session.TotalElapsedTimeScaled(); elapsed > 0 && !math.IsNaN(elapsed) {
		stats.DurationTotalSeconds.setExplicit(int(math.Round(elapsed)))
	}

	moving := sessionDurationMovingSeconds(session)
	if moving > 0 {
		stats.DurationSeconds.setExplicit(moving)
	}

	if dist := session.TotalDistanceScaled(); dist > 0 && !math.IsNaN(dist) {
		_ = dist
	}

	if speed := session.EnhancedMaxSpeedScaled(); validSpeedMps(speed) {
		stats.SpeedMaxKmh.setExplicit(roundFloat(mpsToKmh(speed)))
	} else if speed := session.MaxSpeedScaled(); validSpeedMps(speed) {
		stats.SpeedMaxKmh.setExplicit(roundFloat(mpsToKmh(speed)))
	}

	if speed := session.EnhancedAvgSpeedScaled(); validSpeedMps(speed) {
		stats.SpeedAvgKmh.setExplicit(roundFloat(mpsToKmh(speed)))
	} else if speed := session.AvgSpeedScaled(); validSpeedMps(speed) {
		stats.SpeedAvgKmh.setExplicit(roundFloat(mpsToKmh(speed)))
	}

	if session.TotalAscent != basetype.Uint16Invalid && session.TotalAscent < 65535 {
		stats.ElevationGain.setExplicit(float64(session.TotalAscent))
	}
	if session.TotalDescent != basetype.Uint16Invalid && session.TotalDescent < 65535 {
		stats.ElevationLoss.setExplicit(float64(session.TotalDescent))
	}

	if alt := session.MinAltitudeScaled(); validAltitude(alt) {
		stats.ElevationLow.setExplicit(roundFloat(alt))
	}
	if alt := session.MaxAltitudeScaled(); validAltitude(alt) {
		stats.ElevationHigh.setExplicit(roundFloat(alt))
	}

	if grade := maxAbsGrade(session.MaxPosGradeScaled(), session.MaxNegGradeScaled()); grade > 0 {
		stats.GradeMax.setExplicit(roundFloat(grade))
	}
	if grade := session.AvgGradeScaled(); validGrade(grade) {
		stats.GradeAvg.setExplicit(roundFloat(math.Abs(grade)))
	}

	if session.MaxCadence != basetype.Uint8Invalid && session.MaxCadence > 0 {
		stats.CadenceMax.setExplicit(float64(session.MaxCadence))
	}
	if session.AvgCadence != basetype.Uint8Invalid && session.AvgCadence > 0 {
		stats.CadenceAvg.setExplicit(float64(session.AvgCadence))
	}

	if session.MaxHeartRate != basetype.Uint8Invalid {
		stats.HeartRateMax.setExplicit(float64(session.MaxHeartRate))
	}
	if session.AvgHeartRate != basetype.Uint8Invalid {
		stats.HeartRateAvg.setExplicit(float64(session.AvgHeartRate))
	}

	if session.MaxPower != basetype.Uint16Invalid && session.MaxPower < 65535 {
		stats.WattsMax.setExplicit(float64(session.MaxPower))
	}
	if session.AvgPower != basetype.Uint16Invalid && session.AvgPower < 65535 {
		stats.WattsAvg.setExplicit(float64(session.AvgPower))
	}

	if session.TotalCalories > 0 && session.TotalCalories < 65535 {
		stats.Calories.setExplicit(float64(session.TotalCalories))
	}

	if session.MaxTemperature != basetype.Sint8Invalid && session.MaxTemperature != 127 {
		stats.TemperatureMax.setExplicit(float64(session.MaxTemperature))
	}
	if session.AvgTemperature != basetype.Sint8Invalid && session.AvgTemperature != 127 {
		stats.TemperatureAvg.setExplicit(float64(session.AvgTemperature))
	}

	if session.TotalCycles != basetype.Uint32Invalid && session.TotalCycles < 0xFFFF0000 {
		switch session.Sport {
		case typedef.SportWalking, typedef.SportRunning:
			stats.StepsTotal.setExplicit(int(session.TotalCycles))
		default:
			stats.CyclesTotal.setExplicit(int(session.TotalCycles))
		}
	}
}

func sessionDurationMovingSeconds(session *mesgdef.Session) int {
	if moving := session.TotalMovingTimeScaled(); moving > 0 && !math.IsNaN(moving) {
		return int(math.Round(moving))
	}
	if timer := session.TotalTimerTimeScaled(); timer > 0 && !math.IsNaN(timer) {
		return int(math.Round(timer))
	}
	return 0
}

func extractFITSetStats(activity *filedef.Activity, stats *Stats) {
	sets := 0
	reps := 0
	for _, mesg := range activity.UnrelatedMessages {
		if mesg.Num != mesgnum.Set {
			continue
		}
		set := mesgdef.NewSet(&mesg)
		sets++
		if set.Repetitions != basetype.Uint16Invalid && set.Repetitions > 0 {
			reps += int(set.Repetitions)
		}
	}
	if sets > 0 {
		stats.SetsTotal.setExplicit(sets)
	}
	if reps > 0 {
		stats.RepsTotal.setExplicit(reps)
	}
}

func fitSamplePoints(records []*mesgdef.Record) []SamplePoint {
	points := make([]SamplePoint, 0, len(records))
	for _, record := range records {
		pt := SamplePoint{
			Lat:     record.PositionLatDegrees(),
			Lng:     record.PositionLongDegrees(),
			Time:    record.Timestamp,
			HasTime: !record.Timestamp.IsZero(),
		}
		if alt := record.AltitudeScaled(); validAltitude(alt) {
			v := alt
			pt.Elevation = &v
		}
		if speed := record.EnhancedSpeedScaled(); validSpeedMps(speed) {
			v := speed
			pt.SpeedMps = &v
		} else if speed := record.SpeedScaled(); validSpeedMps(speed) {
			v := speed
			pt.SpeedMps = &v
		}
		if record.HeartRate != basetype.Uint8Invalid {
			v := float64(record.HeartRate)
			pt.HeartRate = &v
		}
		if record.Cadence != basetype.Uint8Invalid {
			v := float64(record.Cadence)
			pt.Cadence = &v
		}
		if record.Power != basetype.Uint16Invalid && record.Power < 65535 {
			v := float64(record.Power)
			pt.Power = &v
		}
		if record.Temperature != basetype.Sint8Invalid && record.Temperature != 127 {
			v := float64(record.Temperature)
			pt.Temperature = &v
		}
		points = append(points, pt)
	}
	return points
}

func applyCalculatedCadence(stats, calc *Stats) {
	if v := floatValue(calc.CadenceMax); v >= 0 {
		stats.CadenceMax.setCalculatedOverride(v)
	}
	if v := floatValue(calc.CadenceAvg); v >= 0 {
		stats.CadenceAvg.setCalculatedOverride(v)
	}
}

func mergeCalculatedStats(stats, calc *Stats) {
	stats.DurationTotalSeconds.setCalculated(intValue(calc.DurationTotalSeconds))
	stats.DurationSeconds.setCalculated(intValue(calc.DurationSeconds))
	stats.SpeedMaxKmh.setCalculated(floatValue(calc.SpeedMaxKmh))
	stats.SpeedAvgKmh.setCalculated(floatValue(calc.SpeedAvgKmh))
	stats.ElevationGain.setCalculated(floatValue(calc.ElevationGain))
	stats.ElevationLoss.setCalculated(floatValue(calc.ElevationLoss))
	stats.ElevationLow.setCalculated(floatValue(calc.ElevationLow))
	stats.ElevationHigh.setCalculated(floatValue(calc.ElevationHigh))
	stats.GradeMax.setCalculated(floatValue(calc.GradeMax))
	stats.GradeAvg.setCalculated(floatValue(calc.GradeAvg))
	stats.CadenceMax.setCalculated(floatValue(calc.CadenceMax))
	stats.CadenceAvg.setCalculated(floatValue(calc.CadenceAvg))
	stats.HeartRateMax.setCalculated(floatValue(calc.HeartRateMax))
	stats.HeartRateAvg.setCalculated(floatValue(calc.HeartRateAvg))
	stats.WattsMax.setCalculated(floatValue(calc.WattsMax))
	stats.WattsAvg.setCalculated(floatValue(calc.WattsAvg))
	stats.TemperatureMax.setCalculated(floatValue(calc.TemperatureMax))
	stats.TemperatureAvg.setCalculated(floatValue(calc.TemperatureAvg))
}

func intValue(stat IntStat) int {
	if stat.Value == nil {
		return -1
	}
	return *stat.Value
}

func floatValue(stat FloatStat) float64 {
	if stat.Value == nil {
		return -1
	}
	return *stat.Value
}

func validSpeedMps(v float64) bool {
	return validFloat(v) && v > 0
}

func validAltitude(v float64) bool {
	return !math.IsNaN(v) && !math.IsInf(v, 0) && v > -500 && v < 9000
}

func validGrade(v float64) bool {
	return !math.IsNaN(v) && !math.IsInf(v, 0)
}

func maxAbsGrade(pos, neg float64) float64 {
	max := 0.0
	if validGrade(pos) {
		max = math.Max(max, math.Abs(pos))
	}
	if validGrade(neg) {
		max = math.Max(max, math.Abs(neg))
	}
	return max
}
