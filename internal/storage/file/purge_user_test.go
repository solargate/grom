package file_test

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/solargate/grom/internal/auth/pat"
	"github.com/solargate/grom/internal/data"
	"github.com/solargate/grom/internal/equipment"
	"github.com/solargate/grom/internal/social"
	"github.com/solargate/grom/internal/storage/file"
	"github.com/solargate/grom/internal/workouts"
)

func TestPurgeUserRemovesDirAndCredentials(t *testing.T) {
	dir := t.TempDir()
	backend, err := file.Open(dir)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	t.Cleanup(func() { _ = backend.Close() })

	alice, err := backend.Users().Create("alice", "Alice", "alice@example.com", "password12")
	if err != nil {
		t.Fatal(err)
	}
	bob, err := backend.Users().Create("bob", "Bob", "bob@example.com", "password12")
	if err != nil {
		t.Fatal(err)
	}

	workout := &workouts.Workout{
		Name:            "Run",
		SportType:       "Run",
		StartDate:       time.Now().UTC(),
		DurationSeconds: 100,
	}
	created, err := backend.Workouts().Create(alice.Nickname, workout)
	if err != nil {
		t.Fatal(err)
	}
	likes := workouts.AddWorkoutLikeUser(nil, workouts.WorkoutLikeUser{
		Handle:   "bob@localhost",
		Nickname: "bob",
		IsLocal:  true,
	})
	if err := backend.Likes().PutLocal(alice.Nickname, created.ID, &likes); err != nil {
		t.Fatal(err)
	}

	handle := "bob@localhost"
	if err := backend.PurgeUser(bob.ID, bob.Nickname, handle); err != nil {
		t.Fatalf("PurgeUser: %v", err)
	}

	if _, err := backend.Users().FindByID(bob.ID); err == nil {
		t.Fatal("expected bob removed")
	}
	if _, err := os.Stat(data.UserDir(dir, "bob")); !os.IsNotExist(err) {
		t.Fatalf("bob dir still present: %v", err)
	}
	got, err := backend.Likes().GetLocal(alice.Nickname, created.ID)
	if err != nil {
		t.Fatal(err)
	}
	if workouts.LikesContainUser(got, handle) {
		t.Fatalf("expected bob like removed: %#v", got)
	}
	if _, err := os.Stat(filepath.Join(dir, "users.yaml")); err != nil {
		t.Fatal(err)
	}
}

func TestPurgeUserScrubsCommentsFollowsEquipmentAndPAT(t *testing.T) {
	dir := t.TempDir()
	backend, err := file.Open(dir)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	t.Cleanup(func() { _ = backend.Close() })

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
	comments := workouts.AddWorkoutComment(nil, workouts.WorkoutComment{
		ID:   "c1",
		User: workouts.WorkoutLikeUser{Handle: handle, Nickname: "bob", IsLocal: true},
		Text: "nice",
	})
	comments = workouts.AddWorkoutComment(&comments, workouts.WorkoutComment{
		ID:   "c2",
		User: workouts.WorkoutLikeUser{Handle: "alice@localhost", Nickname: "alice", IsLocal: true},
		Text: "thanks",
	})
	if err := backend.Comments().PutLocal(alice.Nickname, created.ID, &comments); err != nil {
		t.Fatal(err)
	}

	if _, err := backend.Social().Create(social.Follow{
		FollowerID:     alice.ID,
		TargetHandle:   handle,
		TargetNickname: "bob",
		TargetIsLocal:  true,
		Status:         social.StatusActive,
		CreatedAt:      time.Now().UTC(),
	}); err != nil {
		t.Fatal(err)
	}
	if _, err := backend.Social().Create(social.Follow{
		FollowerID:     bob.ID,
		TargetHandle:   "alice@localhost",
		TargetNickname: "alice",
		TargetIsLocal:  true,
		Status:         social.StatusActive,
		CreatedAt:      time.Now().UTC(),
	}); err != nil {
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

	gotComments, err := backend.Comments().GetLocal(alice.Nickname, created.ID)
	if err != nil {
		t.Fatal(err)
	}
	if gotComments.CommentsNum != 1 || gotComments.Comments[0].ID != "c2" {
		t.Fatalf("expected only alice comment left: %#v", gotComments)
	}

	aliceFollowing, err := backend.Social().ListByFollower(alice.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(aliceFollowing) != 0 {
		t.Fatalf("expected follows involving bob removed from alice: %#v", aliceFollowing)
	}
	bobFollowing, err := backend.Social().ListByFollower(bob.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(bobFollowing) != 0 {
		t.Fatalf("expected bob follows removed: %#v", bobFollowing)
	}

	eq, err := backend.Equipment().List(bob.Nickname)
	if err != nil {
		t.Fatal(err)
	}
	if len(eq) != 0 {
		t.Fatalf("expected bob equipment gone: %#v", eq)
	}
	pats, err := backend.PAT().ListByUser(bob.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(pats) != 0 {
		t.Fatalf("expected bob PATs gone: %#v", pats)
	}
}

func TestPurgeUserRejectsEmptyIdentity(t *testing.T) {
	dir := t.TempDir()
	backend, err := file.Open(dir)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	t.Cleanup(func() { _ = backend.Close() })

	if err := backend.PurgeUser("", "bob", "bob@localhost"); err == nil {
		t.Fatal("expected error for empty user id")
	}
	if err := backend.PurgeUser("id", "", "bob@localhost"); err == nil {
		t.Fatal("expected error for empty nickname")
	}
}
