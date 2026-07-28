package tracks_test

import (
	"math"
	"testing"
	"time"

	"github.com/solargate/grom/internal/tracks"
)

func TestHeartRateSeriesWithGPS(t *testing.T) {
	t0 := time.Date(2026, 7, 14, 8, 0, 0, 0, time.UTC)
	points := []tracks.SamplePoint{
		{Lat: 10, Lng: 10, Time: t0, HasTime: true, HeartRate: floatPtr(120)},
		{Lat: 10.001, Lng: 10, Time: t0.Add(10 * time.Second), HasTime: true, HeartRate: floatPtr(130)},
	}
	series := tracks.HeartRateSeries(points, true)
	if len(series) != 2 {
		t.Fatalf("len = %d", len(series))
	}
	if series[0].BPM != 120 || series[1].BPM != 130 {
		t.Fatalf("bpm = %v, %v", series[0].BPM, series[1].BPM)
	}
	if !series[0].HasDistance || series[0].DistanceM != 0 {
		t.Fatalf("series[0] distance = %#v", series[0])
	}
	if !series[1].HasDistance || series[1].DistanceM < 110 || series[1].DistanceM > 112 {
		t.Fatalf("series[1].DistanceM = %v", series[1].DistanceM)
	}
}

func TestHeartRateSeriesWithoutGPSOmitsDistance(t *testing.T) {
	t0 := time.Date(2026, 7, 14, 8, 0, 0, 0, time.UTC)
	d0 := 0.0
	d1 := 50.0
	points := []tracks.SamplePoint{
		{Time: t0, HasTime: true, HeartRate: floatPtr(110), DistanceM: &d0},
		{Time: t0.Add(time.Minute), HasTime: true, HeartRate: floatPtr(140), DistanceM: &d1},
	}
	series := tracks.HeartRateSeries(points, false)
	if len(series) != 2 {
		t.Fatalf("len = %d", len(series))
	}
	if series[0].HasDistance || series[1].HasDistance {
		t.Fatalf("expected no distance, got %#v %#v", series[0], series[1])
	}
	if series[0].DistanceM != 0 || series[1].DistanceM != 0 {
		t.Fatalf("DistanceM should stay 0 without GPS")
	}
}

func TestHeartRateSeriesSkipsZeroNaNUntimed(t *testing.T) {
	t0 := time.Date(2026, 7, 14, 8, 0, 0, 0, time.UTC)
	nan := math.NaN()
	points := []tracks.SamplePoint{
		{Time: t0, HasTime: true, HeartRate: floatPtr(0)},
		{Time: t0.Add(time.Second), HasTime: true, HeartRate: &nan},
		{Time: t0.Add(2 * time.Second), HasTime: false, HeartRate: floatPtr(100)},
		{Time: t0.Add(3 * time.Second), HasTime: true},
		{Time: t0.Add(4 * time.Second), HasTime: true, HeartRate: floatPtr(150)},
	}
	series := tracks.HeartRateSeries(points, false)
	if len(series) != 1 || series[0].BPM != 150 {
		t.Fatalf("got %+v", series)
	}
}

func TestHeartRateSeriesEmpty(t *testing.T) {
	if got := tracks.HeartRateSeries(nil, true); got != nil {
		t.Fatalf("got %v", got)
	}
	if got := tracks.HeartRateSeries([]tracks.SamplePoint{}, false); got != nil {
		t.Fatalf("got %v", got)
	}
}
