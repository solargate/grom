package users

import (
	"reflect"
	"testing"
	"time"

	"github.com/solargate/grom/internal/workouts"
	"gopkg.in/yaml.v3"
)

func TestWorkoutHistoryYAML(t *testing.T) {
	profile := Profile{LastSportType: "Run"}
	start := time.Date(2026, 11, 27, 0, 0, 0, 0, time.UTC)
	profile.RebuildWorkoutStats([]workouts.Workout{
		{StartDate: start, SportType: "WeightTraining", DurationSeconds: 1800},
		{StartDate: start.AddDate(0, 0, -2), SportType: "Ride", DurationSeconds: 3600, Distance: 24500},
		{StartDate: start, SportType: "Run", DurationSeconds: 2400, Distance: 7200},
	})
	want := `total_workouts: 3
last_sport_type: Run
workouts:
    "2026-11-27":
        - sport_type: Run
          duration_seconds: 2400
          distance_m: 7200
        - sport_type: WeightTraining
          duration_seconds: 1800
    "2026-11-25":
        - sport_type: Ride
          duration_seconds: 3600
          distance_m: 24500
`
	raw, err := yaml.Marshal(profile)
	if err != nil {
		t.Fatal(err)
	}
	if string(raw) != want {
		t.Fatalf("unexpected profile YAML:\n%s", raw)
	}
	var decoded Profile
	if err := yaml.Unmarshal(raw, &decoded); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(decoded, profile) {
		t.Fatalf("decoded profile = %+v, want %+v", decoded, profile)
	}
}
