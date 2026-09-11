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
	Name                 string
	SportType            string // Grom sport type id when known from the track; empty if unknown
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

// ApplyToWorkout fills empty workout core metrics from the track.
// Non-zero client values (start date, duration, distance) are preserved.
func (d *Data) ApplyToWorkout(startDate *time.Time, durationSeconds *int, distanceMeters *float64) {
	if d == nil {
		return
	}
	if d.StartTime != nil && startDate.IsZero() {
		*startDate = *d.StartTime
	}
	if d.DurationSeconds != nil && *durationSeconds <= 0 {
		*durationSeconds = *d.DurationSeconds
	}
	if d.DistanceMeters != nil && *distanceMeters <= 0 {
		*distanceMeters = *d.DistanceMeters
	}
}

// ApplyDurationTotal fills duration_total_seconds from the track only when unset.
func (d *Data) ApplyDurationTotal(durationTotalSeconds *int) {
	if d == nil || durationTotalSeconds == nil || *durationTotalSeconds > 0 {
		return
	}
	if d.DurationTotalSeconds != nil {
		*durationTotalSeconds = *d.DurationTotalSeconds
	} else if d.Stats.DurationTotalSeconds.Value != nil {
		*durationTotalSeconds = *d.Stats.DurationTotalSeconds.Value
	}
}
