package notifications

import (
	"context"
	"sync"
	"testing"

	"github.com/solargate/grom/internal/users"
	"github.com/solargate/grom/internal/workouts"
)

type recordingDeliverer struct {
	mu    sync.Mutex
	calls []deliveredCall
}

type deliveredCall struct {
	userID string
	event  Event
}

func (r *recordingDeliverer) SendToUser(_ context.Context, userID string, event Event) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.calls = append(r.calls, deliveredCall{userID: userID, event: event})
}

func (r *recordingDeliverer) snapshot() []deliveredCall {
	r.mu.Lock()
	defer r.mu.Unlock()
	out := make([]deliveredCall, len(r.calls))
	copy(out, r.calls)
	return out
}

type stubUsers struct {
	byNick map[string]*users.User
}

func (s *stubUsers) FindByNickname(nickname string) (*users.User, error) {
	if u, ok := s.byNick[nickname]; ok {
		return u, nil
	}
	return nil, users.ErrUserNotFound
}
func (s *stubUsers) FindByEmail(string) (*users.User, error) { return nil, users.ErrUserNotFound }
func (s *stubUsers) FindByID(string) (*users.User, error)    { return nil, users.ErrUserNotFound }
func (s *stubUsers) Search(string, string, int) ([]users.User, error) {
	return nil, nil
}
func (s *stubUsers) ListAll() ([]users.User, error) { return nil, nil }
func (s *stubUsers) Create(string, string, string, string) (*users.User, error) {
	return nil, users.ErrUserNotFound
}
func (s *stubUsers) UpdateProfile(string, string) (*users.User, error) {
	return nil, users.ErrUserNotFound
}
func (s *stubUsers) UpdatePassword(string, string) error { return users.ErrUserNotFound }
func (s *stubUsers) SetLastEquipmentForSport(string, string, []string) error {
	return nil
}
func (s *stubUsers) RemoveEquipmentFromLastSets(string, string) error { return nil }
func (s *stubUsers) GetProfile(string) (*users.Profile, error) {
	return &users.Profile{}, nil
}
func (s *stubUsers) PutProfile(string, users.Profile) error               { return nil }
func (s *stubUsers) SetLastSportType(string, string) error                { return nil }
func (s *stubUsers) TouchUsedSportType(string, string) error              { return nil }
func (s *stubUsers) PruneUsedSportTypes(string, map[string]struct{}) error { return nil }
func (s *stubUsers) Delete(string) error                                  { return users.ErrUserNotFound }

type stubWorkouts struct {
	byKey map[string]*workouts.Workout
}

func (s *stubWorkouts) Get(nickname, workoutID string) (*workouts.Workout, error) {
	if w, ok := s.byKey[nickname+"/"+workoutID]; ok {
		return w, nil
	}
	return nil, workouts.ErrWorkoutNotFound
}

func setupNotifierFixture(t *testing.T) (*Notifier, *recordingDeliverer, string, string, string) {
	t.Helper()
	ownerID := "user-bob"
	ownerNick := "bob"
	workoutID := "w1"
	usersStore := &stubUsers{byNick: map[string]*users.User{
		"bob": {ID: ownerID, Nickname: "bob", Name: "Bob"},
	}}
	workoutStore := &stubWorkouts{byKey: map[string]*workouts.Workout{
		"bob/w1": {ID: workoutID, Name: "Morning run"},
	}}
	rec := &recordingDeliverer{}
	n := NewNotifier(usersStore, workoutStore, rec)
	return n, rec, ownerID, ownerNick, workoutID
}

func TestNotifyWorkoutLikedHappyPath(t *testing.T) {
	n, rec, ownerID, ownerNick, workoutID := setupNotifierFixture(t)
	n.NotifyWorkoutLiked(ownerNick, workoutID, workouts.WorkoutLikeUser{
		Handle: "alice@localhost", Nickname: "alice", Name: "Alice", IsLocal: true,
	})
	calls := rec.snapshot()
	if len(calls) != 1 {
		t.Fatalf("calls = %#v", calls)
	}
	if calls[0].userID != ownerID {
		t.Fatalf("userID = %q, want %q", calls[0].userID, ownerID)
	}
	ev := calls[0].event
	if ev.Type != TypeWorkoutLiked {
		t.Fatalf("type = %q", ev.Type)
	}
	if ev.ActorDisplayName != "Alice" || ev.ActorHandle != "alice@localhost" {
		t.Fatalf("actor: %#v", ev)
	}
	if ev.WorkoutID != workoutID || ev.WorkoutTitle != "Morning run" || ev.Owner != ownerNick {
		t.Fatalf("workout fields: %#v", ev)
	}
	if ev.Slot != SlotLiked(ownerNick, workoutID) || ev.EventID == "" {
		t.Fatalf("slot/event: %#v", ev)
	}
}

