package workouts

import "github.com/solargate/grom/internal/tracks"

type MergeMode int

const (
	// MergeModeTrackCreate applies track stats with priority over client-provided values.
	MergeModeTrackCreate MergeMode = iota
	// MergeModeTrackAttach fills only empty workout metrics (preserves CSV/import data).
	MergeModeTrackAttach
)

func MergeTrackStats(workout *Workout, data *tracks.Data, mode MergeMode) {
	if workout == nil || data == nil {
		return
	}

	stats := data.Stats
	distanceMeters := workout.Distance
	if data.DistanceMeters != nil && *data.DistanceMeters > 0 {
		distanceMeters = *data.DistanceMeters
	}
	stats.FinalizePace(&distanceMeters)

	mergeIntStat(&workout.DurationSeconds, stats.DurationSeconds, mode, true)
	mergeIntStat(&workout.DurationTotalSeconds, stats.DurationTotalSeconds, mode, true)

	mergeFloatPtr(&workout.SpeedMaxKmh, stats.SpeedMaxKmh, mode, true)
	mergeFloatPtr(&workout.SpeedAvgKmh, stats.SpeedAvgKmh, mode, true)
	mergeFloatPtr(&workout.ElevationGain, stats.ElevationGain, mode, false)
	mergeFloatPtr(&workout.ElevationLoss, stats.ElevationLoss, mode, false)
	mergeFloatPtr(&workout.ElevationLow, stats.ElevationLow, mode, false)
	mergeFloatPtr(&workout.ElevationHigh, stats.ElevationHigh, mode, false)
	mergeFloatPtr(&workout.GradeMax, stats.GradeMax, mode, false)
	mergeFloatPtr(&workout.GradeAvg, stats.GradeAvg, mode, false)
	mergeRoundedFloatPtr(&workout.CadenceMax, stats.CadenceMax, mode, false)
	mergeRoundedFloatPtr(&workout.CadenceAvg, stats.CadenceAvg, mode, false)
	mergeRoundedFloatPtr(&workout.HeartRateMax, stats.HeartRateMax, mode, false)
	mergeRoundedFloatPtr(&workout.HeartRateAvg, stats.HeartRateAvg, mode, false)
	mergeRoundedFloatPtr(&workout.WattsMax, stats.WattsMax, mode, false)
	mergeRoundedFloatPtr(&workout.WattsAvg, stats.WattsAvg, mode, false)
	mergeRoundedFloatPtr(&workout.Calories, stats.Calories, mode, false)
	mergeFloatPtr(&workout.TemperatureMax, stats.TemperatureMax, mode, false)
	mergeFloatPtr(&workout.TemperatureAvg, stats.TemperatureAvg, mode, false)
	mergeIntPtr(&workout.StepsTotal, stats.StepsTotal, mode, false)
	mergeIntPtr(&workout.CyclesTotal, stats.CyclesTotal, mode, false)
	mergeIntPtr(&workout.SetsTotal, stats.SetsTotal, mode, false)
	mergeIntPtr(&workout.RepsTotal, stats.RepsTotal, mode, false)
	mergeStringPtr(&workout.TempAvgKmm, stats.TempAvgKmm, mode, false)
}

func mergeIntStat(target *int, stat tracks.IntStat, mode MergeMode, allowCalculated bool) {
	if stat.Value == nil || stat.Source == tracks.SourceNone {
		return
	}
	if stat.Source == tracks.SourceCalculated && !allowCalculated {
		return
	}
	if mode == MergeModeTrackAttach && *target > 0 {
		return
	}
	if mode == MergeModeTrackCreate {
		if stat.Source == tracks.SourceExplicit {
			*target = *stat.Value
			return
		}
		if *target <= 0 {
			*target = *stat.Value
		}
		return
	}
	if *target <= 0 {
		*target = *stat.Value
	}
}

func mergeFloatPtr(target **float64, stat tracks.FloatStat, mode MergeMode, trackPreferred bool) {
	if stat.Value == nil || stat.Source == tracks.SourceNone {
		return
	}
	if mode == MergeModeTrackAttach && isSetFloatPtr(target) {
		return
	}
	if mode == MergeModeTrackCreate {
		if trackPreferred || stat.Source == tracks.SourceExplicit || !isSetFloatPtr(target) {
			v := *stat.Value
			*target = &v
		}
		return
	}
	if !isSetFloatPtr(target) {
		v := *stat.Value
		*target = &v
	}
}

func mergeRoundedFloatPtr(target **float64, stat tracks.FloatStat, mode MergeMode, trackPreferred bool) {
	if stat.Value == nil || stat.Source == tracks.SourceNone {
		return
	}
	rounded := float64(int(*stat.Value + 0.5))
	clone := stat
	clone.Value = &rounded
	mergeFloatPtr(target, clone, mode, trackPreferred)
}

func mergeIntPtr(target **int, stat tracks.IntStat, mode MergeMode, trackPreferred bool) {
	if stat.Value == nil || stat.Source == tracks.SourceNone {
		return
	}
	if mode == MergeModeTrackAttach && isSetIntPtr(target) {
		return
	}
	if mode == MergeModeTrackCreate {
		if trackPreferred || stat.Source == tracks.SourceExplicit || !isSetIntPtr(target) {
			v := *stat.Value
			*target = &v
		}
		return
	}
	if !isSetIntPtr(target) {
		v := *stat.Value
		*target = &v
	}
}

func mergeStringPtr(target **string, stat tracks.StringStat, mode MergeMode, trackPreferred bool) {
	if stat.Value == nil || stat.Source == tracks.SourceNone {
		return
	}
	if mode == MergeModeTrackAttach && isSetStringPtr(target) {
		return
	}
	if mode == MergeModeTrackCreate {
		if trackPreferred || stat.Source == tracks.SourceExplicit || !isSetStringPtr(target) {
			v := *stat.Value
			*target = &v
		}
		return
	}
	if !isSetStringPtr(target) {
		v := *stat.Value
		*target = &v
	}
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
