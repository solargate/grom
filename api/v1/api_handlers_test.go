package v1_test

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"image"
	"image/png"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	v1 "github.com/solargate/grom/api/v1"
	"github.com/solargate/grom/internal/workouts"
)

func TestServerInfoAndStatus(t *testing.T) {
	ta := setupTestApp(t)

	w := ta.doJSON(t, http.MethodGet, "/api/v1/status", nil, "")
	expectStatus(t, w, http.StatusOK)
	status := decodeObject(t, w)
	if status["message"] != "OK" {
		t.Fatalf("unexpected status: %#v", status)
	}

	w = ta.doJSON(t, http.MethodGet, "/api/v1/server-info", nil, "")
	expectStatus(t, w, http.StatusOK)
	info := decodeObject(t, w)
	if info["federation_enabled"] != false {
		t.Fatalf("expected federation_enabled=false, got %#v", info)
	}
}

func TestUpdateMeAndDeleteAvatar(t *testing.T) {
	ta := setupTestApp(t)
	ta.register(t, "alice", "alice@example.com", "password12")
	token, _ := ta.login(t, "alice@example.com", "password12")

	w := ta.doJSON(t, http.MethodPatch, "/api/v1/auth/me", map[string]string{
		"name": "Alice Updated",
	}, token)
	expectStatus(t, w, http.StatusOK)
	me := decodeObject(t, w)
	if me["name"] != "Alice Updated" {
		t.Fatalf("unexpected name: %#v", me)
	}

	pngData := readTestdata(t, "images/avatar-square.png")
	w = ta.doMultipart(t, http.MethodPut, "/api/v1/auth/me/avatar", token, nil,
		map[string][]filePart{
			"avatar": {{filename: "avatar.png", data: pngData}},
		},
	)
	expectStatus(t, w, http.StatusOK)

	w = ta.doJSON(t, http.MethodDelete, "/api/v1/auth/me/avatar", nil, token)
	expectStatus(t, w, http.StatusOK)
	me = decodeObject(t, w)
	if me["has_avatar"] != false {
		t.Fatalf("expected avatar cleared: %#v", me)
	}

	w = ta.doJSON(t, http.MethodGet, "/api/v1/users/alice/avatar", nil, token)
	expectStatus(t, w, http.StatusNotFound)
}

func TestEquipmentListAndUpdate(t *testing.T) {
	ta := setupTestApp(t)
	ta.register(t, "alice", "alice@example.com", "password12")
	token, _ := ta.login(t, "alice@example.com", "password12")

	w := ta.doJSON(t, http.MethodPost, "/api/v1/equipment", map[string]any{
		"type":      "bike",
		"name":      "Gravel",
		"bike_type": "gravel",
		"brand":     "Canyon",
	}, token)
	expectStatus(t, w, http.StatusCreated)
	created := decodeObject(t, w)
	id, _ := created["id"].(string)

	w = ta.doJSON(t, http.MethodGet, "/api/v1/equipment", nil, token)
	expectStatus(t, w, http.StatusOK)
	items := decodeList(t, w)
	if len(items) != 1 || items[0]["id"] != id {
		t.Fatalf("unexpected list: %#v", items)
	}

	w = ta.doJSON(t, http.MethodPut, "/api/v1/equipment/"+id, map[string]any{
		"type":      "bike",
		"name":      "Road bike",
		"bike_type": "road",
		"brand":     "Canyon",
	}, token)
	expectStatus(t, w, http.StatusOK)
	updated := decodeObject(t, w)
	if updated["name"] != "Road bike" || updated["bike_type"] != "road" {
		t.Fatalf("unexpected update: %#v", updated)
	}
}

func TestSocialFollowingAndFollowers(t *testing.T) {
	ta := setupTestApp(t)
	ta.register(t, "alice", "alice@example.com", "password12")
	ta.register(t, "bob", "bob@example.com", "password12")
	aliceToken, _ := ta.login(t, "alice@example.com", "password12")
	bobToken, _ := ta.login(t, "bob@example.com", "password12")

	w := ta.doJSON(t, http.MethodPost, "/api/v1/social/follow", map[string]string{
		"handle": "bob",
	}, aliceToken)
	expectStatus(t, w, http.StatusCreated)

	w = ta.doJSON(t, http.MethodGet, "/api/v1/social/following", nil, aliceToken)
	expectStatus(t, w, http.StatusOK)
	following := decodeList(t, w)
	if len(following) != 1 || following[0]["target_nickname"] != "bob" {
		t.Fatalf("unexpected following: %#v", following)
	}

	w = ta.doJSON(t, http.MethodGet, "/api/v1/social/followers", nil, bobToken)
	expectStatus(t, w, http.StatusOK)
	followers := decodeList(t, w)
	if len(followers) != 1 || followers[0]["follower_nickname"] != "alice" {
		t.Fatalf("unexpected followers: %#v", followers)
	}
}

