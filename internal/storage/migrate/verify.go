package migrate

import (
	"context"
	"crypto/sha256"
	"fmt"
	"maps"
	"slices"

	"github.com/solargate/grom/internal/storage"
	"github.com/solargate/grom/internal/storage/keys"
	"github.com/solargate/grom/internal/users"
	"github.com/solargate/grom/internal/workouts"
	"gopkg.in/yaml.v3"
)

type migrationRecords map[[4]string][sha256.Size]byte

func (records migrationRecords) add(key [4]string, value any) error {
	data, err := yaml.Marshal(value)
	if err != nil {
		return fmt.Errorf("encode %s for verification: %w", key[0], err)
	}
	records[key] = sha256.Sum256(data)
	return nil
}

func verifyContent(src, dst storage.Backend, location, domain string) error {
	expected, err := collectRecords(src, location, domain, false)
	if err != nil {
		return fmt.Errorf("verify source records: %w", err)
	}
	actual, err := collectRecords(dst, location, domain, true)
	if err != nil {
		return fmt.Errorf("verify target records: %w", err)
	}
	for _, key := range slices.SortedFunc(maps.Keys(expected), func(a, b [4]string) int { return slices.Compare(a[:], b[:]) }) {
		if value, ok := actual[key]; !ok || value != expected[key] {
			// Keys can contain endpoints or other private data.
			return fmt.Errorf("verify failed: missing or different %s record", key[0])
		}
	}
	if len(actual) != len(expected) {
		return fmt.Errorf("verify failed: target contains extra records")
	}
	return nil
}

func collectRecords(backend storage.Backend, location, domain string, verifyStoredStats bool) (migrationRecords, error) {
	records := make(migrationRecords)
	usersList, err := backend.Users().ListAll()
	if err != nil {
		return nil, err
	}
	charts, err := chartStores(backend)
	if err != nil {
		return nil, err
	}
	for _, user := range usersList {
		user.CreatedAt = user.CreatedAt.UTC()
		if err := records.add([4]string{"user", user.ID}, user); err != nil {
			return nil, err
		}
		profile, err := backend.Users().GetProfile(user.ID)
		if err != nil {
			return nil, err
		}
		equipment, err := backend.Equipment().List(user.Nickname)
		if err != nil {
			return nil, err
		}
		for _, item := range equipment {
			if err := records.add([4]string{"equipment", user.Nickname, item.ID}, item); err != nil {
				return nil, err
			}
		}
		followers, err := backend.Federation().Followers().List(user.Nickname)
		if err != nil {
			return nil, err
		}
		for _, follower := range followers {
			if err := records.add([4]string{"follower", user.Nickname, follower.ActorURI}, follower); err != nil {
				return nil, err
			}
		}
		items, err := backend.Workouts().List(user.Nickname)
		if err != nil {
			return nil, err
		}
		if profile == nil {
			profile = &users.Profile{}
		}
		expected := *profile
		expected.RebuildWorkoutStats(items)
		if verifyStoredStats && (profile.TotalWorkouts == nil || *profile.TotalWorkouts != *expected.TotalWorkouts) {
			return nil, fmt.Errorf("workout total differs from saved workouts for user %q", user.Nickname)
		}
		if verifyStoredStats && (!profile.HasWorkoutStats() || !maps.EqualFunc(*profile.Workouts, *expected.Workouts, slices.Equal)) {
			return nil, fmt.Errorf("workout history differs from saved workouts for user %q", user.Nickname)
		}
		if err := records.add([4]string{"profile", user.ID}, expected); err != nil {
			return nil, err
		}
		for _, item := range items {
			if err := records.addWorkout(charts, user.Nickname, "", item); err != nil {
				return nil, err
			}
			likes, err := backend.Likes().GetLocal(user.Nickname, item.ID)
			if err != nil {
				return nil, err
			}
			if likes != nil && likes.Likes > 0 {
				if err := records.add([4]string{"local likes", user.Nickname, item.ID}, workouts.NormalizeWorkoutLikes(likes)); err != nil {
					return nil, err
				}
			}
			comments, err := backend.Comments().GetLocal(user.Nickname, item.ID)
			if err != nil {
				return nil, err
			}
			if comments != nil && comments.CommentsNum > 0 {
				if err := records.add([4]string{"local comments", user.Nickname, item.ID}, workouts.NormalizeWorkoutComments(comments)); err != nil {
					return nil, err
				}
			}
		}
	}
	follows, err := listFollows(backend)
	if err != nil {
		return nil, err
	}
	for _, follow := range follows {
		follow.CreatedAt = follow.CreatedAt.UTC()
		if err := records.add([4]string{"follow", follow.ID}, follow); err != nil {
			return nil, err
		}
	}
	authors, inbox, err := loadFederationInbox(backend, location)
	if err != nil {
		return nil, err
	}
	for viewer, byOwner := range authors {
		for owner, author := range byOwner {
			if err := records.add([4]string{"inbox author", viewer, owner}, author); err != nil {
				return nil, err
			}
		}
	}
	for viewer, byOwner := range inbox {
		for owner, items := range byOwner {
			for _, item := range items {
				if err := records.addWorkout(charts, viewer, owner, item); err != nil {
					return nil, err
				}
			}
		}
	}
	err = visitLikeActivities(backend, authors, inbox, domain, func(actor, objectID, activityID string) error {
		if actor == "" || objectID == "" || activityID == "" {
			return nil
		}
		return records.add([4]string{"like activity", actor, objectID}, activityID)
	})
	if err != nil {
		return nil, err
	}
	return records, records.addAdditional(backend)
}

