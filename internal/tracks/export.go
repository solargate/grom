package tracks

import (
	"bytes"
	"fmt"
	"math"

	"github.com/muktihari/fit/decoder"
	"github.com/muktihari/fit/kit/datetime"
	"github.com/muktihari/fit/profile/basetype"
	"github.com/muktihari/fit/profile/filedef"
	"github.com/muktihari/fit/profile/mesgdef"
	"github.com/tkrajina/gpxgo/gpx"
)

func ExportGPX(originalData []byte, storageFilename, trackName string) ([]byte, error) {
	switch storageFilename {
	case TrackFileGPX:
		return originalData, nil
	case TrackFileFIT:
		return fitToGPX(originalData, trackName)
	default:
		return nil, ErrInvalidFormat
	}
}

func fitToGPX(data []byte, trackName string) ([]byte, error) {
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

	points := make([]gpx.GPXPoint, 0, len(activity.Records))
	for _, record := range activity.Records {
		pt, ok := recordToGPXPoint(record)
		if !ok {
			continue
		}
		points = append(points, pt)
	}
	if len(points) == 0 {
		return nil, ErrEmptyTrack
	}

	gpxDoc := &gpx.GPX{
		Version: "1.1",
		Creator: "Grom",
		XMLNs:   "http://www.topografix.com/2009/GPX",
		Tracks: []gpx.GPXTrack{{
			Name: trackName,
			Segments: []gpx.GPXTrackSegment{{
				Points: points,
			}},
		}},
	}

	return gpxDoc.ToXml(gpx.ToXmlParams{Version: "1.1", Indent: true})
}

func recordToGPXPoint(record *mesgdef.Record) (gpx.GPXPoint, bool) {
	lat := record.PositionLatDegrees()
	lon := record.PositionLongDegrees()
	if !validCoord(lat, lon) {
		return gpx.GPXPoint{}, false
	}

	point := gpx.GPXPoint{
		Point: gpx.Point{
			Latitude:  lat,
			Longitude: lon,
		},
	}
	if elev, ok := recordElevation(record); ok {
		point.Elevation = *gpx.NewNullableFloat64(elev)
	}
	if !record.Timestamp.Before(datetime.Epoch()) {
		point.Timestamp = record.Timestamp
	}
	return point, true
}

func recordElevation(record *mesgdef.Record) (float64, bool) {
	if record.EnhancedAltitude != basetype.Uint32Invalid {
		elev := record.EnhancedAltitudeScaled()
		if !math.IsNaN(elev) {
			return elev, true
		}
	}
	if record.Altitude != basetype.Uint16Invalid {
		elev := record.AltitudeScaled()
		if !math.IsNaN(elev) {
			return elev, true
		}
	}
	return 0, false
}
