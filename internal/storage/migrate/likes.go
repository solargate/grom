package migrate

import (
	"fmt"
	"strings"

	"github.com/solargate/grom/internal/federation"
	"github.com/solargate/grom/internal/storage"
	storebbolt "github.com/solargate/grom/internal/storage/bbolt"
	"github.com/solargate/grom/internal/storage/file"
	"github.com/solargate/grom/internal/workouts"
)

func copyLocalLikes(src, dst storage.Backend, nickname string, w *workouts.Workout) error {
	if w == nil {
		return nil
	}
	likes, err := src.Likes().GetLocal(nickname, w.ID)
	if err != nil {
		return fmt.Errorf("read local likes: %w", err)
	}
	if likes == nil || likes.Likes == 0 {
		return nil
	}
	if err := dst.Likes().PutLocal(nickname, w.ID, likes); err != nil {
		return fmt.Errorf("write local likes: %w", err)
	}
	return nil
}

func copyFederatedLikes(src, dst storage.Backend) (int, error) {
	entries, err := listFederatedLikes(src)
	if err != nil {
		return 0, err
	}
	for _, e := range entries {
		likes := e.Likes
		if err := dst.Likes().PutFederated(e.ViewerNickname, e.OwnerHandle, e.WorkoutID, &likes); err != nil {
			return 0, fmt.Errorf("write federated likes %s/%s/%s: %w", e.ViewerNickname, e.OwnerHandle, e.WorkoutID, err)
		}
	}
	return len(entries), nil
}

func copyLikeActivities(src, dst storage.Backend, authors map[string]map[string]federation.AuthorMeta, inbox map[string]map[string][]workouts.Workout) (int, error) {
	seen := make(map[string]struct{})
	copied := 0

	put := func(actor, objectID, activityID string) error {
		if actor == "" || objectID == "" || activityID == "" {
			return nil
		}
		key := actor + "|" + objectID
		if _, ok := seen[key]; ok {
			return nil
		}
		if err := dst.Likes().PutLikeActivityID(actor, objectID, activityID); err != nil {
			return err
		}
		seen[key] = struct{}{}
		copied++
		return nil
	}

	entries, err := listLikeActivities(src)
	if err != nil {
		return 0, err
	}
	for _, e := range entries {
		if err := put(e.ActorNickname, e.ObjectID, e.ActivityID); err != nil {
			return copied, fmt.Errorf("write like activity: %w", err)
		}
	}

	// Reconstruct object IDs from inbox so legacy file activity blobs (plain activity id,
	// no object_id) still migrate via GetLikeActivityID hash lookup.
	for viewer, byOwner := range inbox {
		for ownerKey, list := range byOwner {
			handle := federation.OwnerHandleFromKey(ownerKey)
			nickname := federation.OwnerNicknameFromKey(ownerKey)
			if meta, ok := authors[viewer][ownerKey]; ok {
				if meta.Handle != "" {
					handle = meta.Handle
				}
				if meta.Nickname != "" {
					nickname = meta.Nickname
				}
			}
			for i := range list {
				objectID := remoteWorkoutObjectID(handle, nickname, list[i].ID)
				if objectID == "" {
					continue
				}
				activityID, err := src.Likes().GetLikeActivityID(viewer, objectID)
				if err != nil {
					return copied, fmt.Errorf("read like activity for %s: %w", objectID, err)
				}
				if err := put(viewer, objectID, activityID); err != nil {
					return copied, fmt.Errorf("write reconstructed like activity: %w", err)
				}
			}
		}
	}
	return copied, nil
}

func countLocalLikes(backend storage.Backend) (int, error) {
	usersList, err := backend.Users().ListAll()
	if err != nil {
		return 0, err
	}
	count := 0
	for _, u := range usersList {
		ws, err := backend.Workouts().List(u.Nickname)
		if err != nil {
			return 0, err
		}
		for i := range ws {
			likes, err := backend.Likes().GetLocal(u.Nickname, ws[i].ID)
			if err != nil {
				return 0, err
			}
			if likes != nil && likes.Likes > 0 {
				count++
			}
		}
	}
	return count, nil
}

func countFederatedLikes(backend storage.Backend) (int, error) {
	entries, err := listFederatedLikes(backend)
	if err != nil {
		return 0, err
	}
	return len(entries), nil
}

