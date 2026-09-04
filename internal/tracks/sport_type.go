package tracks

import (
	"strings"

	"github.com/muktihari/fit/profile/filedef"
	"github.com/muktihari/fit/profile/typedef"
	"github.com/tkrajina/gpxgo/gpx"
)

// MapSportTypeString maps a free-form sport label (GPX type, FIT profile name, etc.)
// to a Grom sport type id. Empty or unknown labels return "".
func MapSportTypeString(raw string) string {
	key := normalizeSportKey(raw)
	if key == "" {
		return ""
	}
	if mapped, ok := sportTypeAliases[key]; ok {
		return mapped
	}
	compact := strings.ReplaceAll(key, " ", "")
	for id := range gromSportTypeIDs {
		if strings.EqualFold(id, raw) || strings.EqualFold(id, compact) {
			return id
		}
	}
	return ""
}

func mapFITSport(sport typedef.Sport, subSport typedef.SubSport) string {
	switch sport {
	case typedef.SportRunning:
		if subSport == typedef.SubSportTrail {
			return "TrailRun"
		}
		return "Run"
	case typedef.SportCycling:
		switch subSport {
		case typedef.SubSportMountain, typedef.SubSportDownhill, typedef.SubSportBmx:
			return "MountainBikeRide"
		case typedef.SubSportGravelCycling, typedef.SubSportCyclocross, typedef.SubSportMixedSurface:
			return "GravelRide"
		case typedef.SubSportHandCycling:
			return "Handcycle"
		case typedef.SubSportEBikeMountain:
			return "EMountainBikeRide"
		default:
			return "Ride"
		}
	case typedef.SportEBiking:
		if subSport == typedef.SubSportEBikeMountain {
			return "EMountainBikeRide"
		}
		return "EBikeRide"
	case typedef.SportWalking:
		return "Walk"
	case typedef.SportHiking, typedef.SportMountaineering:
		return "Hike"
	case typedef.SportSwimming:
		return "Swim"
	case typedef.SportRowing:
		return "Rowing"
	case typedef.SportKayaking, typedef.SportPaddling, typedef.SportRafting:
		return "Kayaking"
	case typedef.SportCanoeing:
		return "Canoeing"
	case typedef.SportStandUpPaddleboarding:
		return "SUP"
	case typedef.SportSurfing:
		return "Surfing"
	case typedef.SportKitesurfing:
		return "Kitesurf"
	case typedef.SportWindsurfing:
		return "Windsurf"
	case typedef.SportSailing:
		return "Sail"
	case typedef.SportCrossCountrySkiing:
		return "NordicSki"
	case typedef.SportAlpineSkiing:
		if subSport == typedef.SubSportBackcountry {
			return "BackcountrySki"
		}
		return "AlpineSki"
	case typedef.SportSnowboarding:
		return "Snowboard"
	case typedef.SportSnowshoeing:
		return "Snowshoe"
	case typedef.SportIceSkating:
		return "IceSkate"
	case typedef.SportInlineSkating:
		return "InlineSkate"
	case typedef.SportGolf:
		return "Golf"
	case typedef.SportBasketball:
		return "Basketball"
	case typedef.SportSoccer:
		return "Soccer"
	case typedef.SportTennis:
		return "Tennis"
	case typedef.SportRockClimbing:
		return "RockClimbing"
	case typedef.SportHiit:
		return "HIIT"
	case typedef.SportDance:
		return "Dance"
	case typedef.SportCricket:
		return "Cricket"
	case typedef.SportHockey:
		return "IceHockey"
	case typedef.SportVolleyball:
		return "Volleyball"
	case typedef.SportWheelchairPushWalk, typedef.SportWheelchairPushRun:
		return "Wheelchair"
	case typedef.SportFitnessEquipment:
		switch subSport {
		case typedef.SubSportElliptical:
			return "Elliptical"
		case typedef.SubSportStairClimbing:
			return "StairStepper"
		case typedef.SubSportPilates:
			return "Pilates"
		case typedef.SubSportYoga:
			return "Yoga"
		case typedef.SubSportStrengthTraining:
			return "WeightTraining"
		case typedef.SubSportHiit, typedef.SubSportAmrap, typedef.SubSportEmom, typedef.SubSportTabata:
			return "HIIT"
		case typedef.SubSportIndoorCycling, typedef.SubSportSpin:
			return "Ride"
		case typedef.SubSportIndoorRowing:
			return "Rowing"
		case typedef.SubSportTreadmill, typedef.SubSportIndoorRunning:
			return "Run"
		case typedef.SubSportIndoorWalking:
			return "Walk"
		default:
			return "Workout"
		}
	case typedef.SportTraining:
		switch subSport {
		case typedef.SubSportStrengthTraining:
			return "WeightTraining"
		case typedef.SubSportYoga, typedef.SubSportFlexibilityTraining:
			return "Yoga"
		case typedef.SubSportPilates:
			return "Pilates"
		case typedef.SubSportHiit, typedef.SubSportAmrap, typedef.SubSportEmom, typedef.SubSportTabata:
			return "HIIT"
		case typedef.SubSportCardioTraining:
			return "Workout"
		default:
			return "Workout"
		}
	case typedef.SportGeneric, typedef.SportTransition, typedef.SportMultisport,
		typedef.SportInvalid, typedef.SportAll:
		return ""
	default:
		return MapSportTypeString(sport.String())
	}
}