func TestWorkoutMapPreviewAndMedia(t *testing.T) {
	ta := setupTestApp(t)
	ta.register(t, "alice", "alice@example.com", "password12")
	token, _ := ta.login(t, "alice@example.com", "password12")

	gpx := readTestdata(t, "tracks/1-sample.gpx")
	var photoBuf bytes.Buffer
	img := image.NewRGBA(image.Rect(0, 0, 32, 32))
	if err := png.Encode(&photoBuf, img); err != nil {
		t.Fatal(err)
	}

	w := ta.doMultipart(t, http.MethodPost, "/api/v1/workouts", token,
		map[string]string{
			"name":       "Photo run",
			"sport_type": "Run",
			"start_date": "2026-07-08T10:00:00Z",
		},
		map[string][]filePart{
			"track":  {{filename: "sample.gpx", data: gpx}},
			"photos": {{filename: "shot.png", data: photoBuf.Bytes()}},
		},
	)
	expectStatus(t, w, http.StatusCreated)
	created := decodeObject(t, w)
	id, _ := created["id"].(string)
	if created["has_map_preview"] != true || created["has_media"] != true {
		t.Fatalf("expected map and media flags: %#v", created)
	}
	mediaFiles, _ := created["media_files"].([]any)
	if len(mediaFiles) != 1 {
		t.Fatalf("expected one media file: %#v", created["media_files"])
	}
	filename, _ := mediaFiles[0].(string)

	w = ta.doJSON(t, http.MethodGet, "/api/v1/workouts/"+id+"/map-preview", nil, token)
	expectStatus(t, w, http.StatusOK)
	if ct := w.Header().Get("Content-Type"); ct != "image/webp" {
		t.Fatalf("map preview content-type = %q", ct)
	}
	if len(w.Body.Bytes()) == 0 {
		t.Fatal("expected map preview bytes")
	}

	w = ta.doJSON(t, http.MethodGet, "/api/v1/workouts/"+id+"/media/"+filename+"/preview", nil, token)
	expectStatus(t, w, http.StatusOK)
	if ct := w.Header().Get("Content-Type"); ct != "image/webp" {
		t.Fatalf("media preview content-type = %q", ct)
	}

	w = ta.doJSON(t, http.MethodGet, "/api/v1/workouts/"+id+"/media/"+filename, nil, token)
	expectStatus(t, w, http.StatusOK)
	if len(w.Body.Bytes()) == 0 {
		t.Fatal("expected original media bytes")
	}

	ta.register(t, "bob", "bob@example.com", "password12")
	bobToken, _ := ta.login(t, "bob@example.com", "password12")

	w = ta.doJSON(t, http.MethodGet, "/api/v1/workouts/"+id+"/media/"+filename+"/preview?owner=alice", nil, bobToken)
	expectStatus(t, w, http.StatusNotFound)
	w = ta.doJSON(t, http.MethodGet, "/api/v1/workouts/"+id+"/media/"+filename+"?owner=alice", nil, bobToken)
	expectStatus(t, w, http.StatusNotFound)

	w = ta.doJSON(t, http.MethodPost, "/api/v1/social/follow", map[string]string{"handle": "alice"}, bobToken)
	expectStatus(t, w, http.StatusCreated)

	w = ta.doJSON(t, http.MethodGet, "/api/v1/workouts/"+id+"/media/"+filename+"/preview?owner=alice", nil, bobToken)
	expectStatus(t, w, http.StatusOK)
	w = ta.doJSON(t, http.MethodGet, "/api/v1/workouts/"+id+"/media/"+filename+"?owner=alice", nil, bobToken)
	expectStatus(t, w, http.StatusOK)
}

func TestAPIDocsRedirect(t *testing.T) {
	ta := setupTestApp(t)
	v1.RegisterAPIDocs(ta.router)

	w := ta.doJSON(t, http.MethodGet, "/api/docs", nil, "")
	if w.Code != http.StatusMovedPermanently {
		t.Fatalf("status = %d, want %d body=%s", w.Code, http.StatusMovedPermanently, w.Body.String())
	}
	if loc := w.Header().Get("Location"); loc != "/api/docs/" {
		t.Fatalf("Location = %q", loc)
	}
}

