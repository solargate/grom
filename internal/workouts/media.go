package workouts

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"os"
	"path/filepath"
	"strings"

	"github.com/deepteams/webp"
	"golang.org/x/image/draw"

	"github.com/solargate/grom/internal/storage/blob"
	"github.com/solargate/grom/internal/storage/keys"
)

const (
	MediaSubdir         = "media"
	PreviewPrefix       = "preview-"
	PreviewMaxEdgePx    = 320
	PreviewWebPQuality  = 80
	MaxPhotosPerWorkout = 20
	MaxPhotoBytes       = 10 << 20
)

var (
	ErrInvalidPhoto     = errors.New("invalid photo")
	ErrPhotoTooLarge    = errors.New("photo file too large")
	ErrTooManyPhotos    = errors.New("too many photos")
	ErrPhotoNotFound    = errors.New("photo not found")
	ErrInvalidPhotoName = errors.New("invalid photo filename")
)

type MediaFileInput struct {
	Filename string
	Data     []byte
}

func SanitizeMediaFilename(name string) (string, error) {
	name = strings.TrimSpace(name)
	name = filepath.Base(name)
	name = strings.ReplaceAll(name, string(os.PathSeparator), "_")
	if name == "" || name == "." || name == ".." {
		return "", ErrInvalidPhotoName
	}
	if strings.Contains(name, "..") {
		return "", ErrInvalidPhotoName
	}
	return name, nil
}

func uniqueMediaFilename(store blob.Store, nickname, dirName, filename string) (string, error) {
	ctx := context.Background()
	key := keys.WorkoutMediaOriginal(nickname, dirName, filename)
	exists, err := store.Exists(ctx, key)
	if err != nil {
		return "", err
	}
	if !exists {
		return filename, nil
	}
	ext := filepath.Ext(filename)
	base := strings.TrimSuffix(filename, ext)
	for i := 2; i < 1000; i++ {
		candidate := fmt.Sprintf("%s (%d)%s", base, i, ext)
		key = keys.WorkoutMediaOriginal(nickname, dirName, candidate)
		exists, err = store.Exists(ctx, key)
		if err != nil {
			return "", err
		}
		if !exists {
			return candidate, nil
		}
	}
	return fmt.Sprintf("%s-%d%s", base, os.Getpid(), ext), nil
}

func decodePhoto(raw []byte) (image.Image, error) {
	if img, err := webp.Decode(bytes.NewReader(raw)); err == nil {
		return img, nil
	}
	img, _, err := image.Decode(bytes.NewReader(raw))
	return img, err
}

func encodePreview(img image.Image) ([]byte, error) {
	bounds := img.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()
	if width <= 0 || height <= 0 {
		return nil, ErrInvalidPhoto
	}

	var resized image.Image = img
	if width > PreviewMaxEdgePx || height > PreviewMaxEdgePx {
		if width >= height {
			newW := PreviewMaxEdgePx
			newH := max(1, height*PreviewMaxEdgePx/width)
			dst := image.NewRGBA(image.Rect(0, 0, newW, newH))
			draw.CatmullRom.Scale(dst, dst.Bounds(), img, bounds, draw.Over, nil)
			resized = dst
		} else {
			newH := PreviewMaxEdgePx
			newW := max(1, width*PreviewMaxEdgePx/height)
			dst := image.NewRGBA(image.Rect(0, 0, newW, newH))
			draw.CatmullRom.Scale(dst, dst.Bounds(), img, bounds, draw.Over, nil)
			resized = dst
		}
	}

	var buf bytes.Buffer
	if err := webp.Encode(&buf, resized, &webp.EncoderOptions{Quality: PreviewWebPQuality}); err != nil {
		return nil, fmt.Errorf("encode preview: %w", err)
	}
	return buf.Bytes(), nil
}

