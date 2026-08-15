package bbolt

import (
	"fmt"
	"strings"

	bolt "go.etcd.io/bbolt"

	"github.com/solargate/grom/internal/workouts"
)

// PurgeUser permanently removes the account and all related local data.
func (b *Backend) PurgeUser(userID, nickname, localHandle string) error {
	if userID == "" || nickname == "" {
		return fmt.Errorf("user id and nickname are required")
	}

	usersList, err := b.users.ListAll()
	if err != nil {
		return fmt.Errorf("list users: %w", err)
	}

	for _, u := range usersList {
		if u.ID == userID {
			continue
		}
		if err := scrubLocalInteractions(b, u.Nickname, localHandle); err != nil {
			return fmt.Errorf("scrub interactions for %q: %w", u.Nickname, err)
		}
		if localHandle == "" {
			continue
		}
		if err := purgeFederatedCachesForOwner(b, u.Nickname, localHandle); err != nil {
			return fmt.Errorf("purge federated caches for viewer %q: %w", u.Nickname, err)
		}
		if err := b.fed.Inbox().DeleteAllForOwner(u.Nickname, localHandle); err != nil {
			return fmt.Errorf("purge federated inbox for viewer %q: %w", u.Nickname, err)
		}
	}

	if err := b.social.DeleteInvolving(userID, localHandle); err != nil {
		return fmt.Errorf("delete follows: %w", err)
	}

	workoutsList, err := b.workouts.List(nickname)
	if err != nil {
		return fmt.Errorf("list workouts: %w", err)
	}
	for _, w := range workoutsList {
		_ = b.likes.DeleteLocal(nickname, w.ID)
		_ = b.comments.DeleteLocal(nickname, w.ID)
		if err := b.workouts.Delete(nickname, w.ID); err != nil {
			return fmt.Errorf("delete workout %s: %w", w.ID, err)
		}
	}

	eqList, err := b.equipment.List(nickname)
	if err != nil {
		return fmt.Errorf("list equipment: %w", err)
	}
	for _, item := range eqList {
		if err := b.equipment.Delete(nickname, item.ID); err != nil {
			return fmt.Errorf("delete equipment %s: %w", item.ID, err)
		}
	}

	if err := b.deleteNicknamePrefixedBuckets(nickname); err != nil {
		return err
	}
	if err := b.patStore.DeleteAllForUser(userID); err != nil {
		return fmt.Errorf("delete personal access tokens: %w", err)
	}
	if err := b.resetTokens.DeleteAllForUser(userID); err != nil {
		return fmt.Errorf("delete reset tokens: %w", err)
	}

	if err := b.users.Delete(userID); err != nil {
		return fmt.Errorf("delete user: %w", err)
	}
	return nil
}

func scrubLocalInteractions(b *Backend, ownerNickname, handle string) error {
	if handle == "" {
		return nil
	}
	items, err := b.workouts.List(ownerNickname)
	if err != nil {
		return err
	}
	for _, w := range items {
		likes, err := b.likes.GetLocal(ownerNickname, w.ID)
		if err != nil {
			return err
		}
		if workouts.LikesContainUser(likes, handle) {
			updated := workouts.RemoveWorkoutLikeUser(likes, handle)
			if err := b.likes.PutLocal(ownerNickname, w.ID, &updated); err != nil {
				return err
			}
		}
		comments, err := b.comments.GetLocal(ownerNickname, w.ID)
		if err != nil {
			return err
		}
		if comments == nil {
			continue
		}
		updatedComments := workouts.RemoveWorkoutCommentsByHandle(comments, handle)
		if updatedComments.CommentsNum != comments.CommentsNum {
			if err := b.comments.PutLocal(ownerNickname, w.ID, &updatedComments); err != nil {
				return err
			}
		}
	}
	return nil
}

func purgeFederatedCachesForOwner(b *Backend, viewerNickname, ownerHandle string) error {
	feed, err := b.fed.Inbox().List(viewerNickname)
	if err != nil {
		return err
	}
	for _, item := range feed {
		if !strings.EqualFold(item.Author.Handle, ownerHandle) {
			continue
		}
		_ = b.likes.DeleteFederated(viewerNickname, ownerHandle, item.ID)
		_ = b.comments.DeleteFederated(viewerNickname, ownerHandle, item.ID)
	}
	return nil
}

func (b *Backend) deleteNicknamePrefixedBuckets(nickname string) error {
	prefix := []byte(nickname + "/")
	activityPrefix := []byte(nickname + "|")
	return b.db.Update(func(tx *bolt.Tx) error {
		for _, bucket := range [][]byte{
			bucketFedFollowers,
			bucketFedInbox,
			bucketFedAuthors,
			bucketFedWorkoutLikes,
			bucketFedWorkoutComments,
			bucketFedSpeedCharts,
			bucketFedHeartRateCharts,
			bucketWorkoutLikes,
			bucketWorkoutComments,
			bucketSpeedCharts,
			bucketHeartRateCharts,
			bucketEquipment,
			bucketWorkouts,
		} {
			if err := deletePrefix(tx.Bucket(bucket), prefix); err != nil {
				return fmt.Errorf("purge bucket %s: %w", bucket, err)
			}
		}
		if err := deletePrefix(tx.Bucket(bucketLikeActivities), activityPrefix); err != nil {
			return fmt.Errorf("purge like activities: %w", err)
		}
		if err := deletePrefix(tx.Bucket(bucketCommentActivities), activityPrefix); err != nil {
			return fmt.Errorf("purge comment activities: %w", err)
		}
		return nil
	})
}

func deletePrefix(b *bolt.Bucket, prefix []byte) error {
	if b == nil {
		return nil
	}
	var keys [][]byte
	c := b.Cursor()
	for k, _ := c.Seek(prefix); k != nil && strings.HasPrefix(string(k), string(prefix)); k, _ = c.Next() {
		keys = append(keys, append([]byte(nil), k...))
	}
	for _, k := range keys {
		if err := b.Delete(k); err != nil {
			return err
		}
	}
	return nil
}
