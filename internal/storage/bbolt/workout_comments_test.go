package bbolt

import (
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
