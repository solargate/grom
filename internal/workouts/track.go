package workouts

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"github.com/solargate/grom/internal/maprender"
	"github.com/solargate/grom/internal/storage/blob"
	"github.com/solargate/grom/internal/storage/keys"
	"github.com/solargate/grom/internal/tracks"
)

const MapPreviewFileName = "map-preview.webp"

type TrackInput struct {
	Filename string
	Data     []byte
	Parsed   *tracks.Data
}

func (svc *Service) CreateWithTrack(nickname string, workout *Workout, track *TrackInput) (*Workout, error) {
	return svc.CreateWithAttachments(nickname, workout, track, nil)
}

func (svc *Service) CreateWithAttachments(nickname string, workout *Workout, track *TrackInput, photos []MediaFileInput) (*Workout, error) {
	return svc.createWithAttachments(nickname, workout, track, photos, false)
}

// ImportWithAttachments preserves the activity summary supplied by the importer.
func (svc *Service) ImportWithAttachments(nickname string, workout *Workout, track *TrackInput, photos []MediaFileInput) (*Workout, error) {
	return svc.createWithAttachments(nickname, workout, track, photos, true)
}

func (svc *Service) createWithAttachments(nickname string, workout *Workout, track *TrackInput, photos []MediaFileInput, preserveSummary bool) (*Workout, error) {
	if track == nil && len(photos) == 0 {
		return svc.Create(nickname, workout)
	}
	if err := validateWorkout(workout); err != nil {
		return nil, err
	}
	if err := svc.ensureExternalIDAvailable(nickname, workout); err != nil {
		return nil, err
	}

	var parsed *tracks.Data
	workout.Device = NormalizeDevice(workout.Device)
	if track != nil {
		trackName, data, err := prepareTrackInput(track)
		if err != nil {
			return nil, err
		}
		parsed = data
		workout.Track = trackName
		if device := deviceForTrack(trackName, parsed); device != DeviceGrom {
			workout.Device = device
		}
		if preserveSummary {
			MergeTrackStats(workout, parsed, MergeModeTrackAttach)
		} else {
			parsed.ApplyToWorkout(&workout.StartDate, &workout.DurationSeconds, &workout.Distance)
			parsed.ApplyDurationTotal(&workout.DurationTotalSeconds)
			MergeTrackStats(workout, parsed, MergeModeTrackCreate)
		}
	}

	if err := validateWorkout(workout); err != nil {
		return nil, err
	}

	_, dirName, cleanup, err := svc.repo.BeginCreate(nickname, workout)
	if err != nil {
		return nil, err
	}

	committed := false
	defer func() {
		if !committed {
			if track != nil {
				svc.deleteStagedCharts(nickname, dirName)
			}
			cleanup()
		}
	}()

	if track != nil {
		if err := svc.writeTrackArtifacts(nickname, dirName, track.Data, parsed, workout); err != nil {
			return nil, err
		}
	}
	saved, err := svc.saveWorkoutMedia(nickname, dirName, photos)
	if err != nil {
		return nil, err
	}
	workout.MediaFiles = saved
	workout.HasMedia = len(saved) > 0
	if err := svc.repo.WriteMetadata(nickname, workout); err != nil {
		return nil, err
	}
	committed = true

	result := *workout
	return &result, nil
}

func (svc *Service) deleteStagedCharts(nickname, dirName string) {
	ctx := context.Background()
	for _, store := range []interface {
		DeleteLocal(context.Context, string, string) error
	}{svc.speedCharts, svc.heartRateCharts, svc.cadenceCharts} {
		if store != nil {
			if err := store.DeleteLocal(ctx, nickname, dirName); err != nil {
				slog.Warn("staged workout chart cleanup failed", "dir", dirName, "err", err)
			}
		}
	}
}

