package v1_test

import (
	"bytes"
	"fmt"
	"image"
	"image/png"
	"net/http"
	"testing"

	"github.com/solargate/grom/internal/config"
	"github.com/solargate/grom/internal/workouts"
)

func TestWorkoutExternalIDCheckAndConflict(t *testing.T) {
	for _, driver := range []config.StorageDriver{config.StorageDriverFile, config.StorageDriverBBolt} {
		t.Run(string(driver), func(t *testing.T) {
			ta := setupTestAppWithDriver(t, driver)
			ta.register(t, "alice", "alice@example.com", "password12")
			token, _ := ta.login(t, "alice@example.com", "password12")

			w := ta.doJSON(t, http.MethodGet, "/api/v1/workouts/external", nil, token)
			expectStatus(t, w, http.StatusBadRequest)

			w = ta.doJSON(t, http.MethodGet, "/api/v1/workouts/external?name=health-sync/strava", nil, token)
			expectStatus(t, w, http.StatusBadRequest)

			w = ta.doJSON(t, http.MethodGet, "/api/v1/workouts/external?name=health-sync/strava&id=ride-1.csv", nil, token)
			expectStatus(t, w, http.StatusOK)
			if decodeObject(t, w)["exists"] != false {
				t.Fatalf("expected exists=false: %s", w.Body.String())
			}

			w = ta.doJSON(t, http.MethodPost, "/api/v1/workouts", map[string]any{
				"name":       "Synced ride",
				"sport_type": "Ride",
				"start_date": "2026-07-30T16:26:00Z",
				"distance":   12000,
				"external_id": map[string]string{
					"name": "health-sync/strava",
					"id":   "ride-1.csv",
				},
			}, token)
			expectStatus(t, w, http.StatusCreated)
			created := decodeObject(t, w)
			ext, _ := created["external_id"].(map[string]any)
			if ext["name"] != "health-sync/strava" || ext["id"] != "ride-1.csv" {
				t.Fatalf("unexpected external_id: %#v", created["external_id"])
			}

			w = ta.doJSON(t, http.MethodGet, "/api/v1/workouts/external?name=health-sync/strava&id=ride-1.csv", nil, token)
			expectStatus(t, w, http.StatusOK)
			if decodeObject(t, w)["exists"] != true {
				t.Fatalf("expected exists=true: %s", w.Body.String())
			}

			w = ta.doJSON(t, http.MethodPost, "/api/v1/workouts", map[string]any{
				"name":       "Duplicate",
				"sport_type": "Ride",
				"start_date": "2026-07-31T16:26:00Z",
				"external_id": map[string]string{
					"name": "health-sync/strava",
					"id":   "ride-1.csv",
				},
			}, token)
			expectStatus(t, w, http.StatusConflict)

			w = ta.doMultipart(t, http.MethodPost, "/api/v1/workouts", token, map[string]string{
				"name":             "Multipart sync",
				"sport_type":       "Ride",
				"start_date":       "2026-08-01T10:00:00Z",
				"external_id_name": "health-sync/garmin",
				"external_id_id":   "ride-2.csv",
			}, nil)
			expectStatus(t, w, http.StatusCreated)
			created = decodeObject(t, w)
			ext, _ = created["external_id"].(map[string]any)
			if ext["name"] != "health-sync/garmin" || ext["id"] != "ride-2.csv" {
				t.Fatalf("unexpected multipart external_id: %#v", created["external_id"])
			}
		})
	}
}

func TestWorkoutMediaLimitAcrossDrivers(t *testing.T) {
	pngData := tinyPNG(t)

	for _, driver := range []config.StorageDriver{config.StorageDriverFile, config.StorageDriverBBolt} {
		t.Run(string(driver), func(t *testing.T) {
			ta := setupTestAppWithDriver(t, driver)
			ta.register(t, "alice", "alice@example.com", "password12")
			token, _ := ta.login(t, "alice@example.com", "password12")

			w := ta.doJSON(t, http.MethodPost, "/api/v1/workouts", map[string]any{
				"name":       "Photos",
				"sport_type": "Run",
				"start_date": "2026-07-08T10:00:00Z",
			}, token)
			expectStatus(t, w, http.StatusCreated)
			id, _ := decodeObject(t, w)["id"].(string)

			parts := make([]filePart, workouts.MaxPhotosPerWorkout)
			for i := range parts {
				parts[i] = filePart{filename: fmt.Sprintf("p-%02d.png", i), data: pngData}
			}
			w = ta.doMultipart(t, http.MethodPost, "/api/v1/workouts/"+id+"/media", token, nil,
				map[string][]filePart{"photos": parts},
			)
			expectStatus(t, w, http.StatusOK)
			mediaFiles, _ := decodeObject(t, w)["media_files"].([]any)
			if len(mediaFiles) != workouts.MaxPhotosPerWorkout {
				t.Fatalf("expected %d media files, got %d", workouts.MaxPhotosPerWorkout, len(mediaFiles))
			}

			w = ta.doMultipart(t, http.MethodPost, "/api/v1/workouts/"+id+"/media", token, nil,
				map[string][]filePart{
					"photos": {{filename: "overflow.png", data: pngData}},
				},
			)
			expectStatus(t, w, http.StatusBadRequest)

			// create with too many photos at once
			overflow := make([]filePart, workouts.MaxPhotosPerWorkout+1)
			for i := range overflow {
				overflow[i] = filePart{filename: fmt.Sprintf("c-%02d.png", i), data: pngData}
			}
			w = ta.doMultipart(t, http.MethodPost, "/api/v1/workouts", token, map[string]string{
				"name":       "Too many",
				"sport_type": "Run",
				"start_date": "2026-07-09T10:00:00Z",
			}, map[string][]filePart{"photos": overflow})
			expectStatus(t, w, http.StatusBadRequest)
		})
	}
}

