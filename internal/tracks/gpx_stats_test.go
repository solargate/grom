package tracks

import (
	"encoding/xml"
	"os"
	"path/filepath"
	"testing"

	"github.com/tkrajina/gpxgo/gpx"
)

func TestExtractGPXStatsDurationFromTimeBounds(t *testing.T) {
	gpxData, err := os.ReadFile(filepath.Join("..", "..", "testdata", "tracks", "1-sample.gpx"))
	if err != nil {
		t.Fatal(err)
	}
	parsed, err := Parse(gpxData, "1-sample.gpx")
	if err != nil {
		t.Fatal(err)
	}
	if parsed.Stats.DurationSeconds.Value == nil || *parsed.Stats.DurationSeconds.Value <= 0 {
		t.Fatalf("expected duration from GPX, got %#v", parsed.Stats.DurationSeconds)
	}
	if parsed.DistanceMeters == nil || *parsed.DistanceMeters <= 0 {
		t.Fatalf("expected distance from GPX, got %#v", parsed.DistanceMeters)
	}
}

func TestGPXSamplePointsSkipInvalidCoords(t *testing.T) {
	g := &gpx.GPX{
		Tracks: []gpx.GPXTrack{{
			Segments: []gpx.GPXTrackSegment{{
				Points: []gpx.GPXPoint{
					{Point: gpx.Point{Latitude: 91, Longitude: 0}},
					{Point: gpx.Point{Latitude: 55.75, Longitude: 37.62}},
				},
			}},
		}},
	}
	points := gpxSamplePoints(g)
	if len(points) != 1 {
		t.Fatalf("points = %d, want 1", len(points))
	}
}

func TestApplyGPXExtensionNodeHeartRate(t *testing.T) {
	sample := SamplePoint{}
	applyGPXExtensionNode(&sample, gpx.ExtensionNode{
		XMLName: xml.Name{Local: "hr"},
		Data:    "145",
	})
	if sample.HeartRate == nil || *sample.HeartRate != 145 {
		t.Fatalf("heart rate = %#v", sample.HeartRate)
	}
}

// Matches the synthetic GPX written by Flutter Strava API import
// (ui/grom/lib/services/strava_api_gpx.dart).
func TestParseGromStravaAPIGPXHeartRate(t *testing.T) {
	const gpxXML = `<?xml version="1.0" encoding="UTF-8"?>
<gpx version="1.1" creator="grom-strava-api" xmlns="http://www.topografix.com/GPX/1/1" xmlns:gpxtpx="http://www.garmin.com/xmlschemas/TrackPointExtension/v1">
<trk>
<name>HR Run</name>
<trkseg>
<trkpt lat="55.75" lon="37.61"><ele>100.0</ele><time>2026-09-05T10:00:00.000Z</time><extensions><gpxtpx:TrackPointExtension><gpxtpx:hr>120</gpxtpx:hr></gpxtpx:TrackPointExtension></extensions></trkpt>
<trkpt lat="55.76" lon="37.62"><ele>102.0</ele><time>2026-09-05T10:01:00.000Z</time><extensions><gpxtpx:TrackPointExtension><gpxtpx:hr>145</gpxtpx:hr></gpxtpx:TrackPointExtension></extensions></trkpt>
</trkseg>
</trk>
</gpx>
`
	parsed, err := Parse([]byte(gpxXML), "strava_55.gpx")
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if len(parsed.HeartRateSeries) < 2 {
		t.Fatalf("HeartRateSeries len = %d, want >= 2", len(parsed.HeartRateSeries))
	}
	if parsed.HeartRateSeries[0].BPM != 120 {
		t.Fatalf("first BPM = %v, want 120", parsed.HeartRateSeries[0].BPM)
	}
	if parsed.HeartRateSeries[1].BPM != 145 {
		t.Fatalf("second BPM = %v, want 145", parsed.HeartRateSeries[1].BPM)
	}
	if parsed.Stats.HeartRateAvg.Value == nil || *parsed.Stats.HeartRateAvg.Value != 132.5 {
		t.Fatalf("heart_rate_avg = %v, want 132.5", parsed.Stats.HeartRateAvg.Value)
	}
	if parsed.Stats.HeartRateMax.Value == nil || *parsed.Stats.HeartRateMax.Value != 145 {
		t.Fatalf("heart_rate_max = %v, want 145", parsed.Stats.HeartRateMax.Value)
	}
}
