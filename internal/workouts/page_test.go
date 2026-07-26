package workouts_test

import (
	"testing"
	"time"

	"github.com/solargate/grom/internal/workouts"
)

func TestCursorRoundTrip(t *testing.T) {
	start := time.Date(2026, 7, 8, 10, 30, 0, 123, time.UTC)
	c := workouts.CursorFromWorkout(workouts.Workout{ID: "38472901", StartDate: start})
	encoded := c.Encode()
	if encoded == "" {
		t.Fatal("expected non-empty cursor")
	}
	decoded, err := workouts.DecodeCursor(encoded)
	if err != nil {
		t.Fatal(err)
	}
	if decoded.ID != "38472901" {
		t.Fatalf("id = %q", decoded.ID)
	}
	if !decoded.StartDate.Equal(start.UTC()) {
		t.Fatalf("start = %v want %v", decoded.StartDate, start.UTC())
	}
}

func TestDecodeCursorRejectsGarbage(t *testing.T) {
	if _, err := workouts.DecodeCursor("not-a-cursor"); err == nil {
		t.Fatal("expected error")
	}
}

func TestAfterCursor(t *testing.T) {
	newer := time.Date(2026, 7, 8, 12, 0, 0, 0, time.UTC)
	older := time.Date(2026, 7, 8, 10, 0, 0, 0, time.UTC)
	cursor := &workouts.Cursor{StartDate: newer, ID: "20000000"}

	if workouts.AfterCursor(newer, "20000000", cursor) {
		t.Fatal("cursor item itself must not be after cursor")
	}
	if workouts.AfterCursor(newer, "30000000", cursor) {
		t.Fatal("newer id at same time must not be after cursor")
	}
	if !workouts.AfterCursor(newer, "10000000", cursor) {
		t.Fatal("older id at same time should be after cursor")
	}
	if !workouts.AfterCursor(older, "99999999", cursor) {
		t.Fatal("older timestamp should be after cursor")
	}
	if !workouts.AfterCursor(newer, "20000000", nil) {
		t.Fatal("nil cursor should accept all")
	}
}

func TestClampLimit(t *testing.T) {
	if got := workouts.ClampLimit(0); got != workouts.DefaultPageLimit {
		t.Fatalf("default: got %d", got)
	}
	if got := workouts.ClampLimit(5); got != 5 {
		t.Fatalf("passthrough: got %d", got)
	}
	if got := workouts.ClampLimit(1000); got != workouts.MaxPageLimit {
		t.Fatalf("max: got %d", got)
	}
}
