package tracks

import (
	"fmt"
	"math"
)

type Source int

const (
	SourceNone Source = iota
	SourceExplicit
	SourceCalculated
)

type IntStat struct {
	Value  *int
	Source Source
}

type FloatStat struct {
	Value  *float64
	Source Source
}

type StringStat struct {
	Value  *string
	Source Source
}

type Stats struct {
	DurationSeconds      IntStat
	DurationTotalSeconds IntStat
	SpeedMaxKmh          FloatStat
	SpeedAvgKmh          FloatStat
	ElevationGain        FloatStat
	ElevationLoss        FloatStat
	ElevationLow         FloatStat
	ElevationHigh        FloatStat
	GradeMax             FloatStat
	GradeAvg             FloatStat
	CadenceMax           FloatStat
	CadenceAvg           FloatStat
	HeartRateMax         FloatStat
	HeartRateAvg         FloatStat
	WattsMax             FloatStat
	WattsAvg             FloatStat
	Calories             FloatStat
	TemperatureMax       FloatStat
	TemperatureAvg       FloatStat
	StepsTotal           IntStat
	CyclesTotal          IntStat
	SetsTotal            IntStat
	RepsTotal            IntStat
	TempAvgKmm           StringStat
}

func (s *IntStat) setExplicit(v int) {
	if v < 0 {
		return
	}
	s.Value = &v
	s.Source = SourceExplicit
}

func (s *IntStat) setCalculated(v int) {
	if v < 0 {
		return
	}
	if s.Source == SourceExplicit {
		return
	}
	s.Value = &v
	s.Source = SourceCalculated
}

func (s *FloatStat) setExplicit(v float64) {
	if !validFloat(v) {
		return
	}
	s.Value = &v
	s.Source = SourceExplicit
}

func (s *FloatStat) setCalculated(v float64) {
	if !validFloat(v) {
		return
	}
	if s.Source == SourceExplicit {
		return
	}
	s.Value = &v
	s.Source = SourceCalculated
}

func (s *FloatStat) setCalculatedOverride(v float64) {
	if !validFloat(v) {
		return
	}
	s.Value = &v
	s.Source = SourceCalculated
}

func (s *StringStat) setExplicit(v string) {
	if v == "" {
		return
	}
	s.Value = &v
	s.Source = SourceExplicit
}

func (s *StringStat) setCalculated(v string) {
	if v == "" {
		return
	}
	if s.Source == SourceExplicit {
		return
	}
	s.Value = &v
	s.Source = SourceCalculated
}

func validFloat(v float64) bool {
	return !math.IsNaN(v) && !math.IsInf(v, 0) && v >= 0
}

func mpsToKmh(mps float64) float64 {
	return mps * 3.6
}

func roundFloat(v float64) float64 {
	return math.Round(v*100) / 100
}

func FormatPaceMinPerKm(durationSeconds int, distanceMeters float64) *string {
	if durationSeconds <= 0 || distanceMeters <= 0 {
		return nil
	}
	paceSec := float64(durationSeconds) / (distanceMeters / 1000)
	if paceSec <= 0 || math.IsInf(paceSec, 0) || math.IsNaN(paceSec) {
		return nil
	}
	mins := int(paceSec) / 60
	secs := int(paceSec) % 60
	formatted := fmt.Sprintf("%d:%02d", mins, secs)
	return &formatted
}

func (s Stats) PaceDurationSeconds() int {
	if s.DurationSeconds.Source != SourceNone && s.DurationSeconds.Value != nil {
		return *s.DurationSeconds.Value
	}
	if s.DurationTotalSeconds.Source != SourceNone && s.DurationTotalSeconds.Value != nil {
		return *s.DurationTotalSeconds.Value
	}
	return 0
}

func (s Stats) PaceDistanceMeters(distanceMeters *float64) float64 {
	if distanceMeters != nil && *distanceMeters > 0 {
		return *distanceMeters
	}
	return 0
}

func (s *Stats) FinalizePace(distanceMeters *float64) {
	if s.TempAvgKmm.Source != SourceNone {
		return
	}
	dur := s.PaceDurationSeconds()
	dist := s.PaceDistanceMeters(distanceMeters)
	if pace := FormatPaceMinPerKm(dur, dist); pace != nil {
		s.TempAvgKmm.setCalculated(*pace)
	}
}
