package tracks

import "math"

const maxRenderPoints = 500

// SimplifyForRender reduces point count for map preview rendering.
func SimplifyForRender(points []LatLng) []LatLng {
	if len(points) <= maxRenderPoints {
		return points
	}
	step := float64(len(points)-1) / float64(maxRenderPoints-1)
	simplified := make([]LatLng, 0, maxRenderPoints)
	for i := 0; i < maxRenderPoints; i++ {
		idx := int(math.Round(float64(i) * step))
		if idx >= len(points) {
			idx = len(points) - 1
		}
		simplified = append(simplified, points[idx])
	}
	return simplified
}
