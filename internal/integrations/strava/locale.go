package strava

import (
	"strconv"
	"strings"
	"time"
	"unicode"
)

type localeHint int

const (
	localeUnknown localeHint = iota
	localeEnglish
	localeRussian
)

var russianMonths = map[string]time.Month{
	"янв":  time.January,
	"фев":  time.February,
	"февр": time.February,
	"мар":  time.March,
	"апр":  time.April,
	"мая":  time.May,
	"май":  time.May,
	"июн":  time.June,
	"июл":  time.July,
	"авг":  time.August,
	"сен":  time.September,
	"сент": time.September,
	"окт":  time.October,
	"ноя":  time.November,
	"нояб": time.November,
	"дек":  time.December,
}

func detectLocale(rows [][]string) localeHint {
	for _, row := range rows {
		sport := strings.ToLower(strings.TrimSpace(fieldAt(row, ColSportType)))
		switch sport {
		case "велосипед", "ходьба", "плавание", "каякинг", "силовая тренировка", "тренировка":
			return localeRussian
		case "ride", "walk", "run", "swim", "workout", "weight training":
			return localeEnglish
		}
		date := fieldAt(row, ColStartDate)
		if strings.Contains(date, " г.") || strings.Contains(date, " г.") {
			return localeRussian
		}
	}
	return localeEnglish
}

func parseStartDate(raw string, hint localeHint) (time.Time, error) {
	raw = normalizeDateString(raw)
	if raw == "" {
		return time.Time{}, errEmptyValue
	}

	layouts := englishDateLayouts()
	if hint == localeRussian {
		if t, err := parseRussianDate(raw); err == nil {
			return t, nil
		}
	}
	for _, layout := range layouts {
		if t, err := time.ParseInLocation(layout, raw, time.Local); err == nil {
			return t, nil
		}
	}
	if hint != localeRussian {
		if t, err := parseRussianDate(raw); err == nil {
			return t, nil
		}
	}
	return time.Time{}, errInvalidDate
}

func normalizeDateString(raw string) string {
	raw = strings.TrimSpace(raw)
	raw = strings.ReplaceAll(raw, "\u202f", " ")
	raw = strings.ReplaceAll(raw, "\u00a0", " ")
	raw = strings.Join(strings.Fields(raw), " ")
	return raw
}

func englishDateLayouts() []string {
	return []string{
		"Jan 2, 2006, 3:04:05 PM",
		"January 2, 2006, 3:04:05 PM",
		"2 Jan 2006, 15:04:05",
		"2006-01-02 15:04:05",
		time.RFC3339,
	}
}

func parseRussianDate(raw string) (time.Time, error) {
	// Example: 7 июл. 2026 г., 13:38:54
	parts := strings.SplitN(raw, ",", 2)
	if len(parts) != 2 {
		return time.Time{}, errInvalidDate
	}
	datePart := strings.TrimSpace(parts[0])
	timePart := strings.TrimSpace(parts[1])

	datePart = strings.TrimSuffix(datePart, " г.")
	datePart = strings.TrimSuffix(datePart, " г")
	fields := strings.Fields(datePart)
	if len(fields) < 3 {
		return time.Time{}, errInvalidDate
	}

	day, err := strconv.Atoi(strings.TrimSuffix(fields[0], "."))
	if err != nil {
		return time.Time{}, errInvalidDate
	}
	monthKey := strings.TrimSuffix(strings.ToLower(fields[1]), ".")
	month, ok := russianMonths[monthKey]
	if !ok {
		return time.Time{}, errInvalidDate
	}
	year, err := strconv.Atoi(fields[2])
	if err != nil {
		return time.Time{}, errInvalidDate
	}

	hour, minute, second := 0, 0, 0
	timeFields := strings.Split(timePart, ":")
	if len(timeFields) >= 1 {
		hour, _ = strconv.Atoi(timeFields[0])
	}
	if len(timeFields) >= 2 {
		minute, _ = strconv.Atoi(timeFields[1])
	}
	if len(timeFields) >= 3 {
		second, _ = strconv.Atoi(strings.TrimSuffix(timeFields[2], "."))
	}

	return time.Date(year, month, day, hour, minute, second, 0, time.Local), nil
}

func parseInt(raw string) (*int, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil, nil
	}
	n, err := parseNumberInt(raw)
	if err != nil {
		return nil, err
	}
	return &n, nil
}

func parseFloat(raw string) (*float64, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil, nil
	}
	f, err := parseNumberFloat(raw)
	if err != nil {
		return nil, err
	}
	return &f, nil
}

func parseBool(raw string) (*bool, error) {
	raw = strings.ToLower(strings.TrimSpace(raw))
	if raw == "" {
		return nil, nil
	}
	switch raw {
	case "true", "1", "yes":
		v := true
		return &v, nil
	case "false", "0", "no":
		v := false
		return &v, nil
	default:
		return nil, errInvalidBool
	}
}

func parseNumberInt(raw string) (int, error) {
	f, err := parseNumberFloat(raw)
	if err != nil {
		return 0, err
	}
	return int(f), nil
}

func parseNumberFloat(raw string) (float64, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return 0, errEmptyValue
	}
	raw = strings.ReplaceAll(raw, " ", "")

	commaCount := strings.Count(raw, ",")
	dotCount := strings.Count(raw, ".")

	switch {
	case commaCount > 0 && dotCount > 0:
		if strings.LastIndex(raw, ",") > strings.LastIndex(raw, ".") {
			raw = strings.ReplaceAll(raw, ".", "")
			raw = strings.ReplaceAll(raw, ",", ".")
		} else {
			raw = strings.ReplaceAll(raw, ",", "")
		}
	case commaCount == 1 && dotCount == 0:
		if parts := strings.Split(raw, ","); len(parts) == 2 && len(parts[1]) <= 2 {
			raw = strings.ReplaceAll(raw, ",", ".")
		} else {
			raw = strings.ReplaceAll(raw, ",", "")
		}
	case commaCount > 1 && dotCount == 0:
		raw = strings.ReplaceAll(raw, ",", "")
	}

	raw = strings.Map(func(r rune) rune {
		if unicode.IsDigit(r) || r == '.' || r == '-' {
			return r
		}
		return -1
	}, raw)

	return strconv.ParseFloat(raw, 64)
}
