package file

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/solargate/grom/internal/workouts"
)

func TestWorkoutLikesStoreLocalRoundTrip(t *testing.T) {
	dir := t.TempDir()
	workoutsStore := NewWorkoutsStore(dir)
	likesStore := NewWorkoutLikesStore(dir)

	created, err := workoutsStore.Create("alice", &workouts.Workout{
		Name: "Morning", SportType: "Run",
		StartDate: time.Date(2026, 7, 8, 10, 0, 0, 0, time.UTC),
	})
	if err != nil {
		t.Fatal(err)
	}

	empty, err := likesStore.GetLocal("alice", created.ID)
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
	if err := likesStore.PutLocal("alice", created.ID, payload); err != nil {
		t.Fatal(err)
	}

	got, err := likesStore.GetLocal("alice", created.ID)
	if err != nil {
		t.Fatal(err)
	}
	if got.Likes != 2 {
		t.Fatalf("Likes = %d, want 2", got.Likes)
	}
	if len(got.Users) != 2 || got.Users[0].Handle != "bob@localhost" {
		t.Fatalf("unexpected users: %#v", got.Users)
	}

	workoutDir, err := workoutsStore.findWorkoutDir("alice", created.ID)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(workoutDir, likesFileName)); err != nil {
		t.Fatalf("expected likes.yaml: %v", err)
	}

	if err := likesStore.DeleteLocal("alice", created.ID); err != nil {
		t.Fatal(err)
	}
	after, err := likesStore.GetLocal("alice", created.ID)
	if err != nil {
		t.Fatal(err)
	}
	if after.Likes != 0 {
		t.Fatalf("after delete Likes = %d", after.Likes)
	}
	if err := likesStore.DeleteLocal("alice", created.ID); err != nil {
		t.Fatalf("idempotent delete: %v", err)
	}
}

func TestWorkoutLikesStoreLocalMissingWorkout(t *testing.T) {
	likesStore := NewWorkoutLikesStore(t.TempDir())
	err := likesStore.PutLocal("alice", "missing1", &workouts.WorkoutLikes{})
	if !errors.Is(err, workouts.ErrWorkoutNotFound) {
		t.Fatalf("PutLocal missing: %v", err)
	}
	_, err = likesStore.GetLocal("alice", "missing1")
	if !errors.Is(err, workouts.ErrWorkoutNotFound) {
		t.Fatalf("GetLocal missing: %v", err)
	}
	if err := likesStore.DeleteLocal("alice", "missing1"); err != nil {
		t.Fatalf("DeleteLocal missing should be nil: %v", err)
	}
}

func TestWorkoutLikesStoreFederatedAndActivityID(t *testing.T) {
	dir := t.TempDir()
	likesStore := NewWorkoutLikesStore(dir)

	ownerHandle := "bob@remote.test"
	payload := &workouts.WorkoutLikes{Users: []workouts.WorkoutLikeUser{
		{Handle: "alice@localhost", Nickname: "alice", Name: "Alice", IsLocal: true},
	}}
	if err := likesStore.PutFederated("alice", ownerHandle, "38472901", payload); err != nil {
		t.Fatal(err)
	}
	got, err := likesStore.GetFederated("alice", ownerHandle, "38472901")
	if err != nil {
		t.Fatal(err)
	}
	if got.Likes != 1 || got.Users[0].Handle != "alice@localhost" {
		t.Fatalf("federated get: %#v", got)
	}

	other, err := likesStore.GetFederated("carol", ownerHandle, "38472901")
	if err != nil {
		t.Fatal(err)
	}
	if other.Likes != 0 {
		t.Fatalf("viewer isolation broken: %#v", other)
	}

	if err := likesStore.DeleteFederated("alice", ownerHandle, "38472901"); err != nil {
		t.Fatal(err)
	}
	after, err := likesStore.GetFederated("alice", ownerHandle, "38472901")
	if err != nil || after.Likes != 0 {
		t.Fatalf("after federated delete: %#v err=%v", after, err)
	}

	objectID := "https://remote.test/users/bob/workouts/38472901"
	id, err := likesStore.GetLikeActivityID("alice", objectID)
	if err != nil || id != "" {
		t.Fatalf("empty activity id: %q err=%v", id, err)
	}
	if err := likesStore.PutLikeActivityID("alice", objectID, "https://localhost/users/alice/activities/1"); err != nil {
		t.Fatal(err)
	}
	id, err = likesStore.GetLikeActivityID("alice", objectID)
	if err != nil || id != "https://localhost/users/alice/activities/1" {
		t.Fatalf("activity id = %q err=%v", id, err)
	}
	if err := likesStore.DeleteLikeActivityID("alice", objectID); err != nil {
		t.Fatal(err)
	}
	id, err = likesStore.GetLikeActivityID("alice", objectID)
	if err != nil || id != "" {
		t.Fatalf("after delete activity id = %q err=%v", id, err)
	}
	if err := likesStore.DeleteLikeActivityID("alice", objectID); err != nil {
		t.Fatalf("idempotent activity delete: %v", err)
	}
}
