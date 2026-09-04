package tracks_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/solargate/grom/internal/tracks"
)

func TestMapSportTypeString(t *testing.T) {
	tests := []struct {
		raw  string
		want string
	}{
		{"running", "Run"},
		{"Ride", "Ride"},
		{"trail_run", "TrailRun"},
		{"unknown-sport-xyz", ""},
		{"", ""},
	}
	for _, tt := range tests {
		if got := tracks.MapSportTypeString(tt.raw); got != tt.want {
			t.Fatalf("MapSportTypeString(%q) = %q, want %q", tt.raw, got, tt.want)
		}
	}
}

func TestParseFITSportTypes(t *testing.T) {
	cases := []struct {
		file string
		want string
	}{
		{"1-ride.fit", "Ride"},
		{"2-walk.fit", "Walk"},
		{"4-pilates.fit", "Pilates"},
		{"5-weight.fit", "Workout"}, // session is training/cardio_training ("Cardio")
	}
	for _, tc := range cases {
		data, err := os.ReadFile(filepath.Join("..", "..", "testdata", "tracks", tc.file))
		if err != nil {
			t.Fatal(err)
		}
		parsed, err := tracks.Parse(data, tc.file)
		if err != nil {
			t.Fatalf("%s: %v", tc.file, err)
		}
		if parsed.SportType != tc.want {
			t.Fatalf("%s sport_type = %q, want %q", tc.file, parsed.SportType, tc.want)
		}
	}
}

func TestParseGPXSportType(t *testing.T) {
	gpx := []byte(`<?xml version="1.0" encoding="UTF-8"?>
<gpx version="1.1" creator="grom-test">
  <trk>
    <name>Trail</name>
    <type>running</type>
    <trkseg>
      <trkpt lat="55.7558" lon="37.6173">
        <time>2026-07-06T08:40:00Z</time>
      </trkpt>
      <trkpt lat="55.7568" lon="37.6273">
        <time>2026-07-06T09:40:00Z</time>
      </trkpt>
    </trkseg>
  </trk>
</gpx>`)
	parsed, err := tracks.Parse(gpx, "trail.gpx")
	if err != nil {
		t.Fatal(err)
	}
	if parsed.SportType != "Run" {
		t.Fatalf("sport_type = %q, want Run", parsed.SportType)
	}
}
