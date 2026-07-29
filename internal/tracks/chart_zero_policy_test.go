package tracks

import (
	"math"
	"testing"
	"time"
)

func TestAcceptSpeedKmhPolicies(t *testing.T) {
	if !AcceptSpeedKmh(0, ChartZeroKeep) {
		t.Fatal("keep should accept 0")
	}
	if AcceptSpeedKmh(0, ChartZeroOmit) {
		t.Fatal("omit should reject 0")
	}
	if AcceptSpeedKmh(math.NaN(), ChartZeroKeep) || AcceptSpeedKmh(math.Inf(1), ChartZeroOmit) {
		t.Fatal("NaN/Inf must be rejected")
	}
	if AcceptSpeedKmh(-1, ChartZeroKeep) {
		t.Fatal("negatives must be rejected")
	}
}

func TestAcceptHeartRateBPMPolicies(t *testing.T) {
	if !AcceptHeartRateBPM(0, ChartZeroKeep) {
		t.Fatal("keep should accept 0")
	}
	if AcceptHeartRateBPM(0, ChartZeroOmit) {
		t.Fatal("omit should reject 0")
	}
}

func TestSpeedSeriesKmhOmitDropsZero(t *testing.T) {
	t0 := time.Date(2026, 7, 14, 8, 0, 0, 0, time.UTC)
	nan := math.NaN()
	points := []SamplePoint{
		{Lat: 10, Lng: 10, Time: t0, HasTime: true, SpeedMps: floatPtr(0)},
		{Lat: 10.001, Lng: 10, Time: t0.Add(time.Second), HasTime: true, SpeedMps: &nan},
		{Lat: 10.002, Lng: 10, Time: t0.Add(2 * time.Second), HasTime: true, SpeedMps: floatPtr(5)},
		{Lat: 10.002, Lng: 10, Time: t0.Add(3 * time.Second), HasTime: true}, // stationary → 0 km/h
	}
	series := speedSeriesKmh(points, ChartZeroOmit)
	if len(series) != 1 || series[0].Kmh != 18 {
		t.Fatalf("got %+v", series)
	}
}

func TestSpeedSeriesKmhKeepIncludesZero(t *testing.T) {
	t0 := time.Date(2026, 7, 14, 8, 0, 0, 0, time.UTC)
	nan := math.NaN()
	points := []SamplePoint{
		{Lat: 10, Lng: 10, Time: t0, HasTime: true, SpeedMps: floatPtr(0)},
		{Lat: 10.001, Lng: 10, Time: t0.Add(time.Second), HasTime: true, SpeedMps: &nan},
		{Lat: 10.002, Lng: 10, Time: t0.Add(2 * time.Second), HasTime: true, SpeedMps: floatPtr(5)},
		{Lat: 10.002, Lng: 10, Time: t0.Add(3 * time.Second), HasTime: true},
	}
	series := speedSeriesKmh(points, ChartZeroKeep)
	if len(series) != 3 {
		t.Fatalf("len = %d, want 3 (zeros kept, NaN dropped)", len(series))
	}
	if series[0].Kmh != 0 || series[1].Kmh != 18 || series[2].Kmh != 0 {
		t.Fatalf("kmh = %v, %v, %v", series[0].Kmh, series[1].Kmh, series[2].Kmh)
	}
}

func TestHeartRateSeriesOmitDropsZero(t *testing.T) {
	t0 := time.Date(2026, 7, 14, 8, 0, 0, 0, time.UTC)
	nan := math.NaN()
	points := []SamplePoint{
		{Time: t0, HasTime: true, HeartRate: floatPtr(0)},
		{Time: t0.Add(time.Second), HasTime: true, HeartRate: &nan},
		{Time: t0.Add(2 * time.Second), HasTime: false, HeartRate: floatPtr(100)},
		{Time: t0.Add(3 * time.Second), HasTime: true},
		{Time: t0.Add(4 * time.Second), HasTime: true, HeartRate: floatPtr(150)},
	}
	series := heartRateSeries(points, false, ChartZeroOmit)
	if len(series) != 1 || series[0].BPM != 150 {
		t.Fatalf("got %+v", series)
	}
}

func TestHeartRateSeriesKeepIncludesZero(t *testing.T) {
	t0 := time.Date(2026, 7, 14, 8, 0, 0, 0, time.UTC)
	nan := math.NaN()
	points := []SamplePoint{
		{Time: t0, HasTime: true, HeartRate: floatPtr(0)},
		{Time: t0.Add(time.Second), HasTime: true, HeartRate: &nan},
		{Time: t0.Add(2 * time.Second), HasTime: false, HeartRate: floatPtr(100)},
		{Time: t0.Add(3 * time.Second), HasTime: true},
		{Time: t0.Add(4 * time.Second), HasTime: true, HeartRate: floatPtr(150)},
	}
	series := heartRateSeries(points, false, ChartZeroKeep)
	if len(series) != 2 || series[0].BPM != 0 || series[1].BPM != 150 {
		t.Fatalf("got %+v", series)
	}
}

func floatPtr(v float64) *float64 { return &v }
