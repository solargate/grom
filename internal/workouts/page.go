package workouts

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"time"
)

const (
	DefaultPageLimit = 20
	MaxPageLimit     = 100
)

// Cursor identifies a position in a start_date DESC, id DESC feed.
type Cursor struct {
	StartDate time.Time `json:"t"`
	ID        string    `json:"id"`
}

// Page is a cursor-paginated list of feed workouts.
type Page struct {
	Items      []FeedWorkout
	NextCursor string
	HasMore    bool
}

// ClampLimit returns a safe page size (default 20, max 100).
func ClampLimit(limit int) int {
	if limit <= 0 {
		return DefaultPageLimit
	}
	if limit > MaxPageLimit {
		return MaxPageLimit
	}
	return limit
}

// CursorFromWorkout builds a cursor from a workout's sort key.
func CursorFromWorkout(w Workout) Cursor {
	return Cursor{StartDate: w.StartDate.UTC(), ID: w.ID}
}

// Encode serializes the cursor as a URL-safe opaque token.
func (c Cursor) Encode() string {
	payload, err := json.Marshal(cursorJSON{
		T:  c.StartDate.UTC().Format(time.RFC3339Nano),
		ID: c.ID,
	})
	if err != nil {
		return ""
	}
	return base64.RawURLEncoding.EncodeToString(payload)
}

// DecodeCursor parses an opaque cursor token.
func DecodeCursor(raw string) (*Cursor, error) {
	if raw == "" {
		return nil, nil
	}
	data, err := base64.RawURLEncoding.DecodeString(raw)
	if err != nil {
		return nil, fmt.Errorf("invalid cursor encoding")
	}
	var payload cursorJSON
	if err := json.Unmarshal(data, &payload); err != nil {
		return nil, fmt.Errorf("invalid cursor payload")
	}
	if payload.ID == "" || payload.T == "" {
		return nil, fmt.Errorf("invalid cursor fields")
	}
	t, err := time.Parse(time.RFC3339Nano, payload.T)
	if err != nil {
		t, err = time.Parse(time.RFC3339, payload.T)
		if err != nil {
			return nil, fmt.Errorf("invalid cursor time")
		}
	}
	return &Cursor{StartDate: t.UTC(), ID: payload.ID}, nil
}

type cursorJSON struct {
	T  string `json:"t"`
	ID string `json:"id"`
}

// FeedNewer reports whether a should appear before b in the feed
// (start_date DESC, id DESC).
func FeedNewer(aStart time.Time, aID string, bStart time.Time, bID string) bool {
	aStart = aStart.UTC()
	bStart = bStart.UTC()
	if !aStart.Equal(bStart) {
		return aStart.After(bStart)
	}
	return aID > bID
}

// AfterCursor reports whether the workout is strictly older than the cursor
// in feed order (i.e. should appear on a later page).
func AfterCursor(start time.Time, id string, cursor *Cursor) bool {
	if cursor == nil {
		return true
	}
	return FeedNewer(cursor.StartDate, cursor.ID, start, id)
}
