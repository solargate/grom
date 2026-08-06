package workouts

import (
	"reflect"
	"testing"
)

func TestNormalizeWorkoutLikes(t *testing.T) {
	t.Run("nil", func(t *testing.T) {
		got := NormalizeWorkoutLikes(nil)
		if got.Likes != 0 {
			t.Fatalf("Likes = %d, want 0", got.Likes)
		}
		if got.Users == nil || len(got.Users) != 0 {
			t.Fatalf("Users = %#v, want empty non-nil slice", got.Users)
		}
	})

	t.Run("drops empty handles and dedupes", func(t *testing.T) {
		got := NormalizeWorkoutLikes(&WorkoutLikes{
			Likes: 99,
			Users: []WorkoutLikeUser{
				{Handle: "bob@a.test", Nickname: "bob", Name: "Bob"},
				{Handle: "", Nickname: "skip"},
				{Handle: "alice@a.test", Nickname: "alice", Name: "Alice"},
				{Handle: "bob@a.test", Nickname: "bob2", Name: "Bob Updated"},
			},
		})
		if got.Likes != 2 {
			t.Fatalf("Likes = %d, want 2", got.Likes)
		}
		if len(got.Users) != 2 {
			t.Fatalf("len(Users) = %d, want 2", len(got.Users))
		}
		if got.Users[0].Handle != "alice@a.test" || got.Users[1].Handle != "bob@a.test" {
			t.Fatalf("sorted handles = %q, %q", got.Users[0].Handle, got.Users[1].Handle)
		}
		if got.Users[1].Name != "Bob Updated" {
			t.Fatalf("duplicate kept name = %q", got.Users[1].Name)
		}
	})
}

func TestLikesContainUser(t *testing.T) {
	likes := &WorkoutLikes{Users: []WorkoutLikeUser{{Handle: "alice@a.test"}}}
	if LikesContainUser(nil, "alice@a.test") {
		t.Fatal("nil likes should be false")
	}
	if LikesContainUser(likes, "") {
		t.Fatal("empty handle should be false")
	}
	if !LikesContainUser(likes, "alice@a.test") {
		t.Fatal("expected contain")
	}
	if LikesContainUser(likes, "bob@a.test") {
		t.Fatal("unexpected contain")
	}
}

func TestAddWorkoutLikeUser(t *testing.T) {
	user := WorkoutLikeUser{Handle: "alice@a.test", Nickname: "alice", Name: "Alice", IsLocal: true}

	got := AddWorkoutLikeUser(nil, user)
	if got.Likes != 1 || !LikesContainUser(&got, user.Handle) {
		t.Fatalf("add to nil: %#v", got)
	}

	again := AddWorkoutLikeUser(&got, user)
	if again.Likes != 1 {
		t.Fatalf("idempotent likes = %d", again.Likes)
	}

	empty := AddWorkoutLikeUser(&got, WorkoutLikeUser{Nickname: "x"})
	if empty.Likes != 1 {
		t.Fatalf("empty handle should not add: %#v", empty)
	}
}

func TestRemoveWorkoutLikeUser(t *testing.T) {
	likes := &WorkoutLikes{Users: []WorkoutLikeUser{
		{Handle: "alice@a.test", Nickname: "alice"},
		{Handle: "bob@a.test", Nickname: "bob"},
	}}

	got := RemoveWorkoutLikeUser(likes, "alice@a.test")
	want := WorkoutLikes{
		Likes: 1,
		Users: []WorkoutLikeUser{{Handle: "bob@a.test", Nickname: "bob"}},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("remove = %#v, want %#v", got, want)
	}

	again := RemoveWorkoutLikeUser(&got, "alice@a.test")
	if again.Likes != 1 {
		t.Fatalf("idempotent remove likes = %d", again.Likes)
	}

	empty := RemoveWorkoutLikeUser(&got, "")
	if empty.Likes != 1 {
		t.Fatalf("empty handle remove = %#v", empty)
	}
}
