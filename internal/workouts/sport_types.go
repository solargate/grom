package workouts

import "slices"

var validSportTypes = map[string]struct{}{
	"Run":               {},
	"Hike":              {},
	"TrailRun":          {},
	"Wheelchair":        {},
	"Walk":              {},
	"Ride":              {},
	"EBikeRide":         {},
	"MountainBikeRide":  {},
	"EMountainBikeRide": {},
	"GravelRide":        {},
	"Velomobile":        {},
	"Handcycle":         {},
	"Canoeing":          {},
	"SUP":               {},
	"Kayaking":          {},
	"Surfing":           {},
	"Kitesurf":          {},
	"Swim":              {},
	"Rowing":            {},
	"Windsurf":          {},
	"Sail":              {},
	"IceSkate":          {},
	"NordicSki":         {},
	"AlpineSki":         {},
	"Snowboard":         {},
	"BackcountrySki":    {},
	"Snowshoe":          {},
	"Workout":           {},
	"Golf":              {},
	"Badminton":         {},
	"Elliptical":        {},
	"Basketball":        {},
	"InlineSkate":       {},
	"Skateboard":        {},
	"Tennis":            {},
	"StairStepper":      {},
	"Padel":             {},
	"RockClimbing":      {},
	"Soccer":            {},
	"Pickleball":        {},
	"WeightTraining":    {},
	"Volleyball":        {},
	"RollerSki":         {},
	"Squash":            {},
	"Crossfit":          {},
	"Yoga":              {},
	"Dance":             {},
	"TableTennis":       {},
	"Pilates":           {},
	"Racquetball":       {},
	"HIIT":              {},
	"Cricket":           {},
}

func IsValidSportType(sportType string) bool {
	_, ok := validSportTypes[sportType]
	return ok
}

func ValidSportTypes() []string {
	types := make([]string, 0, len(validSportTypes))
	for sportType := range validSportTypes {
		types = append(types, sportType)
	}
	slices.Sort(types)
	return types
}
