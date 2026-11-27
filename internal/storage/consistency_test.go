package storage_test

import (
	"errors"
	"fmt"
	"path/filepath"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/solargate/grom/internal/config"
	"github.com/solargate/grom/internal/storage"
	"github.com/solargate/grom/internal/tracks"
	"github.com/solargate/grom/internal/users"
	"github.com/solargate/grom/internal/workouts"
)

func forEachDriver(t *testing.T, run func(*testing.T, storage.Backend)) {
	t.Helper()
	for _, driver := range []config.StorageDriver{config.StorageDriverFile, config.StorageDriverBBolt} {
		t.Run(string(driver), func(t *testing.T) {
			dir := t.TempDir()
			backend, err := storage.Open(config.StorageConfig{
				Driver: driver, ResolvedLocation: dir, ResolvedBBoltPath: filepath.Join(dir, "grom.db"),
			})
			if err != nil {
				t.Fatal(err)
			}
			t.Cleanup(func() { _ = backend.Close() })
			run(t, backend)
		})
	}
}

func TestConcurrentExternalIDCommitAndAttachmentRetry(t *testing.T) {
	forEachDriver(t, func(t *testing.T, backend storage.Backend) {
		user, err := backend.Users().Create("alice", "Alice", "alice@example.com", "password123")
		if err != nil {
			t.Fatal(err)
		}
		if _, err := backend.GetProfileWithWorkoutTotal(user.ID); err != nil {
			t.Fatal(err)
		}
		draft := workouts.Workout{
			Name: "Run", SportType: "Run", StartDate: time.Now().UTC(),
			ExternalID: &workouts.ExternalID{Name: "strava", ID: "activity"},
		}
		brokenPhoto := []workouts.MediaFileInput{{Filename: "photo.jpg", Data: []byte("broken")}}
		if _, err := backend.Workouts().CreateWithAttachments("alice", &draft, nil, brokenPhoto); err == nil {
			t.Fatal("invalid attachment unexpectedly succeeded")
		}
		if items, err := backend.Workouts().List("alice"); err != nil || len(items) != 0 {
			t.Fatalf("partial workout visible: %v, %v", items, err)
		}
		assertStoredWorkoutTotal(t, backend, user.ID, 0)

		var created atomic.Int32
		var wg sync.WaitGroup
		start := make(chan struct{})
		// A parsed track takes the staged commit path and a plain workout takes the direct path.
		for i := range 12 {
			wg.Go(func() {
				<-start
				item := workouts.Workout{
					Name: "Run", SportType: "Run", StartDate: time.Now().UTC(),
					ExternalID: draft.ExternalID,
				}
				var err error
				if i%2 == 0 {
					_, err = backend.Workouts().CreateWithTrack("alice", &item, &workouts.TrackInput{
						Filename: "track.gpx", Data: []byte("<gpx/>"), Parsed: &tracks.Data{},
					})
				} else {
					_, err = backend.Workouts().Create("alice", &item)
				}
				if err == nil {
					created.Add(1)
				} else if !errors.Is(err, workouts.ErrExternalIDExists) {
					t.Errorf("create %d: %v", i, err)
				}
			})
		}
		close(start)
		wg.Wait()
		items, err := backend.Workouts().List("alice")
		if err != nil || created.Load() != 1 || len(items) != 1 {
			t.Fatalf("got %d successful writes and %d workouts, err=%v", created.Load(), len(items), err)
		}
		assertStoredWorkoutTotal(t, backend, user.ID, 1)
	})
}

func assertStoredWorkoutTotal(t *testing.T, backend storage.Backend, userID string, want int) {
	t.Helper()
	profile, err := backend.Users().GetProfile(userID)
	if err != nil {
		t.Fatal(err)
	}
	if profile.TotalWorkouts == nil {
		t.Fatal("workout total was not stored")
	}
	if got := *profile.TotalWorkouts; got != want {
		t.Fatalf("total_workouts = %d, want %d", got, want)
	}
}

