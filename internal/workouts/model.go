package workouts

import "time"

type WorkoutEquipment struct {
	ID   string `yaml:"id" json:"id"`
	Name string `yaml:"name" json:"name"`
	Type string `yaml:"type,omitempty" json:"type,omitempty"`
}

// ExternalID identifies a workout in an external service it was imported from.
type ExternalID struct {
	Name string `yaml:"name" json:"name"`
	ID   string `yaml:"id" json:"id"`
}

type Workout struct {
	ID                   string             `yaml:"id" json:"id"`
	Name                 string             `yaml:"name" json:"name"`
	Description          string             `yaml:"description,omitempty" json:"description,omitempty"`
	SportType            string             `yaml:"sport_type" json:"sport_type"`
	StartDate            time.Time          `yaml:"start_date" json:"start_date"`
	Device               string             `yaml:"device,omitempty" json:"device,omitempty"`
	DurationSeconds      int                `yaml:"duration_seconds" json:"duration_seconds"`
	Distance             float64            `yaml:"distance" json:"distance"`
	DurationTotalSeconds int                `yaml:"duration_total_seconds,omitempty" json:"duration_total_seconds,omitempty"`
	TempAvgKmm           *string            `yaml:"temp_avg_kmm,omitempty" json:"temp_avg_kmm,omitempty"`
	RelativeEffort       *float64           `yaml:"relative_effort,omitempty" json:"relative_effort,omitempty"`
	RegularTrack         *bool              `yaml:"regular_track,omitempty" json:"regular_track,omitempty"`
	SpeedMaxKmh          *float64           `yaml:"speed_max_kmh,omitempty" json:"speed_max_kmh,omitempty"`
	SpeedAvgKmh          *float64           `yaml:"speed_avg_kmh,omitempty" json:"speed_avg_kmh,omitempty"`
	ElevationGain        *float64           `yaml:"elevation_gain,omitempty" json:"elevation_gain,omitempty"`
	ElevationLoss        *float64           `yaml:"elevation_loss,omitempty" json:"elevation_loss,omitempty"`
	ElevationLow         *float64           `yaml:"elevation_low,omitempty" json:"elevation_low,omitempty"`
	ElevationHigh        *float64           `yaml:"elevation_high,omitempty" json:"elevation_high,omitempty"`
	GradeMax             *float64           `yaml:"grade_max,omitempty" json:"grade_max,omitempty"`
	GradeAvg             *float64           `yaml:"grade_avg,omitempty" json:"grade_avg,omitempty"`
	CadenceMax           *float64           `yaml:"cadence_max,omitempty" json:"cadence_max,omitempty"`
	CadenceAvg           *float64           `yaml:"cadence_avg,omitempty" json:"cadence_avg,omitempty"`
	HeartRateMax         *float64           `yaml:"heart_rate_max,omitempty" json:"heart_rate_max,omitempty"`
	HeartRateAvg         *float64           `yaml:"heart_rate_avg,omitempty" json:"heart_rate_avg,omitempty"`
	WattsMax             *float64           `yaml:"watts_max,omitempty" json:"watts_max,omitempty"`
	WattsAvg             *float64           `yaml:"watts_avg,omitempty" json:"watts_avg,omitempty"`
	Calories             *float64           `yaml:"calories,omitempty" json:"calories,omitempty"`
	TemperatureMax       *float64           `yaml:"temperature_max,omitempty" json:"temperature_max,omitempty"`
	TemperatureAvg       *float64           `yaml:"temperature_avg,omitempty" json:"temperature_avg,omitempty"`
	StepsTotal           *int               `yaml:"steps_total,omitempty" json:"steps_total,omitempty"`
	CyclesTotal          *int               `yaml:"cycles_total,omitempty" json:"cycles_total,omitempty"`
	SetsTotal            *int               `yaml:"sets_total,omitempty" json:"sets_total,omitempty"`
	RepsTotal            *int               `yaml:"reps_total,omitempty" json:"reps_total,omitempty"`
	Track                string             `yaml:"track,omitempty" json:"track,omitempty"`
	ExternalID           *ExternalID        `yaml:"external_id,omitempty" json:"external_id,omitempty"`
	Equipment            []WorkoutEquipment `yaml:"equipment,omitempty" json:"equipment,omitempty"`
	MediaFiles           []string           `yaml:"media_files,omitempty" json:"media_files,omitempty"`
	LikesCount           int                `yaml:"likes_count,omitempty" json:"likes_count,omitempty"`
	LikedUsers           []WorkoutLikeUser  `yaml:"liked_users,omitempty" json:"liked_users,omitempty"`
	HasMapPreview        bool               `yaml:"-" json:"has_map_preview"`
	HasMedia             bool               `yaml:"-" json:"has_media"`
}

// SpeedSample is a per-point speed value bound to an absolute UTC timestamp (km/h)
// and cumulative distance from the start of the track (meters).
type SpeedSample struct {
	Time      time.Time
	SpeedKmh  float64
	DistanceM float64
}

// HeartRateSample is a per-point heart-rate value bound to an absolute UTC timestamp (bpm)
// and optional cumulative distance from the start of the track (meters; nil when no GPS).
type HeartRateSample struct {
	Time      time.Time
	BPM       float64
	DistanceM *float64
}
