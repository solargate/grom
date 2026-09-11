package workouts

import (
	"testing"

	"github.com/solargate/grom/internal/tracks"
)

func floatPtr(v float64) *float64 {
	return &v
}

func TestMergeTrackStatsCreatePreservesClientSpeed(t *testing.T) {
	workout := &Workout{
		SpeedMaxKmh: floatPtr(10),
		SpeedAvgKmh: floatPtr(8),
	}
	stats := tracks.Stats{}
	setExplicitFloatStat(&stats.SpeedMaxKmh, 32.4)
	setExplicitFloatStat(&stats.SpeedAvgKmh, 17.5)

	data := &tracks.Data{Stats: stats}
	MergeTrackStats(workout, data, MergeModeTrackCreate)

	if *workout.SpeedMaxKmh != 10 {
		t.Fatalf("speed_max = %v, want client 10", *workout.SpeedMaxKmh)
	}
	if *workout.SpeedAvgKmh != 8 {
		t.Fatalf("speed_avg = %v, want client 8", *workout.SpeedAvgKmh)
	}
}

func TestMergeTrackStatsCreateFillsEmptySpeed(t *testing.T) {
	workout := &Workout{}
	stats := tracks.Stats{}
	setExplicitFloatStat(&stats.SpeedMaxKmh, 32.4)
	setExplicitFloatStat(&stats.SpeedAvgKmh, 17.5)

	data := &tracks.Data{Stats: stats}
	MergeTrackStats(workout, data, MergeModeTrackCreate)

	if workout.SpeedMaxKmh == nil || *workout.SpeedMaxKmh != 32.4 {
		t.Fatalf("speed_max = %v, want 32.4", workout.SpeedMaxKmh)
	}
	if workout.SpeedAvgKmh == nil || *workout.SpeedAvgKmh != 17.5 {
		t.Fatalf("speed_avg = %v, want 17.5", workout.SpeedAvgKmh)
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

func TestMergeTrackStatsCreatePreservesClientDuration(t *testing.T) {
	workout := &Workout{
		DurationSeconds:      1800,
		DurationTotalSeconds: 2000,
	}
	stats := tracks.Stats{}
	setExplicitIntStat(&stats.DurationSeconds, 2041)
	setExplicitIntStat(&stats.DurationTotalSeconds, 3832)

	data := &tracks.Data{Stats: stats}
	MergeTrackStats(workout, data, MergeModeTrackCreate)

	if workout.DurationSeconds != 1800 {
		t.Fatalf("duration_seconds = %d, want client 1800", workout.DurationSeconds)
	}
	if workout.DurationTotalSeconds != 2000 {
		t.Fatalf("duration_total_seconds = %d, want client 2000", workout.DurationTotalSeconds)
	}
}

func TestMergeTrackStatsCreatePreservesClientElevation(t *testing.T) {
	workout := &Workout{
		ElevationGain: floatPtr(516),
		ElevationLow:  floatPtr(10),
		ElevationHigh: floatPtr(200),
	}
	stats := tracks.Stats{}
	setExplicitFloatStat(&stats.ElevationGain, 100)
	setExplicitFloatStat(&stats.ElevationLow, 1)
	setExplicitFloatStat(&stats.ElevationHigh, 50)

	data := &tracks.Data{Stats: stats}
	MergeTrackStats(workout, data, MergeModeTrackCreate)

	if *workout.ElevationGain != 516 {
		t.Fatalf("elevation_gain = %v, want 516", *workout.ElevationGain)
	}
	if *workout.ElevationLow != 10 {
		t.Fatalf("elevation_low = %v, want 10", *workout.ElevationLow)
	}
	if *workout.ElevationHigh != 200 {
		t.Fatalf("elevation_high = %v, want 200", *workout.ElevationHigh)
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
