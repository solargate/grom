package workouts_test

import (
	"math"
	"testing"
	"time"

	"github.com/solargate/grom/internal/tracks"
	"github.com/solargate/grom/internal/workouts"
)

func TestCadenceChartRoundTrip(t *testing.T) {
	t0 := time.Date(2026, 7, 14, 8, 0, 0, 0, time.UTC)
	d := 50.0
	samples := []workouts.CadenceSample{
		{Time: t0, Cadence: 80, DistanceM: &d},
		{Time: t0.Add(time.Minute), Cadence: 90},
	}
	data, err := workouts.MarshalCadenceChart(samples)
	if err != nil {
		t.Fatal(err)
	}
	got, err := workouts.UnmarshalCadenceChart(data)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 2 || got[0].Cadence != 80 || got[1].DistanceM != nil {
		t.Fatalf("got %#v", got)
	}
	if got[0].DistanceM == nil || *got[0].DistanceM != 50 {
		t.Fatalf("distance = %#v", got[0].DistanceM)
	}
}

func TestBuildCadenceChartSamplesKeepsZeroOmitsNaN(t *testing.T) {
	t0 := time.Date(2026, 7, 14, 8, 0, 0, 0, time.UTC)
	parsed := &tracks.Data{
		CadenceSeries: []tracks.CadencePoint{
			{Time: t0, Cadence: 0, HasDistance: true},
			{Time: t0.Add(time.Second), Cadence: math.NaN(), HasDistance: true},
			{Time: t0.Add(2 * time.Second), Cadence: 88, DistanceM: 10, HasDistance: true},
		},
	}
	got := workouts.BuildCadenceChartSamples(parsed)
	if len(got) != 2 || got[0].Cadence != 0 || got[1].Cadence != 88 {
		t.Fatalf("got %#v", got)
	}
}

func TestCadenceChartBinaryRoundTrip(t *testing.T) {
	t0 := time.Date(2026, 7, 14, 8, 0, 0, 0, time.UTC)
	d := 12.5
	samples := []workouts.CadenceSample{
		{Time: t0, Cadence: 70, DistanceM: &d},
		{Time: t0.Add(time.Minute), Cadence: 95, DistanceM: &d},
	}
	data, err := workouts.MarshalCadenceChartBinary(samples)
	if err != nil {
		t.Fatal(err)
	}
	got, err := workouts.UnmarshalCadenceChartBinary(data)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 2 || got[0].Cadence != 70 || got[1].Cadence != 95 {
		t.Fatalf("got %#v", got)
	}
}
