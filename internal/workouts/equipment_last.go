package workouts

// LastEquipmentIDsForSport returns equipment IDs from the most recent workout
// with the given sport type (start_date DESC, id DESC). Missing or empty
// equipment yields a nil/empty slice.
func LastEquipmentIDsForSport(items []Workout, sportType string) []string {
	if sportType == "" {
		return nil
	}
	var best *Workout
	for i := range items {
		if items[i].SportType != sportType {
			continue
		}
		if best == nil || FeedNewer(items[i].StartDate, items[i].ID, best.StartDate, best.ID) {
			best = &items[i]
		}
	}
	if best == nil {
		return nil
	}
	if len(best.Equipment) == 0 {
		return nil
	}
	ids := make([]string, 0, len(best.Equipment))
	for _, item := range best.Equipment {
		if item.ID == "" {
			continue
		}
		ids = append(ids, item.ID)
	}
	return ids
}
