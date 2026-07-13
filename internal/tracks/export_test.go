package tracks_test

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/muktihari/fit/encoder"
	"github.com/muktihari/fit/profile/filedef"
	"github.com/muktihari/fit/profile/mesgdef"
	"github.com/muktihari/fit/profile/typedef"
	"github.com/solargate/grom/internal/tracks"
)

func TestExportGPXPassthrough(t *testing.T) {
	data, err := os.ReadFile(filepath.Join("testdata", "sample.gpx"))
	if err != nil {
		t.Fatal(err)
	}

	exported, err := tracks.ExportGPX(data, tracks.TrackFileGPX, "Morning run")
	if err != nil {
		t.Fatalf("ExportGPX() error = %v", err)
	}
	if string(exported) != string(data) {
		t.Fatal("expected GPX passthrough without modification")
	}
}

func TestExportGPXFromFIT(t *testing.T) {
	start := time.Date(2026, 7, 6, 8, 40, 0, 0, time.UTC)
	activity := filedef.NewActivity()
	activity.FileId.Type = typedef.FileActivity
	record1 := &mesgdef.Record{Timestamp: start}
	record1.SetPositionLatDegrees(55.7558).SetPositionLongDegrees(37.6173).SetAltitudeScaled(120)
	record2 := &mesgdef.Record{Timestamp: start.Add(10 * time.Minute)}
	record2.SetPositionLatDegrees(55.7568).SetPositionLongDegrees(37.6273).SetAltitudeScaled(125)
	activity.Records = []*mesgdef.Record{record1, record2}

	fit := activity.ToFIT(nil)
	var buf bytes.Buffer
	if err := encoder.New(&buf).Encode(&fit); err != nil {
		t.Fatalf("Encode() error = %v", err)
	}

	exported, err := tracks.ExportGPX(buf.Bytes(), tracks.TrackFileFIT, "Converted workout")
	if err != nil {
		t.Fatalf("ExportGPX() error = %v", err)
	}
	xml := string(exported)
	if !strings.Contains(xml, "<gpx") || !strings.Contains(xml, "<trkpt") {
		t.Fatalf("expected GPX track points, got: %s", xml[:min(200, len(xml))])
	}
	if !strings.Contains(xml, "Converted workout") {
		t.Fatal("expected workout name in GPX track")
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
