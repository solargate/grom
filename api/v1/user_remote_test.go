package v1_test

import (
	"encoding/base64"
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

func TestRemoteUserWorkoutsMapPreviewAndAvatarFromOutbox(t *testing.T) {
	ta := setupFederationTestApp(t)
	ta.register(t, "alice", "alice@example.com", "password12")
	aliceToken, _ := ta.login(t, "alice@example.com", "password12")

	workoutID := "mapav123"
	objectURL := ""
	pngData := readTestdata(t, "images/avatar-square.png")
	gpx := readTestdata(t, "tracks/1-sample.gpx")
	var remote *httptest.Server
	remote = httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		base := "https://" + r.Host
		switch {
		case strings.Contains(r.URL.Path, "/.well-known/webfinger"):
			w.Header().Set("Content-Type", "application/jrd+json")
			_ = json.NewEncoder(w).Encode(map[string]any{
				"subject": r.URL.Query().Get("resource"),
				"links": []map[string]any{{
					"rel":  "self",
					"type": "application/activity+json",
					"href": base + "/users/bob",
				}},
			})
		case r.URL.Path == "/users/bob":
			w.Header().Set("Content-Type", "application/activity+json")
			_ = json.NewEncoder(w).Encode(map[string]any{
				"@context":          "https://www.w3.org/ns/activitystreams",
				"id":                base + "/users/bob",
				"type":              "Person",
				"preferredUsername": "bob",
				"name":              "Bob Remote",
				"inbox":             base + "/users/bob/inbox",
				"outbox":            base + "/users/bob/outbox",
				"icon": map[string]any{
					"type": "Image",
					"url":  base + "/users/bob/avatar",
				},
			})
		case r.URL.Path == "/users/bob/avatar":
			w.Header().Set("Content-Type", "image/png")
			_, _ = w.Write(pngData)
		case r.URL.Path == "/users/bob/outbox":
			w.Header().Set("Content-Type", "application/activity+json")
			_ = json.NewEncoder(w).Encode(map[string]any{
				"@context":   "https://www.w3.org/ns/activitystreams",
				"id":         base + "/users/bob/outbox",
				"type":       "OrderedCollection",
				"totalItems": 1,
				"orderedItems": []any{
					map[string]any{
						"id":    objectURL + "/activity",
						"type":  "Create",
						"actor": base + "/users/bob",
						"object": map[string]any{
							"id":              objectURL,
							"type":            "Workout",
							"name":            "Remote mapped run",
							"sportType":       "Run",
							"startDate":       time.Date(2026, 7, 8, 10, 0, 0, 0, time.UTC).Format(time.RFC3339),
							"durationSeconds": 1800,
							"distance":        5000.0,
							"track":           "track.gpx",
							"hasMapPreview":   true,
						},
					},
				},
			})
		case r.URL.Path == "/users/bob/workouts/"+workoutID:
			w.Header().Set("Content-Type", "application/activity+json")
			_ = json.NewEncoder(w).Encode(map[string]any{
				"@context":        "https://www.w3.org/ns/activitystreams",
				"id":              objectURL,
				"type":            "Workout",
				"name":            "Remote mapped run",
				"sportType":       "Run",
				"startDate":       time.Date(2026, 7, 8, 10, 0, 0, 0, time.UTC).Format(time.RFC3339),
				"durationSeconds": 1800,
				"distance":        5000.0,
				"track":           "track.gpx",
				"trackData":       base64.StdEncoding.EncodeToString(gpx),
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

	w := ta.doJSON(t, http.MethodGet, "/api/v1/users/"+escaped+"/workouts", nil, aliceToken)
	expectStatus(t, w, http.StatusOK)
	page := decodeObject(t, w)
	items, _ := page["items"].([]any)
	if len(items) != 1 {
		t.Fatalf("expected 1 remote workout, got %#v", page)
	}
	item, _ := items[0].(map[string]any)
	if item["has_map_preview"] != true {
		t.Fatalf("expected has_map_preview from outbox flag: %#v", item)
	}
	gotOID, _ := item["object_id"].(string)
	if gotOID != objectURL {
		t.Fatalf("object_id = %q, want %q", gotOID, objectURL)
	}
	author, _ := item["author"].(map[string]any)
	if author["has_avatar"] != true {
		t.Fatalf("expected author.has_avatar: %#v", author)
	}
	avatarURL, _ := author["avatar_url"].(string)
	if !strings.HasPrefix(avatarURL, "/api/v1/federation/authors/") {
		t.Fatalf("author.avatar_url = %q, want same-origin federation path", avatarURL)
	}
	if author["name"] != "Bob Remote" {
		t.Fatalf("author.name = %#v", author["name"])
	}

	previewPath := "/api/v1/workouts/" + workoutID + "/map-preview?owner=bob&object_id=" + url.QueryEscape(objectURL)
	w = ta.doJSON(t, http.MethodGet, previewPath, nil, aliceToken)
	expectStatus(t, w, http.StatusOK)
	if ct := w.Header().Get("Content-Type"); !strings.Contains(ct, "image/webp") {
		t.Fatalf("map preview content-type = %q", ct)
	}
	if len(w.Body.Bytes()) == 0 {
		t.Fatal("expected non-empty map preview body")
	}
}

func TestRemoteUserSearchAvatarIsSameOrigin(t *testing.T) {
	ta := setupFederationTestApp(t)
	ta.register(t, "alice", "alice@example.com", "password12")
	aliceToken, _ := ta.login(t, "alice@example.com", "password12")

	pngData := readTestdata(t, "images/avatar-square.png")
	var remote *httptest.Server
	remote = httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		base := "https://" + r.Host
		switch {
		case strings.Contains(r.URL.Path, "/.well-known/webfinger"):
			w.Header().Set("Content-Type", "application/jrd+json")
			_ = json.NewEncoder(w).Encode(map[string]any{
				"subject": r.URL.Query().Get("resource"),
				"links": []map[string]any{{
					"rel":  "self",
					"type": "application/activity+json",
					"href": base + "/users/bob",
				}},
			})
		case r.URL.Path == "/users/bob":
			w.Header().Set("Content-Type", "application/activity+json")
			_ = json.NewEncoder(w).Encode(map[string]any{
				"@context":          "https://www.w3.org/ns/activitystreams",
				"id":                base + "/users/bob",
				"type":              "Person",
				"preferredUsername": "bob",
				"name":              "Bob Remote",
				"inbox":             base + "/users/bob/inbox",
				"outbox":            base + "/users/bob/outbox",
				"icon": map[string]any{
					"type": "Image",
					"url":  base + "/users/bob/avatar",
				},
			})
		case r.URL.Path == "/users/bob/avatar":
			w.Header().Set("Content-Type", "image/png")
			_, _ = w.Write(pngData)
		default:
			http.NotFound(w, r)
		}
	}))
	defer remote.Close()

	host := remote.Listener.Addr().String()
	ta.app.SetFederationHTTPClient(remote.Client())

	handle := "bob@" + host
	w := ta.doJSON(t, http.MethodGet, "/api/v1/users/search?q="+url.QueryEscape(handle), nil, aliceToken)
	expectStatus(t, w, http.StatusOK)
	results := decodeList(t, w)
	if len(results) != 1 {
		t.Fatalf("expected 1 search result, got %#v", results)
	}
	got := results[0]
	if got["is_local"] != false {
		t.Fatalf("expected remote: %#v", got)
	}
	if got["has_avatar"] != true {
		t.Fatalf("expected has_avatar: %#v", got)
	}
	avatarURL, _ := got["avatar_url"].(string)
	wantPrefix := "/api/v1/federation/authors/"
	if !strings.HasPrefix(avatarURL, wantPrefix) {
		t.Fatalf("avatar_url = %q, want same-origin prefix %q", avatarURL, wantPrefix)
	}
	if strings.HasPrefix(avatarURL, "http://") || strings.HasPrefix(avatarURL, "https://") {
		t.Fatalf("avatar_url must not be absolute remote URL: %q", avatarURL)
	}

	escaped := url.PathEscape(handle)
	w = ta.doJSON(t, http.MethodGet, "/api/v1/users/"+escaped, nil, aliceToken)
	expectStatus(t, w, http.StatusOK)
	profile := decodeObject(t, w)
	profileAvatar, _ := profile["avatar_url"].(string)
	if !strings.HasPrefix(profileAvatar, wantPrefix) {
		t.Fatalf("profile avatar_url = %q, want same-origin prefix %q", profileAvatar, wantPrefix)
	}
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

func TestRemoteUserWorkoutsPagination(t *testing.T) {
	ta := setupFederationTestApp(t)
	ta.register(t, "alice", "alice@example.com", "password12")
	aliceToken, _ := ta.login(t, "alice@example.com", "password12")

	idNewer := "newer001"
	idOlder := "older002"
	objectNewer := ""
	objectOlder := ""
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
			})
		case r.URL.Path == "/users/bob/outbox":
			_ = json.NewEncoder(w).Encode(map[string]any{
				"@context":   "https://www.w3.org/ns/activitystreams",
				"id":         base + "/users/bob/outbox",
				"type":       "OrderedCollection",
				"totalItems": 2,
				"orderedItems": []any{
					map[string]any{
						"id":   objectNewer + "/activity",
						"type": "Create",
						"actor": base + "/users/bob",
						"object": map[string]any{
							"id":              objectNewer,
							"type":            "Workout",
							"name":            "Newer run",
							"sportType":       "Run",
							"startDate":       time.Date(2026, 7, 9, 10, 0, 0, 0, time.UTC).Format(time.RFC3339),
							"durationSeconds": 1800,
							"distance":        5000.0,
						},
					},
					map[string]any{
						"id":   objectOlder + "/activity",
						"type": "Create",
						"actor": base + "/users/bob",
						"object": map[string]any{
							"id":              objectOlder,
							"type":            "Workout",
							"name":            "Older run",
							"sportType":       "Run",
							"startDate":       time.Date(2026, 7, 8, 10, 0, 0, 0, time.UTC).Format(time.RFC3339),
							"durationSeconds": 1800,
							"distance":        4000.0,
						},
					},
				},
			})
		default:
			http.NotFound(w, r)
		}
	}))
	defer remote.Close()

	host := remote.Listener.Addr().String()
	objectNewer = "https://" + host + "/users/bob/workouts/" + idNewer
	objectOlder = "https://" + host + "/users/bob/workouts/" + idOlder
	ta.app.SetFederationHTTPClient(remote.Client())

	escaped := url.PathEscape("bob@" + host)
	w := ta.doJSON(t, http.MethodGet, "/api/v1/users/"+escaped+"/workouts?limit=1", nil, aliceToken)
	expectStatus(t, w, http.StatusOK)
	page := decodeObject(t, w)
	items, _ := page["items"].([]any)
	if len(items) != 1 {
		t.Fatalf("page1 items: %#v", page)
	}
	first, _ := items[0].(map[string]any)
	if first["name"] != "Newer run" {
		t.Fatalf("want newer first: %#v", first)
	}
	if page["has_more"] != true {
		t.Fatalf("expected has_more: %#v", page)
	}
	cursor, _ := page["next_cursor"].(string)
	if cursor == "" {
		t.Fatalf("expected next_cursor: %#v", page)
	}

	w = ta.doJSON(t, http.MethodGet, "/api/v1/users/"+escaped+"/workouts?limit=1&cursor="+url.QueryEscape(cursor), nil, aliceToken)
	expectStatus(t, w, http.StatusOK)
	page2 := decodeObject(t, w)
	items2, _ := page2["items"].([]any)
	if len(items2) != 1 {
		t.Fatalf("page2 items: %#v", page2)
	}
	second, _ := items2[0].(map[string]any)
	if second["name"] != "Older run" {
		t.Fatalf("want older on page2: %#v", second)
	}
	if second["object_id"] == first["object_id"] {
		t.Fatalf("duplicate across pages: %#v %#v", first, second)
	}
}

