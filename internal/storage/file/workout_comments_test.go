package file

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/solargate/grom/internal/workouts"
)

func TestWorkoutCommentsStoreLocalRoundTrip(t *testing.T) {
	dir := t.TempDir()
	workoutsStore := NewWorkoutsStore(dir)
	commentsStore := NewWorkoutCommentsStore(dir)

	created, err := workoutsStore.Create("alice", &workouts.Workout{
		Name: "Morning", SportType: "Run",
		StartDate: time.Date(2026, 7, 8, 10, 0, 0, 0, time.UTC),
	})
	if err != nil {
		t.Fatal(err)
	}

	empty, err := commentsStore.GetLocal("alice", created.ID)
	if err != nil {
		t.Fatal(err)
	}
	if empty.CommentsNum != 0 || len(empty.Comments) != 0 {
		t.Fatalf("expected empty comments: %#v", empty)
	}

	payload := &workouts.WorkoutComments{Comments: []workouts.WorkoutComment{
		{
			ID: "c1",
			User: workouts.WorkoutLikeUser{
				Handle: "bob@localhost", Nickname: "bob", Name: "Bob", IsLocal: true,
			},
			Datetime: time.Date(2026, 8, 6, 12, 0, 0, 0, time.UTC),
			Text:     "Nice!",
			NoteID:   "https://localhost/users/bob/notes/c1",
		},
	}}
	if err := commentsStore.PutLocal("alice", created.ID, payload); err != nil {
		t.Fatal(err)
	}

	got, err := commentsStore.GetLocal("alice", created.ID)
	if err != nil {
		t.Fatal(err)
	}
	if got.CommentsNum != 1 || got.Comments[0].Text != "Nice!" {
		t.Fatalf("unexpected: %#v", got)
	}

	workoutDir, err := workoutsStore.findWorkoutDir("alice", created.ID)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(workoutDir, commentsFileName)); err != nil {
		t.Fatalf("expected comments.yaml: %v", err)
	}

	if err := commentsStore.DeleteLocal("alice", created.ID); err != nil {
		t.Fatal(err)
	}
	after, err := commentsStore.GetLocal("alice", created.ID)
	if err != nil {
		t.Fatal(err)
	}
	if after.CommentsNum != 0 {
		t.Fatalf("after delete CommentsNum = %d", after.CommentsNum)
	}
}

func TestWorkoutCommentsStoreLocalMissingWorkout(t *testing.T) {
	commentsStore := NewWorkoutCommentsStore(t.TempDir())
	err := commentsStore.PutLocal("alice", "missing1", &workouts.WorkoutComments{})
	if !errors.Is(err, workouts.ErrWorkoutNotFound) {
		t.Fatalf("PutLocal missing: %v", err)
	}
}

func TestWorkoutCommentsStoreFederatedAndActivity(t *testing.T) {
	dir := t.TempDir()
	store := NewWorkoutCommentsStore(dir)
	comments := &workouts.WorkoutComments{Comments: []workouts.WorkoutComment{
		{
			ID: "c1",
			User: workouts.WorkoutLikeUser{
				Handle: "alice@localhost", Nickname: "alice", Name: "Alice", IsLocal: true,
			},
			Datetime: time.Now().UTC(),
			Text:     "hi",
			NoteID:   "https://localhost/users/alice/notes/c1",
		},
	}}
	if err := store.PutFederated("alice", "bob@remote.test", "wid1", comments); err != nil {
		t.Fatal(err)
	}
	got, err := store.GetFederated("alice", "bob@remote.test", "wid1")
	if err != nil || got.CommentsNum != 1 {
		t.Fatalf("fed get: %#v %v", got, err)
	}
	noteID := "https://localhost/users/alice/notes/c1"
	if err := store.PutCommentActivityID("alice", noteID, "https://localhost/users/alice/activities/1"); err != nil {
		t.Fatal(err)
	}
	aid, err := store.GetCommentActivityID("alice", noteID)
	if err != nil || aid != "https://localhost/users/alice/activities/1" {
		t.Fatalf("activity: %q %v", aid, err)
	}
}
