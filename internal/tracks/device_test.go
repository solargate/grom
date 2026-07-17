package tracks

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/muktihari/fit/encoder"
	"github.com/muktihari/fit/profile/filedef"
	"github.com/muktihari/fit/profile/mesgdef"
	"github.com/muktihari/fit/profile/typedef"
)

func TestExtractDeviceFromProductName(t *testing.T) {
	activity := filedef.NewActivity()
	activity.DeviceInfos = []*mesgdef.DeviceInfo{
		{
			DeviceIndex:  typedef.DeviceIndexCreator,
			ProductName:  "Garmin Edge 530",
			Manufacturer: typedef.ManufacturerGarmin,
		},
	}

	if got := extractDevice(activity); got != "Garmin Edge 530" {
		t.Fatalf("extractDevice() = %q, want %q", got, "Garmin Edge 530")
	}
}

func TestExtractDeviceFromManufacturerProduct(t *testing.T) {
	activity := filedef.NewActivity()
	activity.FileId.Manufacturer = typedef.ManufacturerGarmin
	activity.FileId.Product = uint16(typedef.GarminProductEdge530)

	if got := extractDevice(activity); got != "Garmin Edge 530" {
		t.Fatalf("extractDevice() = %q, want %q", got, "Garmin Edge 530")
	}
}

func TestExtractDeviceCombinesWahooProductName(t *testing.T) {
	activity := filedef.NewActivity()
	activity.DeviceInfos = []*mesgdef.DeviceInfo{
		{
			DeviceIndex:  typedef.DeviceIndexCreator,
			ProductName:  "ELEMNT",
			Manufacturer: typedef.ManufacturerWahooFitness,
			Product:      28,
		},
	}

	if got := extractDevice(activity); got != "Wahoo ELEMNT" {
		t.Fatalf("extractDevice() = %q, want %q", got, "Wahoo ELEMNT")
	}
}

func TestExtractDeviceKeepsProductNameWithBrand(t *testing.T) {
	activity := filedef.NewActivity()
	activity.DeviceInfos = []*mesgdef.DeviceInfo{
		{
			DeviceIndex:  typedef.DeviceIndexCreator,
			ProductName:  "Wahoo SPEED",
			Manufacturer: typedef.ManufacturerWahooFitness,
		},
	}

	if got := extractDevice(activity); got != "Wahoo SPEED" {
		t.Fatalf("extractDevice() = %q, want %q", got, "Wahoo SPEED")
	}
}

func TestParseWahooElemntDevice(t *testing.T) {
	data, err := os.ReadFile(filepath.Join("..", "..", "testdata", "tracks", "1-ride.fit"))
	if err != nil {
		t.Fatalf("sample FIT not available: %v", err)
	}

	parsed, err := Parse(data, "1-ride.fit")
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}
	if parsed.Device == nil {
		t.Fatal("expected device from FIT")
	}
	if *parsed.Device != "Wahoo ELEMNT" {
		t.Fatalf("device = %q, want %q", *parsed.Device, "Wahoo ELEMNT")
	}
}

func TestParseFITDevice(t *testing.T) {
	start := time.Date(2026, 7, 6, 8, 40, 0, 0, time.UTC)
	activity := filedef.NewActivity()
	activity.FileId.Type = typedef.FileActivity
	activity.FileId.Manufacturer = typedef.ManufacturerGarmin
	activity.FileId.Product = uint16(typedef.GarminProductEdge530)
	activity.DeviceInfos = []*mesgdef.DeviceInfo{
		{
			Timestamp:    start,
			DeviceIndex:  typedef.DeviceIndexCreator,
			ProductName:  "Garmin Edge 530",
			Manufacturer: typedef.ManufacturerGarmin,
			Product:      uint16(typedef.GarminProductEdge530),
		},
	}
	activity.Sessions = []*mesgdef.Session{
		{
			Timestamp:        start,
			StartTime:        start,
			TotalElapsedTime: 3600000,
			TotalTimerTime:   3600000,
			TotalDistance:    5200000,
			Sport:            typedef.SportRunning,
			Event:            typedef.EventSession,
			EventType:        typedef.EventTypeStop,
		},
	}
	activity.Activity = &mesgdef.Activity{
		Timestamp:   start,
		NumSessions: 1,
		Type:        typedef.ActivityManual,
		Event:       typedef.EventActivity,
		EventType:   typedef.EventTypeStop,
	}

	fit := activity.ToFIT(nil)
	var buf bytes.Buffer
	if err := encoder.New(&buf).Encode(&fit); err != nil {
		t.Fatalf("Encode() error = %v", err)
	}

	parsed, err := Parse(buf.Bytes(), "activity.fit")
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}
	if parsed.Device == nil {
		t.Fatal("expected device from FIT")
	}
	if *parsed.Device != "Garmin Edge 530" {
		t.Fatalf("device = %q, want %q", *parsed.Device, "Garmin Edge 530")
	}
}
