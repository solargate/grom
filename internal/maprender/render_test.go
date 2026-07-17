package maprender

import (
	"bytes"
	"image"
	"testing"

	"github.com/chai2010/webp"
	"github.com/solargate/grom/internal/tracks"
)

func TestTightBounds(t *testing.T) {
	if got := tightBounds(nil); got != nil {
		t.Fatal("expected nil for empty points")
	}
	if got := tightBounds([]tracks.LatLng{}); got != nil {
		t.Fatal("expected nil for empty slice")
	}

	one := tightBounds([]tracks.LatLng{{Lat: 55.75, Lng: 37.61}})
	if one == nil {
		t.Fatal("expected bounds for one point")
	}

	multi := tightBounds([]tracks.LatLng{
		{Lat: 55.0, Lng: 37.0},
		{Lat: 56.0, Lng: 38.0},
	})
	if multi == nil {
		t.Fatal("expected bounds for multi points")
	}
}

func TestRenderPreviewTooFewPoints(t *testing.T) {
	data, err := RenderPreview([]tracks.LatLng{{Lat: 1, Lng: 2}})
	if err != nil {
		t.Fatal(err)
	}
	if data != nil {
		t.Fatal("expected nil for <2 points")
	}
}

func TestRenderPreviewProducesWebP(t *testing.T) {
	points := []tracks.LatLng{
		{Lat: 55.75, Lng: 37.61},
		{Lat: 55.76, Lng: 37.62},
		{Lat: 55.77, Lng: 37.63},
	}
	data, err := RenderPreview(points)
	if err != nil {
		t.Fatalf("RenderPreview: %v", err)
	}
	if len(data) == 0 {
		t.Fatal("expected webp bytes")
	}
	img, err := webp.Decode(bytes.NewReader(data))
	if err != nil {
		t.Fatalf("webp.Decode: %v", err)
	}
	bounds := img.Bounds()
	if bounds.Dx() != PreviewWidth || bounds.Dy() != PreviewHeight {
		t.Fatalf("size = %dx%d, want %dx%d", bounds.Dx(), bounds.Dy(), PreviewWidth, PreviewHeight)
	}
	_ = image.Image(img)
}