func TestRemoteUserWorkoutLiveTrackAndSpeed(t *testing.T) {
	ta := setupFederationTestApp(t)
	ta.register(t, "alice", "alice@example.com", "password12")
	aliceToken, _ := ta.login(t, "alice@example.com", "password12")

	workoutID := "livefetch"
	objectURL := ""
	gpx := readTestdata(t, "tracks/1-sample.gpx")
	var remote *httptest.Server
	remote = httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		base := "https://" + r.Host
		switch {
		case strings.Contains(r.URL.Path, "/.well-known/webfinger"):
			w.Header().Set("Content-Type", "application/jrd+json")
			_ = json.NewEncoder(w).Encode(map[string]any{
				"subject": r.URL.Query().Get("resource"),
				"links": []map[string]any{{
					"rel":  "self",
					"type": "application/activity+json",
					"href": base + "/users/bob",
				}},
			})
		case r.URL.Path == "/users/bob":
			w.Header().Set("Content-Type", "application/activity+json")
			_ = json.NewEncoder(w).Encode(map[string]any{
				"@context":          "https://www.w3.org/ns/activitystreams",
				"id":                base + "/users/bob",
				"type":              "Person",
				"preferredUsername": "bob",
				"name":              "Bob Remote",
				"inbox":             base + "/users/bob/inbox",
				"outbox":            base + "/users/bob/outbox",
			})
		case r.URL.Path == "/users/bob/outbox":
			w.Header().Set("Content-Type", "application/activity+json")
			_ = json.NewEncoder(w).Encode(map[string]any{
				"@context":   "https://www.w3.org/ns/activitystreams",
				"id":         base + "/users/bob/outbox",
				"type":       "OrderedCollection",
				"totalItems": 1,
				"orderedItems": []any{
					map[string]any{
						"id":    objectURL + "/activity",
						"type":  "Create",
						"actor": base + "/users/bob",
						"object": map[string]any{
							"id":              objectURL,
							"type":            "Workout",
							"name":            "Remote GPX run",
							"sportType":       "Run",
							"startDate":       time.Date(2026, 7, 8, 10, 0, 0, 0, time.UTC).Format(time.RFC3339),
							"durationSeconds": 1800,
							"distance":        5000.0,
							"track":           "track.gpx",
						},
					},
				},
			})
		case r.URL.Path == "/users/bob/workouts/"+workoutID:
			w.Header().Set("Content-Type", "application/activity+json")
			_ = json.NewEncoder(w).Encode(map[string]any{
				"@context":        "https://www.w3.org/ns/activitystreams",
				"id":              objectURL,
				"type":            "Workout",
				"name":            "Remote GPX run",
				"sportType":       "Run",
				"startDate":       time.Date(2026, 7, 8, 10, 0, 0, 0, time.UTC).Format(time.RFC3339),
				"durationSeconds": 1800,
				"distance":        5000.0,
				"track":           "track.gpx",
				"trackData":       base64.StdEncoding.EncodeToString(gpx),
			})
		default:
			http.NotFound(w, r)
		}
	}))
	defer remote.Close()

	host := remote.Listener.Addr().String()
	objectURL = "https://" + host + "/users/bob/workouts/" + workoutID
	ta.app.SetFederationHTTPClient(remote.Client())

	escaped := url.PathEscape("bob@" + host)
	w := ta.doJSON(t, http.MethodGet, "/api/v1/users/"+escaped+"/workouts", nil, aliceToken)
	expectStatus(t, w, http.StatusOK)
	page := decodeObject(t, w)
	items, _ := page["items"].([]any)
	if len(items) != 1 {
		t.Fatalf("expected 1 workout: %#v", page)
	}
	item, _ := items[0].(map[string]any)
	gotOID, _ := item["object_id"].(string)
	if gotOID != objectURL {
		t.Fatalf("object_id = %q", gotOID)
	}

	trackPath := "/api/v1/workouts/" + workoutID + "/track?owner=bob&format=gpx&object_id=" + url.QueryEscape(objectURL)
	w = ta.doJSON(t, http.MethodGet, trackPath, nil, aliceToken)
	expectStatus(t, w, http.StatusOK)
	if len(w.Body.Bytes()) == 0 {
		t.Fatal("expected GPX track body")
	}
	ct := w.Header().Get("Content-Type")
	if !strings.Contains(ct, "gpx") && !strings.Contains(ct, "xml") {
		t.Fatalf("track content-type = %q", ct)
	}

	speedPath := "/api/v1/workouts/" + workoutID + "/speed?owner=bob&object_id=" + url.QueryEscape(objectURL)
	w = ta.doJSON(t, http.MethodGet, speedPath, nil, aliceToken)
	expectStatus(t, w, http.StatusOK)
	speed := decodeObject(t, w)
	samples, _ := speed["samples"].([]any)
	if len(samples) < 2 {
		t.Fatalf("expected live speed samples: %#v", speed)
	}

	badPath := "/api/v1/workouts/" + workoutID + "/speed?owner=bob&object_id=" + url.QueryEscape("https://"+host+"/users/bob/workouts/missing")
	w = ta.doJSON(t, http.MethodGet, badPath, nil, aliceToken)
	expectStatus(t, w, http.StatusNotFound)
}

