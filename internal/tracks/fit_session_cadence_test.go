package tracks

import (
	"testing"

	"github.com/muktihari/fit/profile/basetype"
	"github.com/muktihari/fit/profile/mesgdef"
)

func TestFITSessionZeroCadenceUsesCalculatedFromRecords(t *testing.T) {
	session := &mesgdef.Session{}
	session.MaxCadence = 0
	session.AvgCadence = 0

	var stats Stats
	extractFITSessionStats(session, &stats)

	if stats.CadenceMax.Source == SourceExplicit {
		t.Fatal("expected zero session MaxCadence to be ignored")
	}
	if stats.CadenceAvg.Source == SourceExplicit {
		t.Fatal("expected zero session AvgCadence to be ignored")
	}

	cad80 := 80.0
	cad100 := 100.0
	cadZero := 0.0
	samples := []SamplePoint{
		{Cadence: &cad80},
		{Cadence: &cad100},
		{Cadence: &cadZero},
	}
	calc := calculateStatsFromSamples(samples)
	mergeCalculatedStats(&stats, &calc)

	if stats.CadenceMax.Value == nil || *stats.CadenceMax.Value != 100 {
		t.Fatalf("cadence_max = %v, want 100", stats.CadenceMax.Value)
	}
	if stats.CadenceAvg.Value == nil || *stats.CadenceAvg.Value != 90 {
		t.Fatalf("cadence_avg = %v, want 90", stats.CadenceAvg.Value)
	}
}

func TestFITSessionPositiveCadencePrefersCalculatedFromRecords(t *testing.T) {
	session := &mesgdef.Session{}
	session.MaxCadence = 95
	session.AvgCadence = 82

	var stats Stats
	extractFITSessionStats(session, &stats)

	if stats.CadenceMax.Source != SourceExplicit || stats.CadenceMax.Value == nil || *stats.CadenceMax.Value != 95 {
		t.Fatalf("cadence_max = %v (source %v), want 95 explicit", stats.CadenceMax.Value, stats.CadenceMax.Source)
	}
	if stats.CadenceAvg.Source != SourceExplicit || stats.CadenceAvg.Value == nil || *stats.CadenceAvg.Value != 82 {
		t.Fatalf("cadence_avg = %v (source %v), want 82 explicit", stats.CadenceAvg.Value, stats.CadenceAvg.Source)
	}

	cad120 := 120.0
	calc := calculateStatsFromSamples([]SamplePoint{{Cadence: &cad120}})
	mergeCalculatedStats(&stats, &calc)
	applyCalculatedCadence(&stats, &calc)

	if stats.CadenceMax.Value == nil || *stats.CadenceMax.Value != 120 {
		t.Fatalf("cadence_max = %v, want 120 from records", stats.CadenceMax.Value)
	}
	if stats.CadenceAvg.Value == nil || *stats.CadenceAvg.Value != 120 {
		t.Fatalf("cadence_avg = %v, want 120 from records", stats.CadenceAvg.Value)
	}
	if stats.CadenceMax.Source != SourceCalculated {
		t.Fatalf("cadence_max source = %v, want calculated", stats.CadenceMax.Source)
	}
}

func TestFITSessionInvalidCadenceUsesCalculated(t *testing.T) {
	session := &mesgdef.Session{}
	session.MaxCadence = basetype.Uint8Invalid
	session.AvgCadence = basetype.Uint8Invalid

	var stats Stats
	extractFITSessionStats(session, &stats)

	cad60 := 60.0
	calc := calculateStatsFromSamples([]SamplePoint{{Cadence: &cad60}})
	mergeCalculatedStats(&stats, &calc)

	if stats.CadenceMax.Value == nil || *stats.CadenceMax.Value != 60 {
		t.Fatalf("cadence_max = %v, want 60", stats.CadenceMax.Value)
	}
	if stats.CadenceAvg.Value == nil || *stats.CadenceAvg.Value != 60 {
		t.Fatalf("cadence_avg = %v, want 60", stats.CadenceAvg.Value)
	}
}
