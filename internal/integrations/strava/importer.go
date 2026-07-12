package strava

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"path/filepath"
	"strings"

	"github.com/solargate/grom/internal/equipment"
	"github.com/solargate/grom/internal/tracks"
	"github.com/solargate/grom/internal/workouts"
)

type PublishWorkoutFunc func(nickname string, workout *workouts.Workout)

type ImportResult struct {
	Imported     int `json:"imported"`
	Skipped      int `json:"skipped"`
	ParseSkipped int `json:"parse_skipped"`
	Errors       int `json:"errors"`
}

type Importer struct {
	workoutStore   *workouts.Store
	equipmentStore *equipment.Store
	archive        *Archive
	onPublish      PublishWorkoutFunc
}

func NewImporter(workoutStore *workouts.Store, equipmentStore *equipment.Store, archive *Archive, onPublish PublishWorkoutFunc) *Importer {
	return &Importer{
		workoutStore:   workoutStore,
		equipmentStore: equipmentStore,
		archive:        archive,
		onPublish:      onPublish,
	}
}

func (imp *Importer) ImportAll(nickname string, progress func(current, total int)) (ImportResult, error) {
	csvData, err := imp.archive.ReadFile("activities.csv")
	if err != nil {
		return ImportResult{}, err
	}
	stats, hint, rows, err := parseActivitiesCSVStats(bytes.NewReader(csvData))
	if err != nil {
		return ImportResult{}, err
	}

	equipmentResolver, err := NewEquipmentResolver(imp.equipmentStore, nickname, imp.archive)
	if err != nil {
		return ImportResult{}, err
	}

	result := ImportResult{ParseSkipped: stats.SkippedRows}
	total := len(rows)
	for i, row := range rows {
		if progress != nil {
			progress(i+1, total)
		}

		if row.StravaActivityID != "" {
			exists, err := imp.workoutStore.HasStravaActivityID(nickname, row.StravaActivityID)
			if err != nil {
				result.Errors++
				continue
			}
			if exists {
				result.Skipped++
				continue
			}
		}

		if err := imp.importOne(nickname, row, hint, equipmentResolver); err != nil {
			result.Errors++
			continue
		}
		result.Imported++
	}
	return result, nil
}

func (imp *Importer) importOne(nickname string, row ActivityRow, hint localeHint, resolver *EquipmentResolver) error {
	workout, err := row.ToWorkout(hint)
	if err != nil {
		return err
	}

	if row.EquipmentName != "" {
		items, err := resolver.Resolve(row.EquipmentName)
		if err != nil {
			return err
		}
		workout.Equipment = toWorkoutEquipment(items)
	}

	created, err := imp.workoutStore.Create(nickname, workout)
	if err != nil {
		return err
	}

	if row.TrackFile != "" {
		trackInput, err := imp.loadTrack(row.TrackFile)
		if err == nil && trackInput != nil {
			created, err = imp.workoutStore.AttachTrack(nickname, created, trackInput)
			if err != nil {
				return err
			}
		}
	}

	photos := parseMediaPaths(row.MediaFiles)
	if len(photos) > 0 {
		mediaFiles, err := imp.loadPhotos(photos)
		if err != nil {
			return err
		}
		if len(mediaFiles) > 0 {
			created, err = imp.workoutStore.AddMedia(nickname, created, mediaFiles)
			if err != nil {
				return err
			}
		}
	}

	if imp.onPublish != nil {
		imp.onPublish(nickname, created)
	}
	return nil
}

func (imp *Importer) loadTrack(relativePath string) (*workouts.TrackInput, error) {
	relativePath = filepath.ToSlash(strings.TrimSpace(relativePath))
	data, err := imp.archive.ReadFile(relativePath)
	if err != nil {
		return nil, err
	}

	filename := filepath.Base(relativePath)
	if strings.HasSuffix(strings.ToLower(filename), ".gz") {
		gzReader, err := gzip.NewReader(bytes.NewReader(data))
		if err != nil {
			return nil, fmt.Errorf("gunzip track: %w", err)
		}
		defer gzReader.Close()
		uncompressed, err := io.ReadAll(gzReader)
		if err != nil {
			return nil, fmt.Errorf("read gunzip track: %w", err)
		}
		data = uncompressed
		filename = strings.TrimSuffix(filename, ".gz")
	}

	parsed, err := tracks.Parse(data, filename)
	if err != nil {
		return nil, err
	}

	return &workouts.TrackInput{
		Filename: filename,
		Data:     data,
		Parsed:   parsed,
	}, nil
}

func (imp *Importer) loadPhotos(relativePaths []string) ([]workouts.MediaFileInput, error) {
	if len(relativePaths) > workouts.MaxPhotosPerWorkout {
		relativePaths = relativePaths[:workouts.MaxPhotosPerWorkout]
	}

	files := make([]workouts.MediaFileInput, 0, len(relativePaths))
	for _, relativePath := range relativePaths {
		relativePath = filepath.ToSlash(strings.TrimSpace(relativePath))
		data, err := imp.archive.ReadFile(relativePath)
		if err != nil {
			continue
		}
		files = append(files, workouts.MediaFileInput{
			Filename: filepath.Base(relativePath),
			Data:     data,
		})
	}
	return files, nil
}
