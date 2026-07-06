package maprender

import (
	"bytes"
	"image"
	"image/color"
	"image/png"
	"math"

	sm "github.com/flopp/go-staticmaps"
	"github.com/golang/geo/s2"
	"github.com/solargate/travka/internal/tracks"
)

const (
	PreviewWidth  = 640
	PreviewHeight = 360
	userAgent     = "Travka/1.0 (https://github.com/solargate/travka)"

	// Values > 1 add padding around the track; 1.0 matches track extent.
	boundsFillFraction       = 1.0
	minBoundsHalfSpanDegrees = 0.0012
	trackLineWeight          = 4.0
)

func RenderPreview(points []tracks.LatLng) ([]byte, error) {
	if len(points) < 2 {
		return nil, nil
	}

	simplified := tracks.SimplifyForRender(points)
	positions := make([]s2.LatLng, len(simplified))
	for i, p := range simplified {
		positions[i] = s2.LatLngFromDegrees(p.Lat, p.Lng)
	}

	ctx := sm.NewContext()
	ctx.SetSize(PreviewWidth, PreviewHeight)
	ctx.SetUserAgent(userAgent)
	ctx.SetTileProvider(sm.NewTileProviderCartoLight())
	ctx.SetCache(nil)

	if bbox := tightBounds(simplified); bbox != nil {
		ctx.SetBoundingBox(*bbox)
	}

	path := sm.NewPath(positions, color.RGBA{R: 0, G: 100, B: 255, A: 255}, trackLineWeight)
	ctx.AddObject(path)

	img, err := ctx.Render()
	if err != nil {
		img, err = renderFallback(positions)
		if err != nil {
			return nil, err
		}
	}

	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func renderFallback(positions []s2.LatLng) (image.Image, error) {
	ctx := sm.NewContext()
	ctx.SetSize(PreviewWidth, PreviewHeight)
	ctx.SetTileProvider(sm.NewTileProviderNone())
	ctx.SetBackground(color.RGBA{R: 240, G: 240, B: 240, A: 255})
	ctx.SetCache(nil)

	path := sm.NewPath(positions, color.RGBA{R: 0, G: 100, B: 255, A: 255}, trackLineWeight)
	ctx.AddObject(path)

	return ctx.Render()
}

func tightBounds(points []tracks.LatLng) *s2.Rect {
	if len(points) == 0 {
		return nil
	}

	minLat, maxLat := points[0].Lat, points[0].Lat
	minLng, maxLng := points[0].Lng, points[0].Lng
	for _, p := range points[1:] {
		minLat = math.Min(minLat, p.Lat)
		maxLat = math.Max(maxLat, p.Lat)
		minLng = math.Min(minLng, p.Lng)
		maxLng = math.Max(maxLng, p.Lng)
	}

	centerLat := (minLat + maxLat) / 2
	centerLng := (minLng + maxLng) / 2
	halfLat := math.Max((maxLat-minLat)*boundsFillFraction/2, minBoundsHalfSpanDegrees)
	halfLng := math.Max((maxLng-minLng)*boundsFillFraction/2, minBoundsHalfSpanDegrees)

	bbox, err := sm.CreateBBox(
		centerLat+halfLat,
		centerLng-halfLng,
		centerLat-halfLat,
		centerLng+halfLng,
	)
	if err != nil {
		return nil
	}
	return bbox
}
