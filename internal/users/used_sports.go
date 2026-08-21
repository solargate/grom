package users

// MoveSportToFront returns list with sport at index 0. Empty sport is a no-op.
// If sport was already present it is moved, not duplicated.
func MoveSportToFront(list []string, sport string) []string {
	if sport == "" {
		return cloneStringSlice(list)
	}
	out := make([]string, 0, len(list)+1)
	out = append(out, sport)
	for _, s := range list {
		if s != sport && s != "" {
			out = append(out, s)
		}
	}
	return out
}

// PruneUsedSports keeps only types present in remaining, preserving order.
func PruneUsedSports(list []string, remaining map[string]struct{}) []string {
	if len(list) == 0 {
		return nil
	}
	out := make([]string, 0, len(list))
	for _, s := range list {
		if s == "" {
			continue
		}
		if _, ok := remaining[s]; ok {
			out = append(out, s)
		}
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

func cloneStringSlice(in []string) []string {
	if len(in) == 0 {
		return nil
	}
	out := make([]string, len(in))
	copy(out, in)
	return out
}
