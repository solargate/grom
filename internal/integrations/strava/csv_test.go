package strava

import (
	"archive/zip"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestReadActivitiesCSVFromSample(t *testing.T) {
	zipPath := filepath.Join("..", "..", "..", "cmd", "grom", "export_60943925.zip")
	if _, err := os.Stat(zipPath); err != nil {
		t.Skip("sample archive not available")
	}

	archive, err := OpenArchive(zipPath)
	if err != nil {
		t.Fatal(err)
	}
	defer archive.Close()

	csvData, err := archive.ReadFile("activities.csv")
	if err != nil {
		t.Fatal(err)
	}

	rows, hint, err := ReadActivitiesCSVFromReader(strings.NewReader(string(csvData)))
	if err != nil {
		t.Fatalf("ReadActivitiesCSVFromReader() error = %v", err)
	}
	if len(rows) != 982 {
		t.Fatalf("parsed %d activities, want 982", len(rows))
	}
	if hint != localeRussian {
		t.Fatalf("locale hint = %v, want russian", hint)
	}

	first := rows[0]
	workout, err := first.ToWorkout(hint)
	if err != nil {
		t.Fatalf("ToWorkout() error = %v", err)
	}
	if workout.SportType != "Ride" {
		t.Fatalf("sport_type = %q, want Ride", workout.SportType)
	}
	if workout.DurationSeconds != 2184 {
		t.Fatalf("duration_seconds = %d, want 2184", workout.DurationSeconds)
	}
	if workout.Distance <= 0 {
		t.Fatalf("expected distance > 0, got %v", workout.Distance)
	}
}

func TestOpenArchiveFromSampleZip(t *testing.T) {
	zipPath := filepath.Join("..", "..", "..", "cmd", "grom", "export_60943925.zip")
	if _, err := os.Stat(zipPath); err != nil {
		t.Skip("sample archive not available")
	}

	archive, err := OpenArchive(zipPath)
	if err != nil {
		t.Fatal(err)
	}
	defer archive.Close()

	data, err := archive.ReadFile("activities/20332562653.fit.gz")
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}
	if len(data) == 0 {
		t.Fatal("expected track data")
	}
}

func extractZipFromFile(archivePath, destDir string) error {
	reader, err := zip.OpenReader(archivePath)
	if err != nil {
		return err
	}
	defer reader.Close()

	if err := os.MkdirAll(destDir, 0700); err != nil {
		return err
	}
	destDir = filepath.Clean(destDir)
	for _, file := range reader.File {
		targetPath := filepath.Join(destDir, file.Name)
		targetPath = filepath.Clean(targetPath)
		if !strings.HasPrefix(targetPath, destDir+string(os.PathSeparator)) && targetPath != destDir {
			continue
		}
		if file.FileInfo().IsDir() {
			_ = os.MkdirAll(targetPath, 0700)
			continue
		}
		_ = os.MkdirAll(filepath.Dir(targetPath), 0700)
		src, err := file.Open()
		if err != nil {
			return err
		}
		dst, err := os.OpenFile(targetPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0600)
		if err != nil {
			src.Close()
			return err
		}
		_, err = io.Copy(dst, src)
		src.Close()
		dst.Close()
		if err != nil {
			return err
		}
	}
	return nil
}

func TestMapSportTypeRussian(t *testing.T) {
	got, err := mapSportType("Велосипед", "")
	if err != nil {
		t.Fatal(err)
	}
	if got != "Ride" {
		t.Fatalf("got %q, want Ride", got)
	}
}

func TestMapSportTypePilatesFromName(t *testing.T) {
	got, err := mapSportType("Тренировка", "Пилатес (день)")
	if err != nil {
		t.Fatal(err)
	}
	if got != "Pilates" {
		t.Fatalf("got %q, want Pilates", got)
	}
}
