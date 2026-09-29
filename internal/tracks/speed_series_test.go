package tracks_test

import (
	"math"
	"testing"
	"time"

	"github.com/solargate/grom/internal/tracks"
)

func TestSpeedSeriesKmhExplicit(t *testing.T) {
	t0 := time.Date(2026, 7, 14, 8, 0, 0, 0, time.UTC)
	mps := 5.0 // 18 km/h
	points := []tracks.SamplePoint{
		{Time: t0, HasTime: true, SpeedMps: &mps},
		{Time: t0.Add(time.Second), HasTime: true, SpeedMps: floatPtr(5.5)},
	}
	series := tracks.SpeedSeriesKmh(points)
	if len(series) != 2 {
		t.Fatalf("len = %d", len(series))
	}
	if series[0].Kmh != 18 {
		t.Fatalf("series[0] = %v", series[0].Kmh)
	}
	if series[1].Kmh != 19.8 {
		t.Fatalf("series[1] = %v", series[1].Kmh)
	}
	if !series[0].Time.Equal(t0) {
		t.Fatalf("time = %v", series[0].Time)
	}
	if series[0].DistanceM != 0 {
		t.Fatalf("series[0].DistanceM = %v, want 0", series[0].DistanceM)
	}
}

func TestSpeedSeriesKmhCalculatedFromDistance(t *testing.T) {
	t0 := time.Date(2026, 7, 14, 8, 0, 0, 0, time.UTC)
	// ~111.19 m north in 10s ≈ 11.119 m/s ≈ 40.03 km/h
	points := []tracks.SamplePoint{
		{Lat: 10, Lng: 10, Time: t0, HasTime: true},
		{Lat: 10.001, Lng: 10, Time: t0.Add(10 * time.Second), HasTime: true},
	}
	series := tracks.SpeedSeriesKmh(points)
	if len(series) != 1 {
		t.Fatalf("len = %d, want 1 (first point has no speed)", len(series))
	}
	if series[0].Kmh < 39 || series[0].Kmh > 41 {
		t.Fatalf("kmh = %v", series[0].Kmh)
	}
	if series[0].DistanceM < 110 || series[0].DistanceM > 112 {
		t.Fatalf("DistanceM = %v, want ~111", series[0].DistanceM)
	}
}

func TestSpeedSeriesKmhSkipsUntimed(t *testing.T) {
	t0 := time.Date(2026, 7, 14, 8, 0, 0, 0, time.UTC)
	mps := 2.0
	points := []tracks.SamplePoint{
		{Time: t0, HasTime: false, SpeedMps: &mps},
		{Time: t0.Add(time.Second), HasTime: true, SpeedMps: &mps},
	}
	series := tracks.SpeedSeriesKmh(points)
	if len(series) != 1 {
		t.Fatalf("len = %d", len(series))
	}
}

func TestSpeedSeriesKmhIncludesAfterLongGap(t *testing.T) {
	t0 := time.Date(2026, 7, 14, 8, 0, 0, 0, time.UTC)
	points := []tracks.SamplePoint{
		{Lat: 10, Lng: 10, Time: t0, HasTime: true, SpeedMps: floatPtr(1)},
		{Lat: 10.001, Lng: 10, Time: t0.Add(200 * time.Second), HasTime: true},
	}
	series := tracks.SpeedSeriesKmh(points)
	if len(series) != 2 {
		t.Fatalf("len = %d, want gap sample included", len(series))
	}
	if series[1].DistanceM < 110 || series[1].DistanceM > 112 {
		t.Fatalf("series[1].DistanceM = %v", series[1].DistanceM)
	}
}

func TestSpeedSeriesKmhAccumulatesThroughUntimed(t *testing.T) {
	t0 := time.Date(2026, 7, 14, 8, 0, 0, 0, time.UTC)
	points := []tracks.SamplePoint{
		{Lat: 10, Lng: 10, Time: t0, HasTime: true, SpeedMps: floatPtr(1)},
		{Lat: 10.001, Lng: 10, HasTime: false}, // untimed, still on path
		{Lat: 10.002, Lng: 10, Time: t0.Add(20 * time.Second), HasTime: true, SpeedMps: floatPtr(1)},
	}
	series := tracks.SpeedSeriesKmh(points)
	if len(series) != 2 {
		t.Fatalf("len = %d", len(series))
	}
	if series[0].DistanceM != 0 {
		t.Fatalf("series[0].DistanceM = %v", series[0].DistanceM)
	}
	// ~222 m for 0.002 deg latitude
	if series[1].DistanceM < 220 || series[1].DistanceM > 225 {
		t.Fatalf("series[1].DistanceM = %v, want ~222", series[1].DistanceM)
	}
}

func TestSpeedSeriesKmhPrefersDeviceDistance(t *testing.T) {
	t0 := time.Date(2026, 7, 14, 8, 0, 0, 0, time.UTC)
	d0 := 0.0
	d1 := 50.0
	points := []tracks.SamplePoint{
		{Lat: 10, Lng: 10, Time: t0, HasTime: true, SpeedMps: floatPtr(1), DistanceM: &d0},
		{Lat: 10.001, Lng: 10, Time: t0.Add(time.Second), HasTime: true, SpeedMps: floatPtr(1), DistanceM: &d1},
	}
	series := tracks.SpeedSeriesKmh(points)
	if len(series) != 2 {
		t.Fatalf("len = %d", len(series))
	}
	if series[1].DistanceM != 50 {
		t.Fatalf("DistanceM = %v, want device 50 (not haversine)", series[1].DistanceM)
	}
}

