package strava

import (
	"bytes"
	"os"
	"testing"
)

func TestParseStatsSampleArchive(t *testing.T) {
	data, err := os.ReadFile("/tmp/activities.csv")
	if err != nil {
		t.Skip("no /tmp/activities.csv")
	}
	stats, hint, activities, err := parseActivitiesCSVStats(bytes.NewReader(data))
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("hint=%v total=%d parsed=%d skipped=%d", hint, stats.TotalRows, stats.ParsedRows, stats.SkippedRows)
	for reason, count := range stats.SkipReasons {
		t.Logf("  skip %d: %s", count, reason)
	}
	if stats.ParsedRows != 982 {
		t.Fatalf("parsed %d want 982", stats.ParsedRows)
	}
	if len(activities) != stats.ParsedRows {
		t.Fatalf("activities len %d != parsed %d", len(activities), stats.ParsedRows)
	}
}
