package workouts

import (
	"encoding/binary"
	"fmt"
	"math"
	"time"
)

const (
	cadenceChartMagic   = "GRCD"
	cadenceChartVersion = 1

	cadenceChartFlagHasDistance = 1 << 0
)

// MarshalCadenceChartBinary encodes chart samples as packed little-endian bytes (bbolt driver).
// Layout: magic "GRCD" | u8 version | u8 flags | u32 n | n×i64 unix_sec | n×f32 cadence
// [| n×f32 distance_m when flags&has_distance]. Missing distances encode as NaN.
func MarshalCadenceChartBinary(samples []CadenceSample) ([]byte, error) {
	if len(samples) == 0 {
		return nil, nil
	}
	n := len(samples)
	var flags uint8
	for _, s := range samples {
		if s.DistanceM != nil {
			flags |= cadenceChartFlagHasDistance
			break
		}
	}
	size := 4 + 1 + 1 + 4 + n*(8+4)
	if flags&cadenceChartFlagHasDistance != 0 {
		size += n * 4
	}
	out := make([]byte, 0, size)
	out = append(out, cadenceChartMagic...)
	out = append(out, cadenceChartVersion, flags)
	out = binary.LittleEndian.AppendUint32(out, uint32(n))
	for _, s := range samples {
		out = binary.LittleEndian.AppendUint64(out, uint64(s.Time.UTC().Unix()))
	}
	for _, s := range samples {
		out = binary.LittleEndian.AppendUint32(out, math.Float32bits(float32(s.Cadence)))
	}
	if flags&cadenceChartFlagHasDistance != 0 {
		for _, s := range samples {
			var bits uint32
			if s.DistanceM != nil {
				bits = math.Float32bits(float32(*s.DistanceM))
			} else {
				bits = math.Float32bits(float32(math.NaN()))
			}
			out = binary.LittleEndian.AppendUint32(out, bits)
		}
	}
	return out, nil
}

// UnmarshalCadenceChartBinary decodes a packed cadence chart payload.
func UnmarshalCadenceChartBinary(data []byte) ([]CadenceSample, error) {
	if len(data) == 0 {
		return nil, nil
	}
	const header = 4 + 1 + 1 + 4
	if len(data) < header {
		return nil, fmt.Errorf("cadence chart binary: truncated header")
	}
	if string(data[:4]) != cadenceChartMagic {
		return nil, fmt.Errorf("cadence chart binary: bad magic")
	}
	if data[4] != cadenceChartVersion {
		return nil, fmt.Errorf("cadence chart binary: unsupported version %d", data[4])
	}
	flags := data[5]
	n := int(binary.LittleEndian.Uint32(data[6:10]))
	if n > CadenceChartMaxPoints {
		return nil, fmt.Errorf("cadence chart binary: invalid count %d", n)
	}
	need := header + n*(8+4)
	hasDist := flags&cadenceChartFlagHasDistance != 0
	if hasDist {
		need += n * 4
	}
	if len(data) < need {
		return nil, fmt.Errorf("cadence chart binary: truncated payload")
	}
	if n == 0 {
		return nil, nil
	}
	off := header
	out := make([]CadenceSample, n)
	for i := 0; i < n; i++ {
		sec := int64(binary.LittleEndian.Uint64(data[off : off+8]))
		out[i].Time = time.Unix(sec, 0).UTC()
		off += 8
	}
	for i := 0; i < n; i++ {
		out[i].Cadence = float64(math.Float32frombits(binary.LittleEndian.Uint32(data[off : off+4])))
		off += 4
	}
	if hasDist {
		for i := 0; i < n; i++ {
			v := math.Float32frombits(binary.LittleEndian.Uint32(data[off : off+4]))
			off += 4
			if !math.IsNaN(float64(v)) {
				d := float64(v)
				out[i].DistanceM = &d
			}
		}
	}
	return out, nil
}
