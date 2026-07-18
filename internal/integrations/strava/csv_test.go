package strava

import (
	"encoding/csv"
	"math"
	"strings"
	"testing"
)

func TestMpsToKmh(t *testing.T) {
	t.Parallel()

	if got := mpsToKmh(nil); got != nil {
		t.Fatalf("nil input: got %v", got)
	}

	mps := 4.581
	got := mpsToKmh(&mps)
	if got == nil {
		t.Fatal("expected non-nil")
	}
	want := 4.581 * 3.6
	if math.Abs(*got-want) > 1e-9 {
		t.Fatalf("got %v, want %v", *got, want)
	}
}

func TestParseActivityRowConvertsSpeedMpsToKmh(t *testing.T) {
	t.Parallel()

	record := make([]string, minActivityColumns)
	record[ColActivityID-1] = "19215512427"
	record[ColStartDate-1] = "Jul 7, 2026, 1:38:54 PM"
	record[ColName-1] = "From office"
	record[ColSportType-1] = "Ride"
	record[ColDurationTotal-1] = "3086"
	record[ColDurationMoving-1] = "2184"
	record[ColDistanceMeters-1] = "10004.7"
	record[ColSpeedMaxMps-1] = "8.18"
	record[ColSpeedAvgMps-1] = "4.581"

	row, err := parseActivityRow(record, localeEnglish)
	if err != nil {
		t.Fatalf("parseActivityRow: %v", err)
	}

	assertFloatPtr(t, "SpeedMaxKmh", row.SpeedMaxKmh, 8.18*3.6)
	assertFloatPtr(t, "SpeedAvgKmh", row.SpeedAvgKmh, 4.581*3.6)

	workout, err := row.ToWorkout(localeEnglish)
	if err != nil {
		t.Fatalf("ToWorkout: %v", err)
	}
	assertFloatPtr(t, "workout.SpeedMaxKmh", workout.SpeedMaxKmh, 8.18*3.6)
	assertFloatPtr(t, "workout.SpeedAvgKmh", workout.SpeedAvgKmh, 4.581*3.6)
}

func TestReadActivitiesCSVFromReaderConvertsSpeeds(t *testing.T) {
	t.Parallel()

	header := make([]string, minActivityColumns)
	for i := range header {
		header[i] = "col"
	}
	row := make([]string, minActivityColumns)
	row[ColActivityID-1] = "1"
	row[ColStartDate-1] = "Jul 7, 2026, 5:06:25 AM"
	row[ColName-1] = "To office"
	row[ColSportType-1] = "Ride"
	row[ColDurationTotal-1] = "3831"
	row[ColDurationMoving-1] = "1971"
	row[ColDistanceMeters-1] = "9897.6"
	row[ColSpeedMaxMps-1] = "9.26"
	row[ColSpeedAvgMps-1] = "5.022"

	var buf strings.Builder
	w := csv.NewWriter(&buf)
	if err := w.Write(header); err != nil {
		t.Fatalf("write header: %v", err)
	}
	if err := w.Write(row); err != nil {
		t.Fatalf("write row: %v", err)
	}
	w.Flush()
	if err := w.Error(); err != nil {
		t.Fatalf("csv flush: %v", err)
	}

	activities, _, err := ReadActivitiesCSVFromReader(strings.NewReader(buf.String()))
	if err != nil {
		t.Fatalf("ReadActivitiesCSVFromReader: %v", err)
	}
	if len(activities) != 1 {
		t.Fatalf("got %d activities, want 1", len(activities))
	}

	assertFloatPtr(t, "SpeedMaxKmh", activities[0].SpeedMaxKmh, 9.26*3.6)
	assertFloatPtr(t, "SpeedAvgKmh", activities[0].SpeedAvgKmh, 5.022*3.6)
}

func TestParseActivityRowEmptySpeedsStayNil(t *testing.T) {
	t.Parallel()

	record := make([]string, minActivityColumns)
	record[ColActivityID-1] = "2"
	record[ColStartDate-1] = "Jul 7, 2026, 1:00:00 PM"
	record[ColName-1] = "Indoor"
	record[ColSportType-1] = "Workout"
	record[ColDurationTotal-1] = "600"
	record[ColDurationMoving-1] = "600"
	record[ColDistanceMeters-1] = "0"

	row, err := parseActivityRow(record, localeEnglish)
	if err != nil {
		t.Fatalf("parseActivityRow: %v", err)
	}
	if row.SpeedMaxKmh != nil || row.SpeedAvgKmh != nil {
		t.Fatalf("expected nil speeds, got max=%v avg=%v", row.SpeedMaxKmh, row.SpeedAvgKmh)
	}
}

func assertFloatPtr(t *testing.T, name string, got *float64, want float64) {
	t.Helper()
	if got == nil {
		t.Fatalf("%s: got nil, want %v", name, want)
	}
	if math.Abs(*got-want) > 1e-9 {
		t.Fatalf("%s: got %v, want %v", name, *got, want)
	}
}
