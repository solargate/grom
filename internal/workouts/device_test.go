package workouts

import (
	"testing"

	"github.com/solargate/grom/internal/tracks"
)

func TestStripStravaFromDevice(t *testing.T) {
	tests := []struct {
		in   string
		want string
	}{
		{"Strava Wahoo ELEMNT", "Wahoo ELEMNT"},
		{"strava Wahoo ELEMNT", "Wahoo ELEMNT"},
		{"Wahoo ELEMNT", "Wahoo ELEMNT"},
		{"Strava", ""},
		{"  Strava   Wahoo   ELEMNT  ", "Wahoo ELEMNT"},
		{"Garmin Edge 530", "Garmin Edge 530"},
		{"", ""},
	}
	for _, tt := range tests {
		if got := stripStravaFromDevice(tt.in); got != tt.want {
			t.Fatalf("stripStravaFromDevice(%q) = %q, want %q", tt.in, got, tt.want)
		}
	}
}

func TestDeviceForTrackStripsStrava(t *testing.T) {
	stravaWahoo := "Strava Wahoo ELEMNT"
	parsed := &tracks.Data{Device: &stravaWahoo}

	if got := deviceForTrack(tracks.TrackFileFIT, parsed); got != "Wahoo ELEMNT" {
		t.Fatalf("deviceForTrack() = %q, want %q", got, "Wahoo ELEMNT")
	}

	onlyStrava := "Strava"
	parsed.Device = &onlyStrava
	if got := deviceForTrack(tracks.TrackFileFIT, parsed); got != DeviceGrom {
		t.Fatalf("deviceForTrack() = %q, want %q", got, DeviceGrom)
	}
}
