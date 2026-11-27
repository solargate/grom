package bbolt

import (
	"path/filepath"
	"slices"
	"testing"
	"time"

	"github.com/solargate/grom/internal/storage/keys"
	"github.com/solargate/grom/internal/workouts"
	bolt "go.etcd.io/bbolt"
)

func TestDeleteAllRemovesWorkoutSidecars(t *testing.T) {
	dir := t.TempDir()
	backend, err := Open(filepath.Join(dir, "grom.db"), dir)
	if err != nil {
		t.Fatal(err)
	}
	defer backend.Close()

	workout, err := backend.Workouts().Create("alice", &workouts.Workout{
		Name: "Run", SportType: "Run", StartDate: time.Now().UTC(),
	})
	if err != nil {
		t.Fatal(err)
	}
	dirName := keys.WorkoutDirName(workout.StartDate, workout.ID)
	sidecars := []struct{ bucket, key []byte }{
		{bucketWorkoutLikes, workoutLikesKey("alice", dirName)},
		{bucketWorkoutComments, workoutCommentsKey("alice", dirName)},
		{bucketSpeedCharts, localSpeedChartKey("alice", dirName)},
		{bucketHeartRateCharts, localHeartRateChartKey("alice", dirName)},
		{bucketCadenceCharts, localCadenceChartKey("alice", dirName)},
	}
	if err := backend.DB().Update(func(tx *bolt.Tx) error {
		for _, entry := range sidecars {
			if err := tx.Bucket(entry.bucket).Put(entry.key, []byte("sidecar")); err != nil {
				return err
			}
		}
		return nil
	}); err != nil {
		t.Fatal(err)
	}

	deleted, err := backend.Workouts().DeleteAll("alice")
	if err != nil || !slices.Equal(deleted, []string{workout.ID}) {
		t.Fatalf("DeleteAll = %v, %v", deleted, err)
	}
	if err := backend.DB().View(func(tx *bolt.Tx) error {
		for _, entry := range sidecars {
			if tx.Bucket(entry.bucket).Get(entry.key) != nil {
				t.Errorf("workout data remains in %s", entry.bucket)
			}
		}
		return nil
	}); err != nil {
		t.Fatal(err)
	}
}
