package file_test

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/solargate/grom/internal/data"
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