func fitActivitySportType(activity *filedef.Activity) string {
	if activity == nil {
		return ""
	}
	for _, session := range activity.Sessions {
		if mapped := mapFITSport(session.Sport, session.SubSport); mapped != "" {
			return mapped
		}
		if name := strings.TrimSpace(session.SportProfileName); name != "" {
			if mapped := MapSportTypeString(name); mapped != "" {
				return mapped
			}
		}
	}
	for _, sport := range activity.Sports {
		if mapped := mapFITSport(sport.Sport, sport.SubSport); mapped != "" {
			return mapped
		}
		if name := strings.TrimSpace(sport.Name); name != "" {
			if mapped := MapSportTypeString(name); mapped != "" {
				return mapped
			}
		}
	}
	return ""
}

func gpxActivitySportType(gpxData *gpx.GPX) string {
	if gpxData == nil {
		return ""
	}
	for _, track := range gpxData.Tracks {
		if mapped := MapSportTypeString(track.Type); mapped != "" {
			return mapped
		}
	}
	return ""
}

func normalizeSportKey(raw string) string {
	raw = strings.ToLower(strings.TrimSpace(raw))
	raw = strings.ReplaceAll(raw, "\u00a0", " ")
	raw = strings.ReplaceAll(raw, "\u202f", " ")
	raw = strings.ReplaceAll(raw, "_", " ")
	raw = strings.ReplaceAll(raw, "-", " ")
	return strings.Join(strings.Fields(raw), " ")
}

