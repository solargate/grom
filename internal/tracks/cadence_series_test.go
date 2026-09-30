package tracks_test

import (
	"math"
	"testing"
	"time"

	"github.com/solargate/grom/internal/tracks"
)

func TestCadenceSeriesWithGPS(t *testing.T) {
	t0 := time.Date(2026, 7, 14, 8, 0, 0, 0, time.UTC)
	points := []tracks.SamplePoint{
		{Lat: 10, Lng: 10, Time: t0, HasTime: true, Cadence: floatPtr(80)},
		{Lat: 10.001, Lng: 10, Time: t0.Add(10 * time.Second), HasTime: true, Cadence: floatPtr(90)},
	}
	series := tracks.CadenceSeries(points, true)
	if len(series) != 2 {
		t.Fatalf("len = %d", len(series))
	}
	if series[0].Cadence != 80 || series[1].Cadence != 90 {
		t.Fatalf("cadence = %v, %v", series[0].Cadence, series[1].Cadence)
	}
	if !series[0].HasDistance || series[0].DistanceM != 0 {
		t.Fatalf("series[0] distance = %#v", series[0])
	}
	if !series[1].HasDistance || series[1].DistanceM < 110 || series[1].DistanceM > 112 {
		t.Fatalf("series[1].DistanceM = %v", series[1].DistanceM)
	}
}

func TestCadenceSeriesWithoutGPSOmitsDistance(t *testing.T) {
	t0 := time.Date(2026, 7, 14, 8, 0, 0, 0, time.UTC)
	points := []tracks.SamplePoint{
		{Time: t0, HasTime: true, Cadence: floatPtr(70)},
		{Time: t0.Add(time.Minute), HasTime: true, Cadence: floatPtr(85)},
	}
	series := tracks.CadenceSeries(points, false)
	if len(series) != 2 {
		t.Fatalf("len = %d", len(series))
	}
	if series[0].HasDistance || series[1].HasDistance {
		t.Fatalf("expected no distance, got %#v %#v", series[0], series[1])
	}
}

func TestCadenceSeriesSkipsNaNUntimedKeepsZero(t *testing.T) {
	t0 := time.Date(2026, 7, 14, 8, 0, 0, 0, time.UTC)
	nan := math.NaN()
	points := []tracks.SamplePoint{
		{Time: t0, HasTime: true, Cadence: floatPtr(0)},
		{Time: t0.Add(time.Second), HasTime: true, Cadence: &nan},
		{Time: t0.Add(2 * time.Second), HasTime: false, Cadence: floatPtr(100)},
		{Time: t0.Add(3 * time.Second), HasTime: true},
		{Time: t0.Add(4 * time.Second), HasTime: true, Cadence: floatPtr(95)},
	}
	series := tracks.CadenceSeries(points, false)
	if len(series) != 2 || series[0].Cadence != 0 || series[1].Cadence != 95 {
		t.Fatalf("got %+v", series)
	}
}

func TestCadenceSeriesEmpty(t *testing.T) {
	if got := tracks.CadenceSeries(nil, true); got != nil {
		t.Fatalf("got %v", got)
	}
	if got := tracks.CadenceSeries([]tracks.SamplePoint{}, false); got != nil {
		t.Fatalf("got %v", got)
	}
}
