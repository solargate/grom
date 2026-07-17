package workouts_test

import (
	"bytes"
	"context"
	"image"
	"image/png"
	"testing"
	"time"

	blobfs "github.com/solargate/grom/internal/storage/blob/fs"
	"github.com/solargate/grom/internal/storage/file"
	"github.com/solargate/grom/internal/storage/keys"
	"github.com/solargate/grom/internal/workouts"
)

func TestStoreAddMediaSanitizesFilename(t *testing.T) {
	dir := t.TempDir()
	svc := newTestService(dir)

	created, err := svc.Create("runner", &workouts.Workout{
		Name:      "Morning",
		SportType: "Run",
		StartDate: mustTime(t, "2026-07-08T10:00:00Z"),
	})
	if err != nil {
		t.Fatal(err)
	}

	img := image.NewRGBA(image.Rect(0, 0, 32, 32))
	var raw bytes.Buffer
	if err := png.Encode(&raw, img); err != nil {
		t.Fatal(err)
	}

	updated, err := svc.AddMedia("runner", created, []workouts.MediaFileInput{
		{Filename: "../evil/photo.png", Data: raw.Bytes()},
	})
	if err != nil {
		t.Fatalf("AddMedia() error = %v", err)
	}
	if len(updated.MediaFiles) != 1 || updated.MediaFiles[0] != "photo.png" {
		t.Fatalf("unexpected media files: %#v", updated.MediaFiles)
	}

	store := file.NewWorkoutsStore(dir)
	dirName, err := store.WorkoutDirName("runner", created.ID)
	if err != nil {
		t.Fatal(err)
	}
	blobs := blobfs.NewStore(dir)
	origKey := keys.WorkoutMediaOriginal("runner", dirName, "photo.png")
	if ok, _ := blobs.Exists(context.Background(), origKey); !ok {
		t.Fatalf("original blob missing at %q", origKey)
	}
	// Path traversal must not create files outside workout media dir.
	evilKey := keys.WorkoutMediaOriginal("runner", dirName, "../evil/photo.png")
	if ok, _ := blobs.Exists(context.Background(), evilKey); ok {
		t.Fatal("unexpected blob for unsanitized path")
	}
}

func TestStoreAddMediaDuplicateNames(t *testing.T) {
	dir := t.TempDir()
	svc := newTestService(dir)

	created, err := svc.Create("runner", &workouts.Workout{
		Name:      "Morning",
		SportType: "Run",
		StartDate: mustTime(t, "2026-07-08T10:00:00Z"),
	})
	if err != nil {
		t.Fatal(err)
	}

	img := image.NewRGBA(image.Rect(0, 0, 32, 32))
	var raw bytes.Buffer
	if err := png.Encode(&raw, img); err != nil {
		t.Fatal(err)
	}
	data := raw.Bytes()

	first, err := svc.AddMedia("runner", created, []workouts.MediaFileInput{
		{Filename: "shot.png", Data: data},
	})
	if err != nil {
		t.Fatal(err)
	}
	second, err := svc.AddMedia("runner", first, []workouts.MediaFileInput{
		{Filename: "shot.png", Data: data},
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(second.MediaFiles) != 2 {
		t.Fatalf("expected 2 media files, got %#v", second.MediaFiles)
	}
	if second.MediaFiles[0] != "shot.png" {
		t.Fatalf("first name = %q, want shot.png", second.MediaFiles[0])
	}
	if second.MediaFiles[1] != "shot (2).png" {
		t.Fatalf("second name = %q, want shot (2).png", second.MediaFiles[1])
	}

	store := file.NewWorkoutsStore(dir)
	dirName, err := store.WorkoutDirName("runner", created.ID)
	if err != nil {
		t.Fatal(err)
	}
	blobs := blobfs.NewStore(dir)
	for _, name := range second.MediaFiles {
		orig := keys.WorkoutMediaOriginal("runner", dirName, name)
		prev := keys.WorkoutMediaPreview("runner", dirName, name)
		if ok, _ := blobs.Exists(context.Background(), orig); !ok {
			t.Fatalf("missing original %q", orig)
		}
		if ok, _ := blobs.Exists(context.Background(), prev); !ok {
			t.Fatalf("missing preview %q", prev)
		}
	}
}

func TestStoreAddMedia(t *testing.T) {
	dir := t.TempDir()
	store := file.NewWorkoutsStore(dir)
	svc := newTestService(dir)

	created, err := svc.Create("runner", &workouts.Workout{
		Name:      "Morning",
		SportType: "Run",
		StartDate: mustTime(t, "2026-07-08T10:00:00Z"),
	})
	if err != nil {
		t.Fatal(err)
	}

	img := image.NewRGBA(image.Rect(0, 0, 32, 32))
	var raw bytes.Buffer
	if err := png.Encode(&raw, img); err != nil {
		t.Fatal(err)
	}

	updated, err := svc.AddMedia("runner", created, []workouts.MediaFileInput{
		{Filename: "a.png", Data: raw.Bytes()},
	})
	if err != nil {
		t.Fatalf("AddMedia() error = %v", err)
	}
	if !updated.HasMedia || len(updated.MediaFiles) != 1 {
		t.Fatalf("unexpected media state: %+v", updated)
	}

	dirName, err := store.WorkoutDirName("runner", created.ID)
	if err != nil {
		t.Fatal(err)
	}
	blobs := blobfs.NewStore(dir)
	previewKey := keys.WorkoutMediaPreview("runner", dirName, "a.png")
	if ok, _ := blobs.Exists(context.Background(), previewKey); !ok {
		t.Fatalf("preview blob missing at %q", previewKey)
	}
}

func mustTime(t *testing.T, value string) time.Time {
	t.Helper()
	parsed, err := time.Parse(time.RFC3339, value)
	if err != nil {
		t.Fatal(err)
	}
	return parsed
}
