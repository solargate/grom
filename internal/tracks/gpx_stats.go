package tracks

import (
	"strconv"
	"strings"

	"github.com/tkrajina/gpxgo/gpx"
)

func extractGPXStats(gpxData *gpx.GPX) Stats {
	samples := gpxSamplePoints(gpxData)
	stats := calculateStatsFromSamples(samples)

	timeBounds := gpxData.TimeBounds()
	if !timeBounds.StartTime.IsZero() && !timeBounds.EndTime.IsZero() {
		dur := int(timeBounds.EndTime.Sub(timeBounds.StartTime).Seconds())
		if dur >= 0 {
			stats.DurationTotalSeconds.setExplicit(dur)
			if stats.DurationSeconds.Source != SourceExplicit && (stats.DurationSeconds.Value == nil || *stats.DurationSeconds.Value == 0) {
				stats.DurationSeconds.setCalculated(dur)
			}
		}
	}

	return stats
}

func gpxSamplePoints(gpxData *gpx.GPX) []SamplePoint {
	points := make([]SamplePoint, 0)
	for _, track := range gpxData.Tracks {
		for _, segment := range track.Segments {
			for _, pt := range segment.Points {
				if !validCoord(pt.Latitude, pt.Longitude) {
					continue
				}
				sample := SamplePoint{
					Lat:     pt.Latitude,
					Lng:     pt.Longitude,
					Time:    pt.Timestamp,
					HasTime: pt.Timestamp.Year() > 1,
				}
				if pt.Elevation.NotNull() {
					v := pt.Elevation.Value()
					sample.Elevation = &v
				}
				for _, node := range pt.Extensions.Nodes {
					applyGPXExtensionNode(&sample, node)
				}
				points = append(points, sample)
			}
		}
	}
	return points
}

func applyGPXExtensionNode(sample *SamplePoint, node gpx.ExtensionNode) {
	name := strings.ToLower(localName(node.XMLName.Local))
	value := strings.TrimSpace(node.Data)
	switch {
	case name == "hr" || name == "heartrate":
		if v, ok := parseExtensionFloat(value); ok && v > 0 {
			sample.HeartRate = &v
		}
	case name == "cad":
		if v, ok := parseExtensionFloat(value); ok && v > 0 {
			sample.Cadence = &v
		}
	case name == "atemp" || name == "temp":
		if v, ok := parseExtensionFloat(value); ok {
			sample.Temperature = &v
		}
	case name == "speed":
		if v, ok := parseExtensionFloat(value); ok && v >= 0 {
			sample.SpeedMps = &v
		}
	case name == "power" || name == "watts":
		if v, ok := parseExtensionFloat(value); ok && v > 0 {
			sample.Power = &v
		}
	}
	for _, child := range node.Nodes {
		applyGPXExtensionNode(sample, child)
	}
}

func localName(name string) string {
	if idx := strings.LastIndex(name, ":"); idx >= 0 {
		return name[idx+1:]
	}
	return name
}

func parseExtensionFloat(raw string) (float64, bool) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return 0, false
	}
	v, err := strconv.ParseFloat(raw, 64)
	if err != nil {
		return 0, false
	}
	return v, true
}
