package strava

// Column indices in activities.csv (1-based). Headers are locale-dependent; always use indices.
const (
	ColActivityID          = 1
	ColStartDate           = 2
	ColName                = 3
	ColSportType           = 4
	ColDescription         = 5
	ColDurationTotal       = 6
	ColRelativeEffort      = 9
	ColRegularTrack        = 10
	ColEquipment           = 12
	ColTrackFile           = 13
	ColDurationMoving      = 17
	ColDistanceMeters      = 18
	ColSpeedMaxMps         = 19 // Strava exports m/s; convert to km/h on import
	ColSpeedAvgMps         = 20
	ColElevationGain       = 21
	ColElevationLoss       = 22
	ColElevationLow        = 23
	ColElevationHigh       = 24
	ColGradeMax            = 25
	ColGradeAvg            = 26
	ColCadenceMax          = 29
	ColCadenceAvg          = 30
	ColHeartRateMax        = 31
	ColHeartRateAvg        = 32
	ColWattsMax            = 33
	ColWattsAvg            = 34
	ColCalories            = 35
	ColTemperatureMax      = 36
	ColTemperatureAvg      = 37
	ColStepsTotal          = 86
	ColCyclesTotal         = 93
	ColSetsTotal           = 101
	ColRepsTotal           = 102
	ColMediaFiles          = 103
)

const minActivityColumns = ColMediaFiles

func fieldAt(row []string, col int) string {
	idx := col - 1
	if idx < 0 || idx >= len(row) {
		return ""
	}
	return row[idx]
}
