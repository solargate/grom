package bbolt_test

import (
	"os"
	"testing"
	"time"

	"github.com/solargate/grom/internal/data"
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