func TestLocalFollowersFollowingCollections(t *testing.T) {
	ta := setupFederationTestAppDefaultAF(t)
	ta.register(t, "alice", "alice@example.com", "password12")
	ta.register(t, "bob", "bob@example.com", "password12")
	aliceToken, _ := ta.login(t, "alice@example.com", "password12")
	bobToken, _ := ta.login(t, "bob@example.com", "password12")
	ta.installInstanceActorKey(t)

	w := ta.doJSON(t, http.MethodPost, "/api/v1/social/follow", map[string]string{"handle": "alice"}, bobToken)
	expectStatus(t, w, http.StatusCreated)
	w = ta.doJSON(t, http.MethodPost, "/api/v1/social/follow", map[string]string{"handle": "bob"}, aliceToken)
	expectStatus(t, w, http.StatusCreated)

	w = ta.signedInstanceGET(t, "/users/alice/followers")
	expectStatus(t, w, http.StatusOK)
	followers := decodeObject(t, w)
	if followers["type"] != "OrderedCollection" {
		t.Fatalf("followers: %#v", followers)
	}
	ordered, _ := followers["orderedItems"].([]any)
	if len(ordered) < 1 {
		t.Fatalf("expected follower actor URI: %#v", followers)
	}
	uri, _ := ordered[0].(string)
	if !strings.Contains(uri, "/users/bob") {
		t.Fatalf("expected bob actor URI, got %q", uri)
	}

	w = ta.signedInstanceGET(t, "/users/alice/following")
	expectStatus(t, w, http.StatusOK)
	following := decodeObject(t, w)
	if following["type"] != "OrderedCollection" {
		t.Fatalf("following: %#v", following)
	}
	ordered, _ = following["orderedItems"].([]any)
	if len(ordered) < 1 {
		t.Fatalf("expected following actor URI: %#v", following)
	}
	uri, _ = ordered[0].(string)
	if !strings.Contains(uri, "/users/bob") {
		t.Fatalf("expected bob in following, got %q", uri)
	}

	req := httptest.NewRequest(http.MethodGet, "/users/alice/followers", nil)
	req.Header.Set("Accept", "application/activity+json")
	w = httptest.NewRecorder()
	ta.router.ServeHTTP(w, req)
	expectStatus(t, w, http.StatusUnauthorized)
}