func TestNotifyWorkoutCommentedHappyPath(t *testing.T) {
	n, rec, ownerID, ownerNick, workoutID := setupNotifierFixture(t)
	n.NotifyWorkoutCommented(ownerNick, workoutID, workouts.WorkoutLikeUser{
		Handle: "alice@localhost", Nickname: "alice", Name: "Alice", IsLocal: true,
	})
	calls := rec.snapshot()
	if len(calls) != 1 || calls[0].userID != ownerID {
		t.Fatalf("calls = %#v", calls)
	}
	ev := calls[0].event
	if ev.Type != TypeWorkoutCommented || ev.Slot != SlotCommented(ownerNick, workoutID) {
		t.Fatalf("event: %#v", ev)
	}
}

func TestNotifyWorkoutSkipsSelfLocalActor(t *testing.T) {
	n, rec, _, ownerNick, workoutID := setupNotifierFixture(t)
	n.NotifyWorkoutLiked(ownerNick, workoutID, workouts.WorkoutLikeUser{
		Handle: "bob@localhost", Nickname: "bob", Name: "Bob", IsLocal: true,
	})
	n.NotifyWorkoutCommented(ownerNick, workoutID, workouts.WorkoutLikeUser{
		Handle: "BOB@localhost", Nickname: "Bob", Name: "Bob", IsLocal: true,
	})
	if calls := rec.snapshot(); len(calls) != 0 {
		t.Fatalf("expected no notify for self: %#v", calls)
	}
}

func TestNotifyWorkoutSkipsEmptyActorAndMissingWorkout(t *testing.T) {
	n, rec, _, ownerNick, workoutID := setupNotifierFixture(t)
	n.NotifyWorkoutLiked(ownerNick, workoutID, workouts.WorkoutLikeUser{})
	n.NotifyWorkoutLiked(ownerNick, "missing", workouts.WorkoutLikeUser{
		Handle: "alice@localhost", Nickname: "alice", IsLocal: true,
	})
	n.NotifyWorkoutLiked("", workoutID, workouts.WorkoutLikeUser{
		Handle: "alice@localhost", Nickname: "alice", IsLocal: true,
	})
	if calls := rec.snapshot(); len(calls) != 0 {
		t.Fatalf("expected skips: %#v", calls)
	}
}

func TestNotifyWorkoutTitleFallbackToID(t *testing.T) {
	usersStore := &stubUsers{byNick: map[string]*users.User{
		"bob": {ID: "user-bob", Nickname: "bob", Name: "Bob"},
	}}
	workoutStore := &stubWorkouts{byKey: map[string]*workouts.Workout{
		"bob/w1": {ID: "w1", Name: "   "},
	}}
	rec := &recordingDeliverer{}
	n := NewNotifier(usersStore, workoutStore, rec)
	n.NotifyWorkoutLiked("bob", "w1", workouts.WorkoutLikeUser{
		Handle: "alice@localhost", Nickname: "alice", Name: "Alice", IsLocal: true,
	})
	calls := rec.snapshot()
	if len(calls) != 1 || calls[0].event.WorkoutTitle != "w1" {
		t.Fatalf("title fallback: %#v", calls)
	}
}

func TestNotifyNewFollowerHappyPath(t *testing.T) {
	n, rec, ownerID, ownerNick, _ := setupNotifierFixture(t)
	n.NotifyNewFollower(ownerNick, workouts.WorkoutLikeUser{
		Handle: "alice@localhost", Nickname: "alice", Name: "Alice", IsLocal: true,
	})
	calls := rec.snapshot()
	if len(calls) != 1 || calls[0].userID != ownerID {
		t.Fatalf("calls = %#v", calls)
	}
	ev := calls[0].event
	if ev.Type != TypeUserFollowed || ev.ActorDisplayName != "Alice" || ev.ActorHandle != "alice@localhost" {
		t.Fatalf("event: %#v", ev)
	}
	if ev.Slot != SlotFollowed(ev.EventID) || ev.EventID == "" {
		t.Fatalf("slot: %#v", ev)
	}
	if ev.WorkoutID != "" || ev.Owner != "" {
		t.Fatalf("follower event should omit workout fields: %#v", ev)
	}
}

func TestNotifyNewFollowerFillsLocalHandle(t *testing.T) {
	n, rec, _, ownerNick, _ := setupNotifierFixture(t)
	n.NotifyNewFollower(ownerNick, workouts.WorkoutLikeUser{
		Nickname: "alice", Name: "Alice", IsLocal: true,
	})
	calls := rec.snapshot()
	if len(calls) != 1 || calls[0].event.ActorHandle != "alice" {
		t.Fatalf("handle fill: %#v", calls)
	}
}

func TestNotifyNewFollowerSkipsSelfAndMissing(t *testing.T) {
	n, rec, _, ownerNick, _ := setupNotifierFixture(t)
	n.NotifyNewFollower(ownerNick, workouts.WorkoutLikeUser{
		Handle: "bob@localhost", Nickname: "bob", Name: "Bob", IsLocal: true,
	})
	n.NotifyNewFollower("missing", workouts.WorkoutLikeUser{
		Handle: "alice@localhost", Nickname: "alice", Name: "Alice", IsLocal: true,
	})
	n.NotifyNewFollower(ownerNick, workouts.WorkoutLikeUser{})
	if calls := rec.snapshot(); len(calls) != 0 {
		t.Fatalf("expected skips: %#v", calls)
	}
}
