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

// UniqueSportTypes returns the set of non-empty sport types present in items.
func UniqueSportTypes(items []Workout) map[string]struct{} {
	out := make(map[string]struct{})
	for i := range items {
		if s := items[i].SportType; s != "" {
			out[s] = struct{}{}
		}
	}
	return out
}
