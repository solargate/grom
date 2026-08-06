package workouts

import (
	"testing"
	"time"
)

func TestNormalizeWorkoutComments(t *testing.T) {
	t.Run("nil", func(t *testing.T) {
		got := NormalizeWorkoutComments(nil)
		if got.CommentsNum != 0 || got.Comments == nil {
			t.Fatalf("got %#v", got)
		}
	})

	t.Run("sorts and drops invalid", func(t *testing.T) {
		t1 := time.Date(2026, 8, 6, 12, 0, 0, 0, time.UTC)
		t2 := time.Date(2026, 8, 6, 13, 0, 0, 0, time.UTC)
		got := NormalizeWorkoutComments(&WorkoutComments{
			CommentsNum: 99,
			Comments: []WorkoutComment{
				{ID: "b", User: WorkoutLikeUser{Handle: "bob@a"}, Datetime: t2, Text: " later "},
				{ID: "", User: WorkoutLikeUser{Handle: "x"}, Text: "no id"},
				{ID: "a", User: WorkoutLikeUser{Handle: "alice@a"}, Datetime: t1, Text: "first"},
				{ID: "a", User: WorkoutLikeUser{Handle: "alice@a"}, Datetime: t1, Text: "dup"},
			},
		})
		if got.CommentsNum != 2 {
			t.Fatalf("CommentsNum = %d", got.CommentsNum)
		}
		if got.Comments[0].ID != "a" || got.Comments[1].ID != "b" {
			t.Fatalf("order = %#v", got.Comments)
		}
		if got.Comments[1].Text != "later" {
			t.Fatalf("trimmed text = %q", got.Comments[1].Text)
		}
	})
}

func TestValidateCommentText(t *testing.T) {
	if err := ValidateCommentText("  "); err != ErrEmptyComment {
		t.Fatalf("empty: %v", err)
	}
	long := make([]rune, MaxCommentTextLength+1)
	for i := range long {
		long[i] = 'a'
	}
	if err := ValidateCommentText(string(long)); err != ErrCommentTooLong {
		t.Fatalf("too long: %v", err)
	}
	if err := ValidateCommentText("ok"); err != nil {
		t.Fatal(err)
	}
}

func TestAddRemoveComment(t *testing.T) {
	c := WorkoutComment{
		ID: "1", User: WorkoutLikeUser{Handle: "a@x"}, Datetime: time.Now().UTC(), Text: "hi",
	}
	got := AddWorkoutComment(nil, c)
	if got.CommentsNum != 1 {
		t.Fatalf("add: %#v", got)
	}
	again := AddWorkoutComment(&got, c)
	if again.CommentsNum != 1 {
		t.Fatalf("dedupe: %#v", again)
	}
	removed := RemoveWorkoutCommentByID(&again, "1")
	if removed.CommentsNum != 0 {
		t.Fatalf("remove: %#v", removed)
	}
}

func TestCanDeleteComment(t *testing.T) {
	c := &WorkoutComment{User: WorkoutLikeUser{Handle: "bob@grom.test"}}
	if !CanDeleteComment("bob@grom.test", "alice", c, "alice@grom.test") {
		t.Fatal("author should delete")
	}
	if !CanDeleteComment("alice@grom.test", "alice", c, "alice@grom.test") {
		t.Fatal("owner should delete")
	}
	if CanDeleteComment("carol@grom.test", "alice", c, "alice@grom.test") {
		t.Fatal("stranger should not delete")
	}
}
