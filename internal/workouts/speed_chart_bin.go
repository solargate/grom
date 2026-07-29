package workouts

import (
	"encoding/binary"
	"fmt"
	"math"
	"time"
)

const (
	speedChartMagic   = "GRSC"
	speedChartVersion = 1
)

// MarshalSpeedChartBinary encodes chart samples as packed little-endian bytes (bbolt driver).
// Layout: magic "GRSC" | u8 version | u32 n | n×i64 unix_sec | n×f32 speed_kmh | n×f32 distance_m.
func MarshalSpeedChartBinary(samples []SpeedSample) ([]byte, error) {
	if len(samples) == 0 {
		return nil, nil
	}
	n := len(samples)
	out := make([]byte, 0, 4+1+4+n*(8+4+4))
	out = append(out, speedChartMagic...)
	out = append(out, speedChartVersion)
	out = binary.LittleEndian.AppendUint32(out, uint32(n))
	for _, s := range samples {
		out = binary.LittleEndian.AppendUint64(out, uint64(s.Time.UTC().Unix()))
	}
	for _, s := range samples {
		out = binary.LittleEndian.AppendUint32(out, math.Float32bits(float32(s.SpeedKmh)))
	}
	for _, s := range samples {
		out = binary.LittleEndian.AppendUint32(out, math.Float32bits(float32(s.DistanceM)))
	}
	return out, nil
}

// UnmarshalSpeedChartBinary decodes a packed speed chart payload.
func UnmarshalSpeedChartBinary(data []byte) ([]SpeedSample, error) {
	if len(data) == 0 {
		return nil, nil
	}
	const header = 4 + 1 + 4
	if len(data) < header {
		return nil, fmt.Errorf("speed chart binary: truncated header")
	}
	if string(data[:4]) != speedChartMagic {
		return nil, fmt.Errorf("speed chart binary: bad magic")
	}
	if data[4] != speedChartVersion {
		return nil, fmt.Errorf("speed chart binary: unsupported version %d", data[4])
	}
	n := int(binary.LittleEndian.Uint32(data[5:9]))
	if n > SpeedChartMaxPoints {
		return nil, fmt.Errorf("speed chart binary: invalid count %d", n)
	}
	need := header + n*(8+4+4)
	if len(data) < need {
		return nil, fmt.Errorf("speed chart binary: truncated payload")
	}
	if n == 0 {
		return nil, nil
	}
	off := header
	out := make([]SpeedSample, n)
	for i := 0; i < n; i++ {
		sec := int64(binary.LittleEndian.Uint64(data[off : off+8]))
		out[i].Time = time.Unix(sec, 0).UTC()
		off += 8
	}
	for i := 0; i < n; i++ {
		out[i].SpeedKmh = float64(math.Float32frombits(binary.LittleEndian.Uint32(data[off : off+4])))
		off += 4
	}
	for i := 0; i < n; i++ {
		out[i].DistanceM = float64(math.Float32frombits(binary.LittleEndian.Uint32(data[off : off+4])))
		off += 4
	}
	return out, nil
}