func TestWorkoutTotalBackfillAndConcurrentChanges(t *testing.T) {
	forEachDriver(t, func(t *testing.T, backend storage.Backend) {
		user, err := backend.Users().Create("alice", "Alice", "alice@example.com", "password123")
		if err != nil {
			t.Fatal(err)
		}
		svc := backend.Workouts()
		workout, err := svc.Create(user.Nickname, &workouts.Workout{
			Name: "Ride", SportType: "Ride", StartDate: time.Now().UTC(),
		})
		if err != nil {
			t.Fatal(err)
		}
		if err := backend.Users().PutProfile(user.ID, users.Profile{LastSportType: "Ride"}); err != nil {
			t.Fatal(err)
		}
		profile, err := backend.GetProfileWithWorkoutTotal(user.ID)
		if err != nil || profile.LastSportType != "Ride" {
			t.Fatalf("backfilled profile = %+v, err = %v", profile, err)
		}
		assertStoredWorkoutTotal(t, backend, user.ID, 1)

		var wg sync.WaitGroup
		const adds = 8
		for i := range adds {
			wg.Go(func() {
				item := workouts.Workout{
					Name: fmt.Sprintf("Run %d", i), SportType: "Run", StartDate: time.Now().UTC(),
				}
				var err error
				if i%2 == 0 {
					item.ExternalID = &workouts.ExternalID{Name: "strava", ID: fmt.Sprint(i)}
					_, err = svc.ImportWithAttachments(user.Nickname, &item, &workouts.TrackInput{
						Filename: "track.gpx", Data: []byte("<gpx/>"), Parsed: &tracks.Data{},
					}, nil)
				} else {
					_, err = svc.Create(user.Nickname, &item)
				}
				if err != nil {
					t.Errorf("creating workout: %v", err)
				}
			})
		}
		wg.Go(func() {
			if err := svc.Delete(user.Nickname, workout.ID); err != nil {
				t.Errorf("deleting workout: %v", err)
			}
			if err := svc.Delete(user.Nickname, workout.ID); !errors.Is(err, workouts.ErrWorkoutNotFound) {
				t.Errorf("repeated deletion: %v", err)
			}
		})
		wg.Wait()
		assertStoredWorkoutTotal(t, backend, user.ID, adds)

		items, err := svc.List(user.Nickname)
		if err != nil {
			t.Fatal(err)
		}
		for _, item := range items {
			wg.Go(func() {
				if err := svc.Delete(user.Nickname, item.ID); err != nil {
					t.Errorf("deleting workout: %v", err)
				}
			})
		}
		wg.Wait()
		assertStoredWorkoutTotal(t, backend, user.ID, 0)
	})
}

func TestConcurrentLocalSocialUpdates(t *testing.T) {
	forEachDriver(t, func(t *testing.T, backend storage.Backend) {
		item, err := backend.Workouts().Create("alice", &workouts.Workout{
			Name: "Run", SportType: "Run", StartDate: time.Now().UTC(),
		})
		if err != nil {
			t.Fatal(err)
		}
		var wg sync.WaitGroup
		for i := range 16 {
			wg.Go(func() {
				id := fmt.Sprint(i)
				actor := workouts.WorkoutLikeUser{Handle: id + "@example.org", Nickname: id}
				if err := backend.Likes().UpdateLocal("alice", item.ID, func(current *workouts.WorkoutLikes) *workouts.WorkoutLikes {
					next := workouts.AddWorkoutLikeUser(current, actor)
					return &next
				}); err != nil {
					t.Error(err)
				}
				if err := backend.Comments().UpdateLocal("alice", item.ID, func(current *workouts.WorkoutComments) *workouts.WorkoutComments {
					next := workouts.AddWorkoutComment(current, workouts.WorkoutComment{ID: id, User: actor, Text: id})
					return &next
				}); err != nil {
					t.Error(err)
				}
			})
		}
		wg.Wait()
		likes, err := backend.Likes().GetLocal("alice", item.ID)
		if err != nil || likes == nil || likes.Likes != 16 {
			t.Fatalf("likes = %v, err=%v", likes, err)
		}
		comments, err := backend.Comments().GetLocal("alice", item.ID)
		if err != nil || comments == nil || comments.CommentsNum != 16 {
			t.Fatalf("comments = %v, err=%v", comments, err)
		}
	})
}

