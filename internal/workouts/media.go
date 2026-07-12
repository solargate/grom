package workouts

import (
	"bytes"
	"errors"
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"os"
	"path/filepath"
	"strings"

	"github.com/chai2010/webp"
	"golang.org/x/image/draw"
)

const (
	MediaSubdir        = "media"
	PreviewPrefix      = "preview-"
	PreviewMaxEdgePx   = 320
	PreviewWebPQuality = 80
	MaxPhotosPerWorkout = 20
	MaxPhotoBytes       = 10 << 20
)

var (
	ErrInvalidPhoto      = errors.New("invalid photo")
	ErrPhotoTooLarge     = errors.New("photo file too large")
	ErrTooManyPhotos     = errors.New("too many photos")
	ErrPhotoNotFound     = errors.New("photo not found")
	ErrInvalidPhotoName  = errors.New("invalid photo filename")
)

type MediaFileInput struct {
	Filename string
	Data     []byte
}

func mediaDir(workoutDir string) string {
	return filepath.Join(workoutDir, MediaSubdir)
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

func uniqueMediaFilename(dir, filename string) string {
	if _, err := os.Stat(filepath.Join(dir, filename)); os.IsNotExist(err) {
		return filename
	}
	ext := filepath.Ext(filename)
	base := strings.TrimSuffix(filename, ext)
	for i := 2; i < 1000; i++ {
		candidate := fmt.Sprintf("%s (%d)%s", base, i, ext)
		if _, err := os.Stat(filepath.Join(dir, candidate)); os.IsNotExist(err) {
			return candidate
		}
	}
	return fmt.Sprintf("%s-%d%s", base, os.Getpid(), ext)
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
	if err := webp.Encode(&buf, resized, &webp.Options{Quality: PreviewWebPQuality}); err != nil {
		return nil, fmt.Errorf("encode preview: %w", err)
	}
	return buf.Bytes(), nil
}

func SaveOriginalAndPreview(workoutDir, filename string, raw []byte) (string, error) {
	return SaveOriginalAndPreviewInDir(mediaDir(workoutDir), filename, raw)
}

func SaveOriginalAndPreviewInDir(dir, filename string, raw []byte) (string, error) {
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

	if err := os.MkdirAll(dir, 0700); err != nil {
		return "", err
	}
	safeName = uniqueMediaFilename(dir, safeName)

	img, err := decodePhoto(raw)
	if err != nil {
		return "", fmt.Errorf("%w: %v", ErrInvalidPhoto, err)
	}

	originalPath := filepath.Join(dir, safeName)
	tmpOriginal := originalPath + ".tmp"
	if err := os.WriteFile(tmpOriginal, raw, 0600); err != nil {
		return "", err
	}
	if err := os.Rename(tmpOriginal, originalPath); err != nil {
		return "", err
	}

	previewData, err := encodePreview(img)
	if err != nil {
		return "", err
	}
	previewPath := filepath.Join(dir, PreviewPrefix+safeName+".webp")
	tmpPreview := previewPath + ".tmp"
	if err := os.WriteFile(tmpPreview, previewData, 0600); err != nil {
		return "", err
	}
	if err := os.Rename(tmpPreview, previewPath); err != nil {
		return "", err
	}

	return safeName, nil
}

func SaveWorkoutMedia(workoutDir string, files []MediaFileInput) ([]string, error) {
	return SaveMediaFilesToDir(mediaDir(workoutDir), files)
}

func SaveMediaFilesToDir(dir string, files []MediaFileInput) ([]string, error) {
	if len(files) == 0 {
		return nil, nil
	}
	if len(files) > MaxPhotosPerWorkout {
		return nil, ErrTooManyPhotos
	}

	saved := make([]string, 0, len(files))
	for _, file := range files {
		name, err := SaveOriginalAndPreviewInDir(dir, file.Filename, file.Data)
		if err != nil {
			return saved, err
		}
		saved = append(saved, name)
	}
	return saved, nil
}

func ListMediaFilesInDir(dir string) ([]string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	files := make([]string, 0)
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if strings.HasPrefix(name, PreviewPrefix) {
			continue
		}
		files = append(files, name)
	}
	return files, nil
}

func ListMediaFiles(workoutDir string) ([]string, error) {
	return ListMediaFilesInDir(mediaDir(workoutDir))
}

