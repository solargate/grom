package v1_test

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/solargate/grom/internal/config"
	"github.com/solargate/grom/internal/data"
	"github.com/solargate/grom/internal/federation"
	"github.com/solargate/grom/internal/storage/bbolt"
	"github.com/solargate/grom/internal/storage/file"
	"github.com/solargate/grom/internal/storage/keys"
	"github.com/solargate/grom/internal/workouts"
)

const workoutCollection = "/api/v1/workouts"

func saveWorkoutForDeletionTest(t *testing.T, ta *testApp, token, start string) string {
	t.Helper()
	w := ta.doJSON(t, http.MethodPost, workoutCollection, map[string]any{
		"name": "Morning run", "sport_type": "Run", "start_date": start,
		"duration_seconds": 600, "distance": 2500,
	}, token)
	expectStatus(t, w, http.StatusCreated)
	return decodeObject(t, w)["id"].(string)
}

func TestDeleteAllWorkoutsOnBothDrivers(t *testing.T) {
	for _, driver := range []config.StorageDriver{config.StorageDriverFile, config.StorageDriverBBolt} {
		t.Run(string(driver), func(t *testing.T) {
			ta := setupTestAppWithDriver(t, driver)
			ta.register(t, "alice", "alice@example.com", "password12")
			ta.register(t, "bob", "bob@example.com", "password12")
			alice, _ := ta.login(t, "alice@example.com", "password12")
			bob, _ := ta.login(t, "bob@example.com", "password12")
			password := map[string]string{"password": "password12"}

			w := ta.doJSON(t, http.MethodPost, "/api/v1/equipment", map[string]string{"type": "shoes", "name": "Trail"}, alice)
			expectStatus(t, w, http.StatusCreated)
			equipmentID := decodeObject(t, w)["id"].(string)
			w = ta.doMultipart(t, http.MethodPost, workoutCollection, alice, map[string]string{
				"name": "Imported run", "sport_type": "Run", "start_date": "2026-07-08T10:00:00Z",
				"equipment_ids": fmt.Sprintf("[%q]", equipmentID), "external_id_name": "strava", "external_id_id": "123",
			}, map[string][]filePart{
				"track":  {{filename: "sample.gpx", data: readTestdata(t, "tracks/1-sample.gpx")}},
				"photos": {{filename: "sample.png", data: tinyPNG(t)}},
			})
			expectStatus(t, w, http.StatusCreated)
			aliceWorkoutID := decodeObject(t, w)["id"].(string)
			aliceWorkout, err := ta.app.Workouts.Get("alice", aliceWorkoutID)
			if err != nil {
				t.Fatal(err)
			}
			workoutDir := filepath.Join(data.UserDir(ta.dir, "alice"), "workouts", keys.WorkoutDirName(aliceWorkout.StartDate, aliceWorkout.ID))
			if _, err := os.Stat(workoutDir); err != nil || aliceWorkout.Track == "" || len(aliceWorkout.MediaFiles) != 1 {
				t.Fatalf("workout attachments were not stored: %+v, %v", aliceWorkout, err)
			}
			saveWorkoutForDeletionTest(t, ta, alice, "2026-07-09T10:00:00Z")
			bobWorkoutID := saveWorkoutForDeletionTest(t, ta, bob, "2026-07-08T10:00:00Z")
			ta.app.EquipmentDistance.Wait()
			gear, err := ta.app.Equipment.FindByID("alice", equipmentID)
			if err != nil || gear.Distance <= 0 {
				t.Fatalf("expected equipment mileage before deletion: %+v, %v", gear, err)
			}

			expectStatus(t, ta.doJSON(t, http.MethodPost, "/api/v1/social/follow", map[string]string{"handle": "bob"}, alice), http.StatusCreated)
			for _, reaction := range []struct {
				owner, workoutID, token string
			}{{"alice", aliceWorkoutID, bob}, {"bob", bobWorkoutID, alice}} {
				path := workoutCollection + "/" + reaction.workoutID
				owner := "?owner=" + reaction.owner
				expectStatus(t, ta.doJSON(t, http.MethodPost, path+"/likes"+owner, nil, reaction.token), http.StatusOK)
				expectStatus(t, ta.doJSON(t, http.MethodPost, path+"/comments"+owner, map[string]string{"text": "Great!"}, reaction.token), http.StatusOK)
			}

			w = ta.doJSON(t, http.MethodPost, "/api/v1/auth/pat", map[string]any{
				"name": "cli", "scopes": []string{"workouts:write"},
			}, alice)
			expectStatus(t, w, http.StatusCreated)
			pat := decodeObject(t, w)["token"].(string)
			expectStatus(t, ta.doJSON(t, http.MethodDelete, workoutCollection, password, ""), http.StatusUnauthorized)
			expectStatus(t, ta.doJSON(t, http.MethodDelete, workoutCollection, password, pat), http.StatusUnauthorized)
			expectStatus(t, ta.doJSON(t, http.MethodDelete, workoutCollection, map[string]string{"password": " "}, alice), http.StatusBadRequest)
			expectStatus(t, ta.doJSON(t, http.MethodDelete, workoutCollection, map[string]string{"password": "wrong"}, alice), http.StatusForbidden)
			expectStatus(t, ta.doJSON(t, http.MethodGet, workoutCollection+"/"+aliceWorkoutID, nil, alice), http.StatusOK)

			// Use the same staging and commit operations as the Strava importer.
			var repo workouts.Repository
			switch backend := ta.app.Backend.(type) {
			case *file.Backend:
				repo = backend.WorkoutsRepo()
			case *bbolt.Backend:
				repo = backend.WorkoutsRepo()
			}
			staged, stagedDir, cleanup, err := repo.BeginCreate("alice", &workouts.Workout{
				Name: "Import in progress", SportType: "Run", StartDate: time.Date(2026, 7, 10, 10, 0, 0, 0, time.UTC),
			})
			if err != nil {
				t.Fatal(err)
			}

			expectStatus(t, ta.doJSON(t, http.MethodDelete, workoutCollection, password, alice), http.StatusNoContent)
			expectStatus(t, ta.doJSON(t, http.MethodDelete, workoutCollection, password, alice), http.StatusNoContent)
			expectStatus(t, ta.doJSON(t, http.MethodGet, "/api/v1/auth/me", nil, alice), http.StatusOK)
			expectStatus(t, ta.doJSON(t, http.MethodGet, workoutCollection+"/"+aliceWorkoutID, nil, alice), http.StatusNotFound)
			if _, err := os.Stat(workoutDir); !os.IsNotExist(err) {
				t.Fatalf("workout directory still present: %v", err)
			}
			if exists, err := ta.app.Workouts.HasExternalID("alice", "strava", "123"); err != nil || exists {
				t.Fatalf("deleted import ID still indexed: %v, %v", exists, err)
			}
			w = ta.doJSON(t, http.MethodGet, "/api/v1/profile", nil, alice)
			expectStatus(t, w, http.StatusOK)
			profile := decodeObject(t, w)
			if profile["total_workouts"] != float64(0) || profile["last_sport_type"] != nil || profile["used_sport_types"] != nil {
				t.Fatalf("expected empty workout history, got %#v", profile)
			}
			lastEquipment, _ := profile["last_equipment_by_sport"].(map[string]any)["Run"].([]any)
			if len(lastEquipment) != 1 || lastEquipment[0] != equipmentID {
				t.Fatalf("equipment preference was lost: %#v", profile)
			}
			gear, err = ta.app.Equipment.FindByID("alice", equipmentID)
			if err != nil || gear.Distance != 0 {
				t.Fatalf("expected retained equipment with zero mileage: %+v, %v", gear, err)
			}
			w = ta.doJSON(t, http.MethodGet, "/api/v1/social/following", nil, alice)
			expectStatus(t, w, http.StatusOK)
			if len(decodeList(t, w)) != 1 {
				t.Fatal("follow was removed")
			}
			w = ta.doJSON(t, http.MethodGet, workoutCollection+"/"+bobWorkoutID, nil, bob)
			expectStatus(t, w, http.StatusOK)
			remaining := decodeObject(t, w)
			if remaining["likes_count"] != float64(1) || remaining["comments_count"] != float64(1) {
				t.Fatalf("reactions on the other user's workout were removed: %#v", remaining)
			}
			if _, err := os.Stat(filepath.Join(data.UserDir(ta.dir, "alice"), "workouts", stagedDir)); err != nil {
				cleanup()
				t.Fatalf("staged import was removed: %v", err)
			}
			if err := repo.WriteMetadata("alice", staged); err != nil {
				cleanup()
				t.Fatal(err)
			}
			expectStatus(t, ta.doJSON(t, http.MethodGet, workoutCollection+"/"+staged.ID, nil, alice), http.StatusOK)
		})
	}
}

