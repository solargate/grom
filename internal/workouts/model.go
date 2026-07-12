package workouts

import "time"

type WorkoutEquipment struct {
	ID   string `yaml:"id" json:"id"`
	Name string `yaml:"name" json:"name"`
	Type string `yaml:"type,omitempty" json:"type,omitempty"`
}

type Workout struct {
	ID              string             `yaml:"id" json:"id"`
	Name            string             `yaml:"name" json:"name"`
	Description     string             `yaml:"description,omitempty" json:"description,omitempty"`
	SportType       string             `yaml:"sport_type" json:"sport_type"`
	StartDate       time.Time          `yaml:"start_date" json:"start_date"`
	Device          string             `yaml:"device,omitempty" json:"device,omitempty"`
	DurationSeconds int                `yaml:"duration_seconds" json:"duration_seconds"`
	Distance        float64            `yaml:"distance" json:"distance"`
	Track           string             `yaml:"track,omitempty" json:"track,omitempty"`
	Equipment       []WorkoutEquipment `yaml:"equipment,omitempty" json:"equipment,omitempty"`
	MediaFiles      []string           `yaml:"media_files,omitempty" json:"media_files,omitempty"`
	HasMapPreview   bool               `yaml:"-" json:"has_map_preview"`
	HasMedia        bool               `yaml:"-" json:"has_media"`
}
