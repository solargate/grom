package workouts_test

import (
	"testing"
	"time"

	"github.com/solargate/grom/internal/workouts"
)

func TestCadenceChartBinaryRoundTripWithDistance(t *testing.T) {
	t0 := time.Date(2026, 7, 14, 8, 0, 1, 0, time.UTC)
	d0 := 0.0
	d1 := 5.2
	samples := []workouts.CadenceSample{
		{Time: t0, Cadence: 70, DistanceM: &d0},
		{Time: t0.Add(time.Second), Cadence: 90, DistanceM: &d1},
	}
	data, err := workouts.MarshalCadenceChartBinary(samples)
	if err != nil {
		t.Fatal(err)
	}
	got, err := workouts.UnmarshalCadenceChartBinary(data)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 2 || got[0].Cadence != 70 || got[1].Cadence != 90 {
		t.Fatalf("got %+v", got)
	}
	if got[0].DistanceM == nil || abs(*got[0].DistanceM-0) > 1e-5 {
		t.Fatalf("dist0 = %v", got[0].DistanceM)
	}
	if got[1].DistanceM == nil || abs(*got[1].DistanceM-5.2) > 1e-4 {
		t.Fatalf("dist1 = %v", got[1].DistanceM)
	}
}

func TestCadenceChartBinaryRoundTripWithoutDistance(t *testing.T) {
	t0 := time.Date(2026, 7, 14, 8, 0, 1, 0, time.UTC)
	samples := []workouts.CadenceSample{
		{Time: t0, Cadence: 70},
		{Time: t0.Add(time.Minute), Cadence: 95},
	}
	data, err := workouts.MarshalCadenceChartBinary(samples)
	if err != nil {
		t.Fatal(err)
	}
	got, err := workouts.UnmarshalCadenceChartBinary(data)
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

func TestCadenceChartBinaryMixedDistance(t *testing.T) {
	t0 := time.Date(2026, 7, 14, 8, 0, 1, 0, time.UTC)
	d := 5.2
	samples := []workouts.CadenceSample{
		{Time: t0, Cadence: 70},
		{Time: t0.Add(time.Second), Cadence: 90, DistanceM: &d},
	}
	data, err := workouts.MarshalCadenceChartBinary(samples)
	if err != nil {
		t.Fatal(err)
	}
	got, err := workouts.UnmarshalCadenceChartBinary(data)
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

func TestCadenceChartBinaryRejectsBadMagic(t *testing.T) {
	_, err := workouts.UnmarshalCadenceChartBinary([]byte("XXXX\x01\x00\x00\x00\x00\x00"))
	if err == nil {
		t.Fatal("expected error")
	}
}
