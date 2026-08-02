package workouts

import "strings"

const DeviceGrom = "Grom App"

// stripStravaFromDevice removes the word "Strava" from a device label (case-insensitive),
// leaving the rest intact (e.g. "Strava Wahoo ELEMNT" → "Wahoo ELEMNT").
// Strava-reexported FIT files often set manufacturer to Strava while keeping the real device name.
func stripStravaFromDevice(device string) string {
	parts := strings.Fields(device)
	out := make([]string, 0, len(parts))
	for _, part := range parts {
		if strings.EqualFold(part, "Strava") {
			continue
		}
		out = append(out, part)
	}
	return strings.Join(out, " ")
}
