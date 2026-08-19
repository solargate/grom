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
	"github.com/solargate/grom/internal/config"
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
	if info["password_reset_enabled"] != false {
		t.Fatalf("expected password_reset_enabled=false, got %#v", info)
	}
	if info["registration"] != "open" {
		t.Fatalf("expected registration=open, got %#v", info)
	}
}

func TestRegisterClosedMode(t *testing.T) {
	ta := setupTestAppWithConfig(t, func(cfg *config.Config) {
		cfg.Server.TLS.Mode = "off"
		cfg.Federation.Enabled = false
		cfg.Federation.Domain = "localhost"
		cfg.Server.Registration = config.RegistrationClosed
	})
	w := ta.doJSON(t, http.MethodPost, "/api/v1/auth/register", map[string]string{
		"nickname": "bob",
		"name":     "Bob",
		"email":    "bob@example.com",
		"password": "password12",
	}, "")
	expectStatus(t, w, http.StatusForbidden)
	body := decodeObject(t, w)
	if msg, _ := body["error"].(string); msg != "registration is disabled on this server" {
		t.Fatalf("unexpected error: %q", msg)
	}
}

func TestRegisterInviteMode(t *testing.T) {
	ta := setupTestAppWithConfig(t, func(cfg *config.Config) {
		cfg.Server.TLS.Mode = "off"
		cfg.Federation.Enabled = false
		cfg.Federation.Domain = "localhost"
		cfg.Server.Registration = config.RegistrationInvite
	})
	w := ta.doJSON(t, http.MethodPost, "/api/v1/auth/register", map[string]string{
		"nickname": "bob",
		"name":     "Bob",
		"email":    "bob@example.com",
		"password": "password12",
	}, "")
	expectStatus(t, w, http.StatusForbidden)
	body := decodeObject(t, w)
	if msg, _ := body["error"].(string); msg != "registration on this server is by invitation only" {
		t.Fatalf("unexpected error: %q", msg)
	}
}