func TestFederationInboxCreateWithTrackAndUpdate(t *testing.T) {
	ta := setupFederationTestApp(t)
	ta.register(t, "alice", "alice@example.com", "password12")
	token, _ := ta.login(t, "alice@example.com", "password12")

	w := ta.doJSON(t, http.MethodGet, "/.well-known/webfinger?resource=acct:missing@localhost", nil, "")
	expectStatus(t, w, http.StatusNotFound)

	req := httptest.NewRequest(http.MethodGet, "/users/missing", nil)
	req.Header.Set("Accept", "application/activity+json")
	w = httptest.NewRecorder()
	ta.router.ServeHTTP(w, req)
	expectStatus(t, w, http.StatusNotFound)

	gpx := readTestdata(t, "tracks/1-sample.gpx")
	createObj := map[string]any{
		"@context": "https://www.w3.org/ns/activitystreams",
		"type":     "Create",
		"actor":    "https://remote.example/users/bob",
		"object": map[string]any{
			"id":              "https://remote.example/users/bob/workouts/22222222",
			"type":            "Note",
			"name":            "Remote GPX",
			"sportType":       "Run",
			"startDate":       "2026-07-08T10:00:00Z",
			"durationSeconds": 1200,
			"distance":        3000.0,
			"track":           "track.gpx",
			"trackData":       base64.StdEncoding.EncodeToString(gpx),
		},
	}
	data, err := json.Marshal(createObj)
	if err != nil {
		t.Fatal(err)
	}
	req = httptest.NewRequest(http.MethodPost, "/users/alice/inbox", bytes.NewReader(data))
	req.Header.Set("Content-Type", "application/activity+json")
	w = httptest.NewRecorder()
	ta.router.ServeHTTP(w, req)
	expectStatus(t, w, http.StatusAccepted)

	_, samples, err := ta.app.Federation.Inbox().GetSpeedChart("alice", "bob", "22222222")
	if err != nil {
		t.Fatalf("inbox GetSpeedChart: %v", err)
	}
	if len(samples) < 1 {
		t.Fatalf("expected federated speed chart samples, got %d", len(samples))
	}

	w = ta.doJSON(t, http.MethodGet, "/api/v1/workouts?scope=feed&limit=20", nil, token)
	expectStatus(t, w, http.StatusOK)
	items, _, _ := decodeWorkoutPage(t, w)
	found := false
	var remoteID string
	for _, item := range items {
		if item["name"] == "Remote GPX" {
			found = true
			remoteID, _ = item["id"].(string)
			break
		}
	}
	if !found || remoteID == "" {
		t.Fatalf("expected federated workout in feed: %#v", items)
	}
	if remoteID != "22222222" {
		t.Fatalf("remote id = %q", remoteID)
	}

	updateObj := map[string]any{
		"@context": "https://www.w3.org/ns/activitystreams",
		"type":     "Update",
		"actor":    "https://remote.example/users/bob",
		"object": map[string]any{
			"id":              "https://remote.example/users/bob/workouts/22222222",
			"type":            "Note",
			"name":            "Remote GPX edited",
			"sportType":       "Run",
			"startDate":       "2026-07-08T10:00:00Z",
			"durationSeconds": 1300,
			"distance":        3100.0,
			"track":           "track.gpx",
			"trackData":       base64.StdEncoding.EncodeToString(gpx),
		},
	}
	data, err = json.Marshal(updateObj)
	if err != nil {
		t.Fatal(err)
	}
	req = httptest.NewRequest(http.MethodPost, "/users/alice/inbox", bytes.NewReader(data))
	req.Header.Set("Content-Type", "application/activity+json")
	w = httptest.NewRecorder()
	ta.router.ServeHTTP(w, req)
	expectStatus(t, w, http.StatusAccepted)

	got, err := ta.app.Federation.Inbox().Get("alice", "bob", "22222222")
	if err != nil {
		t.Fatalf("inbox Get after update: %v", err)
	}
	if got.Name != "Remote GPX edited" {
		t.Fatalf("updated name = %q", got.Name)
	}

	deleteObj := map[string]any{
		"@context": "https://www.w3.org/ns/activitystreams",
		"type":     "Delete",
		"actor":    "https://remote.example/users/bob",
		"object":   "https://remote.example/users/bob/workouts/22222222",
	}
	data, err = json.Marshal(deleteObj)
	if err != nil {
		t.Fatal(err)
	}
	req = httptest.NewRequest(http.MethodPost, "/users/alice/inbox", bytes.NewReader(data))
	req.Header.Set("Content-Type", "application/activity+json")
	w = httptest.NewRecorder()
	ta.router.ServeHTTP(w, req)
	expectStatus(t, w, http.StatusAccepted)

	if _, err := ta.app.Federation.Inbox().Get("alice", "bob", "22222222"); !errors.Is(err, workouts.ErrWorkoutNotFound) {
		t.Fatalf("expected deleted workout, err=%v", err)
	}
}

