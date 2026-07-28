package tracks

import "time"

const (
	MaxTrackSizeBytes = 20 << 20 // 20 MiB
	TrackFileGPX      = "track.gpx"
	TrackFileFIT      = "track.fit"
)

type LatLng struct {
	Lat float64
	Lng float64
}

type Data struct {
	StartTime            *time.Time
	DurationSeconds      *int
	DurationTotalSeconds *int
	DistanceMeters       *float64
	Device               *string
	Points               []LatLng
	SpeedSeries          []SpeedPoint
	HeartRateSeries      []HeartRatePoint
	Stats                Stats
}

func (d *Data) HasGPS() bool {
	return d != nil && len(d.Points) >= 2
}

func (d *Data) ApplyToWorkout(startDate *time.Time, durationSeconds *int, distanceMeters *float64) {
	if d == nil {
		return
	}
	if d.StartTime != nil {
		*startDate = *d.StartTime
	}
	if d.DurationSeconds != nil {
		*durationSeconds = *d.DurationSeconds
	}
	if d.DistanceMeters != nil {
		*distanceMeters = *d.DistanceMeters
	}
}

func (d *Data) ApplyDurationTotal(durationTotalSeconds *int) {
	if d == nil || durationTotalSeconds == nil {
		return
	}
	if d.DurationTotalSeconds != nil {
		*durationTotalSeconds = *d.DurationTotalSeconds
	} else if d.Stats.DurationTotalSeconds.Value != nil {
		*durationTotalSeconds = *d.Stats.DurationTotalSeconds.Value
	}
}
