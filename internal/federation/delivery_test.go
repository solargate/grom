package federation

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/solargate/grom/internal/workouts"
)

func TestDeliverWorkoutDelete(t *testing.T) {
	var received []map[string]any
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatal(err)
		}
		var activity map[string]any
		if err := json.Unmarshal(body, &activity); err != nil {
			t.Fatal(err)
		}
		received = append(received, activity)
		w.WriteHeader(http.StatusAccepted)
	}))
	defer server.Close()

	delivery := &Delivery{client: server.Client()}
	if err := delivery.DeliverWorkoutDelete("bob", "38472901", []string{server.URL}); err != nil {
		t.Fatalf("DeliverWorkoutDelete() error = %v", err)
	}
	if len(received) != 1 {
		t.Fatalf("expected 1 activity, got %d", len(received))
	}
	if received[0]["type"] != "Delete" {
		t.Fatalf("type = %v", received[0]["type"])
	}
	object, _ := received[0]["object"].(string)
	if object != workoutObjectURL("bob", "38472901") {
		t.Fatalf("object = %q", object)
	}
}

func TestDeliverWorkoutIncludesStats(t *testing.T) {
	var received []map[string]any
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatal(err)
		}
		var activity map[string]any
		if err := json.Unmarshal(body, &activity); err != nil {
			t.Fatal(err)
		}
		received = append(received, activity)
		w.WriteHeader(http.StatusAccepted)
	}))
	defer server.Close()

	pace := "5:30"
	elevation := 120.0
	speed := 10.5
	hr := 145.0
	steps := 8000
	calories := 520.0
	workout := &workouts.Workout{
		ID:                   "38472901",
		Name:                 "Morning run",
		Description:          "Easy",
		SportType:            "Run",
		StartDate:            time.Date(2026, 7, 8, 10, 0, 0, 0, time.UTC),
		Device:               "Grom",
		DurationSeconds:      3600,
		DurationTotalSeconds: 3900,
		Distance:             10000,
		TempAvgKmm:           &pace,
		ElevationGain:        &elevation,
		SpeedAvgKmh:          &speed,
		HeartRateAvg:         &hr,
		StepsTotal:           &steps,
		Calories:             &calories,
		Track:                "track.gpx",
	}

	delivery := &Delivery{client: server.Client()}
	if err := delivery.DeliverWorkout("bob", workout, []string{server.URL}, nil, nil); err != nil {
		t.Fatalf("DeliverWorkout() error = %v", err)
	}
	if len(received) != 1 {
		t.Fatalf("expected 1 activity, got %d", len(received))
	}
	object, ok := received[0]["object"].(map[string]any)
	if !ok {
		t.Fatalf("object type = %T", received[0]["object"])
	}
	if object["durationTotalSeconds"] != float64(3900) {
		t.Fatalf("durationTotalSeconds = %v", object["durationTotalSeconds"])
	}
	if object["tempAvgKmm"] != "5:30" {
		t.Fatalf("tempAvgKmm = %v", object["tempAvgKmm"])
	}
	if object["elevationGain"] != 120.0 {
		t.Fatalf("elevationGain = %v", object["elevationGain"])
	}
	if object["speedAvgKmh"] != 10.5 {
		t.Fatalf("speedAvgKmh = %v", object["speedAvgKmh"])
	}
	if object["heartRateAvg"] != 145.0 {
		t.Fatalf("heartRateAvg = %v", object["heartRateAvg"])
	}
	if object["stepsTotal"] != float64(8000) {
		t.Fatalf("stepsTotal = %v", object["stepsTotal"])
	}
	if object["calories"] != 520.0 {
		t.Fatalf("calories = %v", object["calories"])
	}
}

func TestDeliverWorkoutUpdate(t *testing.T) {
	var received []map[string]any
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatal(err)
		}
		var activity map[string]any
		if err := json.Unmarshal(body, &activity); err != nil {
			t.Fatal(err)
		}
		received = append(received, activity)
		w.WriteHeader(http.StatusAccepted)
	}))
	defer server.Close()

	workout := &workouts.Workout{
		ID:              "38472901",
		Name:            "Updated run",
		Description:     "Edited",
		SportType:       "Run",
		StartDate:       time.Date(2026, 7, 8, 10, 0, 0, 0, time.UTC),
		DurationSeconds: 3600,
		Distance:        10000,
	}

	delivery := &Delivery{client: server.Client()}
	if err := delivery.DeliverWorkoutUpdate("bob", workout, []string{server.URL}, nil, nil); err != nil {
		t.Fatalf("DeliverWorkoutUpdate() error = %v", err)
	}
	if len(received) != 1 {
		t.Fatalf("expected 1 activity, got %d", len(received))
	}
	if received[0]["type"] != "Update" {
		t.Fatalf("type = %v", received[0]["type"])
	}
	object, ok := received[0]["object"].(map[string]any)
	if !ok {
		t.Fatalf("object type = %T", received[0]["object"])
	}
	if object["name"] != "Updated run" {
		t.Fatalf("name = %v", object["name"])
	}
	if object["id"] != workoutObjectURL("bob", "38472901") {
		t.Fatalf("object id = %v", object["id"])
	}
	mediaItems, ok := object["mediaItems"].([]any)
	if !ok {
		t.Fatalf("mediaItems type = %T", object["mediaItems"])
	}
	if len(mediaItems) != 0 {
		t.Fatalf("expected empty mediaItems, got %#v", mediaItems)
	}
}
