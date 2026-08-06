package bbolt

import (
	"errors"
	"testing"
	"time"

	"github.com/solargate/grom/internal/workouts"
)

func TestWorkoutCommentsStoreLocalRoundTrip(t *testing.T) {
	dir := t.TempDir()
	b, err := Open(dir+"/grom.db", dir)
	if err != nil {
		t.Fatal(err)
	}
	defer b.Close()

	created, err := b.Workouts().Create("alice", &workouts.Workout{
		Name: "Morning", SportType: "Run",
		StartDate: time.Date(2026, 7, 8, 10, 0, 0, 0, time.UTC),
	})
	if err != nil {
		t.Fatal(err)
	}

	comments := b.Comments()
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
	if err := comments.PutLocal("alice", created.ID, payload); err != nil {
		t.Fatal(err)
	}
	got, err := comments.GetLocal("alice", created.ID)
	if err != nil || got.CommentsNum != 1 {
		t.Fatalf("got %#v err=%v", got, err)
	}
	if err := comments.DeleteLocal("alice", created.ID); err != nil {
		t.Fatal(err)
	}
	after, err := comments.GetLocal("alice", created.ID)
	if err != nil || after.CommentsNum != 0 {
		t.Fatalf("after delete %#v", after)
	}
}

func TestWorkoutCommentsStoreLocalMissingWorkout(t *testing.T) {
	dir := t.TempDir()
	b, err := Open(dir+"/grom.db", dir)
	if err != nil {
		t.Fatal(err)
	}
	defer b.Close()

	err = b.Comments().PutLocal("alice", "missing1", &workouts.WorkoutComments{})
	if !errors.Is(err, workouts.ErrWorkoutNotFound) {
		t.Fatalf("PutLocal missing: %v", err)
	}
}

func TestWorkoutCommentsStoreFederatedAndActivity(t *testing.T) {
	dir := t.TempDir()
	b, err := Open(dir+"/grom.db", dir)
	if err != nil {
		t.Fatal(err)
	}
	defer b.Close()

	comments := b.Comments()
	payload := &workouts.WorkoutComments{Comments: []workouts.WorkoutComment{
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
	if err := comments.PutFederated("alice", "bob@remote.test", "wid1", payload); err != nil {
		t.Fatal(err)
	}
	got, err := comments.GetFederated("alice", "bob@remote.test", "wid1")
	if err != nil || got.CommentsNum != 1 {
		t.Fatalf("fed get: %#v %v", got, err)
	}
	noteID := "https://localhost/users/alice/notes/c1"
	if err := comments.PutCommentActivityID("alice", noteID, "https://localhost/users/alice/activities/1"); err != nil {
		t.Fatal(err)
	}
	aid, err := comments.GetCommentActivityID("alice", noteID)
	if err != nil || aid != "https://localhost/users/alice/activities/1" {
		t.Fatalf("activity: %q %v", aid, err)
	}
	if err := comments.DeleteCommentActivityID("alice", noteID); err != nil {
		t.Fatal(err)
	}
	aid, err = comments.GetCommentActivityID("alice", noteID)
	if err != nil || aid != "" {
		t.Fatalf("after delete activity: %q %v", aid, err)
	}
	if err := comments.DeleteFederated("alice", "bob@remote.test", "wid1"); err != nil {
		t.Fatal(err)
	}
	got, err = comments.GetFederated("alice", "bob@remote.test", "wid1")
	if err != nil || got.CommentsNum != 0 {
		t.Fatalf("after fed delete: %#v %v", got, err)
	}
}