func (records migrationRecords) addAdditional(backend storage.Backend) error {
	likes, err := listFederatedLikes(backend)
	if err != nil {
		return err
	}
	for _, item := range likes {
		if err := records.add([4]string{"federated likes", item.ViewerNickname, item.OwnerHandle, item.WorkoutID}, workouts.NormalizeWorkoutLikes(&item.Likes)); err != nil {
			return err
		}
	}
	comments, err := listFederatedComments(backend)
	if err != nil {
		return err
	}
	for _, item := range comments {
		if err := records.add([4]string{"federated comments", item.ViewerNickname, item.OwnerHandle, item.WorkoutID}, workouts.NormalizeWorkoutComments(&item.Comments)); err != nil {
			return err
		}
	}
	activities, err := listCommentActivities(backend)
	if err != nil {
		return err
	}
	for _, item := range activities {
		if err := records.add([4]string{"comment activity", item.ActorNickname, item.NoteID}, item.ActivityID); err != nil {
			return err
		}
	}
	tokens, err := listPersonalAccessTokens(backend)
	if err != nil {
		return err
	}
	for _, item := range tokens {
		if err := records.add([4]string{"personal access token", item.ID}, item); err != nil {
			return err
		}
	}
	subs, err := listPushSubscriptions(backend)
	if err != nil {
		return err
	}
	for _, item := range subs {
		if err := records.add([4]string{"push subscription", item.UserID, item.InstallationID}, item); err != nil {
			return err
		}
	}
	return nil
}

func (records migrationRecords) addWorkout(charts chartStoreSet, viewer, owner string, item workouts.Workout) error {
	item.StartDate = item.StartDate.UTC()
	if err := records.add([4]string{"workout", viewer, owner, item.ID}, item); err != nil {
		return err
	}
	if item.Track == "" {
		return nil
	}
	key := [4]string{"speed chart", viewer, owner, item.ID}
	ctx := context.Background()
	dir := keys.WorkoutDirName(item.StartDate, item.ID)
	if owner == "" {
		if err := addChart(records, key, func() ([]workouts.SpeedSample, error) {
			return charts.speed.ReadLocal(ctx, viewer, dir)
		}, workouts.MarshalSpeedChartBinary); err != nil {
			return err
		}
		key[0] = "heart rate chart"
		if err := addChart(records, key, func() ([]workouts.HeartRateSample, error) {
			return charts.hr.ReadLocal(ctx, viewer, dir)
		}, workouts.MarshalHeartRateChartBinary); err != nil {
			return err
		}
		key[0] = "cadence chart"
		return addChart(records, key, func() ([]workouts.CadenceSample, error) {
			return charts.cadence.ReadLocal(ctx, viewer, dir)
		}, workouts.MarshalCadenceChartBinary)
	}
	if err := addChart(records, key, func() ([]workouts.SpeedSample, error) {
		return charts.speed.ReadFederated(ctx, viewer, owner, item.ID)
	}, workouts.MarshalSpeedChartBinary); err != nil {
		return err
	}
	key[0] = "heart rate chart"
	if err := addChart(records, key, func() ([]workouts.HeartRateSample, error) {
		return charts.hr.ReadFederated(ctx, viewer, owner, item.ID)
	}, workouts.MarshalHeartRateChartBinary); err != nil {
		return err
	}
	key[0] = "cadence chart"
	return addChart(records, key, func() ([]workouts.CadenceSample, error) {
		return charts.cadence.ReadFederated(ctx, viewer, owner, item.ID)
	}, workouts.MarshalCadenceChartBinary)
}

func addChart[T any](records migrationRecords, key [4]string, read func() ([]T, error), encode func([]T) ([]byte, error)) error {
	samples, err := read()
	if err != nil || len(samples) == 0 {
		return err
	}
	data, err := encode(samples)
	if err != nil {
		return err
	}
	records[key] = sha256.Sum256(data)
	return nil
}