func countLikeActivities(backend storage.Backend, authors map[string]map[string]federation.AuthorMeta, inbox map[string]map[string][]workouts.Workout) (int, error) {
	seen := make(map[string]struct{})
	add := func(actor, objectID string) {
		if actor == "" || objectID == "" {
			return
		}
		seen[actor+"|"+objectID] = struct{}{}
	}

	entries, err := listLikeActivities(backend)
	if err != nil {
		return 0, err
	}
	for _, e := range entries {
		add(e.ActorNickname, e.ObjectID)
	}

	for viewer, byOwner := range inbox {
		for ownerKey, list := range byOwner {
			handle := federation.OwnerHandleFromKey(ownerKey)
			nickname := federation.OwnerNicknameFromKey(ownerKey)
			if meta, ok := authors[viewer][ownerKey]; ok {
				if meta.Handle != "" {
					handle = meta.Handle
				}
				if meta.Nickname != "" {
					nickname = meta.Nickname
				}
			}
			for i := range list {
				objectID := remoteWorkoutObjectID(handle, nickname, list[i].ID)
				if objectID == "" {
					continue
				}
				activityID, err := backend.Likes().GetLikeActivityID(viewer, objectID)
				if err != nil {
					return 0, err
				}
				if activityID != "" {
					add(viewer, objectID)
				}
			}
		}
	}
	return len(seen), nil
}

type federatedLikeEntry struct {
	ViewerNickname string
	OwnerHandle    string
	WorkoutID      string
	Likes          workouts.WorkoutLikes
}

type likeActivityEntry struct {
	ActorNickname string
	ObjectID      string
	ActivityID    string
}

func listFederatedLikes(backend storage.Backend) ([]federatedLikeEntry, error) {
	switch b := backend.(type) {
	case *file.Backend:
		raw, err := b.Likes().(*file.WorkoutLikesStore).ListAllFederated()
		if err != nil {
			return nil, err
		}
		out := make([]federatedLikeEntry, len(raw))
		for i, e := range raw {
			out[i] = federatedLikeEntry{
				ViewerNickname: e.ViewerNickname,
				OwnerHandle:    e.OwnerHandle,
				WorkoutID:      e.WorkoutID,
				Likes:          e.Likes,
			}
		}
		return out, nil
	case *storebbolt.Backend:
		raw, err := b.Likes().(*storebbolt.WorkoutLikesStore).ListAllFederated()
		if err != nil {
			return nil, err
		}
		out := make([]federatedLikeEntry, len(raw))
		for i, e := range raw {
			out[i] = federatedLikeEntry{
				ViewerNickname: e.ViewerNickname,
				OwnerHandle:    e.OwnerHandle,
				WorkoutID:      e.WorkoutID,
				Likes:          e.Likes,
			}
		}
		return out, nil
	default:
		return nil, fmt.Errorf("unsupported backend type %T", backend)
	}
}

func listLikeActivities(backend storage.Backend) ([]likeActivityEntry, error) {
	switch b := backend.(type) {
	case *file.Backend:
		raw, err := b.Likes().(*file.WorkoutLikesStore).ListAllLikeActivities()
		if err != nil {
			return nil, err
		}
		out := make([]likeActivityEntry, len(raw))
		for i, e := range raw {
			out[i] = likeActivityEntry{
				ActorNickname: e.ActorNickname,
				ObjectID:      e.ObjectID,
				ActivityID:    e.ActivityID,
			}
		}
		return out, nil
	case *storebbolt.Backend:
		raw, err := b.Likes().(*storebbolt.WorkoutLikesStore).ListAllLikeActivities()
		if err != nil {
			return nil, err
		}
		out := make([]likeActivityEntry, len(raw))
		for i, e := range raw {
			out[i] = likeActivityEntry{
				ActorNickname: e.ActorNickname,
				ObjectID:      e.ObjectID,
				ActivityID:    e.ActivityID,
			}
		}
		return out, nil
	default:
		return nil, fmt.Errorf("unsupported backend type %T", backend)
	}
}

func remoteWorkoutObjectID(ownerHandle, ownerNickname, workoutID string) string {
	domain := ""
	if idx := strings.LastIndex(ownerHandle, "@"); idx >= 0 && idx < len(ownerHandle)-1 {
		domain = ownerHandle[idx+1:]
	}
	if domain == "" || ownerNickname == "" || workoutID == "" {
		return ""
	}
	return fmt.Sprintf("https://%s/users/%s/workouts/%s", domain, ownerNickname, workoutID)
}
