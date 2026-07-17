package avatars

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

	"github.com/chai2010/webp"
	"github.com/solargate/grom/internal/data"
	"github.com/solargate/grom/internal/storage/blob"
	"github.com/solargate/grom/internal/storage/keys"
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

func HasStore(store blob.Store, nickname string) bool {
	ok, err := store.Exists(context.Background(), keys.UserAvatar(nickname))
	return err == nil && ok
}

func FieldsStore(store blob.Store, nickname string) (hasAvatar bool, avatarURL string) {
	if !HasStore(store, nickname) {
		return false, ""
	}
	return true, APIPath(nickname)
}

// Fields checks avatar presence on the filesystem (legacy helper for callers without blob store).
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

func SaveStore(store blob.Store, nickname string, raw []byte) error {
	return SaveKey(store, keys.UserAvatar(nickname), raw)
}

func SaveKey(store blob.Store, key string, raw []byte) error {
	encoded, err := prepareAvatarBytes(raw)
	if err != nil {
		return err
	}
	_, err = blob.PutBytes(context.Background(), store, key, encoded, blob.PutOptions{ContentType: "image/webp"})
	return err
}

func Save(dataDir, nickname string, raw []byte) error {
	userDir := data.UserDir(dataDir, nickname)
	if err := os.MkdirAll(userDir, 0700); err != nil {
		return err
	}
	return SaveFile(Path(dataDir, nickname), raw)
}

func SaveFile(path string, raw []byte) error {
	encoded, err := prepareAvatarBytes(raw)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return err
	}
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, encoded, 0600); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}

func LoadStore(store blob.Store, nickname string) ([]byte, error) {
	data, err := blob.ReadAll(context.Background(), store, keys.UserAvatar(nickname))
	if err != nil {
		if os.IsNotExist(err) {
			return nil, ErrAvatarNotFound
		}
		return nil, err
	}
	return data, nil
}

func Load(dataDir, nickname string) ([]byte, error) {
	path := Path(dataDir, nickname)
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, ErrAvatarNotFound
		}
		return nil, err
	}
	return data, nil
}

func DeleteStore(store blob.Store, nickname string) error {
	err := store.Delete(context.Background(), keys.UserAvatar(nickname))
	if err != nil {
		if os.IsNotExist(err) {
			return ErrAvatarNotFound
		}
		return err
	}
	return nil
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

func prepareAvatarBytes(raw []byte) ([]byte, error) {
	if len(raw) == 0 {
		return nil, ErrInvalidAvatar
	}
	if len(raw) > MaxUploadBytes {
		return nil, ErrAvatarTooLarge
	}

	img, err := decodeImage(raw)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidAvatar, err)
	}

	bounds := img.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()
	if width < 64 || height < 64 {
		return nil, ErrInvalidAvatar
	}
	if width != height {
		return nil, ErrAvatarNotSquare
	}

	if width != AvatarSize || height != AvatarSize {
		resized := image.NewRGBA(image.Rect(0, 0, AvatarSize, AvatarSize))
		draw.CatmullRom.Scale(resized, resized.Bounds(), img, bounds, draw.Over, nil)
		img = resized
	}

	var buf bytes.Buffer
	if err := webp.Encode(&buf, img, &webp.Options{Quality: 85}); err != nil {
		return nil, fmt.Errorf("encode avatar: %w", err)
	}
	return buf.Bytes(), nil
}

func decodeImage(raw []byte) (image.Image, error) {
	if img, err := webp.Decode(bytes.NewReader(raw)); err == nil {
		return img, nil
	}
	img, _, err := image.Decode(bytes.NewReader(raw))
	return img, err
}
