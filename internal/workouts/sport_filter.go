package workouts

// MatchSportType reports whether sportType passes an optional allow-set.
// A nil allow-set means no filter (always true).
func MatchSportType(sportType string, allow map[string]struct{}) bool {
	if allow == nil {
		return true
	}
	_, ok := allow[sportType]
	return ok
}
