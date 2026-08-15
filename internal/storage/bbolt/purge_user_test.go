package bbolt_test

import (
	"os"
	"testing"
	"time"

	"github.com/solargate/grom/internal/auth/pat"
	"github.com/solargate/grom/internal/data"
	"github.com/solargate/grom/internal/equipment"
	"github.com/solargate/grom/internal/social"
	"github.com/solargate/grom/internal/workouts"
)

func TestPurgeUserRemovesUserAndWorkouts(t *testing.T) {
	backend := openTestBackend(t)
	alice, err := backend.Users().Create("alice", "Alice", "alice@example.com", "password12")
	if err != nil {
		t.Fatal(err)
	}
	created, err := backend.Workouts().Create(alice.Nickname, &workouts.Workout{
		Name:            "Run",
		SportType:       "Run",
		StartDate:       time.Now().UTC(),
		DurationSeconds: 100,
	})
	if err != nil {
		t.Fatal(err)
	}

	if err := backend.PurgeUser(alice.ID, alice.Nickname, "alice@localhost"); err != nil {
		t.Fatalf("PurgeUser: %v", err)
	}
	if _, err := backend.Users().FindByID(alice.ID); err == nil {
		t.Fatal("expected user removed")
	}
	if _, err := backend.Workouts().Get(alice.Nickname, created.ID); err == nil {
		t.Fatal("expected workout removed")
	}
	if _, err := os.Stat(data.UserDir(backend.Location(), "alice")); !os.IsNotExist(err) {
		t.Fatalf("expected user dir removed, err=%v", err)
	}
}

func TestPurgeUserScrubsLikesCommentsFollowsEquipmentAndPAT(t *testing.T) {
	backend := openTestBackend(t)
	alice, err := backend.Users().Create("alice", "Alice", "alice@example.com", "password12")
	if err != nil {
		t.Fatal(err)
	}
	bob, err := backend.Users().Create("bob", "Bob", "bob@example.com", "password12")
	if err != nil {
		t.Fatal(err)
	}

	created, err := backend.Workouts().Create(alice.Nickname, &workouts.Workout{
		Name: "Run", SportType: "Run", StartDate: time.Now().UTC(), DurationSeconds: 100,
	})
	if err != nil {
		t.Fatal(err)
	}
	handle := "bob@localhost"
	likes := workouts.AddWorkoutLikeUser(nil, workouts.WorkoutLikeUser{
		Handle: handle, Nickname: "bob", IsLocal: true,
	})
	if err := backend.Likes().PutLocal(alice.Nickname, created.ID, &likes); err != nil {
		t.Fatal(err)
	}
	comments := workouts.AddWorkoutComment(nil, workouts.WorkoutComment{
		ID: "c1", User: workouts.WorkoutLikeUser{Handle: handle, Nickname: "bob", IsLocal: true}, Text: "nice",
	})
	comments = workouts.AddWorkoutComment(&comments, workouts.WorkoutComment{
		ID: "c2", User: workouts.WorkoutLikeUser{Handle: "alice@localhost", Nickname: "alice", IsLocal: true}, Text: "thanks",
	})
	if err := backend.Comments().PutLocal(alice.Nickname, created.ID, &comments); err != nil {
		t.Fatal(err)
	}

	if _, err := backend.Social().Create(social.Follow{
		FollowerID: alice.ID, TargetHandle: handle, TargetNickname: "bob",
		TargetIsLocal: true, Status: social.StatusActive, CreatedAt: time.Now().UTC(),
	}); err != nil {
		t.Fatal(err)
	}
	if _, err := backend.Social().Create(social.Follow{
		FollowerID: bob.ID, TargetHandle: "alice@localhost", TargetNickname: "alice",
		TargetIsLocal: true, Status: social.StatusActive, CreatedAt: time.Now().UTC(),
	}); err != nil {
		t.Fatal(err)
	}

	bobWorkout, err := backend.Workouts().Create(bob.Nickname, &workouts.Workout{
		Name: "Bob run", SportType: "Run", StartDate: time.Now().UTC(), DurationSeconds: 50,
	})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := backend.Equipment().Create(bob.Nickname, &equipment.Equipment{
		Type: equipment.TypeShoes, Name: "Trail",
	}); err != nil {
		t.Fatal(err)
	}
	if err := backend.PAT().Create(pat.TokenRecord{
		ID: "pat-bob", TokenHash: "hash-bob", TokenPrefix: "grom_pat_bb",
		UserID: bob.ID, Name: "cli", Scopes: []string{pat.ScopeWorkoutsRead},
		CreatedAt: time.Now().UTC(),
	}); err != nil {
		t.Fatal(err)
	}

	if err := backend.PurgeUser(bob.ID, bob.Nickname, handle); err != nil {
		t.Fatalf("PurgeUser: %v", err)
	}

	gotLikes, err := backend.Likes().GetLocal(alice.Nickname, created.ID)
	if err != nil {
		t.Fatal(err)
	}
	if workouts.LikesContainUser(gotLikes, handle) {
		t.Fatalf("expected bob like removed: %#v", gotLikes)
	}
	gotComments, err := backend.Comments().GetLocal(alice.Nickname, created.ID)
	if err != nil {
		t.Fatal(err)
	}
	if gotComments.CommentsNum != 1 || gotComments.Comments[0].ID != "c2" {
		t.Fatalf("expected only alice comment left: %#v", gotComments)
	}
	if list, err := backend.Social().ListByFollower(alice.ID); err != nil || len(list) != 0 {
		t.Fatalf("alice follows: %#v err=%v", list, err)
	}
	if list, err := backend.Social().ListByFollower(bob.ID); err != nil || len(list) != 0 {
		t.Fatalf("bob follows: %#v err=%v", list, err)
	}
	if _, err := backend.Workouts().Get(bob.Nickname, bobWorkout.ID); err == nil {
		t.Fatal("expected bob workout removed")
	}
	if eq, err := backend.Equipment().List(bob.Nickname); err != nil || len(eq) != 0 {
		t.Fatalf("bob equipment: %#v err=%v", eq, err)
	}
	if pats, err := backend.PAT().ListByUser(bob.ID); err != nil || len(pats) != 0 {
		t.Fatalf("bob pats: %#v err=%v", pats, err)
	}
	if _, err := os.Stat(data.UserDir(backend.Location(), "bob")); !os.IsNotExist(err) {
		t.Fatalf("expected bob dir removed, err=%v", err)
	}
}

func TestPurgeUserRejectsEmptyIdentity(t *testing.T) {
	backend := openTestBackend(t)
	if err := backend.PurgeUser("", "bob", "bob@localhost"); err == nil {
		t.Fatal("expected error for empty user id")
	}
	if err := backend.PurgeUser("id", "", "bob@localhost"); err == nil {
		t.Fatal("expected error for empty nickname")
	}
}
