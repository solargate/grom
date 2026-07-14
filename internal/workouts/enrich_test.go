package workouts

import (
	"testing"

	"github.com/solargate/grom/internal/tracks"
)

func floatPtr(v float64) *float64 {
	return &v
}

func TestMergeTrackStatsCreatePrefersTrackSpeed(t *testing.T) {
	workout := &Workout{
		SpeedMaxKmh: floatPtr(10),
		SpeedAvgKmh: floatPtr(8),
	}
	stats := tracks.Stats{}
	setExplicitFloatStat(&stats.SpeedMaxKmh, 32.4)
	setExplicitFloatStat(&stats.SpeedAvgKmh, 17.5)

	data := &tracks.Data{Stats: stats}
	MergeTrackStats(workout, data, MergeModeTrackCreate)

	if *workout.SpeedMaxKmh != 32.4 {
		t.Fatalf("speed_max = %v", *workout.SpeedMaxKmh)
	}
	if *workout.SpeedAvgKmh != 17.5 {
		t.Fatalf("speed_avg = %v", *workout.SpeedAvgKmh)
	}
}

func TestMergeTrackStatsAttachPreservesExisting(t *testing.T) {
	workout := &Workout{
		DurationSeconds: 2184,
		Distance:        10004.7,
		SpeedMaxKmh:     floatPtr(20),
		ElevationGain:   floatPtr(150),
	}
	stats := tracks.Stats{}
	setExplicitIntStat(&stats.DurationSeconds, 999)
	setExplicitFloatStat(&stats.SpeedMaxKmh, 40)
	setExplicitFloatStat(&stats.ElevationGain, 300)
	setExplicitFloatStat(&stats.HeartRateMax, 190)

	data := &tracks.Data{Stats: stats}
	MergeTrackStats(workout, data, MergeModeTrackAttach)

	if workout.DurationSeconds != 2184 {
		t.Fatalf("duration changed to %d", workout.DurationSeconds)
	}
	if *workout.SpeedMaxKmh != 20 {
		t.Fatalf("speed_max changed to %v", *workout.SpeedMaxKmh)
	}
	if *workout.ElevationGain != 150 {
		t.Fatalf("elevation_gain changed to %v", *workout.ElevationGain)
	}
	if workout.HeartRateMax == nil || *workout.HeartRateMax != 190 {
		t.Fatalf("heart_rate_max = %v, want 190", workout.HeartRateMax)
	}
}

func TestMergeTrackStatsCreateDurationPriority(t *testing.T) {
	workout := &Workout{
		DurationSeconds:      1800,
		DurationTotalSeconds: 2000,
	}
	stats := tracks.Stats{}
	setExplicitIntStat(&stats.DurationSeconds, 2041)
	setExplicitIntStat(&stats.DurationTotalSeconds, 3832)

	data := &tracks.Data{Stats: stats}
	MergeTrackStats(workout, data, MergeModeTrackCreate)

	if workout.DurationSeconds != 2041 {
		t.Fatalf("duration_seconds = %d, want 2041", workout.DurationSeconds)
	}
	if workout.DurationTotalSeconds != 3832 {
		t.Fatalf("duration_total_seconds = %d, want 3832", workout.DurationTotalSeconds)
	}
}

func TestMergeTrackStatsTempAvgKmm(t *testing.T) {
	workout := &Workout{
		DurationSeconds: 600,
		Distance:        2000,
	}
	stats := tracks.Stats{}
	setCalculatedStringStat(&stats.TempAvgKmm, "5:00")

	data := &tracks.Data{Stats: stats}
	MergeTrackStats(workout, data, MergeModeTrackCreate)

	if workout.TempAvgKmm == nil || *workout.TempAvgKmm != "5:00" {
		t.Fatalf("temp_avg_kmm = %v", workout.TempAvgKmm)
	}
}

func setExplicitFloatStat(stat *tracks.FloatStat, v float64) {
	stat.Value = &v
	stat.Source = tracks.SourceExplicit
}

func setExplicitIntStat(stat *tracks.IntStat, v int) {
	stat.Value = &v
	stat.Source = tracks.SourceExplicit
}

func setCalculatedStringStat(stat *tracks.StringStat, v string) {
	stat.Value = &v
	stat.Source = tracks.SourceCalculated
}
