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
	if track == nil {
		return svc.Create(nickname, workout)
	}

	if err := validateWorkout(workout); err != nil {
		return nil, err
	}

	trackName, parsed, err := prepareTrackInput(track)
	if err != nil {
		return nil, err
	}

	startDate := workout.StartDate
	durationSeconds := workout.DurationSeconds
	durationTotalSeconds := workout.DurationTotalSeconds
	distanceMeters := workout.Distance
	parsed.ApplyToWorkout(&startDate, &durationSeconds, &distanceMeters)
	parsed.ApplyDurationTotal(&durationTotalSeconds)

	workout.StartDate = startDate
	workout.DurationSeconds = durationSeconds
	workout.DurationTotalSeconds = durationTotalSeconds
	workout.Distance = distanceMeters
	workout.Track = trackName
	workout.Device = deviceForTrack(trackName, parsed)

	MergeTrackStats(workout, parsed, MergeModeTrackCreate)

	if err := validateWorkout(workout); err != nil {
		return nil, err
	}

	_, dirName, cleanup, err := svc.repo.BeginCreate(nickname, workout)
	if err != nil {
		return nil, err
	}

	if err := svc.writeTrackArtifacts(nickname, dirName, track.Data, parsed, workout); err != nil {
		cleanup()
		return nil, err
	}

	if err := svc.repo.WriteMetadata(nickname, workout); err != nil {
		cleanup()
		return nil, err
	}

	result := *workout
	return &result, nil
}

// AttachTrack saves a track file and map preview for an existing workout without
// overwriting start_date, duration_seconds, or distance from the track.
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

	if err := svc.repo.WriteMetadata(nickname, workout); err != nil {
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
		if device := strings.TrimSpace(*parsed.Device); device != "" {
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