func (svc *Service) saveOriginalAndPreview(nickname, dirName, filename string, raw []byte) (string, error) {
	safeName, err := SanitizeMediaFilename(filename)
	if err != nil {
		return "", err
	}
	if len(raw) == 0 {
		return "", ErrInvalidPhoto
	}
	if len(raw) > MaxPhotoBytes {
		return "", ErrPhotoTooLarge
	}

	safeName, err = uniqueMediaFilename(svc.blobs, nickname, dirName, safeName)
	if err != nil {
		return "", err
	}

	img, err := decodePhoto(raw)
	if err != nil {
		return "", fmt.Errorf("%w: %v", ErrInvalidPhoto, err)
	}

	ctx := context.Background()
	originalKey := keys.WorkoutMediaOriginal(nickname, dirName, safeName)
	if _, err := blob.PutBytes(ctx, svc.blobs, originalKey, raw, blob.PutOptions{ContentType: MediaContentType(safeName)}); err != nil {
		return "", err
	}

	previewData, err := encodePreview(img)
	if err != nil {
		return "", err
	}
	previewKey := keys.WorkoutMediaPreview(nickname, dirName, safeName)
	if _, err := blob.PutBytes(ctx, svc.blobs, previewKey, previewData, blob.PutOptions{ContentType: "image/webp"}); err != nil {
		return "", err
	}

	return safeName, nil
}

func (svc *Service) saveWorkoutMedia(nickname, dirName string, files []MediaFileInput) ([]string, error) {
	if len(files) == 0 {
		return nil, nil
	}
	if len(files) > MaxPhotosPerWorkout {
		return nil, ErrTooManyPhotos
	}

	saved := make([]string, 0, len(files))
	for _, file := range files {
		name, err := svc.saveOriginalAndPreview(nickname, dirName, file.Filename, file.Data)
		if err != nil {
			return saved, err
		}
		saved = append(saved, name)
	}
	return saved, nil
}

func (svc *Service) AddMedia(nickname string, workout *Workout, files []MediaFileInput) (*Workout, error) {
	if workout == nil || workout.ID == "" {
		return nil, ErrInvalidWorkout
	}

	stored, err := svc.repo.Get(nickname, workout.ID)
	if err != nil {
		return nil, err
	}
	*workout = *stored

	dirName, err := svc.repo.WorkoutDirName(nickname, workout.ID)
	if err != nil {
		return nil, err
	}

	existing := workout.MediaFiles
	if len(existing)+len(files) > MaxPhotosPerWorkout {
		return nil, ErrTooManyPhotos
	}

	saved, err := svc.saveWorkoutMedia(nickname, dirName, files)
	if err != nil {
		return nil, err
	}

	workout.MediaFiles = append(existing, saved...)
	workout.HasMedia = len(workout.MediaFiles) > 0

	if err := svc.repo.WriteMetadata(nickname, workout); err != nil {
		return nil, err
	}

	svc.enrichWorkout(nickname, workout)
	result := *workout
	return &result, nil
}

func (svc *Service) MediaOriginal(nickname, workoutID, filename string) ([]byte, string, error) {
	if _, err := svc.repo.Get(nickname, workoutID); err != nil {
		return nil, "", err
	}
	dirName, err := svc.repo.WorkoutDirName(nickname, workoutID)
	if err != nil {
		return nil, "", err
	}
	safeName, err := SanitizeMediaFilename(filename)
	if err != nil {
		return nil, "", err
	}
	data, err := blob.ReadAll(context.Background(), svc.blobs, keys.WorkoutMediaOriginal(nickname, dirName, safeName))
	if err != nil {
		if os.IsNotExist(err) {
			return nil, "", ErrPhotoNotFound
		}
		return nil, "", err
	}
	return data, MediaContentType(safeName), nil
}

func (svc *Service) MediaPreview(nickname, workoutID, filename string) ([]byte, error) {
	if _, err := svc.repo.Get(nickname, workoutID); err != nil {
		return nil, err
	}
	dirName, err := svc.repo.WorkoutDirName(nickname, workoutID)
	if err != nil {
		return nil, err
	}
	safeName, err := SanitizeMediaFilename(filename)
	if err != nil {
		return nil, err
	}
	data, err := blob.ReadAll(context.Background(), svc.blobs, keys.WorkoutMediaPreview(nickname, dirName, safeName))
	if err != nil {
		if os.IsNotExist(err) {
			return nil, ErrPhotoNotFound
		}
		return nil, err
	}
	return data, nil
}

