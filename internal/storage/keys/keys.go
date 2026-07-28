package keys

import (
	"path/filepath"
	"strings"
	"time"

	"github.com/solargate/grom/internal/data"
)

const (
	MapPreviewFileName = "map-preview.webp"
	MediaSubdir        = "media"
	PreviewPrefix      = "preview-"
	SpeedChartFileJSON     = "speed-chart.json"
	HeartRateChartFileJSON = "heartrate-chart.json"
)

// WorkoutDirName returns the workout directory basename ({startDate}-{id}).
func WorkoutDirName(startDate time.Time, id string) string {
	iso := startDate.UTC().Format("2006-01-02T15:04:05Z")
	iso = strings.ReplaceAll(iso, ":", "")
	return iso + "-" + id
}

// UserAvatar returns the logical storage key for a local user avatar.
func UserAvatar(nickname string) string {
	return filepath.Join(data.UsersSubdir, nickname, data.AvatarFileName)
}

// WorkoutTrack returns the logical storage key for a workout track file.
func WorkoutTrack(nickname, workoutDirName, trackName string) string {
	return filepath.Join(data.UsersSubdir, nickname, "workouts", workoutDirName, trackName)
}

// WorkoutMapPreview returns the logical storage key for a workout map preview.
func WorkoutMapPreview(nickname, workoutDirName string) string {
	return filepath.Join(data.UsersSubdir, nickname, "workouts", workoutDirName, MapPreviewFileName)
}

// WorkoutSpeed returns the logical storage key for a workout speed sidecar file.
func WorkoutSpeed(nickname, workoutDirName, filename string) string {
	return filepath.Join(data.UsersSubdir, nickname, "workouts", workoutDirName, filename)
}

// WorkoutMediaOriginal returns the logical storage key for an original workout photo.
func WorkoutMediaOriginal(nickname, workoutDirName, filename string) string {
	return filepath.Join(data.UsersSubdir, nickname, "workouts", workoutDirName, MediaSubdir, filename)
}

// WorkoutMediaPreview returns the logical storage key for a workout photo preview.
func WorkoutMediaPreview(nickname, workoutDirName, filename string) string {
	return filepath.Join(
		data.UsersSubdir,
		nickname,
		"workouts",
		workoutDirName,
		MediaSubdir,
		PreviewPrefix+filename+".webp",
	)
}

// UserActorKey returns the logical storage key for a federation actor private key.
func UserActorKey(nickname string) string {
	return filepath.Join(data.UsersSubdir, nickname, "federation", "actor_key.pem")
}

// FederatedInboxOwnerDir returns the inbox owner directory key prefix.
func FederatedInboxOwnerDir(viewerNickname, ownerKey string) string {
	return filepath.Join(data.UsersSubdir, viewerNickname, "federation", "inbox", "workouts", ownerKey)
}

func FederatedInboxTrack(viewerNickname, ownerKey, workoutID, trackName string) string {
	return filepath.Join(FederatedInboxOwnerDir(viewerNickname, ownerKey), workoutID+"_"+trackName)
}

func FederatedInboxMapPreview(viewerNickname, ownerKey, workoutID string) string {
	return filepath.Join(FederatedInboxOwnerDir(viewerNickname, ownerKey), workoutID+"_"+MapPreviewFileName)
}

// FederatedInboxSpeed returns the logical storage key for a federated workout speed sidecar.
func FederatedInboxSpeed(viewerNickname, ownerKey, workoutID, filename string) string {
	return filepath.Join(FederatedInboxOwnerDir(viewerNickname, ownerKey), workoutID+"_"+filename)
}

func FederatedInboxMediaOriginal(viewerNickname, ownerKey, workoutID, filename string) string {
	return filepath.Join(FederatedInboxOwnerDir(viewerNickname, ownerKey), workoutID+"_"+MediaSubdir, filename)
}

func FederatedInboxMediaPreview(viewerNickname, ownerKey, workoutID, filename string) string {
	return filepath.Join(
		FederatedInboxOwnerDir(viewerNickname, ownerKey),
		workoutID+"_"+MediaSubdir,
		PreviewPrefix+filename+".webp",
	)
}

func FederatedInboxAvatar(viewerNickname, ownerKey string) string {
	return filepath.Join(FederatedInboxOwnerDir(viewerNickname, ownerKey), data.AvatarFileName)
}
