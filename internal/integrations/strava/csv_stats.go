package strava

import (
	"encoding/csv"
	"fmt"
	"io"
)

type CSVParseStats struct {
	TotalRows   int
	ParsedRows  int
	SkippedRows int
	SkipReasons map[string]int
}

func parseActivitiesCSVStats(r io.Reader) (CSVParseStats, localeHint, []ActivityRow, error) {
	stats := CSVParseStats{SkipReasons: make(map[string]int)}
	reader := csv.NewReader(r)
	reader.LazyQuotes = true
	reader.FieldsPerRecord = -1
	if _, err := reader.Read(); err != nil {
		return stats, localeUnknown, nil, fmt.Errorf("read activities header: %w", err)
	}

	rawRows := make([][]string, 0)
	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return stats, localeUnknown, nil, fmt.Errorf("read activities row: %w", err)
		}
		stats.TotalRows++
		if len(record) < minActivityColumns {
			stats.SkippedRows++
			stats.SkipReasons["too_few_columns"]++
			continue
		}
		rawRows = append(rawRows, record)
	}

	hint := detectLocale(rawRows)
	activities := make([]ActivityRow, 0, len(rawRows))
	for _, record := range rawRows {
		row, err := parseActivityRow(record, hint)
		if err != nil {
			stats.SkippedRows++
			stats.SkipReasons[err.Error()]++
			continue
		}
		activities = append(activities, row)
		stats.ParsedRows++
	}
	return stats, hint, activities, nil
}
