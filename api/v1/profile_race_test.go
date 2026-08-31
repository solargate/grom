package v1_test

import (
	"net/http"
	"sync"
	"testing"
	"time"
)

func TestProfileUsedSportTypesRaceRefreshAndTouch(t *testing.T) {
	ta := setupTestApp(t)
	ta.register(t, "alice", "alice@example.com", "password12")
	token, user := ta.login(t, "alice@example.com", "password12")
	nickname, _ := user["nickname"].(string)
	userID, _ := user["id"].(string)
	if nickname == "" || userID == "" {
		t.Fatal("missing user fields")
	}

	// Seed one sport type via workout create.
	w := ta.doJSON(t, http.MethodPost, "/api/v1/workouts", map[string]any{
		"name":             "Seed run",
		"sport_type":       "Run",
		"start_date":       time.Date(2026, 8, 1, 10, 0, 0, 0, time.UTC).Format(time.RFC3339),
		"duration_seconds": 1800,
		"distance":         5000,
	}, token)
	expectStatus(t, w, http.StatusCreated)

	const workers = 24
	var wg sync.WaitGroup
	wg.Add(workers * 2)
	for i := 0; i < workers; i++ {
		go func(n int) {
			defer wg.Done()
			sport := "Walk"
			if n%2 == 0 {
				sport = "Ride"
			}
			body := map[string]any{
				"name":             "Concurrent",
				"sport_type":       sport,
				"start_date":       time.Date(2026, 8, 2, 10, 0, 0, 0, time.UTC).Add(time.Duration(n) * time.Minute).Format(time.RFC3339),
				"duration_seconds": 1200,
				"distance":         3000,
			}
			_ = ta.doJSON(t, http.MethodPost, "/api/v1/workouts", body, token)
		}(i)
		go func() {
			defer wg.Done()
			_ = ta.app.RefreshLastSportType(nickname, userID)
		}()
	}
	wg.Wait()

	w = ta.doJSON(t, http.MethodGet, "/api/v1/profile", nil, token)
	expectStatus(t, w, http.StatusOK)
	profile := decodeObject(t, w)
	raw, _ := profile["used_sport_types"].([]any)
	got := make([]string, 0, len(raw))
	for _, item := range raw {
		if s, ok := item.(string); ok {
			got = append(got, s)
		}
	}
	has := func(sport string) bool {
		for _, s := range got {
			if s == sport {
				return true
			}
		}
		return false
	}
	if !has("Run") || !has("Walk") || !has("Ride") {
		t.Fatalf("used_sport_types missing concurrent sports: %#v", got)
	}
}
