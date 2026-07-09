package workouts

import "time"

type WorkoutEquipment struct {
	ID   string `yaml:"id" json:"id"`
	Name string `yaml:"name" json:"name"`
}

type Workout struct {
	ID              string             `yaml:"id" json:"id"`
	Name            string             `yaml:"name" json:"name"`
	Description     string             `yaml:"description,omitempty" json:"description,omitempty"`
	SportType       string             `yaml:"sport_type" json:"sport_type"`
	StartDate       time.Time          `yaml:"start_date" json:"start_date"`
	DurationSeconds int                `yaml:"duration_seconds" json:"duration_seconds"`
	Distance        float64            `yaml:"distance" json:"distance"`
	Track           string             `yaml:"track,omitempty" json:"track,omitempty"`
	Equipment       []WorkoutEquipment `yaml:"equipment,omitempty" json:"equipment,omitempty"`
	HasMapPreview   bool               `yaml:"-" json:"has_map_preview"`
}
