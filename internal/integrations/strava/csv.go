package strava

import (
	"io"
	"os"
	"strings"

	"github.com/solargate/grom/internal/workouts"
)

type ActivityRow struct {
	StravaActivityID     string
	StartDate            string
	Name                 string
	SportTypeRaw         string
	Description          string
	DurationTotalSeconds int
	RelativeEffort       *float64
	RegularTrack         *bool
	EquipmentName        string
	TrackFile            string
	DurationSeconds      int
	Distance             float64
	SpeedMaxKmh          *float64
	SpeedAvgKmh          *float64
	ElevationGain        *float64
	ElevationLoss        *float64
	ElevationLow         *float64
	ElevationHigh        *float64
	GradeMax             *float64
	GradeAvg             *float64
	CadenceMax           *float64
	CadenceAvg           *float64
	HeartRateMax         *float64
	HeartRateAvg         *float64
	WattsMax             *float64
	WattsAvg             *float64
	Calories             *float64
	TemperatureMax       *float64
	TemperatureAvg       *float64
	StepsTotal           *int
	CyclesTotal          *int
	SetsTotal            *int
	RepsTotal            *int
	MediaFiles           string
}

func ReadActivitiesCSV(path string) ([]ActivityRow, localeHint, error) {
	file, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, localeUnknown, errNoActivities
		}
		return nil, localeUnknown, err
	}
	defer file.Close()
	return ReadActivitiesCSVFromReader(file)
}

func ReadActivitiesCSVFromReader(r io.Reader) ([]ActivityRow, localeHint, error) {
	_, hint, activities, err := parseActivitiesCSVStats(r)
	return activities, hint, err
}

func parseActivityRow(record []string, hint localeHint) (ActivityRow, error) {
	startDateRaw := fieldAt(record, ColStartDate)
	startDate, err := parseStartDate(startDateRaw, hint)
	if err != nil {
		return ActivityRow{}, err
	}

	durationTotal, err := parseNumberInt(fieldAt(record, ColDurationTotal))
	if err != nil {
		durationTotal = 0
	}
	durationMoving, err := parseNumberInt(fieldAt(record, ColDurationMoving))
	if err != nil {
		durationMoving = 0
	}
	distance, err := parseNumberFloat(fieldAt(record, ColDistanceMeters))
	if err != nil {
		distance = 0
	}

	relativeEffort, _ := parseFloat(fieldAt(record, ColRelativeEffort))
	regularTrack, _ := parseBool(fieldAt(record, ColRegularTrack))
	speedMaxMps, _ := parseFloat(fieldAt(record, ColSpeedMaxMps))
	speedAvgMps, _ := parseFloat(fieldAt(record, ColSpeedAvgMps))
	elevGain, _ := parseFloat(fieldAt(record, ColElevationGain))
	elevLoss, _ := parseFloat(fieldAt(record, ColElevationLoss))
	elevLow, _ := parseFloat(fieldAt(record, ColElevationLow))
	elevHigh, _ := parseFloat(fieldAt(record, ColElevationHigh))
	gradeMax, _ := parseFloat(fieldAt(record, ColGradeMax))
	gradeAvg, _ := parseFloat(fieldAt(record, ColGradeAvg))
	cadenceMax, _ := parseFloat(fieldAt(record, ColCadenceMax))
	cadenceAvg, _ := parseFloat(fieldAt(record, ColCadenceAvg))
	hrMax, _ := parseFloat(fieldAt(record, ColHeartRateMax))
	hrAvg, _ := parseFloat(fieldAt(record, ColHeartRateAvg))
	wattsMax, _ := parseFloat(fieldAt(record, ColWattsMax))
	wattsAvg, _ := parseFloat(fieldAt(record, ColWattsAvg))
	calories, _ := parseFloat(fieldAt(record, ColCalories))
	tempMax, _ := parseFloat(fieldAt(record, ColTemperatureMax))
	tempAvg, _ := parseFloat(fieldAt(record, ColTemperatureAvg))
	steps, _ := parseInt(fieldAt(record, ColStepsTotal))
	cycles, _ := parseInt(fieldAt(record, ColCyclesTotal))
	sets, _ := parseInt(fieldAt(record, ColSetsTotal))
	reps, _ := parseInt(fieldAt(record, ColRepsTotal))

	_ = startDate // used via return in ToWorkout

	return ActivityRow{
		StravaActivityID:     strings.TrimSpace(fieldAt(record, ColActivityID)),
		StartDate:            startDateRaw,
		Name:                 strings.TrimSpace(fieldAt(record, ColName)),
		SportTypeRaw:         strings.TrimSpace(fieldAt(record, ColSportType)),
		Description:          strings.TrimSpace(fieldAt(record, ColDescription)),
		DurationTotalSeconds: durationTotal,
		RelativeEffort:       relativeEffort,
		RegularTrack:         regularTrack,
		EquipmentName:        strings.TrimSpace(fieldAt(record, ColEquipment)),
		TrackFile:            strings.TrimSpace(fieldAt(record, ColTrackFile)),
		DurationSeconds:      durationMoving,
		Distance:             distance,
		SpeedMaxKmh:          mpsToKmh(speedMaxMps),
		SpeedAvgKmh:          mpsToKmh(speedAvgMps),
		ElevationGain:        elevGain,
		ElevationLoss:        elevLoss,
		ElevationLow:         elevLow,
		ElevationHigh:        elevHigh,
		GradeMax:             gradeMax,
		GradeAvg:             gradeAvg,
		CadenceMax:           cadenceMax,
		CadenceAvg:           cadenceAvg,
		HeartRateMax:         hrMax,
		HeartRateAvg:         hrAvg,
		WattsMax:             wattsMax,
		WattsAvg:             wattsAvg,
		Calories:             calories,
		TemperatureMax:       tempMax,
		TemperatureAvg:       tempAvg,
		StepsTotal:           steps,
		CyclesTotal:          cycles,
		SetsTotal:            sets,
		RepsTotal:            reps,
		MediaFiles:           strings.TrimSpace(fieldAt(record, ColMediaFiles)),
	}, nil
}

