package strava

import (
	"strings"

	"github.com/solargate/grom/internal/workouts"
)

var sportTypeAliases = map[string]string{
	// English
	"run": "Run",
	"ride": "Ride",
	"walk": "Walk",
	"nordic walk": "NordicWalk",
	"nordic walking": "NordicWalk",
	"hike": "Hike",
	"swim": "Swim",
	"workout": "Workout",
	"weight training": "WeightTraining",
	"virtual ride": "Ride",
	"virtual run": "Run",
	"ebikeride": "EBikeRide",
	"e-bike ride": "EBikeRide",
	"emountainbikeride": "EMountainBikeRide",
	"mountain bike ride": "MountainBikeRide",
	"gravel ride": "GravelRide",
	"trail run": "TrailRun",
	"canoeing": "Canoeing",
	"kayaking": "Kayaking",
	"packraft": "Packraft",
	"stand up paddling": "SUP",
	"stand-up paddling": "SUP",
	"surfing": "Surfing",
	"kitesurf": "Kitesurf",
	"rowing": "Rowing",
	"windsurf": "Windsurf",
	"sail": "Sail",
	"ice skate": "IceSkate",
	"nordic ski": "NordicSki",
	"alpine ski": "AlpineSki",
	"snowboard": "Snowboard",
	"backcountry ski": "BackcountrySki",
	"ice hockey": "IceHockey",
	"hockey": "IceHockey",
	"snowshoe": "Snowshoe",
	"golf": "Golf",
	"badminton": "Badminton",
	"elliptical": "Elliptical",
	"basketball": "Basketball",
	"inline skate": "InlineSkate",
	"skateboard": "Skateboard",
	"tennis": "Tennis",
	"stair stepper": "StairStepper",
	"padel": "Padel",
	"rock climbing": "RockClimbing",
	"soccer": "Soccer",
	"pickleball": "Pickleball",
	"volleyball": "Volleyball",
	"roller ski": "RollerSki",
	"squash": "Squash",
	"crossfit": "Crossfit",
	"yoga": "Yoga",
	"dance": "Dance",
	"table tennis": "TableTennis",
	"pilates": "Pilates",
	"racquetball": "Racquetball",
	"hiit": "HIIT",
	"high intensity interval training": "HIIT",
	"highintensityintervaltraining": "HIIT",
	"cricket": "Cricket",
	"wheelchair": "Wheelchair",
	"handcycle": "Handcycle",
	"velomobile": "Velomobile",
	"water sport": "Workout",

	// Russian
	"бег": "Run",
	"велосипед": "Ride",
	"ходьба": "Walk",
	"скандинавская ходьба": "NordicWalk",
	"поход": "Hike",
	"плавание": "Swim",
	"тренировка": "Workout",
	"силовая тренировка": "WeightTraining",
	"виртуальный заезд": "Ride",
	"виртуальный бег": "Run",
	"заезд на электровелосипеде": "EBikeRide",
	"горный велозаезд": "MountainBikeRide",
	"гравийный велозаезд": "GravelRide",
	"трейлран": "TrailRun",
	"канoeing": "Canoeing",
	"каякинг": "Kayaking",
	"пакрафт": "Packraft",
	"гребля на сапе": "SUP",
	"серфинг": "Surfing",
	"кайтсерфинг": "Kitesurf",
	"гребля": "Rowing",
	"виндсерфинг": "Windsurf",
	"парусный спорт": "Sail",
	"катание на коньках": "IceSkate",
	"лыжи классические": "NordicSki",
	"горные лыжи": "AlpineSki",
	"сноуборд": "Snowboard",
	"скитур": "BackcountrySki",
	"хоккей": "IceHockey",
	"снегоступы": "Snowshoe",
	"гольф": "Golf",
	"бадминтон": "Badminton",
	"эллипс": "Elliptical",
	"баскетбол": "Basketball",
	"ролики": "InlineSkate",
	"скейтборд": "Skateboard",
	"теннис": "Tennis",
	"степпер": "StairStepper",
	"скалолазание": "RockClimbing",
	"футбол": "Soccer",
	"волейбол": "Volleyball",
	"сквош": "Squash",
	"кроссфит": "Crossfit",
	"йога": "Yoga",
	"танцы": "Dance",
	"настольный теннис": "TableTennis",
	"пилатес": "Pilates",
	"крикет": "Cricket",
	"инвалидная коляска": "Wheelchair",
	"ручной велосипед": "Handcycle",
	"веломобиль": "Velomobile",
	"силовая": "WeightTraining",
}

var nameSportKeywords = []struct {
	keyword string
	sport   string
}{
	{"пакрафт", "Packraft"},
	{"packraft", "Packraft"},
	{"пилатес", "Pilates"},
	{"pilates", "Pilates"},
	{"йога", "Yoga"},
	{"yoga", "Yoga"},
	{"кроссфит", "Crossfit"},
	{"crossfit", "Crossfit"},
	{"hiit", "HIIT"},
	{"растяжка", "Yoga"},
	{"stretching", "Yoga"},
	{"танцы", "Dance"},
	{"dance", "Dance"},
	{"эллипс", "Elliptical"},
	{"elliptical", "Elliptical"},
	{"степпер", "StairStepper"},
	{"stair stepper", "StairStepper"},
}

func normalizeSportKey(raw string) string {
	raw = strings.ToLower(strings.TrimSpace(raw))
	raw = strings.ReplaceAll(raw, "\u00a0", " ")
	raw = strings.ReplaceAll(raw, "\u202f", " ")
	return strings.Join(strings.Fields(raw), " ")
}

func inferSportFromName(name string) string {
	key := normalizeSportKey(name)
	if key == "" {
		return ""
	}
	for _, item := range nameSportKeywords {
		if strings.Contains(key, item.keyword) {
			return item.sport
		}
	}
	return ""
}

func mapSportType(raw, name string) (string, error) {
	key := normalizeSportKey(raw)
	if key == "" {
		return "", errInvalidSport
	}
	if mapped, ok := sportTypeAliases[key]; ok {
		if mapped == "Workout" {
			if inferred := inferSportFromName(name); inferred != "" {
				return inferred, nil
			}
		}
		return mapped, nil
	}
	// Direct grom sport type passthrough.
	candidate := strings.ReplaceAll(key, " ", "")
	for _, sport := range workouts.ValidSportTypes() {
		if strings.EqualFold(sport, raw) || strings.EqualFold(sport, candidate) {
			return sport, nil
		}
	}
	if inferred := inferSportFromName(name); inferred != "" {
		return inferred, nil
	}
	return "Workout", nil
}