func HasMediaInDir(dir string) bool {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return false
	}
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if strings.HasPrefix(name, PreviewPrefix) {
			continue
		}
		return true
	}
	return false
}

func HasMedia(workoutDir string) bool {
	return HasMediaInDir(mediaDir(workoutDir))
}

func MediaOriginalPathInDir(dir, filename string) string {
	return filepath.Join(dir, filename)
}

func MediaPreviewPathInDir(dir, filename string) string {
	return filepath.Join(dir, PreviewPrefix+filename+".webp")
}

func MediaOriginalPath(workoutDir, filename string) string {
	return MediaOriginalPathInDir(mediaDir(workoutDir), filename)
}

func MediaPreviewPath(workoutDir, filename string) string {
	return MediaPreviewPathInDir(mediaDir(workoutDir), filename)
}

func (s *Store) AddMedia(nickname string, workout *Workout, files []MediaFileInput) (*Workout, error) {
	if workout == nil || workout.ID == "" {
		return nil, ErrInvalidWorkout
	}
	dir, err := s.findWorkoutDir(nickname, workout.ID)
	if err != nil {
		return nil, err
	}

	existing, err := ListMediaFiles(dir)
	if err != nil {
		return nil, err
	}
	if len(existing)+len(files) > MaxPhotosPerWorkout {
		return nil, ErrTooManyPhotos
	}

	saved, err := SaveWorkoutMedia(dir, files)
	if err != nil {
		return nil, err
	}

	workout.MediaFiles = append(existing, saved...)
	workout.HasMedia = len(workout.MediaFiles) > 0

	dirName := filepath.Base(dir)
	filePath := filepath.Join(dir, dirName+".yaml")
	if err := writeWorkoutYAML(filePath, workout); err != nil {
		return nil, err
	}

	result := *workout
	return &result, nil
}

func (s *Store) WorkoutDir(nickname, workoutID string) (string, error) {
	return s.findWorkoutDir(nickname, workoutID)
}

func (s *Store) MediaOriginalFile(nickname, workoutID, filename string) (string, error) {
	dir, err := s.findWorkoutDir(nickname, workoutID)
	if err != nil {
		return "", err
	}
	safeName, err := SanitizeMediaFilename(filename)
	if err != nil {
		return "", err
	}
	path := MediaOriginalPath(dir, safeName)
	if _, err := os.Stat(path); err != nil {
		if os.IsNotExist(err) {
			return "", ErrPhotoNotFound
		}
		return "", err
	}
	return path, nil
}

func (s *Store) MediaPreviewFile(nickname, workoutID, filename string) (string, error) {
	dir, err := s.findWorkoutDir(nickname, workoutID)
	if err != nil {
		return "", err
	}
	safeName, err := SanitizeMediaFilename(filename)
	if err != nil {
		return "", err
	}
	path := MediaPreviewPath(dir, safeName)
	if _, err := os.Stat(path); err != nil {
		if os.IsNotExist(err) {
			return "", ErrPhotoNotFound
		}
		return "", err
	}
	return path, nil
}

func (s *Store) ReadMediaPayload(nickname, workoutID string) ([]MediaFileInput, error) {
	dir, err := s.findWorkoutDir(nickname, workoutID)
	if err != nil {
		return nil, err
	}
	workout, err := readWorkoutFromDir(dir)
	if err != nil {
		return nil, err
	}
	files := workout.MediaFiles
	if len(files) == 0 {
		files, err = ListMediaFiles(dir)
		if err != nil {
			return nil, err
		}
	}
	result := make([]MediaFileInput, 0, len(files))
	for _, name := range files {
		path := MediaOriginalPath(dir, name)
		data, err := os.ReadFile(path)
		if err != nil {
			return nil, err
		}
		result = append(result, MediaFileInput{Filename: name, Data: data})
	}
	return result, nil
}

func populateWorkoutMedia(workout *Workout, workoutDir string) {
	if workout == nil {
		return
	}
	if len(workout.MediaFiles) == 0 {
		if files, err := ListMediaFiles(workoutDir); err == nil && len(files) > 0 {
			workout.MediaFiles = files
		}
	}
	workout.HasMedia = len(workout.MediaFiles) > 0 || HasMedia(workoutDir)
	if workout.HasMedia && len(workout.MediaFiles) == 0 {
		if files, err := ListMediaFiles(workoutDir); err == nil {
			workout.MediaFiles = files
		}
	}
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
