package file

import (
	"os"
	"strings"
	"testing"
	"time"

	"github.com/solargate/grom/internal/workouts"
)

func TestWorkoutTotalRecoversAfterProfileWriteFailure(t *testing.T) {
	b, err := Open(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = b.Close() })
	user, err := b.users.Create("alice", "Alice", "alice@example.com", "password123")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := b.GetProfileWithWorkoutTotal(user.ID); err != nil {
		t.Fatal(err)
	}
	if err := b.users.SetLastSportType(user.ID, "Run"); err != nil {
		t.Fatal(err)
	}

	profilePath := b.users.profilePath(user.Nickname)
	item := &workouts.Workout{
		ID: "123456", Name: "Run", SportType: "Run", StartDate: time.Date(2026, 11, 27, 0, 0, 0, 0, time.UTC),
		DurationSeconds: 600, Distance: 1800,
	}
	func() {
		workoutCommitMu.Lock()
		defer workoutCommitMu.Unlock()
		err = b.workoutRepo.withWorkoutChange(user.Nickname, nil, item, func() error {
			// Block the final profile write after the invalidation has succeeded.
			if err := os.Mkdir(profilePath+".tmp", 0700); err != nil {
				return err
			}
			_, err := b.workoutRepo.saveNewWorkout(user.Nickname, item)
			return err
		})
	}()
	if err != nil {
		t.Fatalf("workout commit: %v", err)
	}
	profile, err := b.users.GetProfile(user.ID)
	if err != nil || profile.TotalWorkouts != nil || profile.Workouts != nil {
		t.Fatalf("expected invalidated statistics, got %+v, err = %v", profile, err)
	}

	if err := os.Remove(profilePath + ".tmp"); err != nil {
		t.Fatal(err)
	}
	profile, err = b.GetProfileWithWorkoutTotal(user.ID)
	if err != nil || !profile.HasWorkoutStats() || *profile.TotalWorkouts != 1 || len((*profile.Workouts)["2026-11-27"]) != 1 {
		t.Fatalf("expected recovered statistics, got %+v, err = %v", profile, err)
	}
	raw, err := os.ReadFile(profilePath)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(string(raw), "total_workouts: 1\nlast_sport_type: Run\n") {
		t.Fatalf("unexpected profile YAML:\n%s", raw)
	}
}