func TestSpeedSeriesKmhDeviceDistanceWithoutGPS(t *testing.T) {
	t0 := time.Date(2026, 7, 14, 8, 0, 0, 0, time.UTC)
	d0 := 0.0
	d1 := 120.5
	points := []tracks.SamplePoint{
		{Time: t0, HasTime: true, SpeedMps: floatPtr(2), DistanceM: &d0},
		{Time: t0.Add(time.Second), HasTime: true, SpeedMps: floatPtr(2), DistanceM: &d1},
	}
	series := tracks.SpeedSeriesKmh(points)
	if len(series) != 2 {
		t.Fatalf("len = %d", len(series))
	}
	if series[0].DistanceM != 0 || series[1].DistanceM != 120.5 {
		t.Fatalf("got distances %v, %v", series[0].DistanceM, series[1].DistanceM)
	}
}

func TestSpeedSeriesKmhSkipsNaNKeepsZero(t *testing.T) {
	t0 := time.Date(2026, 7, 14, 8, 0, 0, 0, time.UTC)
	nan := math.NaN()
	points := []tracks.SamplePoint{
		{Lat: 10, Lng: 10, Time: t0, HasTime: true, SpeedMps: floatPtr(0)},
		{Lat: 10.001, Lng: 10, Time: t0.Add(time.Second), HasTime: true, SpeedMps: &nan},
		{Lat: 10.002, Lng: 10, Time: t0.Add(2 * time.Second), HasTime: true, SpeedMps: floatPtr(5)},
		{Lat: 10.002, Lng: 10, Time: t0.Add(3 * time.Second), HasTime: true}, // stationary → 0 km/h
	}
	series := tracks.SpeedSeriesKmh(points)
	if len(series) != 3 {
		t.Fatalf("len = %d, want 3 (zeros kept per SpeedChartZeroPolicy)", len(series))
	}
	if series[0].Kmh != 0 || series[1].Kmh != 18 || series[2].Kmh != 0 {
		t.Fatalf("kmh = %v, %v, %v", series[0].Kmh, series[1].Kmh, series[2].Kmh)
	}
}

func TestSpeedSeriesKmhEmpty(t *testing.T) {
	if got := tracks.SpeedSeriesKmh(nil); got != nil {
		t.Fatalf("got %v", got)
	}
}

func TestSpeedSeriesKmhOmitsTimedSamplesWithoutMotionData(t *testing.T) {
	t0 := time.Date(2026, 7, 14, 8, 0, 0, 0, time.UTC)
	hr := 120.0
	points := []tracks.SamplePoint{
		{Time: t0, HasTime: true, HeartRate: &hr},
		{Time: t0.Add(30 * time.Second), HasTime: true, HeartRate: floatPtr(130)},
		{Time: t0.Add(60 * time.Second), HasTime: true, HeartRate: floatPtr(140)},
	}
	if got := tracks.SpeedSeriesKmh(points); got != nil {
		t.Fatalf("got %v, want nil for HR-only samples", got)
	}
}

func TestSpeedSeriesKmhDerivedFromDeviceDistanceWithoutGPS(t *testing.T) {
	t0 := time.Date(2026, 7, 14, 8, 0, 0, 0, time.UTC)
	d0 := 0.0
	d1 := 100.0
	points := []tracks.SamplePoint{
		{Time: t0, HasTime: true, DistanceM: &d0},
		{Time: t0.Add(10 * time.Second), HasTime: true, DistanceM: &d1},
	}
	series := tracks.SpeedSeriesKmh(points)
	if len(series) != 1 {
		t.Fatalf("len = %d, want 1", len(series))
	}
	// 100 m / 10 s = 10 m/s = 36 km/h
	if series[0].Kmh < 35.9 || series[0].Kmh > 36.1 {
		t.Fatalf("kmh = %v, want 36", series[0].Kmh)
	}
}

func TestSpeedSeriesKmhKeepsExplicitZeroBetweenNonZero(t *testing.T) {
	t0 := time.Date(2026, 7, 14, 8, 0, 0, 0, time.UTC)
	points := []tracks.SamplePoint{
		{Time: t0, HasTime: true, SpeedMps: floatPtr(2)},
		{Time: t0.Add(time.Second), HasTime: true, SpeedMps: floatPtr(0)},
		{Time: t0.Add(2 * time.Second), HasTime: true, SpeedMps: floatPtr(3)},
	}
	series := tracks.SpeedSeriesKmh(points)
	if len(series) != 3 {
		t.Fatalf("len = %d, want 3", len(series))
	}
	if series[0].Kmh != 7.2 || series[1].Kmh != 0 || series[2].Kmh != 10.8 {
		t.Fatalf("kmh = %v, %v, %v", series[0].Kmh, series[1].Kmh, series[2].Kmh)
	}
}

func floatPtr(v float64) *float64 { return &v }