func (svc *Service) ReadMediaPayload(nickname, workoutID string) ([]MediaFileInput, error) {
	workout, err := svc.repo.Get(nickname, workoutID)
	if err != nil {
		return nil, err
	}
	dirName, err := svc.repo.WorkoutDirName(nickname, workoutID)
	if err != nil {
		return nil, err
	}

	files := workout.MediaFiles
	result := make([]MediaFileInput, 0, len(files))
	for _, name := range files {
		data, err := blob.ReadAll(context.Background(), svc.blobs, keys.WorkoutMediaOriginal(nickname, dirName, name))
		if err != nil {
			return nil, err
		}
		result = append(result, MediaFileInput{Filename: name, Data: data})
	}
	return result, nil
}

func MediaContentType(filename string) string {
	switch strings.ToLower(filepath.Ext(filename)) {
	case ".jpg", ".jpeg":
		return "image/jpeg"
	case ".png":
		return "image/png"
	case ".webp":
		return "image/webp"
	case ".gif":
		return "image/gif"
	case ".heic", ".heif":
		return "image/heic"
	default:
		return "application/octet-stream"
	}
}

func SaveFederatedMedia(blobs blob.Store, viewerNickname, ownerKey, workoutID string, files []MediaFileInput) ([]string, error) {
	if len(files) == 0 {
		return nil, nil
	}
	if len(files) > MaxPhotosPerWorkout {
		return nil, ErrTooManyPhotos
	}

	saved := make([]string, 0, len(files))
	for _, file := range files {
		name, err := saveFederatedPhoto(blobs, viewerNickname, ownerKey, workoutID, file.Filename, file.Data)
		if err != nil {
			return saved, err
		}
		saved = append(saved, name)
	}
	return saved, nil
}

func saveFederatedPhoto(blobs blob.Store, viewerNickname, ownerKey, workoutID, filename string, raw []byte) (string, error) {
	safeName, err := SanitizeMediaFilename(filename)
	if err != nil {
		return "", err
	}
	if len(raw) == 0 {
		return "", ErrInvalidPhoto
	}
	if len(raw) > MaxPhotoBytes {
		return "", ErrPhotoTooLarge
	}

	ctx := context.Background()
	safeName, err = uniqueFederatedMediaFilename(blobs, viewerNickname, ownerKey, workoutID, safeName)
	if err != nil {
		return "", err
	}

	img, err := decodePhoto(raw)
	if err != nil {
		return "", fmt.Errorf("%w: %v", ErrInvalidPhoto, err)
	}

	originalKey := keys.FederatedInboxMediaOriginal(viewerNickname, ownerKey, workoutID, safeName)
	if _, err := blob.PutBytes(ctx, blobs, originalKey, raw, blob.PutOptions{ContentType: MediaContentType(safeName)}); err != nil {
		return "", err
	}

	previewData, err := encodePreview(img)
	if err != nil {
		return "", err
	}
	previewKey := keys.FederatedInboxMediaPreview(viewerNickname, ownerKey, workoutID, safeName)
	if _, err := blob.PutBytes(ctx, blobs, previewKey, previewData, blob.PutOptions{ContentType: "image/webp"}); err != nil {
		return "", err
	}

	return safeName, nil
}

func uniqueFederatedMediaFilename(blobs blob.Store, viewerNickname, ownerKey, workoutID, filename string) (string, error) {
	ctx := context.Background()
	key := keys.FederatedInboxMediaOriginal(viewerNickname, ownerKey, workoutID, filename)
	exists, err := blobs.Exists(ctx, key)
	if err != nil {
		return "", err
	}
	if !exists {
		return filename, nil
	}
	ext := filepath.Ext(filename)
	base := strings.TrimSuffix(filename, ext)
	for i := 2; i < 1000; i++ {
		candidate := fmt.Sprintf("%s (%d)%s", base, i, ext)
		key = keys.FederatedInboxMediaOriginal(viewerNickname, ownerKey, workoutID, candidate)
		exists, err = blobs.Exists(ctx, key)
		if err != nil {
			return "", err
		}
		if !exists {
			return candidate, nil
		}
	}
	return fmt.Sprintf("%s-%d%s", base, os.Getpid(), ext), nil
}

// SaveWorkoutMedia saves photos using the service blob store (used by federation inbox wiring).
func SaveWorkoutMedia(svc *Service, nickname, dirName string, files []MediaFileInput) ([]string, error) {
	return svc.saveWorkoutMedia(nickname, dirName, files)
}
