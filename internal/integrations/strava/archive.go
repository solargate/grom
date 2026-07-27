package strava

import (
	"archive/zip"
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

const zipMagic = "PK"

// Archive provides read access to a Strava export ZIP without extracting it.
type Archive struct {
	path   string
	reader *zip.ReadCloser
	index  map[string]*zip.File
}

func OpenArchive(path string) (*Archive, error) {
	if err := validateZipFile(path); err != nil {
		return nil, err
	}

	reader, err := zip.OpenReader(path)
	if err != nil {
		return nil, fmt.Errorf("open zip archive: %w", err)
	}

	index := make(map[string]*zip.File, len(reader.File))
	for _, file := range reader.File {
		index[normalizeArchivePath(file.Name)] = file
	}

	if _, ok := index["activities.csv"]; !ok {
		reader.Close()
		return nil, errNoActivities
	}

	return &Archive{
		path:   path,
		reader: reader,
		index:  index,
	}, nil
}

func (a *Archive) Close() error {
	if a == nil || a.reader == nil {
		return nil
	}
	err := a.reader.Close()
	a.reader = nil
	return err
}

func (a *Archive) ReadFile(relativePath string) ([]byte, error) {
	file, err := a.Open(relativePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	return io.ReadAll(file)
}

func (a *Archive) Has(relativePath string) bool {
	if a == nil || a.reader == nil {
		return false
	}
	_, ok := a.index[normalizeArchivePath(relativePath)]
	return ok
}

func (a *Archive) Open(relativePath string) (io.ReadCloser, error) {
	if a == nil || a.reader == nil {
		return nil, fmt.Errorf("archive is closed")
	}
	key := normalizeArchivePath(relativePath)
	file, ok := a.index[key]
	if !ok {
		return nil, fmt.Errorf("file not found in archive: %s", relativePath)
	}
	return file.Open()
}

func normalizeArchivePath(name string) string {
	name = filepath.ToSlash(strings.TrimSpace(name))
	name = strings.TrimPrefix(name, "./")
	return name
}

func validateZipFile(path string) error {
	info, err := os.Stat(path)
	if err != nil {
		return err
	}
	if info.Size() < 4 {
		return fmt.Errorf("archive file is empty or too small")
	}

	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()

	header := make([]byte, 4)
	if _, err := io.ReadFull(file, header); err != nil {
		return fmt.Errorf("read archive header: %w", err)
	}
	if !bytes.HasPrefix(header, []byte(zipMagic)) {
		return fmt.Errorf("file is not a valid zip archive")
	}
	return nil
}

func ValidateSavedArchive(path string) error {
	if err := validateZipFile(path); err != nil {
		return err
	}
	archive, err := OpenArchive(path)
	if err != nil {
		return err
	}
	defer archive.Close()
	if _, err := archive.ReadFile("activities.csv"); err != nil {
		return fmt.Errorf("archive validation failed: %w", err)
	}
	return nil
}
