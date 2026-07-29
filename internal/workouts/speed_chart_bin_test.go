package workouts_test

import (
	"testing"
	"time"

	"github.com/solargate/grom/internal/workouts"
)

func TestSpeedChartBinaryRoundTrip(t *testing.T) {
	t0 := time.Date(2026, 7, 14, 8, 0, 1, 0, time.UTC)
	samples := []workouts.SpeedSample{
		{Time: t0, SpeedKmh: 18.4, DistanceM: 0},
		{Time: t0.Add(time.Second), SpeedKmh: 18.7, DistanceM: 5.2},
	}
	data, err := workouts.MarshalSpeedChartBinary(samples)
	if err != nil {
		t.Fatal(err)
	}
	if len(data) == 0 {
		t.Fatal("expected non-empty payload")
	}
	got, err := workouts.UnmarshalSpeedChartBinary(data)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 2 {
		t.Fatalf("len = %d", len(got))
	}
	if !got[0].Time.Equal(t0) || !got[1].Time.Equal(t0.Add(time.Second)) {
		t.Fatalf("times %+v", got)
	}
	if abs(got[0].SpeedKmh-18.4) > 1e-5 || abs(got[1].SpeedKmh-18.7) > 1e-5 {
		t.Fatalf("speeds %+v", got)
	}
	if abs(got[0].DistanceM-0) > 1e-5 || abs(got[1].DistanceM-5.2) > 1e-4 {
		t.Fatalf("distances %+v", got)
	}
}

func TestSpeedChartBinaryEmpty(t *testing.T) {
	data, err := workouts.MarshalSpeedChartBinary(nil)
	if err != nil || data != nil {
		t.Fatalf("marshal nil: %v %#v", err, data)
	}
	got, err := workouts.UnmarshalSpeedChartBinary(nil)
	if err != nil || got != nil {
		t.Fatalf("unmarshal nil: %v %#v", err, got)
	}
}

func TestSpeedChartBinaryRejectsBadMagic(t *testing.T) {
	_, err := workouts.UnmarshalSpeedChartBinary([]byte("XXXX\x01\x00\x00\x00\x00"))
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestSpeedChartBinaryRejectsTruncated(t *testing.T) {
	t0 := time.Date(2026, 7, 14, 8, 0, 1, 0, time.UTC)
	data, err := workouts.MarshalSpeedChartBinary([]workouts.SpeedSample{
		{Time: t0, SpeedKmh: 10, DistanceM: 1},
	})
	if err != nil {
		t.Fatal(err)
	}
	_, err = workouts.UnmarshalSpeedChartBinary(data[:len(data)-1])
	if err == nil {
		t.Fatal("expected truncated error")
	}
}

func abs(v float64) float64 {
	if v < 0 {
		return -v
	}
	return v
}
