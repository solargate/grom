package users

import (
	"cmp"
	"math"
	"slices"
	"time"

	"github.com/solargate/grom/internal/workouts"
	"gopkg.in/yaml.v3"
)

// WorkoutHistory groups summaries of the owner's saved workouts by UTC start day.
type WorkoutHistory map[string][]WorkoutSummary

type WorkoutSummary struct {
	SportType       string  `yaml:"sport_type" json:"sport_type"`
	DurationSeconds int     `yaml:"duration_seconds" json:"duration_seconds"`
	DistanceM       float64 `yaml:"distance_m,omitempty" json:"distance_m,omitempty"`
}

func (h WorkoutHistory) MarshalYAML() (any, error) {
	var node yaml.Node
	// Encode the ordinary map once so the YAML field tags and value types apply.
	if err := node.Encode(map[string][]WorkoutSummary(h)); err != nil {
		return nil, err
	}
	type day struct{ key, entries *yaml.Node }
	days := make([]day, 0, len(h))
	for i := 0; i < len(node.Content); i += 2 {
		days = append(days, day{node.Content[i], node.Content[i+1]})
	}
	slices.SortFunc(days, func(a, b day) int { return cmp.Compare(b.key.Value, a.key.Value) })
	for i, date := range days {
		node.Content[2*i] = date.key
		node.Content[2*i+1] = date.entries
	}
	return &node, nil
}

func (h *WorkoutHistory) Clone() *WorkoutHistory {
	if h == nil {
		return nil
	}
	copy := make(WorkoutHistory, len(*h))
	for day, rows := range *h {
		copy[day] = slices.Clone(rows)
	}
	return &copy
}

func workoutSummary(w *workouts.Workout) (string, WorkoutSummary) {
	if w == nil {
		return "", WorkoutSummary{}
	}
	row := WorkoutSummary{SportType: w.SportType, DurationSeconds: w.DurationSeconds}
	if w.Distance > 0 && !math.IsInf(w.Distance, 1) {
		row.DistanceM = w.Distance
	}
	return w.StartDate.UTC().Format(time.DateOnly), row
}

func compareWorkoutSummaries(a, b WorkoutSummary) int {
	return cmp.Or(cmp.Compare(a.SportType, b.SportType),
		cmp.Compare(a.DurationSeconds, b.DurationSeconds), cmp.Compare(a.DistanceM, b.DistanceM))
}

func (p *Profile) HasWorkoutStats() bool {
	return p.TotalWorkouts != nil && p.Workouts != nil && *p.Workouts != nil
}

func (p *Profile) RebuildWorkoutStats(items []workouts.Workout) {
	history := make(WorkoutHistory)
	for i := range items {
		day, row := workoutSummary(&items[i])
		history[day] = append(history[day], row)
	}
	for _, rows := range history {
		slices.SortFunc(rows, compareWorkoutSummaries)
	}
	total := len(items)
	p.TotalWorkouts = &total
	p.Workouts = &history
}

func SameWorkoutSummary(before, after *workouts.Workout) bool {
	oldDay, oldRow := workoutSummary(before)
	newDay, newRow := workoutSummary(after)
	return oldDay == newDay && oldRow == newRow
}

// ApplyWorkoutChange returns false when the statistics must be rebuilt.
func (p *Profile) ApplyWorkoutChange(before, after *workouts.Workout) bool {
	if !p.HasWorkoutStats() {
		return false
	}
	history := *p.Workouts
	total := *p.TotalWorkouts
	if before != nil {
		day, row := workoutSummary(before)
		rows := history[day]
		i, found := slices.BinarySearchFunc(rows, row, compareWorkoutSummaries)
		if !found || total < 1 {
			return false
		}
		if len(rows) == 1 {
			delete(history, day)
		} else {
			history[day] = slices.Delete(rows, i, i+1)
		}
		total--
	}
	if after != nil {
		day, row := workoutSummary(after)
		rows := history[day]
		i, _ := slices.BinarySearchFunc(rows, row, compareWorkoutSummaries)
		history[day] = slices.Insert(rows, i, row)
		total++
	}
	p.TotalWorkouts = &total
	return true
}
