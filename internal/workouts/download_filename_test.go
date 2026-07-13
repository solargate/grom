package workouts_test

import (
	"strings"
	"testing"

	"github.com/solargate/grom/internal/workouts"
)

func TestSanitizeDownloadBasename(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"", "workout"},
		{"Morning run", "Morning run"},
		{"Утренняя пробежка", "Утренняя пробежка"},
		{"bad/name", "bad_name"},
		{"bad\\name", "bad_name"},
		{"   ", "workout"},
	}

	for _, tc := range tests {
		if got := workouts.SanitizeDownloadBasename(tc.input); got != tc.want {
			t.Fatalf("SanitizeDownloadBasename(%q) = %q, want %q", tc.input, got, tc.want)
		}
	}
}

func TestTrackDownloadFilename(t *testing.T) {
	if got := workouts.TrackDownloadFilename("Morning run", ".gpx"); got != "Morning run.gpx" {
		t.Fatalf("got %q", got)
	}
	if got := workouts.TrackDownloadFilename("Утренняя пробежка", "fit"); got != "Утренняя пробежка.fit" {
		t.Fatalf("got %q", got)
	}
}

func TestContentDispositionAttachment(t *testing.T) {
	header := workouts.ContentDispositionAttachment("Утренняя.gpx")
	if header == "" {
		t.Fatal("expected non-empty header")
	}
	if !strings.Contains(header, `filename="`) || !strings.Contains(header, ".gpx") {
		t.Fatalf("unexpected ascii filename in header: %q", header)
	}
	if !strings.Contains(header, "filename*=UTF-8''") {
		t.Fatalf("unexpected utf-8 filename in header: %q", header)
	}
}
