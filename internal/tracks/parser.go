package tracks

import (
	"bytes"
	"fmt"
	"math"
	"path/filepath"
	"strings"
	"time"

	"github.com/muktihari/fit/decoder"
	"github.com/muktihari/fit/kit/datetime"
	"github.com/muktihari/fit/profile/basetype"
	"github.com/muktihari/fit/profile/filedef"
	"github.com/tkrajina/gpxgo/gpx"
)

var (
	ErrInvalidFormat = fmt.Errorf("invalid track format: only FIT and GPX are supported")
	ErrTrackTooLarge = fmt.Errorf("track file exceeds size limit")
	ErrEmptyTrack    = fmt.Errorf("track file is empty")
	ErrInvalidTrack  = fmt.Errorf("invalid track file")
)

func TrackFileName(filename string) (string, error) {
	ext := strings.ToLower(filepath.Ext(filename))
	switch ext {
	case ".gpx":
		return TrackFileGPX, nil
	case ".fit":
		return TrackFileFIT, nil
	default:
		return "", ErrInvalidFormat
	}
}

func Parse(data []byte, filename string) (*Data, error) {
	if len(data) == 0 {
		return nil, ErrEmptyTrack
	}
	if len(data) > MaxTrackSizeBytes {
		return nil, ErrTrackTooLarge
	}

	ext := strings.ToLower(filepath.Ext(filename))
	switch ext {
	case ".gpx":
		return parseGPX(data)
	case ".fit":
		return parseFIT(data)
	default:
		return nil, ErrInvalidFormat
	}
}

func parseGPX(data []byte) (*Data, error) {
	gpxData, err := gpx.Parse(bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidTrack, err)
	}

	result := &Data{}
	points := make([]LatLng, 0)

	for _, track := range gpxData.Tracks {
		for _, segment := range track.Segments {
			for _, pt := range segment.Points {
				if !validCoord(pt.Latitude, pt.Longitude) {
					continue
				}
				points = append(points, LatLng{Lat: pt.Latitude, Lng: pt.Longitude})
			}
		}
	}
	result.Points = points

	timeBounds := gpxData.TimeBounds()
	if !timeBounds.StartTime.IsZero() {
		start := timeBounds.StartTime
		result.StartTime = &start
	}

	stats, speedSeries := extractGPXStats(gpxData)
	result.Stats = stats
	result.SpeedSeries = speedSeries
	populateLegacyDurationFields(result)

	length := gpxData.Length2D()
	if length > 0 {
		result.DistanceMeters = &length
	}

	result.Stats.FinalizePace(result.DistanceMeters)
	return result, nil
}

func parseFIT(data []byte) (*Data, error) {
	lis := filedef.NewListener()
	defer lis.Close()

	dec := decoder.New(bytes.NewReader(data),
		decoder.WithMesgListener(lis),
		decoder.WithBroadcastOnly(),
	)
	if _, err := dec.Decode(); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidTrack, err)
	}

	activity, ok := lis.File().(*filedef.Activity)
	if !ok {
		return nil, fmt.Errorf("%w: not an activity file", ErrInvalidTrack)
	}

	result := &Data{}
	points := make([]LatLng, 0, len(activity.Records))

	var firstTimestamp time.Time
	var lastTimestamp time.Time
	hasTimestamp := false

	for _, record := range activity.Records {
		lat := record.PositionLatDegrees()
		lon := record.PositionLongDegrees()
		if !validCoord(lat, lon) {
			continue
		}
		points = append(points, LatLng{Lat: lat, Lng: lon})

		if !record.Timestamp.Before(datetime.Epoch()) {
			if !hasTimestamp {
				firstTimestamp = record.Timestamp
				hasTimestamp = true
			}
			lastTimestamp = record.Timestamp
		}
	}
	result.Points = points

	if len(activity.Sessions) > 0 {
		session := activity.Sessions[0]
		if !session.StartTime.Before(datetime.Epoch()) {
			start := session.StartTime
			result.StartTime = &start
		}
		dist := session.TotalDistanceScaled()
		if dist > 0 {
			result.DistanceMeters = &dist
		}
	}

	if result.StartTime == nil && hasTimestamp {
		start := firstTimestamp
		result.StartTime = &start
	}

	if result.DistanceMeters == nil && len(points) >= 2 {
		dist := haversineDistance(points)
		if dist > 0 {
			result.DistanceMeters = &dist
		}
	}

	stats, speedSeries := extractFITStats(activity)
	result.Stats = stats
	result.SpeedSeries = speedSeries
	populateLegacyDurationFields(result)

	if result.DurationSeconds == nil && hasTimestamp && !lastTimestamp.Before(firstTimestamp) {
		dur := int(lastTimestamp.Sub(firstTimestamp).Seconds())
		if dur >= 0 {
			result.DurationSeconds = &dur
			result.Stats.DurationSeconds.setCalculated(dur)
		}
	}

	if device := extractDevice(activity); device != "" {
		result.Device = &device
	}

	result.Stats.FinalizePace(result.DistanceMeters)
	return result, nil
}

func populateLegacyDurationFields(result *Data) {
	if result.Stats.DurationSeconds.Value != nil {
		v := *result.Stats.DurationSeconds.Value
		result.DurationSeconds = &v
	}
	if result.Stats.DurationTotalSeconds.Value != nil {
		v := *result.Stats.DurationTotalSeconds.Value
		result.DurationTotalSeconds = &v
	}
}

func validCoord(lat, lon float64) bool {
	if math.IsNaN(lat) || math.IsNaN(lon) {
		return false
	}
	if math.Float64bits(lat) == basetype.Float64Invalid || math.Float64bits(lon) == basetype.Float64Invalid {
		return false
	}
	if lat < -90 || lat > 90 || lon < -180 || lon > 180 {
		return false
	}
	if lat == 0 && lon == 0 {
		return false
	}
	return true
}

func haversineDistance(points []LatLng) float64 {
	if len(points) < 2 {
		return 0
	}
	var total float64
	for i := 1; i < len(points); i++ {
		total += haversine(points[i-1], points[i], earthRadiusMeters)
	}
	return total
}

func haversine(a, b LatLng, radius float64) float64 {
	lat1 := a.Lat * math.Pi / 180
	lat2 := b.Lat * math.Pi / 180
	dLat := (b.Lat - a.Lat) * math.Pi / 180
	dLng := (b.Lng - a.Lng) * math.Pi / 180

	sinDLat := math.Sin(dLat / 2)
	sinDLng := math.Sin(dLng / 2)
	h := sinDLat*sinDLat + math.Cos(lat1)*math.Cos(lat2)*sinDLng*sinDLng
	return 2 * radius * math.Asin(math.Sqrt(h))
}
