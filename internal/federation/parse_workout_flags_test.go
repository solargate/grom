package federation

import (
	"testing"
)

func TestParseFederatedWorkoutObjectHonorsLightweightFlags(t *testing.T) {
	workout, track, media, err := ParseWorkoutObject(map[string]any{
		"id":              "https://remote.test/users/bob/workouts/abcd1234",
		"type":            "Workout",
		"name":            "Remote run",
		"sportType":       "Run",
		"startDate":       "2026-07-08T10:00:00Z",
		"durationSeconds": 1800,
		"distance":        5000.0,
		"track":           "track.gpx",
		"hasMapPreview":   true,
		"hasMedia":        true,
		"mediaFiles":      []any{"photo1.jpg", "photo2.jpg"},
	})
	if err != nil {
		t.Fatalf("ParseWorkoutObject: %v", err)
	}
	if workout == nil {
		t.Fatal("expected workout")
	}
	if !workout.HasMapPreview {
		t.Fatal("expected HasMapPreview from lightweight outbox flag")
	}
	if !workout.HasMedia {
		t.Fatal("expected HasMedia from lightweight outbox flag")
	}
	if len(workout.MediaFiles) != 2 || workout.MediaFiles[0] != "photo1.jpg" {
		t.Fatalf("MediaFiles = %#v", workout.MediaFiles)
	}
	if len(track) != 0 {
		t.Fatalf("expected no track bytes, got %d", len(track))
	}
	if len(media) != 0 {
		t.Fatalf("expected no media bytes, got %d", len(media))
	}
}
