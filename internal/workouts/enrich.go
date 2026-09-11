package workouts

import "github.com/solargate/grom/internal/tracks"

type MergeMode int

const (
	// MergeModeTrackCreate fills empty workout metrics from the track; client-provided
	// values (distance, durations, speeds, elevation, etc.) are preserved.
	MergeModeTrackCreate MergeMode = iota
	// MergeModeTrackAttach fills only empty workout metrics (preserves CSV/import data).
	MergeModeTrackAttach
)

func MergeTrackStats(workout *Workout, data *tracks.Data, mode MergeMode) {
	if workout == nil || data == nil {
		return
	}
	_ = mode // create and attach both preserve client-provided metrics

	stats := data.Stats
	distanceMeters := workout.Distance
	if distanceMeters <= 0 && data.DistanceMeters != nil && *data.DistanceMeters > 0 {
		distanceMeters = *data.DistanceMeters
	}
	stats.FinalizePace(&distanceMeters)

	mergeIntStat(&workout.DurationSeconds, stats.DurationSeconds, true)
	mergeIntStat(&workout.DurationTotalSeconds, stats.DurationTotalSeconds, true)

	mergeFloatPtr(&workout.SpeedMaxKmh, stats.SpeedMaxKmh)
	mergeFloatPtr(&workout.SpeedAvgKmh, stats.SpeedAvgKmh)
	mergeFloatPtr(&workout.ElevationGain, stats.ElevationGain)
	mergeFloatPtr(&workout.ElevationLoss, stats.ElevationLoss)
	mergeFloatPtr(&workout.ElevationLow, stats.ElevationLow)
	mergeFloatPtr(&workout.ElevationHigh, stats.ElevationHigh)
	mergeFloatPtr(&workout.GradeMax, stats.GradeMax)
	mergeFloatPtr(&workout.GradeAvg, stats.GradeAvg)
	mergeRoundedFloatPtr(&workout.CadenceMax, stats.CadenceMax)
	mergeRoundedFloatPtr(&workout.CadenceAvg, stats.CadenceAvg)
	mergeRoundedFloatPtr(&workout.HeartRateMax, stats.HeartRateMax)
	mergeRoundedFloatPtr(&workout.HeartRateAvg, stats.HeartRateAvg)
	mergeRoundedFloatPtr(&workout.WattsMax, stats.WattsMax)
	mergeRoundedFloatPtr(&workout.WattsAvg, stats.WattsAvg)
	mergeRoundedFloatPtr(&workout.Calories, stats.Calories)
	mergeFloatPtr(&workout.TemperatureMax, stats.TemperatureMax)
	mergeFloatPtr(&workout.TemperatureAvg, stats.TemperatureAvg)
	mergeIntPtr(&workout.StepsTotal, stats.StepsTotal)
	mergeIntPtr(&workout.CyclesTotal, stats.CyclesTotal)
	mergeIntPtr(&workout.SetsTotal, stats.SetsTotal)
	mergeIntPtr(&workout.RepsTotal, stats.RepsTotal)
	mergeStringPtr(&workout.TempAvgKmm, stats.TempAvgKmm)
}

func mergeIntStat(target *int, stat tracks.IntStat, allowCalculated bool) {
	if stat.Value == nil || stat.Source == tracks.SourceNone {
		return
	}
	if stat.Source == tracks.SourceCalculated && !allowCalculated {
		return
	}
	if *target > 0 {
		return
	}
	*target = *stat.Value
}

func mergeFloatPtr(target **float64, stat tracks.FloatStat) {
	if stat.Value == nil || stat.Source == tracks.SourceNone {
		return
	}
	if isSetFloatPtr(target) {
		return
	}
	v := *stat.Value
	*target = &v
}

func mergeRoundedFloatPtr(target **float64, stat tracks.FloatStat) {
	if stat.Value == nil || stat.Source == tracks.SourceNone {
		return
	}
	rounded := float64(int(*stat.Value + 0.5))
	clone := stat
	clone.Value = &rounded
	mergeFloatPtr(target, clone)
}

func mergeIntPtr(target **int, stat tracks.IntStat) {
	if stat.Value == nil || stat.Source == tracks.SourceNone {
		return
	}
	if isSetIntPtr(target) {
		return
	}
	v := *stat.Value
	*target = &v
}

func mergeStringPtr(target **string, stat tracks.StringStat) {
	if stat.Value == nil || stat.Source == tracks.SourceNone {
		return
	}
	if isSetStringPtr(target) {
		return
	}
	v := *stat.Value
	*target = &v
}

func isSetFloatPtr(v **float64) bool {
	return v != nil && *v != nil
}

func isSetIntPtr(v **int) bool {
	return v != nil && *v != nil
}

func isSetStringPtr(v **string) bool {
	return v != nil && *v != nil && **v != ""
}