func TestDeleteAllWorkoutsCanRetryPartialFailure(t *testing.T) {
	if runtime.GOOS == "windows" || os.Geteuid() == 0 {
		t.Skip("requires an unprivileged user and Unix file permissions")
	}
	for _, driver := range []config.StorageDriver{config.StorageDriverFile, config.StorageDriverBBolt} {
		t.Run(string(driver), func(t *testing.T) {
			ta := setupTestAppWithDriver(t, driver)
			ta.register(t, "alice", "alice@example.com", "password12")
			token, _ := ta.login(t, "alice@example.com", "password12")
			password := map[string]string{"password": "password12"}
			saveWorkoutForDeletionTest(t, ta, token, "2026-07-08T10:00:00Z")
			blockedID := saveWorkoutForDeletionTest(t, ta, token, "2026-07-09T10:00:00Z")
			saveWorkoutForDeletionTest(t, ta, token, "2026-07-10T10:00:00Z")
			blockedWorkout, err := ta.app.Workouts.Get("alice", blockedID)
			if err != nil {
				t.Fatal(err)
			}
			blockedDir := filepath.Join(data.UserDir(ta.dir, "alice"), "workouts", keys.WorkoutDirName(blockedWorkout.StartDate, blockedID), "blocked")
			if err := os.MkdirAll(blockedDir, 0700); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(filepath.Join(blockedDir, "attachment"), []byte("data"), 0600); err != nil {
				t.Fatal(err)
			}
			if err := os.Chmod(blockedDir, 0500); err != nil {
				t.Fatal(err)
			}
			t.Cleanup(func() { _ = os.Chmod(blockedDir, 0700) })

			expectStatus(t, ta.doJSON(t, http.MethodDelete, workoutCollection, password, token), http.StatusInternalServerError)
			w := ta.doJSON(t, http.MethodGet, "/api/v1/profile", nil, token)
			expectStatus(t, w, http.StatusOK)
			if profile := decodeObject(t, w); profile["total_workouts"] != float64(2) {
				t.Fatalf("expected two workouts after partial deletion, got %#v", profile)
			}
			expectStatus(t, ta.doJSON(t, http.MethodGet, workoutCollection+"/"+blockedID, nil, token), http.StatusOK)
			if err := os.Chmod(blockedDir, 0700); err != nil {
				t.Fatal(err)
			}
			expectStatus(t, ta.doJSON(t, http.MethodDelete, workoutCollection, password, token), http.StatusNoContent)
			if _, err := ta.app.Workouts.Get("alice", blockedID); !errors.Is(err, workouts.ErrWorkoutNotFound) {
				t.Fatalf("blocked workout survived retry: %v", err)
			}
			if _, err := os.Stat(blockedDir); !os.IsNotExist(err) {
				t.Fatalf("blocked attachment survived retry: %v", err)
			}
		})
	}
}

