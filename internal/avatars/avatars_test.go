package avatars_test

import (
	"bytes"
	"image"
	"image/color"
	"image/png"
	"os"
	"testing"

	"github.com/deepteams/webp"
	"github.com/solargate/grom/internal/avatars"
	blobfs "github.com/solargate/grom/internal/storage/blob/fs"
)

func encodeSquarePNG(t *testing.T, size int) []byte {
	t.Helper()
	img := image.NewRGBA(image.Rect(0, 0, size, size))
	for y := 0; y < size; y++ {
		for x := 0; x < size; x++ {
			img.Set(x, y, color.RGBA{R: 120, G: 80, B: 200, A: 255})
		}
	}
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		t.Fatal(err)
	}
	return buf.Bytes()
}

func TestSaveSquarePNG(t *testing.T) {
	dir := t.TempDir()
	raw := encodeSquarePNG(t, 256)

	if err := avatars.Save(dir, "alice", raw); err != nil {
		t.Fatal(err)
	}
	if !avatars.Has(dir, "alice") {
		t.Fatal("expected avatar file to exist")
	}

	path := avatars.Path(dir, "alice")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	img, err := webp.Decode(bytes.NewReader(data))
	if err != nil {
		t.Fatalf("expected webp: %v", err)
	}
	b := img.Bounds()
	if b.Dx() != avatars.AvatarSize || b.Dy() != avatars.AvatarSize {
		t.Fatalf("size = %dx%d, want %d", b.Dx(), b.Dy(), avatars.AvatarSize)
	}
}

func TestSaveRejectsNonSquare(t *testing.T) {
	dir := t.TempDir()
	img := image.NewRGBA(image.Rect(0, 0, 256, 128))
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		t.Fatal(err)
	}
	if err := avatars.Save(dir, "alice", buf.Bytes()); err != avatars.ErrAvatarNotSquare {
		t.Fatalf("expected ErrAvatarNotSquare, got %v", err)
	}
}

func TestSaveStoreRoundTrip(t *testing.T) {
	dir := t.TempDir()
	store := blobfs.NewStore(dir)
	raw := encodeSquarePNG(t, 128)

	if err := avatars.SaveStore(store, "bob", raw); err != nil {
		t.Fatal(err)
	}
	if !avatars.HasStore(store, "bob") {
		t.Fatal("expected HasStore true")
	}
	data, err := avatars.LoadStore(store, "bob")
	if err != nil {
		t.Fatal(err)
	}
	img, err := webp.Decode(bytes.NewReader(data))
	if err != nil {
		t.Fatal(err)
	}
	if img.Bounds().Dx() != avatars.AvatarSize {
		t.Fatalf("width = %d", img.Bounds().Dx())
	}
	if err := avatars.DeleteStore(store, "bob"); err != nil {
		t.Fatal(err)
	}
	if avatars.HasStore(store, "bob") {
		t.Fatal("expected deleted")
	}
}

func TestAPIPathAndPublicURL(t *testing.T) {
	if got := avatars.APIPath("alice"); got != "/api/v1/users/alice/avatar" {
		t.Fatalf("APIPath = %q", got)
	}
	if got := avatars.PublicURL("example.com", "alice"); got != "https://example.com/users/alice/avatar" {
		t.Fatalf("PublicURL = %q", got)
	}
}
