package migrate

import (
	"fmt"

	"github.com/solargate/grom/internal/storage"
	storebbolt "github.com/solargate/grom/internal/storage/bbolt"
	"github.com/solargate/grom/internal/storage/file"
	"github.com/solargate/grom/internal/workouts"
)

func copyLocalComments(src, dst storage.Backend, nickname string, w *workouts.Workout) error {
	if w == nil {
		return nil
	}
	comments, err := src.Comments().GetLocal(nickname, w.ID)
	if err != nil {
		return fmt.Errorf("read local comments: %w", err)
	}
	if comments == nil || comments.CommentsNum == 0 {
		return nil
	}
	if err := dst.Comments().PutLocal(nickname, w.ID, comments); err != nil {
		return fmt.Errorf("write local comments: %w", err)
	}
	return nil
}

func copyFederatedComments(src, dst storage.Backend) (int, error) {
	entries, err := listFederatedComments(src)
	if err != nil {
		return 0, err
	}
	for _, e := range entries {
		comments := e.Comments
		if err := dst.Comments().PutFederated(e.ViewerNickname, e.OwnerHandle, e.WorkoutID, &comments); err != nil {
			return 0, fmt.Errorf("write federated comments %s/%s/%s: %w", e.ViewerNickname, e.OwnerHandle, e.WorkoutID, err)
		}
	}
	return len(entries), nil
}

func copyCommentActivities(src, dst storage.Backend) (int, error) {
	entries, err := listCommentActivities(src)
	if err != nil {
		return 0, err
	}
	seen := make(map[string]struct{})
	copied := 0
	for _, e := range entries {
		if e.ActorNickname == "" || e.NoteID == "" || e.ActivityID == "" {
			continue
		}
		key := e.ActorNickname + "|" + e.NoteID
		if _, ok := seen[key]; ok {
			continue
		}
		if err := dst.Comments().PutCommentActivityID(e.ActorNickname, e.NoteID, e.ActivityID); err != nil {
			return copied, fmt.Errorf("write comment activity: %w", err)
		}
		seen[key] = struct{}{}
		copied++
	}
	return copied, nil
}

func countLocalComments(backend storage.Backend) (int, error) {
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
			comments, err := backend.Comments().GetLocal(u.Nickname, ws[i].ID)
			if err != nil {
				return 0, err
			}
			if comments != nil && comments.CommentsNum > 0 {
				count++
			}
		}
	}
	return count, nil
}

func countFederatedComments(backend storage.Backend) (int, error) {
	entries, err := listFederatedComments(backend)
	if err != nil {
		return 0, err
	}
	return len(entries), nil
}

func countCommentActivities(backend storage.Backend) (int, error) {
	entries, err := listCommentActivities(backend)
	if err != nil {
		return 0, err
	}
	return len(entries), nil
}

type federatedCommentEntry struct {
	ViewerNickname string
	OwnerHandle    string
	WorkoutID      string
	Comments       workouts.WorkoutComments
}

type commentActivityEntry struct {
	ActorNickname string
	NoteID        string
	ActivityID    string
}

func listFederatedComments(backend storage.Backend) ([]federatedCommentEntry, error) {
	switch b := backend.(type) {
	case *file.Backend:
		raw, err := b.Comments().(*file.WorkoutCommentsStore).ListAllFederated()
		if err != nil {
			return nil, err
		}
		out := make([]federatedCommentEntry, len(raw))
		for i, e := range raw {
			out[i] = federatedCommentEntry{
				ViewerNickname: e.ViewerNickname,
				OwnerHandle:    e.OwnerHandle,
				WorkoutID:      e.WorkoutID,
				Comments:       e.Comments,
			}
		}
		return out, nil
	case *storebbolt.Backend:
		raw, err := b.Comments().(*storebbolt.WorkoutCommentsStore).ListAllFederated()
		if err != nil {
			return nil, err
		}
		out := make([]federatedCommentEntry, len(raw))
		for i, e := range raw {
			out[i] = federatedCommentEntry{
				ViewerNickname: e.ViewerNickname,
				OwnerHandle:    e.OwnerHandle,
				WorkoutID:      e.WorkoutID,
				Comments:       e.Comments,
			}
		}
		return out, nil
	default:
		return nil, fmt.Errorf("unsupported backend type %T", backend)
	}
}

func listCommentActivities(backend storage.Backend) ([]commentActivityEntry, error) {
	switch b := backend.(type) {
	case *file.Backend:
		raw, err := b.Comments().(*file.WorkoutCommentsStore).ListAllCommentActivities()
		if err != nil {
			return nil, err
		}
		out := make([]commentActivityEntry, len(raw))
		for i, e := range raw {
			out[i] = commentActivityEntry{
				ActorNickname: e.ActorNickname,
				NoteID:        e.NoteID,
				ActivityID:    e.ActivityID,
			}
		}
		return out, nil
	case *storebbolt.Backend:
		raw, err := b.Comments().(*storebbolt.WorkoutCommentsStore).ListAllCommentActivities()
		if err != nil {
			return nil, err
		}
		out := make([]commentActivityEntry, len(raw))
		for i, e := range raw {
			out[i] = commentActivityEntry{
				ActorNickname: e.ActorNickname,
				NoteID:        e.NoteID,
				ActivityID:    e.ActivityID,
			}
		}
		return out, nil
	default:
		return nil, fmt.Errorf("unsupported backend type %T", backend)
	}
}