func TestListWorkoutsInvalidScopeAndLocalFeed(t *testing.T) {
	ta := setupTestApp(t)
	ta.register(t, "alice", "alice@example.com", "password12")
	ta.register(t, "bob", "bob@example.com", "password12")
	aliceToken, _ := ta.login(t, "alice@example.com", "password12")
	bobToken, _ := ta.login(t, "bob@example.com", "password12")

	w := ta.doJSON(t, http.MethodGet, "/api/v1/workouts?scope=invalid", nil, aliceToken)
	expectStatus(t, w, http.StatusBadRequest)

	w = ta.doJSON(t, http.MethodPost, "/api/v1/workouts", map[string]any{
		"name":       "Alice run",
		"sport_type": "Run",
		"start_date": "2026-07-08T10:00:00Z",
	}, aliceToken)
	expectStatus(t, w, http.StatusCreated)
	aliceID, _ := decodeObject(t, w)["id"].(string)

	w = ta.doJSON(t, http.MethodPost, "/api/v1/workouts", map[string]any{
		"name":       "Bob ride",
		"sport_type": "Ride",
		"start_date": "2026-07-09T10:00:00Z",
	}, bobToken)
	expectStatus(t, w, http.StatusCreated)

	w = ta.doJSON(t, http.MethodPost, "/api/v1/social/follow", map[string]string{"handle": "alice"}, bobToken)
	expectStatus(t, w, http.StatusCreated)

	w = ta.doJSON(t, http.MethodGet, "/api/v1/workouts?scope=feed&limit=20", nil, bobToken)
	expectStatus(t, w, http.StatusOK)
	items, _, _ := decodeWorkoutPage(t, w)
	foundAlice := false
	for _, item := range items {
		if item["id"] == aliceID {
			foundAlice = true
			if item["owner"] != "alice" {
				t.Fatalf("expected owner alice, got %#v", item)
			}
		}
	}
	if !foundAlice {
		t.Fatalf("expected followed alice workout in feed: %#v", items)
	}
}

func TestEquipmentDistanceUpdatesAfterWorkoutCreateAndDelete(t *testing.T) {
	ta := setupTestApp(t)
	ta.register(t, "alice", "alice@example.com", "password12")
	token, _ := ta.login(t, "alice@example.com", "password12")

	w := ta.doJSON(t, http.MethodPost, "/api/v1/equipment", map[string]any{
		"type": "bike",
		"name": "Road",
	}, token)
	expectStatus(t, w, http.StatusCreated)
	eqID, _ := decodeObject(t, w)["id"].(string)

	w = ta.doJSON(t, http.MethodPost, "/api/v1/workouts", map[string]any{
		"name":          "Ride",
		"sport_type":    "Ride",
		"start_date":    "2026-07-08T10:00:00Z",
		"distance":      15000,
		"equipment_ids": []string{eqID},
	}, token)
	expectStatus(t, w, http.StatusCreated)
	workoutID, _ := decodeObject(t, w)["id"].(string)

	ta.app.EquipmentDistance.Wait()

	w = ta.doJSON(t, http.MethodGet, "/api/v1/equipment", nil, token)
	expectStatus(t, w, http.StatusOK)
	eqList := decodeList(t, w)
	if len(eqList) != 1 || eqList[0]["distance"] != float64(15000) {
		t.Fatalf("expected equipment distance 15000, got %#v", eqList)
	}

	w = ta.doJSON(t, http.MethodDelete, "/api/v1/workouts/"+workoutID, nil, token)
	expectStatus(t, w, http.StatusNoContent)
	ta.app.EquipmentDistance.Wait()

	w = ta.doJSON(t, http.MethodGet, "/api/v1/equipment", nil, token)
	expectStatus(t, w, http.StatusOK)
	eqList = decodeList(t, w)
	if len(eqList) != 1 || eqList[0]["distance"] != float64(0) {
		t.Fatalf("expected equipment distance 0 after delete, got %#v", eqList)
	}
}

func tinyPNG(t *testing.T) []byte {
	t.Helper()
	img := image.NewRGBA(image.Rect(0, 0, 8, 8))
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		t.Fatal(err)
	}
	return buf.Bytes()
}
