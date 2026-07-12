package workouts

import (
	"bytes"
	"image"
	"image/png"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestSaveOriginalAndPreview(t *testing.T) {
	dir := t.TempDir()

	img := image.NewRGBA(image.Rect(0, 0, 800, 600))
	var raw bytes.Buffer
	if err := png.Encode(&raw, img); err != nil {
		t.Fatal(err)
	}

	name, err := SaveOriginalAndPreview(dir, "../evil/photo.png", raw.Bytes())
	if err != nil {
		t.Fatalf("SaveOriginalAndPreview() error = %v", err)
	}
	if name != "photo.png" {
		t.Fatalf("filename = %q", name)
	}

	originalPath := MediaOriginalPath(dir, name)
	if _, err := os.Stat(originalPath); err != nil {
		t.Fatalf("original missing: %v", err)
	}
	previewPath := MediaPreviewPath(dir, name)
	if _, err := os.Stat(previewPath); err != nil {
		t.Fatalf("preview missing: %v", err)
	}
}

func TestSaveWorkoutMediaDuplicateNames(t *testing.T) {
	dir := t.TempDir()
	img := image.NewRGBA(image.Rect(0, 0, 64, 64))
	var raw bytes.Buffer
	if err := png.Encode(&raw, img); err != nil {
		t.Fatal(err)
	}
	data := raw.Bytes()

	first, err := SaveOriginalAndPreview(dir, "shot.png", data)
	if err != nil {
		t.Fatal(err)
	}
	second, err := SaveOriginalAndPreview(dir, "shot.png", data)
	if err != nil {
		t.Fatal(err)
	}
	if first == second {
		t.Fatalf("expected unique names, both %q", first)
	}
}

func TestStoreAddMedia(t *testing.T) {
	dir := t.TempDir()
	store := NewStore(dir)

	created, err := store.Create("runner", &Workout{
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

	updated, err := store.AddMedia("runner", created, []MediaFileInput{
		{Filename: "a.png", Data: raw.Bytes()},
	})
	if err != nil {
		t.Fatalf("AddMedia() error = %v", err)
	}
	if !updated.HasMedia || len(updated.MediaFiles) != 1 {
		t.Fatalf("unexpected media state: %+v", updated)
	}

	workoutDir, err := store.WorkoutDir("runner", created.ID)
	if err != nil {
		t.Fatal(err)
	}
	preview := filepath.Join(workoutDir, MediaSubdir, PreviewPrefix+"a.png.webp")
	if _, err := os.Stat(preview); err != nil {
		t.Fatalf("preview file missing: %v", err)
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
