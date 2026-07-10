package avatars

import (
	"bytes"
	"image"
	"image/color"
	"image/png"
	"testing"
)

func TestSaveSquarePNG(t *testing.T) {
	dir := t.TempDir()
	img := image.NewRGBA(image.Rect(0, 0, 256, 256))
	for y := 0; y < 256; y++ {
		for x := 0; x < 256; x++ {
			img.Set(x, y, color.RGBA{R: 120, G: 80, B: 200, A: 255})
		}
	}
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		t.Fatal(err)
	}

	if err := Save(dir, "alice", buf.Bytes()); err != nil {
		t.Fatal(err)
	}
	if !Has(dir, "alice") {
		t.Fatal("expected avatar file to exist")
	}
}

func TestSaveRejectsNonSquare(t *testing.T) {
	dir := t.TempDir()
	img := image.NewRGBA(image.Rect(0, 0, 256, 128))
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		t.Fatal(err)
	}
	if err := Save(dir, "alice", buf.Bytes()); err != ErrAvatarNotSquare {
		t.Fatalf("expected ErrAvatarNotSquare, got %v", err)
	}
}