func TestWorkoutPaginationUsesFractionalSeconds(t *testing.T) {
	forEachDriver(t, func(t *testing.T, backend storage.Backend) {
		second := time.Now().UTC().Truncate(time.Second)
		create := func() *workouts.Workout {
			t.Helper()
			item, err := backend.Workouts().Create("alice", &workouts.Workout{Name: "Run", SportType: "Run", StartDate: second})
			if err != nil {
				t.Fatal(err)
			}
			return item
		}
		newer, older := create(), create()
		if newer.ID > older.ID {
			newer, older = older, newer
		}
		newer.StartDate = second.Add(900 * time.Millisecond)
		older.StartDate = second.Add(100 * time.Millisecond)
		for _, item := range []*workouts.Workout{newer, older} {
			if _, err := backend.Workouts().Update("alice", item.ID, item); err != nil {
				t.Fatal(err)
			}
		}
		page, more, err := backend.Workouts().ListPage("alice", nil, 1, nil)
		if err != nil || !more || len(page) != 1 || page[0].ID != newer.ID {
			t.Fatalf("first page = %v, more=%v, err=%v", page, more, err)
		}
		cursor := workouts.CursorFromWorkout(page[0])
		page, more, err = backend.Workouts().ListPage("alice", &cursor, 1, nil)
		if err != nil || more || len(page) != 1 || page[0].ID != older.ID {
			t.Fatalf("second page = %v, more=%v, err=%v", page, more, err)
		}
	})
}

func TestFederatedFeedCursorResolvesEqualIDsAndTimes(t *testing.T) {
	forEachDriver(t, func(t *testing.T, backend storage.Backend) {
		inbox := backend.Federation().Inbox()
		item := workouts.Workout{
			ID: "42", Name: "Remote run", SportType: "Run", StartDate: time.Now().UTC(),
		}
		for _, owner := range []string{"bob@first.test", "bob@second.test"} {
			if err := inbox.Save("alice", owner, &item, nil, nil, nil); err != nil {
				t.Fatal(err)
			}
		}
		page, more, err := inbox.ListPage("alice", nil, 1)
		if err != nil || !more || len(page) != 1 || page[0].Author.Handle != "bob@second.test" {
			t.Fatalf("first page = %v, more=%v, err=%v", page, more, err)
		}
		cursor := workouts.CursorFromWorkout(page[0].Workout)
		cursor.Owner = &page[0].SortOwner
		decoded, err := workouts.DecodeCursor(cursor.Encode())
		if err != nil {
			t.Fatal(err)
		}
		page, more, err = inbox.ListPage("alice", decoded, 1)
		if err != nil || more || len(page) != 1 || page[0].Author.Handle != "bob@first.test" {
			t.Fatalf("second page = %v, more=%v, err=%v", page, more, err)
		}
	})
}

func TestExternalIDComponentsDoNotCollide(t *testing.T) {
	forEachDriver(t, func(t *testing.T, backend storage.Backend) {
		for _, external := range []workouts.ExternalID{{Name: "source/a", ID: "b"}, {Name: "source", ID: "a/b"}} {
			_, err := backend.Workouts().Create("alice", &workouts.Workout{
				Name: "Run", SportType: "Run", StartDate: time.Now().UTC(), ExternalID: &external,
			})
			if err != nil {
				t.Fatal(err)
			}
		}
	})
}
