package v1_test

import (
	"context"
	"net/http"
	"sync"
	"testing"

	"github.com/solargate/grom/internal/notifications"
)

type recordingDeliverer struct {
	mu    sync.Mutex
	calls []struct {
		userID string
		event  notifications.Event
	}
}

func (r *recordingDeliverer) SendToUser(_ context.Context, userID string, event notifications.Event) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.calls = append(r.calls, struct {
		userID string
		event  notifications.Event
	}{userID: userID, event: event})
}

func (r *recordingDeliverer) events() []notifications.Event {
	r.mu.Lock()
	defer r.mu.Unlock()
	out := make([]notifications.Event, len(r.calls))
	for i := range r.calls {
		out[i] = r.calls[i].event
	}
	return out
}

func (r *recordingDeliverer) snapshot() []struct {
	userID string
	event  notifications.Event
} {
	r.mu.Lock()
	defer r.mu.Unlock()
	out := make([]struct {
		userID string
		event  notifications.Event
	}, len(r.calls))
	copy(out, r.calls)
	return out
}

func installRecordingNotifier(t *testing.T, ta *testApp) *recordingDeliverer {
	t.Helper()
	rec := &recordingDeliverer{}
	ta.app.Notifier = notifications.NewNotifier(ta.app.Users, ta.app.Workouts, rec)
	return rec
}

func TestLocalLikeAndCommentNotifyOnlyOnNew(t *testing.T) {
	ta := setupTestApp(t)
	rec := installRecordingNotifier(t, ta)

	ta.register(t, "alice", "alice@example.com", "password12")
	ta.register(t, "bob", "bob@example.com", "password12")
	aliceToken, _ := ta.login(t, "alice@example.com", "password12")
	bobToken, bobUser := ta.login(t, "bob@example.com", "password12")
	bobID, _ := bobUser["id"].(string)

	w := ta.doJSON(t, http.MethodPost, "/api/v1/social/follow", map[string]string{"handle": "bob"}, aliceToken)
	expectStatus(t, w, http.StatusCreated)

	w = ta.doJSON(t, http.MethodPost, "/api/v1/workouts", map[string]any{
		"name":             "Bob run",
		"sport_type":       "Run",
		"start_date":       "2026-07-08T10:00:00Z",
		"duration_seconds": 1800,
		"distance":         5000,
	}, bobToken)
	expectStatus(t, w, http.StatusCreated)
	workoutID, _ := decodeObject(t, w)["id"].(string)

	// Local follow already notified once.
	followEvents := 0
	for _, ev := range rec.events() {
		if ev.Type == notifications.TypeUserFollowed {
			followEvents++
		}
	}
	if followEvents != 1 {
		t.Fatalf("follow notify count = %d events=%#v", followEvents, rec.events())
	}

	likePath := "/api/v1/workouts/" + workoutID + "/likes?owner=bob"
	w = ta.doJSON(t, http.MethodPost, likePath, nil, aliceToken)
	expectStatus(t, w, http.StatusOK)
	w = ta.doJSON(t, http.MethodPost, likePath, nil, aliceToken)
	expectStatus(t, w, http.StatusOK)

	commentPath := "/api/v1/workouts/" + workoutID + "/comments?owner=bob"
	w = ta.doJSON(t, http.MethodPost, commentPath, map[string]string{"text": "Nice!"}, aliceToken)
	expectStatus(t, w, http.StatusOK)

	// Owner comments on own workout — must not notify.
	w = ta.doJSON(t, http.MethodPost, "/api/v1/workouts/"+workoutID+"/comments", map[string]string{
		"text": "Thanks!",
	}, bobToken)
	expectStatus(t, w, http.StatusOK)

	liked := 0
	commented := 0
	for _, call := range rec.snapshot() {
		if call.userID != bobID {
			continue
		}
		switch call.event.Type {
		case notifications.TypeWorkoutLiked:
			liked++
			if call.event.WorkoutID != workoutID || call.event.ActorDisplayName != "alice" {
				t.Fatalf("like event: %#v", call.event)
			}
		case notifications.TypeWorkoutCommented:
			commented++
			if call.event.ActorDisplayName != "alice" {
				t.Fatalf("comment event: %#v", call.event)
			}
		}
	}
	if liked != 1 {
		t.Fatalf("liked notify = %d, want 1 (idempotent like)", liked)
	}
	if commented != 1 {
		t.Fatalf("commented notify = %d, want 1 (owner self-comment skipped)", commented)
	}
}

func TestLocalFollowNotifyIdempotent(t *testing.T) {
	ta := setupTestApp(t)
	rec := installRecordingNotifier(t, ta)

	ta.register(t, "alice", "alice@example.com", "password12")
	ta.register(t, "bob", "bob@example.com", "password12")
	aliceToken, _ := ta.login(t, "alice@example.com", "password12")

	w := ta.doJSON(t, http.MethodPost, "/api/v1/social/follow", map[string]string{"handle": "bob"}, aliceToken)
	expectStatus(t, w, http.StatusCreated)
	w = ta.doJSON(t, http.MethodPost, "/api/v1/social/follow", map[string]string{"handle": "bob"}, aliceToken)
	expectStatus(t, w, http.StatusCreated)

	followed := 0
	for _, ev := range rec.events() {
		if ev.Type == notifications.TypeUserFollowed {
			followed++
		}
	}
	if followed != 1 {
		t.Fatalf("follow notify = %d, want 1", followed)
	}
}
