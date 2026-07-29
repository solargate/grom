package workouts_test

import (
	"testing"
	"time"

	"github.com/solargate/grom/internal/workouts"
)

func TestHeartRateChartBinaryRoundTripWithDistance(t *testing.T) {
	t0 := time.Date(2026, 7, 14, 8, 0, 1, 0, time.UTC)
	d0 := 0.0
	d1 := 5.2
	samples := []workouts.HeartRateSample{
		{Time: t0, BPM: 120, DistanceM: &d0},
		{Time: t0.Add(time.Second), BPM: 130, DistanceM: &d1},
	}
	data, err := workouts.MarshalHeartRateChartBinary(samples)
	if err != nil {
		t.Fatal(err)
	}
	got, err := workouts.UnmarshalHeartRateChartBinary(data)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 2 || got[0].BPM != 120 || got[1].BPM != 130 {
		t.Fatalf("got %+v", got)
	}
	if got[0].DistanceM == nil || abs(*got[0].DistanceM-0) > 1e-5 {
		t.Fatalf("dist0 = %v", got[0].DistanceM)
	}
	if got[1].DistanceM == nil || abs(*got[1].DistanceM-5.2) > 1e-4 {
		t.Fatalf("dist1 = %v", got[1].DistanceM)
	}
}

func TestHeartRateChartBinaryRoundTripWithoutDistance(t *testing.T) {
	t0 := time.Date(2026, 7, 14, 8, 0, 1, 0, time.UTC)
	samples := []workouts.HeartRateSample{
		{Time: t0, BPM: 110},
		{Time: t0.Add(time.Minute), BPM: 140},
	}
	data, err := workouts.MarshalHeartRateChartBinary(samples)
	if err != nil {
		t.Fatal(err)
	}
	got, err := workouts.UnmarshalHeartRateChartBinary(data)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 2 {
		t.Fatalf("len = %d", len(got))
	}
	for i, s := range got {
		if s.DistanceM != nil {
			t.Fatalf("sample[%d] DistanceM = %v, want nil", i, *s.DistanceM)
		}
	}
}

func TestHeartRateChartBinaryMixedDistance(t *testing.T) {
	t0 := time.Date(2026, 7, 14, 8, 0, 1, 0, time.UTC)
	d := 5.2
	samples := []workouts.HeartRateSample{
		{Time: t0, BPM: 120},
		{Time: t0.Add(time.Second), BPM: 130, DistanceM: &d},
	}
	data, err := workouts.MarshalHeartRateChartBinary(samples)
	if err != nil {
		t.Fatal(err)
	}
	got, err := workouts.UnmarshalHeartRateChartBinary(data)
	if err != nil {
		t.Fatal(err)
	}
	if got[0].DistanceM != nil {
		t.Fatalf("got[0].DistanceM = %v, want nil", got[0].DistanceM)
	}
	if got[1].DistanceM == nil || abs(*got[1].DistanceM-5.2) > 1e-4 {
		t.Fatalf("got[1].DistanceM = %v", got[1].DistanceM)
	}
}

func TestHeartRateChartBinaryRejectsBadMagic(t *testing.T) {
	_, err := workouts.UnmarshalHeartRateChartBinary([]byte("XXXX\x01\x00\x00\x00\x00\x00"))
	if err == nil {
		t.Fatal("expected error")
	}
}
