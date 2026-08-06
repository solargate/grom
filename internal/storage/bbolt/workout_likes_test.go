package bbolt_test

import (
	"errors"
	"testing"
	"time"

	"github.com/solargate/grom/internal/workouts"
)

func TestWorkoutLikesStoreLocalRoundTrip(t *testing.T) {
	b := openTestBackend(t)
	created, err := b.Workouts().Create("alice", &workouts.Workout{
		Name: "Morning", SportType: "Run",
		StartDate: time.Date(2026, 7, 8, 10, 0, 0, 0, time.UTC),
	})
	if err != nil {
		t.Fatal(err)
	}
	likes := b.Likes()

	empty, err := likes.GetLocal("alice", created.ID)
	if err != nil {
		t.Fatal(err)
	}
	if empty.Likes != 0 || len(empty.Users) != 0 {
		t.Fatalf("expected empty likes: %#v", empty)
	}

	payload := &workouts.WorkoutLikes{Users: []workouts.WorkoutLikeUser{
		{Handle: "bob@localhost", Nickname: "bob", Name: "Bob", IsLocal: true},
		{Handle: "carol@remote.test", Nickname: "carol", Name: "Carol", IsLocal: false},
	}}
	if err := likes.PutLocal("alice", created.ID, payload); err != nil {
		t.Fatal(err)
	}
	got, err := likes.GetLocal("alice", created.ID)
	if err != nil {
		t.Fatal(err)
	}
	if got.Likes != 2 || len(got.Users) != 2 {
		t.Fatalf("unexpected likes: %#v", got)
	}
	if got.Users[0].Handle != "bob@localhost" {
		t.Fatalf("expected sorted handles, got %#v", got.Users)
	}

	if err := likes.DeleteLocal("alice", created.ID); err != nil {
		t.Fatal(err)
	}
	after, err := likes.GetLocal("alice", created.ID)
	if err != nil || after.Likes != 0 {
		t.Fatalf("after delete: %#v err=%v", after, err)
	}
	if err := likes.DeleteLocal("alice", created.ID); err != nil {
		t.Fatalf("idempotent delete: %v", err)
	}
}

func TestWorkoutLikesStoreLocalMissingWorkout(t *testing.T) {
	b := openTestBackend(t)
	likes := b.Likes()
	err := likes.PutLocal("alice", "missing1", &workouts.WorkoutLikes{})
	if !errors.Is(err, workouts.ErrWorkoutNotFound) {
		t.Fatalf("PutLocal missing: %v", err)
	}
	_, err = likes.GetLocal("alice", "missing1")
	if !errors.Is(err, workouts.ErrWorkoutNotFound) {
		t.Fatalf("GetLocal missing: %v", err)
	}
	if err := likes.DeleteLocal("alice", "missing1"); err != nil {
		t.Fatalf("DeleteLocal missing should be nil: %v", err)
	}
}

func TestWorkoutLikesStoreFederatedAndActivityID(t *testing.T) {
	b := openTestBackend(t)
	likes := b.Likes()
	ownerHandle := "bob@remote.test"

	payload := &workouts.WorkoutLikes{Users: []workouts.WorkoutLikeUser{
		{Handle: "alice@localhost", Nickname: "alice", Name: "Alice", IsLocal: true},
	}}
	if err := likes.PutFederated("alice", ownerHandle, "38472901", payload); err != nil {
		t.Fatal(err)
	}
	got, err := likes.GetFederated("alice", ownerHandle, "38472901")
	if err != nil {
		t.Fatal(err)
	}
	if got.Likes != 1 || got.Users[0].Handle != "alice@localhost" {
		t.Fatalf("federated get: %#v", got)
	}

	other, err := likes.GetFederated("carol", ownerHandle, "38472901")
	if err != nil || other.Likes != 0 {
		t.Fatalf("viewer isolation: %#v err=%v", other, err)
	}

	if err := likes.DeleteFederated("alice", ownerHandle, "38472901"); err != nil {
		t.Fatal(err)
	}
	after, err := likes.GetFederated("alice", ownerHandle, "38472901")
	if err != nil || after.Likes != 0 {
		t.Fatalf("after federated delete: %#v err=%v", after, err)
	}

	objectID := "https://remote.test/users/bob/workouts/38472901"
	id, err := likes.GetLikeActivityID("alice", objectID)
	if err != nil || id != "" {
		t.Fatalf("empty activity id: %q err=%v", id, err)
	}
	if err := likes.PutLikeActivityID("alice", objectID, "https://localhost/users/alice/activities/1"); err != nil {
		t.Fatal(err)
	}
	id, err = likes.GetLikeActivityID("alice", objectID)
	if err != nil || id != "https://localhost/users/alice/activities/1" {
		t.Fatalf("activity id = %q err=%v", id, err)
	}
	if err := likes.DeleteLikeActivityID("alice", objectID); err != nil {
		t.Fatal(err)
	}
	id, err = likes.GetLikeActivityID("alice", objectID)
	if err != nil || id != "" {
		t.Fatalf("after delete activity id = %q err=%v", id, err)
	}
}