// AttachTrack saves a track file and map preview for an existing workout without
// overwriting client-provided start_date, durations, distance, speeds, or elevation.
func (svc *Service) AttachTrack(nickname string, workout *Workout, track *TrackInput) (*Workout, error) {
	if workout == nil || workout.ID == "" {
		return nil, ErrInvalidWorkout
	}
	if track == nil {
		return workout, nil
	}

	trackName, parsed, err := prepareTrackInput(track)
	if err != nil {
		return nil, err
	}

	dirName, err := svc.repo.WorkoutDirName(nickname, workout.ID)
	if err != nil {
		return nil, err
	}

	workout.Track = trackName
	if device := deviceForTrack(trackName, parsed); device != DeviceGrom {
		workout.Device = device
	}

	if err := svc.writeTrackArtifacts(nickname, dirName, track.Data, parsed, workout); err != nil {
		return nil, err
	}

	MergeTrackStats(workout, parsed, MergeModeTrackAttach)

	if err := svc.repo.UpdateMetadata(nickname, workout); err != nil {
		return nil, err
	}

	result := *workout
	return &result, nil
}

func prepareTrackInput(track *TrackInput) (string, *tracks.Data, error) {
	trackName, err := tracks.TrackFileName(track.Filename)
	if err != nil {
		return "", nil, err
	}

	parsed := track.Parsed
	if parsed == nil {
		parsed, err = tracks.Parse(track.Data, track.Filename)
		if err != nil {
			return "", nil, err
		}
	}

	return trackName, parsed, nil
}

func (svc *Service) writeTrackArtifacts(nickname, dirName string, trackData []byte, parsed *tracks.Data, workout *Workout) error {
	ctx := context.Background()
	trackKey := keys.WorkoutTrack(nickname, dirName, workout.Track)
	if _, err := blob.PutBytes(ctx, svc.blobs, trackKey, trackData, blob.PutOptions{}); err != nil {
		return fmt.Errorf("write track: %w", err)
	}

	if err := svc.writeSpeedChart(nickname, dirName, parsed); err != nil {
		slog.Error("speed series write failed", "workout_id", workout.ID, "err", err)
	}

	if err := svc.writeHeartRateChart(nickname, dirName, parsed); err != nil {
		slog.Error("heart rate series write failed", "workout_id", workout.ID, "err", err)
	}

	if err := svc.writeCadenceChart(nickname, dirName, parsed); err != nil {
		slog.Error("cadence series write failed", "workout_id", workout.ID, "err", err)
	}

	if parsed != nil && parsed.HasGPS() {
		if preview, err := maprender.RenderPreview(parsed.Points); err != nil {
			slog.Error("map preview render failed", "workout_id", workout.ID, "err", err)
		} else if len(preview) > 0 {
			previewKey := keys.WorkoutMapPreview(nickname, dirName)
			if _, err := blob.PutBytes(ctx, svc.blobs, previewKey, preview, blob.PutOptions{ContentType: "image/webp"}); err != nil {
				slog.Error("map preview write failed", "workout_id", workout.ID, "err", err)
			} else {
				workout.HasMapPreview = true
			}
		}
	}

	return nil
}

func deviceForTrack(trackName string, parsed *tracks.Data) string {
	if trackName == tracks.TrackFileFIT && parsed != nil && parsed.Device != nil {
		if device := stripStravaFromDevice(strings.TrimSpace(*parsed.Device)); device != "" {
			return device
		}
	}
	return DeviceGrom
}

func (svc *Service) TrackFile(nickname, workoutID string) ([]byte, string, string, error) {
	workout, err := svc.repo.Get(nickname, workoutID)
	if err != nil {
		return nil, "", "", err
	}
	if workout.Track == "" {
		return nil, "", "", ErrWorkoutNotFound
	}

	dirName, err := svc.repo.WorkoutDirName(nickname, workoutID)
	if err != nil {
		return nil, "", "", err
	}

	trackKey := keys.WorkoutTrack(nickname, dirName, workout.Track)
	data, err := blob.ReadAll(context.Background(), svc.blobs, trackKey)
	if err != nil {
		return nil, "", "", fmt.Errorf("read track: %w", err)
	}
	return data, workout.Track, workout.Name, nil
}

func (svc *Service) MapPreview(nickname, workoutID string) ([]byte, error) {
	if _, err := svc.repo.Get(nickname, workoutID); err != nil {
		return nil, err
	}
	dirName, err := svc.repo.WorkoutDirName(nickname, workoutID)
	if err != nil {
		return nil, err
	}
	previewKey := keys.WorkoutMapPreview(nickname, dirName)
	data, err := blob.ReadAll(context.Background(), svc.blobs, previewKey)
	if err != nil {
		return nil, err
	}
	return data, nil
}
