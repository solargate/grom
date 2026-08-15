package file

import (
	"fmt"
	"strings"

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
	if err := b.patStore.DeleteAllForUser(userID); err != nil {
		return fmt.Errorf("delete personal access tokens: %w", err)
	}
	if err := b.resetTokens.DeleteAllForUser(userID); err != nil {
		return fmt.Errorf("delete reset tokens: %w", err)
	}

	// RemoveAll of the user directory covers workouts, equipment, profile, avatar,
	// federation keys/followers/inbox/outbox for this nickname (file driver).
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