var sportTypeAliases = map[string]string{
	"run":               "Run",
	"running":           "Run",
	"ride":              "Ride",
	"cycling":           "Ride",
	"bike":              "Ride",
	"biking":            "Ride",
	"walk":              "Walk",
	"walking":           "Walk",
	"nordic walk":       "NordicWalk",
	"nordic walking":    "NordicWalk",
	"hike":              "Hike",
	"hiking":            "Hike",
	"mountaineering":    "Hike",
	"swim":              "Swim",
	"swimming":          "Swim",
	"workout":           "Workout",
	"training":          "Workout",
	"other":             "Workout",
	"weight training":   "WeightTraining",
	"strength training": "WeightTraining",
	"virtual ride":      "Ride",
	"virtual run":       "Run",
	"ebike":             "EBikeRide",
	"ebiking":           "EBikeRide",
	"e bike":            "EBikeRide",
	"e bike ride":       "EBikeRide",
	"ebikeride":         "EBikeRide",
	"emountainbikeride": "EMountainBikeRide",
	"e mountain bike":   "EMountainBikeRide",
	"mountain bike":     "MountainBikeRide",
	"mountain bike ride": "MountainBikeRide",
	"mtb":               "MountainBikeRide",
	"gravel":            "GravelRide",
	"gravel ride":       "GravelRide",
	"gravel cycling":    "GravelRide",
	"trail run":         "TrailRun",
	"trail running":     "TrailRun",
	"canoeing":          "Canoeing",
	"kayaking":          "Kayaking",
	"paddling":          "Kayaking",
	"rafting":           "Kayaking",
	"packraft":          "Packraft",
	"stand up paddling": "SUP",
	"stand up paddleboarding": "SUP",
	"sup":               "SUP",
	"surfing":           "Surfing",
	"kitesurf":          "Kitesurf",
	"kitesurfing":       "Kitesurf",
	"rowing":            "Rowing",
	"windsurf":          "Windsurf",
	"windsurfing":       "Windsurf",
	"sail":              "Sail",
	"sailing":           "Sail",
	"ice skate":         "IceSkate",
	"ice skating":       "IceSkate",
	"nordic ski":        "NordicSki",
	"cross country skiing": "NordicSki",
	"alpine ski":        "AlpineSki",
	"alpine skiing":     "AlpineSki",
	"snowboard":         "Snowboard",
	"snowboarding":      "Snowboard",
	"backcountry ski":   "BackcountrySki",
	"ice hockey":        "IceHockey",
	"hockey":            "IceHockey",
	"snowshoe":          "Snowshoe",
	"snowshoeing":       "Snowshoe",
	"golf":              "Golf",
	"badminton":         "Badminton",
	"elliptical":        "Elliptical",
	"basketball":        "Basketball",
	"inline skate":      "InlineSkate",
	"inline skating":    "InlineSkate",
	"skateboard":        "Skateboard",
	"tennis":            "Tennis",
	"stair stepper":     "StairStepper",
	"stair climbing":    "StairStepper",
	"padel":             "Padel",
	"rock climbing":     "RockClimbing",
	"soccer":            "Soccer",
	"pickleball":        "Pickleball",
	"volleyball":        "Volleyball",
	"roller ski":        "RollerSki",
	"squash":            "Squash",
	"crossfit":          "Crossfit",
	"yoga":              "Yoga",
	"dance":             "Dance",
	"table tennis":      "TableTennis",
	"pilates":           "Pilates",
	"racquetball":       "Racquetball",
	"hiit":              "HIIT",
	"cricket":           "Cricket",
	"wheelchair":        "Wheelchair",
	"handcycle":         "Handcycle",
	"hand cycling":      "Handcycle",
	"velomobile":        "Velomobile",
}

// Keep in sync with internal/workouts/sport_types.go (no import: avoids cycle via maprender).
var gromSportTypeIDs = map[string]struct{}{
	"Run": {}, "Hike": {}, "TrailRun": {}, "Wheelchair": {}, "Walk": {}, "NordicWalk": {},
	"Ride": {}, "EBikeRide": {}, "MountainBikeRide": {}, "EMountainBikeRide": {},
	"GravelRide": {}, "Velomobile": {}, "Handcycle": {}, "Canoeing": {}, "SUP": {},
	"Kayaking": {}, "Packraft": {}, "Surfing": {}, "Kitesurf": {}, "Swim": {}, "Rowing": {},
	"Windsurf": {}, "Sail": {}, "IceSkate": {}, "NordicSki": {}, "AlpineSki": {},
	"Snowboard": {}, "BackcountrySki": {}, "IceHockey": {}, "Snowshoe": {}, "Workout": {},
	"Golf": {}, "Badminton": {}, "Elliptical": {}, "Basketball": {}, "InlineSkate": {},
	"Skateboard": {}, "Tennis": {}, "StairStepper": {}, "Padel": {}, "RockClimbing": {},
	"Soccer": {}, "Pickleball": {}, "WeightTraining": {}, "Volleyball": {}, "RollerSki": {},
	"Squash": {}, "Crossfit": {}, "Yoga": {}, "Dance": {}, "TableTennis": {}, "Pilates": {},
	"Racquetball": {}, "HIIT": {}, "Cricket": {},
}