func TestServerInfoRegistrationClosed(t *testing.T) {
	ta := setupTestAppWithConfig(t, func(cfg *config.Config) {
		cfg.Server.TLS.Mode = "off"
		cfg.Federation.Enabled = false
		cfg.Federation.Domain = "localhost"
		cfg.Server.Registration = config.RegistrationClosed
	})
	w := ta.doJSON(t, http.MethodGet, "/api/v1/server-info", nil, "")
	expectStatus(t, w, http.StatusOK)
	info := decodeObject(t, w)
	if info["registration"] != "closed" {
		t.Fatalf("expected registration=closed, got %#v", info)
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

func TestWorkoutMediaAddAndDelete(t *testing.T) {
	ta := setupTestApp(t)
	ta.register(t, "alice", "alice@example.com", "password12")
	token, _ := ta.login(t, "alice@example.com", "password12")

	var photoBuf bytes.Buffer
	img := image.NewRGBA(image.Rect(0, 0, 32, 32))
	if err := png.Encode(&photoBuf, img); err != nil {
		t.Fatal(err)
	}
	photoBytes := photoBuf.Bytes()

	w := ta.doJSON(t, http.MethodPost, "/api/v1/workouts", map[string]any{
		"name":       "Media edit",
		"sport_type": "Run",
		"start_date": "2026-07-08T10:00:00Z",
	}, token)
	expectStatus(t, w, http.StatusCreated)
	created := decodeObject(t, w)
	id, _ := created["id"].(string)

	w = ta.doMultipart(t, http.MethodPost, "/api/v1/workouts/"+id+"/media", token, nil,
		map[string][]filePart{
			"photos": {
				{filename: "a.png", data: photoBytes},
				{filename: "b.png", data: photoBytes},
			},
		},
	)
	expectStatus(t, w, http.StatusOK)
	withMedia := decodeObject(t, w)
	if withMedia["has_media"] != true {
		t.Fatalf("expected has_media: %#v", withMedia)
	}
	mediaFiles, _ := withMedia["media_files"].([]any)
	if len(mediaFiles) != 2 {
		t.Fatalf("expected 2 media files: %#v", withMedia["media_files"])
	}
	first, _ := mediaFiles[0].(string)
	second, _ := mediaFiles[1].(string)

	w = ta.doJSON(t, http.MethodDelete, "/api/v1/workouts/"+id+"/media/"+first, nil, token)
	expectStatus(t, w, http.StatusOK)
	afterDelete := decodeObject(t, w)
	remaining, _ := afterDelete["media_files"].([]any)
	if len(remaining) != 1 || remaining[0] != second {
		t.Fatalf("expected only %q left: %#v", second, afterDelete["media_files"])
	}

	w = ta.doJSON(t, http.MethodGet, "/api/v1/workouts/"+id+"/media/"+first+"/preview", nil, token)
	expectStatus(t, w, http.StatusNotFound)
	w = ta.doJSON(t, http.MethodGet, "/api/v1/workouts/"+id+"/media/"+second+"/preview", nil, token)
	expectStatus(t, w, http.StatusOK)

	w = ta.doJSON(t, http.MethodDelete, "/api/v1/workouts/"+id+"/media/"+second, nil, token)
	expectStatus(t, w, http.StatusOK)
	cleared := decodeObject(t, w)
	if cleared["has_media"] != false {
		t.Fatalf("expected has_media false: %#v", cleared)
	}
	if files, _ := cleared["media_files"].([]any); len(files) != 0 {
		t.Fatalf("expected empty media_files: %#v", cleared["media_files"])
	}

	w = ta.doJSON(t, http.MethodDelete, "/api/v1/workouts/"+id+"/media/missing.png", nil, token)
	expectStatus(t, w, http.StatusNotFound)

	w = ta.doMultipart(t, http.MethodPost, "/api/v1/workouts/"+id+"/media", token, nil, nil)
	expectStatus(t, w, http.StatusBadRequest)
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

func TestFederationInboxMediaReplaceOnUpdate(t *testing.T) {
	ta := setupFederationTestApp(t)
	ta.register(t, "alice", "alice@example.com", "password12")

	var photoA, photoB bytes.Buffer
	img := image.NewRGBA(image.Rect(0, 0, 24, 24))
	if err := png.Encode(&photoA, img); err != nil {
		t.Fatal(err)
	}
	if err := png.Encode(&photoB, img); err != nil {
		t.Fatal(err)
	}

	createObj := map[string]any{
		"@context": "https://www.w3.org/ns/activitystreams",
		"type":     "Create",
		"actor":    "https://remote.example/users/bob",
		"object": map[string]any{
			"id":              "https://remote.example/users/bob/workouts/33333333",
			"type":            "Note",
			"name":            "Remote with photos",
			"sportType":       "Run",
			"startDate":       "2026-07-08T11:00:00Z",
			"durationSeconds": 900,
			"distance":        2000.0,
			"mediaItems": []any{
				map[string]any{
					"filename":  "keep.png",
					"mediaType": "image/png",
					"data":      base64.StdEncoding.EncodeToString(photoA.Bytes()),
				},
				map[string]any{
					"filename":  "drop.png",
					"mediaType": "image/png",
					"data":      base64.StdEncoding.EncodeToString(photoB.Bytes()),
				},
			},
		},
	}
	data, err := json.Marshal(createObj)
	if err != nil {
		t.Fatal(err)
	}
	req := httptest.NewRequest(http.MethodPost, "/users/alice/inbox", bytes.NewReader(data))
	req.Header.Set("Content-Type", "application/activity+json")
	w := httptest.NewRecorder()
	ta.router.ServeHTTP(w, req)
	expectStatus(t, w, http.StatusAccepted)

	got, err := ta.app.Federation.Inbox().Get("alice", "bob", "33333333")
	if err != nil {
		t.Fatalf("inbox Get: %v", err)
	}
	if len(got.MediaFiles) != 2 {
		t.Fatalf("expected 2 media files, got %#v", got.MediaFiles)
	}

	preview, err := ta.app.Federation.Inbox().MediaPreview("alice", "bob", "33333333", "drop.png")
	if err != nil || len(preview) == 0 {
		t.Fatalf("expected drop.png preview before update, err=%v", err)
	}

	updateObj := map[string]any{
		"@context": "https://www.w3.org/ns/activitystreams",
		"type":     "Update",
		"actor":    "https://remote.example/users/bob",
		"object": map[string]any{
			"id":              "https://remote.example/users/bob/workouts/33333333",
			"type":            "Note",
			"name":            "Remote with photos edited",
			"sportType":       "Run",
			"startDate":       "2026-07-08T11:00:00Z",
			"durationSeconds": 950,
			"distance":        2100.0,
			"mediaItems": []any{
				map[string]any{
					"filename":  "keep.png",
					"mediaType": "image/png",
					"data":      base64.StdEncoding.EncodeToString(photoA.Bytes()),
				},
				map[string]any{
					"filename":  "new.png",
					"mediaType": "image/png",
					"data":      base64.StdEncoding.EncodeToString(photoB.Bytes()),
				},
			},
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

	got, err = ta.app.Federation.Inbox().Get("alice", "bob", "33333333")
	if err != nil {
		t.Fatalf("inbox Get after update: %v", err)
	}
	if got.Name != "Remote with photos edited" {
		t.Fatalf("updated name = %q", got.Name)
	}
	if len(got.MediaFiles) != 2 {
		t.Fatalf("expected 2 media files after update, got %#v", got.MediaFiles)
	}
	want := map[string]bool{"keep.png": true, "new.png": true}
	for _, name := range got.MediaFiles {
		if !want[name] {
			t.Fatalf("unexpected media file %q in %#v", name, got.MediaFiles)
		}
		delete(want, name)
	}
	if len(want) != 0 {
		t.Fatalf("missing media files: %#v", want)
	}

	if _, err := ta.app.Federation.Inbox().MediaPreview("alice", "bob", "33333333", "drop.png"); !errors.Is(err, workouts.ErrPhotoNotFound) {
		t.Fatalf("expected drop.png removed, err=%v", err)
	}
	if preview, err := ta.app.Federation.Inbox().MediaPreview("alice", "bob", "33333333", "new.png"); err != nil || len(preview) == 0 {
		t.Fatalf("expected new.png preview, err=%v", err)
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
