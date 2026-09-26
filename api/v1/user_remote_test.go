package v1_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"
)

func TestRemoteUserProfileWorkoutsViaOutbox(t *testing.T) {
	ta := setupFederationTestApp(t)
	ta.register(t, "alice", "alice@example.com", "password12")
	aliceToken, _ := ta.login(t, "alice@example.com", "password12")

	workoutID := "abcd1234"
	objectURL := ""
	var remote *httptest.Server
	remote = httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/activity+json")
		base := "https://" + r.Host
		switch {
		case strings.Contains(r.URL.Path, "/.well-known/webfinger"):
			_ = json.NewEncoder(w).Encode(map[string]any{
				"subject": r.URL.Query().Get("resource"),
				"links": []map[string]any{{
					"rel":  "self",
					"type": "application/activity+json",
					"href": base + "/users/bob",
				}},
			})
		case r.URL.Path == "/users/bob":
			_ = json.NewEncoder(w).Encode(map[string]any{
				"@context":          "https://www.w3.org/ns/activitystreams",
				"id":                base + "/users/bob",
				"type":              "Person",
				"preferredUsername": "bob",
				"name":              "Bob Remote",
				"inbox":             base + "/users/bob/inbox",
				"outbox":            base + "/users/bob/outbox",
				"followers":         base + "/users/bob/followers",
				"following":         base + "/users/bob/following",
			})
		case r.URL.Path == "/users/bob/outbox":
			_ = json.NewEncoder(w).Encode(map[string]any{
				"@context":   "https://www.w3.org/ns/activitystreams",
				"id":         base + "/users/bob/outbox",
				"type":       "OrderedCollection",
				"totalItems": 1,
				"orderedItems": []any{
					map[string]any{
						"id":   objectURL + "/activity",
						"type": "Create",
						"actor": base + "/users/bob",
						"object": map[string]any{
							"id":              objectURL,
							"type":            "Workout",
							"name":            "Remote run",
							"sportType":       "Run",
							"startDate":       time.Date(2026, 7, 8, 10, 0, 0, 0, time.UTC).Format(time.RFC3339),
							"durationSeconds": 1800,
							"distance":        5000.0,
						},
					},
				},
			})
		case r.URL.Path == "/users/bob/workouts/"+workoutID:
			_ = json.NewEncoder(w).Encode(map[string]any{
				"@context":        "https://www.w3.org/ns/activitystreams",
				"id":              objectURL,
				"type":            "Workout",
				"name":            "Remote run",
				"sportType":       "Run",
				"startDate":       time.Date(2026, 7, 8, 10, 0, 0, 0, time.UTC).Format(time.RFC3339),
				"durationSeconds": 1800,
				"distance":        5000.0,
			})
		case r.URL.Path == "/users/bob/following":
			_ = json.NewEncoder(w).Encode(map[string]any{
				"@context":     "https://www.w3.org/ns/activitystreams",
				"id":           base + "/users/bob/following",
				"type":         "OrderedCollection",
				"totalItems":   1,
				"orderedItems": []any{base + "/users/carol"},
			})
		case r.URL.Path == "/users/bob/followers":
			_ = json.NewEncoder(w).Encode(map[string]any{
				"@context":     "https://www.w3.org/ns/activitystreams",
				"id":           base + "/users/bob/followers",
				"type":         "OrderedCollection",
				"totalItems":   0,
				"orderedItems": []any{},
			})
		default:
			http.NotFound(w, r)
		}
	}))
	defer remote.Close()

	host := remote.Listener.Addr().String()
	objectURL = "https://" + host + "/users/bob/workouts/" + workoutID
	ta.app.SetFederationHTTPClient(remote.Client())

	handle := "bob@" + host
	escaped := url.PathEscape(handle)

	w := ta.doJSON(t, http.MethodGet, "/api/v1/users/"+escaped, nil, aliceToken)
	expectStatus(t, w, http.StatusOK)
	profile := decodeObject(t, w)
	if profile["nickname"] != "bob" || profile["is_local"] != false {
		t.Fatalf("unexpected remote profile: %#v", profile)
	}
	if profile["name"] != "Bob Remote" {
		t.Fatalf("expected Bob Remote name: %#v", profile)
	}

	w = ta.doJSON(t, http.MethodGet, "/api/v1/users/"+escaped+"/workouts", nil, aliceToken)
	expectStatus(t, w, http.StatusOK)
	page := decodeObject(t, w)
	items, _ := page["items"].([]any)
	if len(items) != 1 {
		t.Fatalf("expected 1 remote workout, got %#v", page)
	}
	item, _ := items[0].(map[string]any)
	if item["name"] != "Remote run" {
		t.Fatalf("unexpected workout: %#v", item)
	}
	gotOID, _ := item["object_id"].(string)
	if gotOID != objectURL {
		t.Fatalf("object_id = %q, want %q", gotOID, objectURL)
	}
	if item["can_like"] != false {
		t.Fatalf("expected can_like false without follow: %#v", item)
	}

	w = ta.doJSON(t, http.MethodGet, "/api/v1/users/"+escaped+"/following", nil, aliceToken)
	expectStatus(t, w, http.StatusOK)
	following := decodeList(t, w)
	if len(following) != 1 {
		t.Fatalf("expected 1 following, got %#v", following)
	}

	w = ta.doJSON(t, http.MethodGet, "/api/v1/users/"+escaped+"/followers", nil, aliceToken)
	expectStatus(t, w, http.StatusOK)
}

func TestLocalOutboxListsCreateWorkouts(t *testing.T) {
	ta := setupFederationTestApp(t)
	ta.register(t, "alice", "alice@example.com", "password12")
	aliceToken, _ := ta.login(t, "alice@example.com", "password12")

	w := ta.doJSON(t, http.MethodPost, "/api/v1/workouts", map[string]any{
		"name": "Alice run", "sport_type": "Run", "start_date": "2026-07-08T10:00:00Z",
		"duration_seconds": 1800, "distance": 5000,
	}, aliceToken)
	expectStatus(t, w, http.StatusCreated)
	id, _ := decodeObject(t, w)["id"].(string)
	if id == "" {
		t.Fatal("expected workout id")
	}

	w = ta.doJSON(t, http.MethodGet, "/users/alice/outbox", nil, "")
	expectStatus(t, w, http.StatusOK)
	outbox := decodeObject(t, w)
	if outbox["type"] != "OrderedCollection" {
		t.Fatalf("unexpected outbox: %#v", outbox)
	}
	if outbox["totalItems"] != float64(1) {
		t.Fatalf("expected totalItems=1: %#v", outbox)
	}
	ordered, _ := outbox["orderedItems"].([]any)
	if len(ordered) != 1 {
		t.Fatalf("expected 1 Create: %#v", outbox)
	}
	act, _ := ordered[0].(map[string]any)
	if act["type"] != "Create" {
		t.Fatalf("expected Create activity: %#v", act)
	}
	obj, _ := act["object"].(map[string]any)
	if obj["type"] != "Workout" || obj["name"] != "Alice run" {
		t.Fatalf("unexpected object: %#v", obj)
	}

	req := httptest.NewRequest(http.MethodGet, "/users/alice/workouts/"+id, nil)
	req.Header.Set("Accept", "application/activity+json")
	w = httptest.NewRecorder()
	ta.router.ServeHTTP(w, req)
	expectStatus(t, w, http.StatusOK)
	pub := decodeObject(t, w)
	if pub["type"] != "Workout" || pub["name"] != "Alice run" {
		t.Fatalf("unexpected public workout object: %#v", pub)
	}
}
