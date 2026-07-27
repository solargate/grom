package strava

import (
	"archive/zip"
	"os"
	"path/filepath"
	"testing"
)

func TestLoadPhotosCountsMissingAndLoadsPresent(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	zipPath := filepath.Join(dir, "export.zip")
	createTestArchive(t, zipPath, map[string][]byte{
		"activities.csv": []byte("ID\n1\n"),
		"media/present.jpg": {
			0xff, 0xd8, 0xff, 0xd9, // minimal JPEG markers
		},
	})

	archive, err := OpenArchive(zipPath)
	if err != nil {
		t.Fatalf("OpenArchive: %v", err)
	}
	defer archive.Close()

	imp := &Importer{archive: archive}
	files, missing, err := imp.loadPhotos([]string{
		"media/present.jpg",
		"media/missing-a.jpg",
		"media/missing-b.jpg",
	})
	if err != nil {
		t.Fatalf("loadPhotos: %v", err)
	}
	if missing != 2 {
		t.Fatalf("missing = %d, want 2", missing)
	}
	if len(files) != 1 {
		t.Fatalf("loaded %d files, want 1", len(files))
	}
	if files[0].Filename != "present.jpg" {
		t.Fatalf("filename = %q, want present.jpg", files[0].Filename)
	}
}

func TestLoadPhotosCountsMissingBeyondAttachLimit(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	zipPath := filepath.Join(dir, "export.zip")
	createTestArchive(t, zipPath, map[string][]byte{
		"activities.csv": []byte("ID\n1\n"),
		"media/a.jpg":    {0xff, 0xd8, 0xff, 0xd9},
		"media/b.jpg":    {0xff, 0xd8, 0xff, 0xd9},
	})

	archive, err := OpenArchive(zipPath)
	if err != nil {
		t.Fatalf("OpenArchive: %v", err)
	}
	defer archive.Close()

	paths := make([]string, 0, 24)
	for i := 0; i < 22; i++ {
		paths = append(paths, "media/missing.jpg")
	}
	paths = append(paths, "media/a.jpg", "media/b.jpg")

	imp := &Importer{archive: archive}
	files, missing, err := imp.loadPhotos(paths)
	if err != nil {
		t.Fatalf("loadPhotos: %v", err)
	}
	if missing != 22 {
		t.Fatalf("missing = %d, want 22", missing)
	}
	if len(files) != 2 {
		t.Fatalf("loaded %d files, want 2", len(files))
	}
}

func TestArchiveHas(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	zipPath := filepath.Join(dir, "export.zip")
	createTestArchive(t, zipPath, map[string][]byte{
		"activities.csv":    []byte("ID\n1\n"),
		"media/present.jpg": {1, 2, 3},
	})

	archive, err := OpenArchive(zipPath)
	if err != nil {
		t.Fatalf("OpenArchive: %v", err)
	}
	defer archive.Close()

	if !archive.Has("media/present.jpg") {
		t.Fatal("expected Has(media/present.jpg)")
	}
	if !archive.Has("./media/present.jpg") {
		t.Fatal("expected Has with ./ prefix")
	}
	if archive.Has("media/absent.jpg") {
		t.Fatal("did not expect Has(media/absent.jpg)")
	}
}

func createTestArchive(t *testing.T, path string, files map[string][]byte) {
	t.Helper()
	f, err := os.Create(path)
	if err != nil {
		t.Fatalf("create zip: %v", err)
	}
	defer f.Close()

	w := zip.NewWriter(f)
	for name, data := range files {
		fw, err := w.Create(name)
		if err != nil {
			t.Fatalf("create entry %s: %v", name, err)
		}
		if _, err := fw.Write(data); err != nil {
			t.Fatalf("write entry %s: %v", name, err)
		}
	}
	if err := w.Close(); err != nil {
		t.Fatalf("close zip writer: %v", err)
	}
}
