package avatars

import (
	"bytes"
	"errors"
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"os"
	"path/filepath"

	"github.com/chai2010/webp"
	"github.com/solargate/grom/internal/data"
	"golang.org/x/image/draw"
)

const (
	MaxUploadBytes = 512 * 1024
	AvatarSize     = 256
)

var (
	ErrInvalidAvatar   = errors.New("invalid avatar image")
	ErrAvatarTooLarge  = errors.New("avatar file too large")
	ErrAvatarNotFound  = errors.New("avatar not found")
	ErrAvatarNotSquare = errors.New("avatar must be square")
)

func Path(dataDir, nickname string) string {
	return data.UserAvatarPath(dataDir, nickname)
}

func APIPath(nickname string) string {
	return fmt.Sprintf("/api/v1/users/%s/avatar", nickname)
}

func PublicURL(domain, nickname string) string {
	if domain == "" {
		domain = "localhost"
	}
	return fmt.Sprintf("https://%s/users/%s/avatar", domain, nickname)
}

func Fields(dataDir, nickname string) (hasAvatar bool, avatarURL string) {
	if !Has(dataDir, nickname) {
		return false, ""
	}
	return true, APIPath(nickname)
}

func Has(dataDir, nickname string) bool {
	_, err := os.Stat(Path(dataDir, nickname))
	return err == nil
}

func Save(dataDir, nickname string, raw []byte) error {
	userDir := data.UserDir(dataDir, nickname)
	if err := os.MkdirAll(userDir, 0700); err != nil {
		return err
	}
	return SaveFile(Path(dataDir, nickname), raw)
}

func SaveFile(path string, raw []byte) error {
	if len(raw) == 0 {
		return ErrInvalidAvatar
	}
	if len(raw) > MaxUploadBytes {
		return ErrAvatarTooLarge
	}

	img, err := decodeImage(raw)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrInvalidAvatar, err)
	}

	bounds := img.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()
	if width < 64 || height < 64 {
		return ErrInvalidAvatar
	}
	if width != height {
		return ErrAvatarNotSquare
	}

	if width != AvatarSize || height != AvatarSize {
		resized := image.NewRGBA(image.Rect(0, 0, AvatarSize, AvatarSize))
		draw.CatmullRom.Scale(resized, resized.Bounds(), img, bounds, draw.Over, nil)
		img = resized
	}

	var buf bytes.Buffer
	if err := webp.Encode(&buf, img, &webp.Options{Quality: 85}); err != nil {
		return fmt.Errorf("encode avatar: %w", err)
	}

	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return err
	}

	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, buf.Bytes(), 0600); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}

func Delete(dataDir, nickname string) error {
	path := Path(dataDir, nickname)
	err := os.Remove(path)
	if err != nil {
		if os.IsNotExist(err) {
			return ErrAvatarNotFound
		}
		return err
	}
	return nil
}

func decodeImage(raw []byte) (image.Image, error) {
	if img, err := webp.Decode(bytes.NewReader(raw)); err == nil {
		return img, nil
	}
	img, _, err := image.Decode(bytes.NewReader(raw))
	return img, err
}
