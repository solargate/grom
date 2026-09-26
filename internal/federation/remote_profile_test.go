package federation

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/solargate/grom/internal/config"
	"github.com/solargate/grom/internal/social"
)

func TestFetchOutboxWorkouts(t *testing.T) {
	prev := config.Cfg
	t.Cleanup(func() { config.Cfg = prev })
	config.Cfg.Federation.Enabled = true
	config.Cfg.Federation.Domain = "grom.test"

	workoutID := "xyz98765"
	var objectURL string
	mux := http.NewServeMux()
	mux.HandleFunc("/.well-known/webfinger", func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{
			"links": []map[string]any{{
				"rel":  "self",
				"href": "https://" + r.Host + "/users/bob",
			}},
		})
	})
	mux.HandleFunc("/users/bob", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/activity+json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"type":   "Person",
			"name":   "Bob",
			"outbox": "https://" + r.Host + "/users/bob/outbox",
		})
	})
	mux.HandleFunc("/users/bob/outbox", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/activity+json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"type": "OrderedCollection",
			"orderedItems": []any{
				map[string]any{
					"type": "Create",
					"object": map[string]any{
						"id":              objectURL,
						"type":            "Workout",
						"name":            "Trail",
						"sportType":       "Hike",
						"startDate":       time.Date(2026, 1, 2, 3, 0, 0, 0, time.UTC).Format(time.RFC3339),
						"durationSeconds": 3600,
						"distance":        10000.0,
					},
				},
				map[string]any{
					"type": "Announce",
					"object": "https://" + r.Host + "/users/other/note/1",
				},
			},
		})
	})
	server := httptest.NewTLSServer(mux)
	defer server.Close()

	host := server.Listener.Addr().String()
	objectURL = "https://" + host + "/users/bob/workouts/" + workoutID

	parsed := social.ParsedHandle{
		Nickname: "bob",
		Domain:   host,
		Handle:   "bob@" + host,
	}
	items, err := FetchOutboxWorkouts(server.Client(), nil, parsed, 10)
	if err != nil {
		t.Fatalf("FetchOutboxWorkouts: %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("expected 1 workout (Announce skipped), got %#v", items)
	}
	if items[0].Workout.Name != "Trail" || items[0].ObjectID != objectURL {
		t.Fatalf("unexpected item: %#v", items[0])
	}
	if !strings.HasSuffix(items[0].Workout.ID, workoutID) && items[0].Workout.ID != workoutID {
		t.Fatalf("unexpected workout id: %q", items[0].Workout.ID)
	}
}