func TestFederationRoutes(t *testing.T) {
	ta := setupFederationTestApp(t)
	ta.register(t, "alice", "alice@example.com", "password12")

	w := ta.doJSON(t, http.MethodGet, "/api/v1/server-info", nil, "")
	expectStatus(t, w, http.StatusOK)
	info := decodeObject(t, w)
	if info["federation_enabled"] != true {
		t.Fatalf("expected federation_enabled=true: %#v", info)
	}

	w = ta.doJSON(t, http.MethodGet, "/.well-known/webfinger?resource=acct:alice@localhost", nil, "")
	expectStatus(t, w, http.StatusOK)
	wf := decodeObject(t, w)
	if wf["subject"] != "acct:alice@localhost" {
		t.Fatalf("unexpected webfinger: %#v", wf)
	}
	links, _ := wf["links"].([]any)
	if len(links) == 0 {
		t.Fatalf("expected webfinger links: %#v", wf)
	}

	req := httptest.NewRequest(http.MethodGet, "/users/alice", nil)
	req.Header.Set("Accept", "application/activity+json")
	w = httptest.NewRecorder()
	ta.router.ServeHTTP(w, req)
	expectStatus(t, w, http.StatusOK)
	actor := decodeObject(t, w)
	if actor["type"] != "Person" || actor["preferredUsername"] != "alice" {
		t.Fatalf("unexpected actor: %#v", actor)
	}
	if inbox, _ := actor["inbox"].(string); !strings.Contains(inbox, "/users/alice/inbox") {
		t.Fatalf("unexpected inbox: %#v", actor["inbox"])
	}

	w = ta.doJSON(t, http.MethodGet, "/users/alice/outbox", nil, "")
	expectStatus(t, w, http.StatusOK)
	outbox := decodeObject(t, w)
	if outbox["type"] != "OrderedCollection" {
		t.Fatalf("unexpected outbox: %#v", outbox)
	}

	followBody := map[string]any{
		"@context": "https://www.w3.org/ns/activitystreams",
		"type":     "Follow",
		"actor":    "https://remote.example/users/bob",
		"object":   "https://localhost/users/alice",
		"id":       "https://remote.example/activities/1",
	}
	data, err := json.Marshal(followBody)
	if err != nil {
		t.Fatal(err)
	}
	req = httptest.NewRequest(http.MethodPost, "/users/alice/inbox", bytes.NewReader(data))
	req.Header.Set("Content-Type", "application/activity+json")
	w = httptest.NewRecorder()
	ta.router.ServeHTTP(w, req)
	expectStatus(t, w, http.StatusAccepted)

	req = httptest.NewRequest(http.MethodPost, "/inbox", bytes.NewReader([]byte(`{}`)))
	req.Header.Set("Content-Type", "application/activity+json")
	w = httptest.NewRecorder()
	ta.router.ServeHTTP(w, req)
	expectStatus(t, w, http.StatusAccepted)

	pngData := readTestdata(t, "images/avatar-square.png")
	token, _ := ta.login(t, "alice@example.com", "password12")
	w = ta.doMultipart(t, http.MethodPut, "/api/v1/auth/me/avatar", token, nil,
		map[string][]filePart{
			"avatar": {{filename: "avatar.png", data: pngData}},
		},
	)
	expectStatus(t, w, http.StatusOK)

	w = ta.doJSON(t, http.MethodGet, "/users/alice/avatar", nil, "")
	expectStatus(t, w, http.StatusOK)
	if ct := w.Header().Get("Content-Type"); ct != "image/webp" {
		t.Fatalf("public avatar content-type = %q", ct)
	}
}
