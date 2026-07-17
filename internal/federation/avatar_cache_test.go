package federation

import (
	"bytes"
	"context"
	"image"
	"image/color"
	"image/png"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/solargate/grom/internal/storage/keys"
	"github.com/solargate/grom/internal/tracks"
	"github.com/solargate/grom/internal/workouts"
)

func TestWorkoutInboxStoreCachesRemoteAvatar(t *testing.T) {
	dir := t.TempDir()
	store := newTestInboxStore(dir)

	img := image.NewRGBA(image.Rect(0, 0, 256, 256))
	for y := 0; y < 256; y++ {
		for x := 0; x < 256; x++ {
			img.Set(x, y, color.RGBA{R: 40, G: 120, B: 200, A: 255})
		}
	}
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		t.Fatal(err)
	}

	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "image/png")
		_, _ = w.Write(buf.Bytes())
	}))
	defer server.Close()

	client := server.Client()
	store.SetHTTPClient(client)

	workout := &workouts.Workout{
		ID:              "38472901",
		Name:            "Remote run",
		SportType:       "Run",
		StartDate:       time.Date(2026, 7, 8, 10, 0, 0, 0, time.UTC),
		DurationSeconds: 4200,
		Distance:        10000,
		Track:           tracks.TrackFileGPX,
	}
	ownerHandle := "test2@192.168.1.251:8445"
	ownerKey := OwnerKeyFromHandle(ownerHandle)
	actor := map[string]any{
		"name": "Test Two",
		"icon": map[string]any{
			"type": "Image",
			"url":  server.URL + "/users/test2/avatar",
		},
	}

	if err := store.Save("solarwind", ownerHandle, workout, nil, nil, actor); err != nil {
		t.Fatal(err)
	}

	ctx := context.Background()
	if !hasFederatedAvatar(ctx, store.blobs, "solarwind", ownerKey) {
		t.Fatal("expected cached federated avatar blob")
	}

	items, err := store.List("solarwind")
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(items))
	}
	if !items[0].Author.HasAvatar {
		t.Fatal("expected has_avatar true")
	}
	wantURL := FederatedAvatarAPIPath(ownerHandle, 1)
	if items[0].Author.AvatarURL != wantURL {
		t.Fatalf("avatar url = %q, want %q", items[0].Author.AvatarURL, wantURL)
	}

	data, err := store.Avatar("solarwind", ownerKey)
	if err != nil {
		t.Fatalf("Avatar() error = %v", err)
	}
	if len(data) == 0 {
		t.Fatal("expected avatar bytes")
	}
}

func TestWorkoutInboxStoreRefreshesAvatarOnNewActivity(t *testing.T) {
	dir := t.TempDir()
	store := newTestInboxStore(dir)

	makePNG := func(r, g, b uint8) []byte {
		img := image.NewRGBA(image.Rect(0, 0, 256, 256))
		for y := 0; y < 256; y++ {
			for x := 0; x < 256; x++ {
				img.Set(x, y, color.RGBA{R: r, G: g, B: b, A: 255})
			}
		}
		var buf bytes.Buffer
		if err := png.Encode(&buf, img); err != nil {
			t.Fatal(err)
		}
		return buf.Bytes()
	}

	avatarURL := "/users/test2/avatar"
	var current []byte
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "image/png")
		_, _ = w.Write(current)
	}))
	defer server.Close()

	client := server.Client()
	store.SetHTTPClient(client)

	ownerHandle := "test2@192.168.1.251:8445"
	ownerKey := OwnerKeyFromHandle(ownerHandle)
	actor := func() map[string]any {
		return map[string]any{
			"name": "Test Two",
			"icon": map[string]any{
				"type": "Image",
				"url":  server.URL + avatarURL,
			},
		}
	}

	current = makePNG(40, 120, 200)
	workout1 := &workouts.Workout{
		ID:        "111",
		Name:      "First",
		SportType: "Run",
		StartDate: time.Date(2026, 7, 8, 10, 0, 0, 0, time.UTC),
	}
	if err := store.Save("solarwind", ownerHandle, workout1, nil, nil, actor()); err != nil {
		t.Fatal(err)
	}

	firstData, err := store.Avatar("solarwind", ownerKey)
	if err != nil {
		t.Fatal(err)
	}

	current = makePNG(200, 80, 40)
	workout2 := &workouts.Workout{
		ID:        "222",
		Name:      "Second",
		SportType: "Run",
		StartDate: time.Date(2026, 7, 9, 10, 0, 0, 0, time.UTC),
	}
	if err := store.Save("solarwind", ownerHandle, workout2, nil, nil, actor()); err != nil {
		t.Fatal(err)
	}

	secondData, err := store.Avatar("solarwind", ownerKey)
	if err != nil {
		t.Fatal(err)
	}
	if bytes.Equal(firstData, secondData) {
		t.Fatal("expected avatar blob to be refreshed")
	}

	ownerDir := store.ownerDir("solarwind", ownerHandle)
	meta, err := readAuthorMeta(ownerDir)
	if err != nil {
		t.Fatal(err)
	}
	if meta.AvatarVersion != 2 {
		t.Fatalf("avatar version = %d, want 2", meta.AvatarVersion)
	}

	items, err := store.List("solarwind")
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 2 {
		t.Fatalf("expected 2 items, got %d", len(items))
	}
	wantURL := FederatedAvatarAPIPath(ownerHandle, 2)
	for _, item := range items {
		if item.Author.AvatarURL != wantURL {
			t.Fatalf("avatar url = %q, want %q", item.Author.AvatarURL, wantURL)
		}
	}
}

func TestWorkoutInboxStoreEnsureAuthorForFollow(t *testing.T) {
	dir := t.TempDir()
	store := newTestInboxStore(dir)

	img := image.NewRGBA(image.Rect(0, 0, 256, 256))
	for y := 0; y < 256; y++ {
		for x := 0; x < 256; x++ {
			img.Set(x, y, color.RGBA{R: 10, G: 150, B: 90, A: 255})
		}
	}
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		t.Fatal(err)
	}

	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "image/png")
		_, _ = w.Write(buf.Bytes())
	}))
	defer server.Close()

	store.SetHTTPClient(server.Client())

	ownerHandle := "test2@192.168.1.251:8445"
	ownerKey := OwnerKeyFromHandle(ownerHandle)
	remoteURL := server.URL + "/users/test2/avatar"
	if err := store.EnsureAuthor("solarwind", ownerHandle, "test2", "Test Two", remoteURL, true); err != nil {
		t.Fatal(err)
	}

	hasAvatar, avatarURL := store.AuthorAvatarFields("solarwind", ownerHandle)
	if !hasAvatar {
		t.Fatal("expected has avatar")
	}
	wantURL := FederatedAvatarAPIPath(ownerHandle, 1)
	if avatarURL != wantURL {
		t.Fatalf("avatar url = %q, want %q", avatarURL, wantURL)
	}

	ctx := context.Background()
	if ok, _ := store.blobs.Exists(ctx, keys.FederatedInboxAvatar("solarwind", ownerKey)); !ok {
		t.Fatal("expected avatar blob")
	}
}
