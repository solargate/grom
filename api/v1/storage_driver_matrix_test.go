package v1_test

import (
	"net/http"
	"testing"
	"time"

	"github.com/solargate/grom/internal/config"
)

func TestWorkoutsCRUDBothStorageDrivers(t *testing.T) {
	for _, driver := range []config.StorageDriver{config.StorageDriverFile, config.StorageDriverBBolt} {
		t.Run(string(driver), func(t *testing.T) {
			ta := setupTestAppWithDriver(t, driver)
			ta.register(t, "alice", "alice@example.com", "password12")
			token, _ := ta.login(t, "alice@example.com", "password12")

			w := ta.doJSON(t, http.MethodPost, "/api/v1/workouts", map[string]any{
				"name":             "Matrix run",
				"sport_type":       "Run",
				"start_date":       time.Date(2026, 8, 10, 9, 0, 0, 0, time.UTC).Format(time.RFC3339),
				"duration_seconds": 3600,
				"distance":         10000,
			}, token)
			expectStatus(t, w, http.StatusCreated)
			created := decodeObject(t, w)
			id, _ := created["id"].(string)

			w = ta.doJSON(t, http.MethodGet, "/api/v1/workouts/"+id, nil, token)
			expectStatus(t, w, http.StatusOK)
			got := decodeObject(t, w)
			if got["name"] != "Matrix run" {
				t.Fatalf("name = %#v", got["name"])
			}

			w = ta.doJSON(t, http.MethodPut, "/api/v1/workouts/"+id, map[string]any{
				"name":             "Matrix run edited",
				"sport_type":       "Run",
				"start_date":       time.Date(2026, 8, 10, 9, 0, 0, 0, time.UTC).Format(time.RFC3339),
				"duration_seconds": 3600,
				"distance":         10000,
			}, token)
			expectStatus(t, w, http.StatusOK)

			w = ta.doJSON(t, http.MethodDelete, "/api/v1/workouts/"+id, nil, token)
			expectStatus(t, w, http.StatusNoContent)
		})
	}
}

func TestProfileBothStorageDrivers(t *testing.T) {
	for _, driver := range []config.StorageDriver{config.StorageDriverFile, config.StorageDriverBBolt} {
		t.Run(string(driver), func(t *testing.T) {
			ta := setupTestAppWithDriver(t, driver)
			ta.register(t, "alice", "alice@example.com", "password12")
			token, _ := ta.login(t, "alice@example.com", "password12")

			w := ta.doJSON(t, http.MethodPost, "/api/v1/workouts", map[string]any{
				"name":             "Profile seed",
				"sport_type":       "Ride",
				"start_date":       time.Date(2026, 8, 11, 9, 0, 0, 0, time.UTC).Format(time.RFC3339),
				"duration_seconds": 1800,
				"distance":         8000,
			}, token)
			expectStatus(t, w, http.StatusCreated)

			w = ta.doJSON(t, http.MethodGet, "/api/v1/profile", nil, token)
			expectStatus(t, w, http.StatusOK)
			profile := decodeObject(t, w)
			raw, _ := profile["used_sport_types"].([]any)
			if len(raw) != 1 || raw[0] != "Ride" {
				t.Fatalf("used_sport_types = %#v", profile["used_sport_types"])
			}
		})
	}
}

func TestWorkoutLikesBothStorageDrivers(t *testing.T) {
	for _, driver := range []config.StorageDriver{config.StorageDriverFile, config.StorageDriverBBolt} {
		t.Run(string(driver), func(t *testing.T) {
			ta := setupTestAppWithDriver(t, driver)
			ta.register(t, "alice", "alice@example.com", "password12")
			ta.register(t, "bob", "bob@example.com", "password12")
			aliceToken, _ := ta.login(t, "alice@example.com", "password12")
			bobToken, _ := ta.login(t, "bob@example.com", "password12")

			w := ta.doJSON(t, http.MethodPost, "/api/v1/social/follow", map[string]string{
				"handle": "alice",
			}, bobToken)
			expectStatus(t, w, http.StatusCreated)

			w = ta.doJSON(t, http.MethodPost, "/api/v1/workouts", map[string]any{
				"name":             "Like me",
				"sport_type":       "Run",
				"start_date":       time.Date(2026, 8, 12, 9, 0, 0, 0, time.UTC).Format(time.RFC3339),
				"duration_seconds": 1200,
				"distance":         4000,
			}, aliceToken)
			expectStatus(t, w, http.StatusCreated)
			id, _ := decodeObject(t, w)["id"].(string)

			w = ta.doJSON(t, http.MethodPost, "/api/v1/workouts/"+id+"/likes?owner=alice", nil, bobToken)
			expectStatus(t, w, http.StatusOK)
			if decodeObject(t, w)["count"] != float64(1) {
				t.Fatalf("count = %s", w.Body.String())
			}
		})
	}
}
