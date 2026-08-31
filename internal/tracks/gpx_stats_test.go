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
