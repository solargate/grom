package workouts

import (
	"encoding/binary"
	"fmt"
	"math"
	"time"
)

const (
	heartRateChartMagic   = "GRHR"
	heartRateChartVersion = 1

	heartRateChartFlagHasDistance = 1 << 0
)

// MarshalHeartRateChartBinary encodes chart samples as packed little-endian bytes (bbolt driver).
// Layout: magic "GRHR" | u8 version | u8 flags | u32 n | n×i64 unix_sec | n×f32 bpm
// [| n×f32 distance_m when flags&has_distance]. Missing distances encode as NaN.
func MarshalHeartRateChartBinary(samples []HeartRateSample) ([]byte, error) {
	if len(samples) == 0 {
		return nil, nil
	}
	n := len(samples)
	var flags uint8
	for _, s := range samples {
		if s.DistanceM != nil {
			flags |= heartRateChartFlagHasDistance
			break
		}
	}
	size := 4 + 1 + 1 + 4 + n*(8+4)
	if flags&heartRateChartFlagHasDistance != 0 {
		size += n * 4
	}
	out := make([]byte, 0, size)
	out = append(out, heartRateChartMagic...)
	out = append(out, heartRateChartVersion, flags)
	out = binary.LittleEndian.AppendUint32(out, uint32(n))
	for _, s := range samples {
		out = binary.LittleEndian.AppendUint64(out, uint64(s.Time.UTC().Unix()))
	}
	for _, s := range samples {
		out = binary.LittleEndian.AppendUint32(out, math.Float32bits(float32(s.BPM)))
	}
	if flags&heartRateChartFlagHasDistance != 0 {
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

// UnmarshalHeartRateChartBinary decodes a packed heart-rate chart payload.
func UnmarshalHeartRateChartBinary(data []byte) ([]HeartRateSample, error) {
	if len(data) == 0 {
		return nil, nil
	}
	const header = 4 + 1 + 1 + 4
	if len(data) < header {
		return nil, fmt.Errorf("heart rate chart binary: truncated header")
	}
	if string(data[:4]) != heartRateChartMagic {
		return nil, fmt.Errorf("heart rate chart binary: bad magic")
	}
	if data[4] != heartRateChartVersion {
		return nil, fmt.Errorf("heart rate chart binary: unsupported version %d", data[4])
	}
	flags := data[5]
	n := int(binary.LittleEndian.Uint32(data[6:10]))
	if n > HeartRateChartMaxPoints {
		return nil, fmt.Errorf("heart rate chart binary: invalid count %d", n)
	}
	need := header + n*(8+4)
	hasDist := flags&heartRateChartFlagHasDistance != 0
	if hasDist {
		need += n * 4
	}
	if len(data) < need {
		return nil, fmt.Errorf("heart rate chart binary: truncated payload")
	}
	if n == 0 {
		return nil, nil
	}
	off := header
	out := make([]HeartRateSample, n)
	for i := 0; i < n; i++ {
		sec := int64(binary.LittleEndian.Uint64(data[off : off+8]))
		out[i].Time = time.Unix(sec, 0).UTC()
		off += 8
	}
	for i := 0; i < n; i++ {
		out[i].BPM = float64(math.Float32frombits(binary.LittleEndian.Uint32(data[off : off+4])))
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