func (row ActivityRow) ToWorkout(hint localeHint) (*workouts.Workout, error) {
	startDate, err := parseStartDate(row.StartDate, hint)
	if err != nil {
		return nil, err
	}
	sportType, err := mapSportType(row.SportTypeRaw, row.Name)
	if err != nil {
		return nil, err
	}
	name := row.Name
	if name == "" {
		name = row.SportTypeRaw
	}

	return &workouts.Workout{
		Name:                 name,
		Description:          row.Description,
		SportType:            sportType,
		StartDate:            startDate,
		DurationSeconds:      row.DurationSeconds,
		Distance:             row.Distance,
		DurationTotalSeconds: row.DurationTotalSeconds,
		RelativeEffort:       row.RelativeEffort,
		RegularTrack:         row.RegularTrack,
		SpeedMaxKmh:          row.SpeedMaxKmh,
		SpeedAvgKmh:          row.SpeedAvgKmh,
		ElevationGain:        row.ElevationGain,
		ElevationLoss:        row.ElevationLoss,
		ElevationLow:         row.ElevationLow,
		ElevationHigh:        row.ElevationHigh,
		GradeMax:             row.GradeMax,
		GradeAvg:             row.GradeAvg,
		CadenceMax:           row.CadenceMax,
		CadenceAvg:           row.CadenceAvg,
		HeartRateMax:         row.HeartRateMax,
		HeartRateAvg:         row.HeartRateAvg,
		WattsMax:             row.WattsMax,
		WattsAvg:             row.WattsAvg,
		Calories:             row.Calories,
		TemperatureMax:       row.TemperatureMax,
		TemperatureAvg:       row.TemperatureAvg,
		StepsTotal:           row.StepsTotal,
		CyclesTotal:          row.CyclesTotal,
		SetsTotal:            row.SetsTotal,
		RepsTotal:            row.RepsTotal,
		StravaActivityID:     row.StravaActivityID,
	}, nil
}

// mpsToKmh converts Strava CSV speeds (metres per second) to km/h.
func mpsToKmh(mps *float64) *float64 {
	if mps == nil {
		return nil
	}
	kmh := *mps * 3.6
	return &kmh
}

func parseMediaPaths(raw string) []string {
	if raw == "" {
		return nil
	}
	parts := strings.Split(raw, "|")
	result := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		if isVideoFile(part) {
			continue
		}
		if isPhotoFile(part) {
			result = append(result, part)
		}
	}
	return result
}

func isVideoFile(path string) bool {
	lower := strings.ToLower(path)
	return strings.HasSuffix(lower, ".mp4") ||
		strings.HasSuffix(lower, ".mov") ||
		strings.HasSuffix(lower, ".avi") ||
		strings.HasSuffix(lower, ".mkv") ||
		strings.HasSuffix(lower, ".m4v")
}

func isPhotoFile(path string) bool {
	lower := strings.ToLower(path)
	return strings.HasSuffix(lower, ".jpg") ||
		strings.HasSuffix(lower, ".jpeg") ||
		strings.HasSuffix(lower, ".png") ||
		strings.HasSuffix(lower, ".webp") ||
		strings.HasSuffix(lower, ".gif") ||
		strings.HasSuffix(lower, ".heic") ||
		strings.HasSuffix(lower, ".heif")
}
