package workouts

// NewestSportType returns the sport type of the newest workout by start_date
// DESC, id DESC (FeedNewer). Empty items yield "".
func NewestSportType(items []Workout) string {
	var best *Workout
	for i := range items {
		if best == nil || FeedNewer(items[i].StartDate, items[i].ID, best.StartDate, best.ID) {
			best = &items[i]
		}
	}
	if best == nil {
		return ""
	}
	return best.SportType
}
