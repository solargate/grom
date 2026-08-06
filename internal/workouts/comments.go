package workouts

import (
	"strings"
	"time"
	"unicode/utf8"
)

const MaxCommentTextLength = 1000

type WorkoutComment struct {
	ID       string           `yaml:"id" json:"id"`
	User     WorkoutLikeUser  `yaml:"user" json:"user"`
	Datetime time.Time        `yaml:"datetime" json:"datetime"`
	Text     string           `yaml:"text" json:"text"`
	NoteID   string           `yaml:"note_id,omitempty" json:"note_id,omitempty"`
}

type WorkoutComments struct {
	CommentsNum int              `yaml:"comments_num" json:"comments_num"`
	Comments    []WorkoutComment `yaml:"comments" json:"comments"`
}

type CommentsRepository interface {
	GetLocal(ownerNickname, workoutID string) (*WorkoutComments, error)
	PutLocal(ownerNickname, workoutID string, comments *WorkoutComments) error
	DeleteLocal(ownerNickname, workoutID string) error

	GetFederated(viewerNickname, ownerHandle, workoutID string) (*WorkoutComments, error)
	PutFederated(viewerNickname, ownerHandle, workoutID string, comments *WorkoutComments) error
	DeleteFederated(viewerNickname, ownerHandle, workoutID string) error

	GetCommentActivityID(actorNickname, noteID string) (string, error)
	PutCommentActivityID(actorNickname, noteID, activityID string) error
	DeleteCommentActivityID(actorNickname, noteID string) error
}

func NormalizeWorkoutComments(comments *WorkoutComments) WorkoutComments {
	if comments == nil {
		return WorkoutComments{Comments: []WorkoutComment{}}
	}
	out := make([]WorkoutComment, 0, len(comments.Comments))
	seen := make(map[string]struct{}, len(comments.Comments))
	for _, c := range comments.Comments {
		if c.ID == "" || c.User.Handle == "" {
			continue
		}
		if _, ok := seen[c.ID]; ok {
			continue
		}
		seen[c.ID] = struct{}{}
		c.Text = strings.TrimSpace(c.Text)
		out = append(out, c)
	}
	sortCommentsByDatetime(out)
	return WorkoutComments{
		CommentsNum: len(out),
		Comments:    out,
	}
}

func sortCommentsByDatetime(comments []WorkoutComment) {
	for i := 1; i < len(comments); i++ {
		j := i
		for j > 0 {
			if comments[j-1].Datetime.Before(comments[j].Datetime) ||
				(comments[j-1].Datetime.Equal(comments[j].Datetime) && comments[j-1].ID < comments[j].ID) {
				break
			}
			comments[j-1], comments[j] = comments[j], comments[j-1]
			j--
		}
	}
}

func ValidateCommentText(text string) error {
	trimmed := strings.TrimSpace(text)
	if trimmed == "" {
		return ErrEmptyComment
	}
	if utf8.RuneCountInString(trimmed) > MaxCommentTextLength {
		return ErrCommentTooLong
	}
	return nil
}

func FindCommentByID(comments *WorkoutComments, commentID string) *WorkoutComment {
	if comments == nil || commentID == "" {
		return nil
	}
	for i := range comments.Comments {
		if comments.Comments[i].ID == commentID {
			return &comments.Comments[i]
		}
	}
	return nil
}

func FindCommentByNoteID(comments *WorkoutComments, noteID string) *WorkoutComment {
	if comments == nil || noteID == "" {
		return nil
	}
	for i := range comments.Comments {
		if comments.Comments[i].NoteID == noteID {
			return &comments.Comments[i]
		}
	}
	return nil
}

func CanDeleteComment(viewerHandle, workoutOwnerNickname string, comment *WorkoutComment, localOwnerHandle string) bool {
	if comment == nil || viewerHandle == "" {
		return false
	}
	if comment.User.Handle == viewerHandle {
		return true
	}
	if workoutOwnerNickname != "" && localOwnerHandle != "" && viewerHandle == localOwnerHandle {
		return true
	}
	return false
}

func AddWorkoutComment(comments *WorkoutComments, comment WorkoutComment) WorkoutComments {
	norm := NormalizeWorkoutComments(comments)
	if comment.ID == "" || comment.User.Handle == "" {
		return norm
	}
	comment.Text = strings.TrimSpace(comment.Text)
	if comment.Text == "" {
		return norm
	}
	if FindCommentByID(&norm, comment.ID) != nil {
		return norm
	}
	norm.Comments = append(norm.Comments, comment)
	return NormalizeWorkoutComments(&norm)
}

func RemoveWorkoutCommentByID(comments *WorkoutComments, commentID string) WorkoutComments {
	norm := NormalizeWorkoutComments(comments)
	if commentID == "" {
		return norm
	}
	filtered := make([]WorkoutComment, 0, len(norm.Comments))
	for _, c := range norm.Comments {
		if c.ID == commentID {
			continue
		}
		filtered = append(filtered, c)
	}
	norm.Comments = filtered
	return NormalizeWorkoutComments(&norm)
}

func RemoveWorkoutCommentByNoteID(comments *WorkoutComments, noteID string) WorkoutComments {
	norm := NormalizeWorkoutComments(comments)
	if noteID == "" {
		return norm
	}
	filtered := make([]WorkoutComment, 0, len(norm.Comments))
	for _, c := range norm.Comments {
		if c.NoteID == noteID {
			continue
		}
		filtered = append(filtered, c)
	}
	norm.Comments = filtered
	return NormalizeWorkoutComments(&norm)
}
