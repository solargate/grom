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
		MediaFiles:      []string{"shot.png"},
		HasMedia:        true,
	}
	media := []workouts.MediaFileInput{
		{Filename: "shot.png", Data: []byte{0x89, 0x50, 0x4e, 0x47}},
	}

	delivery := &Delivery{client: server.Client()}
	if err := delivery.DeliverWorkoutUpdate("bob", workout, []string{server.URL}, nil, media); err != nil {
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
	if len(mediaItems) != 1 {
		t.Fatalf("expected 1 mediaItems, got %#v", mediaItems)
	}
	item, ok := mediaItems[0].(map[string]any)
	if !ok || item["filename"] != "shot.png" {
		t.Fatalf("unexpected media item: %#v", mediaItems[0])
	}
}

func TestPostActivityNon2xxReturnsError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	delivery := &Delivery{client: server.Client()}
	err := delivery.DeliverWorkoutDelete("bob", "38472901", []string{server.URL})
	if err == nil {
		t.Fatal("expected delivery error for non-2xx inbox response")
	}
}

func TestDeliverWorkoutStopsOnFirstInboxFailure(t *testing.T) {
	var successCalls int
	okServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		successCalls++
		w.WriteHeader(http.StatusAccepted)
	}))
	defer okServer.Close()

	failServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
	}))
	defer failServer.Close()

	workout := &workouts.Workout{
		ID:              "38472901",
		Name:            "Run",
		SportType:       "Run",
		StartDate:       time.Date(2026, 7, 8, 10, 0, 0, 0, time.UTC),
		DurationSeconds: 3600,
		Distance:        10000,
	}

	delivery := &Delivery{client: okServer.Client()}
	err := delivery.DeliverWorkout("alice", workout, []string{failServer.URL, okServer.URL}, nil, nil)
	if err == nil {
		t.Fatal("expected error when first inbox rejects activity")
	}
	if successCalls != 0 {
		t.Fatalf("second inbox should not be called after first failure, successCalls=%d", successCalls)
	}
}
