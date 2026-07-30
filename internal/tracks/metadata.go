package tracks

import "time"

// ParseTrackMetadata is the API-facing metadata extracted from a track file.
type ParseTrackMetadata struct {
	Name                 string   `json:"name,omitempty"`
	StartDate            string   `json:"start_date,omitempty"`
	Device               string   `json:"device,omitempty"`
	DurationSeconds      int      `json:"duration_seconds,omitempty"`
	DurationTotalSeconds int      `json:"duration_total_seconds,omitempty"`
	Distance             float64  `json:"distance,omitempty"`
	HasGPS               bool     `json:"has_gps"`
	SpeedMaxKmh          *float64 `json:"speed_max_kmh,omitempty"`
	SpeedAvgKmh          *float64 `json:"speed_avg_kmh,omitempty"`
	ElevationGain        *float64 `json:"elevation_gain,omitempty"`
	ElevationLoss        *float64 `json:"elevation_loss,omitempty"`
	ElevationLow         *float64 `json:"elevation_low,omitempty"`
	ElevationHigh        *float64 `json:"elevation_high,omitempty"`
	GradeMax             *float64 `json:"grade_max,omitempty"`
	GradeAvg             *float64 `json:"grade_avg,omitempty"`
	CadenceMax           *float64 `json:"cadence_max,omitempty"`
	CadenceAvg           *float64 `json:"cadence_avg,omitempty"`
	HeartRateMax         *float64 `json:"heart_rate_max,omitempty"`
	HeartRateAvg         *float64 `json:"heart_rate_avg,omitempty"`
	WattsMax             *float64 `json:"watts_max,omitempty"`
	WattsAvg             *float64 `json:"watts_avg,omitempty"`
	Calories             *float64 `json:"calories,omitempty"`
	TemperatureMax       *float64 `json:"temperature_max,omitempty"`
	TemperatureAvg       *float64 `json:"temperature_avg,omitempty"`
	StepsTotal           *int     `json:"steps_total,omitempty"`
	CyclesTotal          *int     `json:"cycles_total,omitempty"`
	SetsTotal            *int     `json:"sets_total,omitempty"`
	RepsTotal            *int     `json:"reps_total,omitempty"`
	TempAvgKmm           *string  `json:"temp_avg_kmm,omitempty"`
}

func (d *Data) Metadata() ParseTrackMetadata {
	if d == nil {
		return ParseTrackMetadata{}
	}
	meta := ParseTrackMetadata{
		Name:   d.Name,
		HasGPS: d.HasGPS(),
	}
	if d.StartTime != nil {
		meta.StartDate = d.StartTime.Format(time.RFC3339)
	}
	if d.Device != nil {
		meta.Device = *d.Device
	}
	if d.DurationSeconds != nil {
		meta.DurationSeconds = *d.DurationSeconds
	}
	if d.DurationTotalSeconds != nil {
		meta.DurationTotalSeconds = *d.DurationTotalSeconds
	} else if d.Stats.DurationTotalSeconds.Value != nil {
		meta.DurationTotalSeconds = *d.Stats.DurationTotalSeconds.Value
	}
	if d.DistanceMeters != nil {
		meta.Distance = *d.DistanceMeters
	}
	meta.SpeedMaxKmh = statFloatPtr(d.Stats.SpeedMaxKmh)
	meta.SpeedAvgKmh = statFloatPtr(d.Stats.SpeedAvgKmh)
	meta.ElevationGain = statFloatPtr(d.Stats.ElevationGain)
	meta.ElevationLoss = statFloatPtr(d.Stats.ElevationLoss)
	meta.ElevationLow = statFloatPtr(d.Stats.ElevationLow)
	meta.ElevationHigh = statFloatPtr(d.Stats.ElevationHigh)
	meta.GradeMax = statFloatPtr(d.Stats.GradeMax)
	meta.GradeAvg = statFloatPtr(d.Stats.GradeAvg)
	meta.CadenceMax = statRoundedFloatPtr(d.Stats.CadenceMax)
	meta.CadenceAvg = statRoundedFloatPtr(d.Stats.CadenceAvg)
	meta.HeartRateMax = statRoundedFloatPtr(d.Stats.HeartRateMax)
	meta.HeartRateAvg = statRoundedFloatPtr(d.Stats.HeartRateAvg)
	meta.WattsMax = statRoundedFloatPtr(d.Stats.WattsMax)
	meta.WattsAvg = statRoundedFloatPtr(d.Stats.WattsAvg)
	meta.Calories = statRoundedFloatPtr(d.Stats.Calories)
	meta.TemperatureMax = statFloatPtr(d.Stats.TemperatureMax)
	meta.TemperatureAvg = statFloatPtr(d.Stats.TemperatureAvg)
	meta.StepsTotal = statIntPtr(d.Stats.StepsTotal)
	meta.CyclesTotal = statIntPtr(d.Stats.CyclesTotal)
	meta.SetsTotal = statIntPtr(d.Stats.SetsTotal)
	meta.RepsTotal = statIntPtr(d.Stats.RepsTotal)
	meta.TempAvgKmm = statStringPtr(d.Stats.TempAvgKmm)
	return meta
}

func statFloatPtr(stat FloatStat) *float64 {
	if stat.Value == nil || stat.Source == SourceNone {
		return nil
	}
	return stat.Value
}

func statRoundedFloatPtr(stat FloatStat) *float64 {
	if stat.Value == nil || stat.Source == SourceNone {
		return nil
	}
	v := float64(int(*stat.Value + 0.5))
	return &v
}

func statIntPtr(stat IntStat) *int {
	if stat.Value == nil || stat.Source == SourceNone {
		return nil
	}
	return stat.Value
}

func statStringPtr(stat StringStat) *string {
	if stat.Value == nil || stat.Source == SourceNone {
		return nil
	}
	return stat.Value
}
