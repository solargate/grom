package tracks_test

import (
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
}

func TestSpeedSeriesKmhCalculatedFromDistance(t *testing.T) {
	t0 := time.Date(2026, 7, 14, 8, 0, 0, 0, time.UTC)
	// ~111.19 m north in 10s ≈ 11.119 m/s ≈ 40.03 km/h
	points := []tracks.SamplePoint{
		{Lat: 0, Lng: 0, Time: t0, HasTime: true},
		{Lat: 0.001, Lng: 0, Time: t0.Add(10 * time.Second), HasTime: true},
	}
	series := tracks.SpeedSeriesKmh(points)
	if len(series) != 1 {
		t.Fatalf("len = %d, want 1 (first point has no speed)", len(series))
	}
	if series[0].Kmh < 39 || series[0].Kmh > 41 {
		t.Fatalf("kmh = %v", series[0].Kmh)
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
		{Lat: 0, Lng: 0, Time: t0, HasTime: true, SpeedMps: floatPtr(1)},
		{Lat: 0.001, Lng: 0, Time: t0.Add(200 * time.Second), HasTime: true},
	}
	series := tracks.SpeedSeriesKmh(points)
	if len(series) != 2 {
		t.Fatalf("len = %d, want gap sample included", len(series))
	}
}

func TestSpeedSeriesKmhEmpty(t *testing.T) {
	if got := tracks.SpeedSeriesKmh(nil); got != nil {
		t.Fatalf("got %v", got)
	}
}

func floatPtr(v float64) *float64 { return &v }
