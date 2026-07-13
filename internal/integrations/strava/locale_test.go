package strava

import (
	"testing"
	"time"
)

func TestParseRussianDateAsUTC(t *testing.T) {
	got, err := parseRussianDate("7 июл. 2026 г., 13:38:54")
	if err != nil {
		t.Fatalf("parseRussianDate() error = %v", err)
	}

	want := time.Date(2026, time.July, 7, 13, 38, 54, 0, time.UTC)
	if !got.Equal(want) {
		t.Fatalf("parseRussianDate() = %v, want %v", got, want)
	}
	if got.Location() != time.UTC {
		t.Fatalf("location = %v, want UTC", got.Location())
	}
}

func TestParseStartDateEnglishAsUTC(t *testing.T) {
	got, err := parseStartDate("Jul 7, 2026, 1:38:54 PM", localeEnglish)
	if err != nil {
		t.Fatalf("parseStartDate() error = %v", err)
	}

	want := time.Date(2026, time.July, 7, 13, 38, 54, 0, time.UTC)
	if !got.Equal(want) {
		t.Fatalf("parseStartDate() = %v, want %v", got, want)
	}
}

func TestParseStartDateRussianAsUTC(t *testing.T) {
	got, err := parseStartDate("7 июл. 2026 г., 13:38:54", localeRussian)
	if err != nil {
		t.Fatalf("parseStartDate() error = %v", err)
	}

	want := time.Date(2026, time.July, 7, 13, 38, 54, 0, time.UTC)
	if !got.Equal(want) {
		t.Fatalf("parseStartDate() = %v, want %v", got, want)
	}
}