func TestDeleteAllWorkoutsPublishesWorkoutDeletes(t *testing.T) {
	ta := setupFederationTestApp(t)
	delivery := &captureDeliveryRT{}
	ta.app.SetFederationHTTPClient(&http.Client{Transport: delivery})
	ta.register(t, "alice", "alice@example.com", "password12")
	token, _ := ta.login(t, "alice@example.com", "password12")
	first := saveWorkoutForDeletionTest(t, ta, token, "2026-07-08T10:00:00Z")
	second := saveWorkoutForDeletionTest(t, ta, token, "2026-07-09T10:00:00Z")
	if _, err := ta.app.Federation.Followers().Add("alice", federation.InboundFollower{
		ActorURI: "https://remote.example/users/carol",
		Inbox:    "https://remote.example/users/carol/inbox",
		Handle:   "carol@remote.example",
	}); err != nil {
		t.Fatal(err)
	}
	expectStatus(t, ta.doJSON(t, http.MethodDelete, workoutCollection, map[string]string{"password": "password12"}, token), http.StatusNoContent)
	if err := ta.app.Close(); err != nil {
		t.Fatal(err)
	}
	got := map[string]bool{}
	for _, activity := range delivery.activities() {
		if activity["type"] != "Delete" || activity["actor"] != "https://localhost/users/alice" {
			t.Fatalf("unexpected activity: %#v", activity)
		}
		got[activity["object"].(string)] = true
	}
	const objectPrefix = "https://localhost/users/alice/workouts/"
	if len(got) != 2 || !got[objectPrefix+first] || !got[objectPrefix+second] {
		t.Fatalf("expected a Delete for each workout, got %v", got)
	}
}
