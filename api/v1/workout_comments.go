package v1

import (
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/solargate/grom/internal/config"
	"github.com/solargate/grom/internal/workouts"
)

type CreateWorkoutCommentRequest struct {
	Text string `json:"text" binding:"required" example:"Great workout!"`
}

type WorkoutCommentUserResponse struct {
	Handle    string `json:"handle" example:"alice@grom.example"`
	Nickname  string `json:"nickname" example:"alice"`
	Name      string `json:"name" example:"Alice"`
	IsLocal   bool   `json:"is_local" example:"true"`
	HasAvatar bool   `json:"has_avatar" example:"true"`
	AvatarURL string `json:"avatar_url,omitempty" example:"/api/v1/users/alice/avatar"`
}

type WorkoutCommentResponse struct {
	ID        string                     `json:"id" example:"550e8400-e29b-41d4-a716-446655440000"`
	User      WorkoutCommentUserResponse `json:"user"`
	Datetime  string                     `json:"datetime" example:"2026-08-06T12:00:00Z"`
	Text      string                     `json:"text" example:"Great workout!"`
	CanDelete bool                       `json:"can_delete" example:"true"`
}

type WorkoutCommentsResponse struct {
	Count    int                      `json:"count" example:"2"`
	Comments []WorkoutCommentResponse `json:"comments"`
}

type WorkoutCommentCreateResponse struct {
	Count   int                     `json:"count" example:"2"`
	Comment WorkoutCommentResponse  `json:"comment"`
}

type WorkoutCommentDeleteResponse struct {
	Count int `json:"count" example:"1"`
}

// getWorkoutComments godoc
// @Summary      List workout comments
// @Description  Return comments on a workout ordered by datetime ascending. Use owner query for followed users' workouts.
// @Tags         workouts
// @Produce      json
// @Security     BearerAuth
// @Param        id     path   string  true   "Workout ID"
// @Param        owner  query  string  false  "Workout owner nickname (required for followed users' workouts)"
// @Success      200  {object}  WorkoutCommentsResponse
// @Failure      401  {object}  ErrorResponse
// @Failure      404  {object}  ErrorResponse
// @Failure      500  {object}  ErrorResponse
// @Router       /workouts/{id}/comments [get]
func (a *App) getWorkoutComments(ctx *gin.Context) {
	viewerNickname, err := a.currentUserNickname(ctx)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, ErrorResponse{Error: "user not found"})
		return
	}
	ownerNickname, workoutID, err := a.resolveWorkoutOwner(ctx, viewerNickname)
	if err != nil {
		if errors.Is(err, workouts.ErrWorkoutNotFound) {
			ctx.JSON(http.StatusNotFound, ErrorResponse{Error: "workout not found"})
			return
		}
		respondInternal(ctx, "failed to resolve workout", err)
		return
	}
	comments, item, err := a.loadWorkoutComments(viewerNickname, ownerNickname, workoutID)
	if err != nil {
		if errors.Is(err, workouts.ErrWorkoutNotFound) {
			ctx.JSON(http.StatusNotFound, ErrorResponse{Error: "workout not found"})
			return
		}
		respondInternal(ctx, "failed to load workout comments", err)
		return
	}
	ownerHandle := a.ownerHandleForLikeOwner(ownerNickname, item)
	viewerHandle := a.Social.LocalHandle(viewerNickname)
	resp := a.toWorkoutCommentsResponse(viewerNickname, viewerHandle, ownerNickname, ownerHandle, comments)
	ctx.JSON(http.StatusOK, resp)
}

// createWorkoutComment godoc
// @Summary      Add workout comment
// @Description  Comment on a workout (own or another user's, local or federated). Max 1000 characters; empty text rejected.
// @Tags         workouts
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id     path   string  true   "Workout ID"
// @Param        owner  query  string  false  "Workout owner nickname (required for followed users' workouts)"
// @Param        body   body   CreateWorkoutCommentRequest  true  "Comment text"
// @Success      200  {object}  WorkoutCommentCreateResponse
// @Failure      400  {object}  ErrorResponse
// @Failure      401  {object}  ErrorResponse
// @Failure      404  {object}  ErrorResponse
// @Failure      500  {object}  ErrorResponse
// @Router       /workouts/{id}/comments [post]
func (a *App) createWorkoutComment(ctx *gin.Context) {
	viewerNickname, err := a.currentUserNickname(ctx)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, ErrorResponse{Error: "user not found"})
		return
	}
	var req CreateWorkoutCommentRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid request body"})
		return
	}
	if err := workouts.ValidateCommentText(req.Text); err != nil {
		ctx.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}
	ownerNickname, workoutID, err := a.resolveWorkoutOwner(ctx, viewerNickname)
	if err != nil {
		if errors.Is(err, workouts.ErrWorkoutNotFound) {
			ctx.JSON(http.StatusNotFound, ErrorResponse{Error: "workout not found"})
			return
		}
		respondInternal(ctx, "failed to resolve workout", err)
		return
	}
	actor, err := a.currentLikeActor(ctx)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, ErrorResponse{Error: "user not found"})
		return
	}

	comments, item, err := a.loadWorkoutComments(viewerNickname, ownerNickname, workoutID)
	if err != nil {
		if errors.Is(err, workouts.ErrWorkoutNotFound) {
			ctx.JSON(http.StatusNotFound, ErrorResponse{Error: "workout not found"})
			return
		}
		respondInternal(ctx, "failed to load workout comments", err)
		return
	}

	commentID := uuid.NewString()
	noteID := commentNoteURL(viewerNickname, commentID)
	comment := workouts.WorkoutComment{
		ID:       commentID,
		User:     actor,
		Datetime: time.Now().UTC(),
		Text:     strings.TrimSpace(req.Text),
		NoteID:   noteID,
	}

	isLocal := item == nil || item.Author.IsLocal
	updated := workouts.AddWorkoutComment(comments, comment)

	if isLocal {
		if err := a.Comments.PutLocal(ownerNickname, workoutID, &updated); err != nil {
			respondInternal(ctx, "failed to store workout comments", err)
			return
		}
		a.publishWorkoutCommentsUpdate(ownerNickname, workoutID)
	} else {
		ownerHandle := item.Author.Handle
		objectID := remoteWorkoutObjectID(ownerHandle, item.Author.Nickname, workoutID)
		if a.federationDelivery == nil || !config.Cfg.Federation.Enabled {
			ctx.JSON(http.StatusBadRequest, ErrorResponse{Error: "federation is not enabled"})
			return
		}
		activityID, deliverErr := a.federationDelivery.DeliverWorkoutComment(viewerNickname, ownerHandle, objectID, noteID, comment.Text, comment.Datetime)
		if deliverErr != nil {
			respondInternal(ctx, "failed to deliver comment", deliverErr)
			return
		}
		if err := a.Comments.PutCommentActivityID(viewerNickname, noteID, activityID); err != nil {
			respondInternal(ctx, "failed to store comment activity", err)
			return
		}
		if err := a.Comments.PutFederated(viewerNickname, ownerHandle, workoutID, &updated); err != nil {
			respondInternal(ctx, "failed to update federated comment cache", err)
			return
		}
	}

	ownerHandle := a.ownerHandleForLikeOwner(ownerNickname, item)
	viewerHandle := a.Social.LocalHandle(viewerNickname)
	created := workouts.FindCommentByID(&updated, commentID)
	respComment := a.toWorkoutCommentResponse(viewerNickname, viewerHandle, ownerNickname, ownerHandle, *created)
	ctx.JSON(http.StatusOK, WorkoutCommentCreateResponse{
		Count:   updated.CommentsNum,
		Comment: respComment,
	})
}

// deleteWorkoutComment godoc
// @Summary      Delete workout comment
// @Description  Delete a comment. Allowed for the comment author or the workout owner.
// @Tags         workouts
// @Produce      json
// @Security     BearerAuth
// @Param        id         path   string  true   "Workout ID"
// @Param        commentId  path   string  true   "Comment ID"
// @Param        owner      query  string  false  "Workout owner nickname (required for followed users' workouts)"
// @Success      200  {object}  WorkoutCommentDeleteResponse
// @Failure      401  {object}  ErrorResponse
// @Failure      403  {object}  ErrorResponse
// @Failure      404  {object}  ErrorResponse
// @Failure      500  {object}  ErrorResponse
// @Router       /workouts/{id}/comments/{commentId} [delete]
func (a *App) deleteWorkoutComment(ctx *gin.Context) {
	viewerNickname, err := a.currentUserNickname(ctx)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, ErrorResponse{Error: "user not found"})
		return
	}
	commentID := ctx.Param("commentId")
	if commentID == "" {
		ctx.JSON(http.StatusBadRequest, ErrorResponse{Error: "comment id required"})
		return
	}
	ownerNickname, workoutID, err := a.resolveWorkoutOwner(ctx, viewerNickname)
	if err != nil {
		if errors.Is(err, workouts.ErrWorkoutNotFound) {
			ctx.JSON(http.StatusNotFound, ErrorResponse{Error: "workout not found"})
			return
		}
		respondInternal(ctx, "failed to resolve workout", err)
		return
	}

	comments, item, err := a.loadWorkoutComments(viewerNickname, ownerNickname, workoutID)
	if err != nil {
		if errors.Is(err, workouts.ErrWorkoutNotFound) {
			ctx.JSON(http.StatusNotFound, ErrorResponse{Error: "workout not found"})
			return
		}
		respondInternal(ctx, "failed to load workout comments", err)
		return
	}
	comment := workouts.FindCommentByID(comments, commentID)
	if comment == nil {
		ctx.JSON(http.StatusNotFound, ErrorResponse{Error: workouts.ErrCommentNotFound.Error()})
		return
	}

	viewerHandle := a.Social.LocalHandle(viewerNickname)
	ownerHandle := a.ownerHandleForLikeOwner(ownerNickname, item)
	if !workouts.CanDeleteComment(viewerHandle, ownerNickname, comment, ownerHandle) {
		ctx.JSON(http.StatusForbidden, ErrorResponse{Error: workouts.ErrCannotDeleteComment.Error()})
		return
	}

	isLocal := item == nil || item.Author.IsLocal
	updated := workouts.RemoveWorkoutCommentByID(comments, commentID)

	if isLocal {
		if err := a.Comments.PutLocal(ownerNickname, workoutID, &updated); err != nil {
			respondInternal(ctx, "failed to store workout comments", err)
			return
		}
		// Notify remote author when owner (or anyone) deletes a federated author's comment.
		if comment.NoteID != "" && !comment.User.IsLocal && a.federationDelivery != nil && config.Cfg.Federation.Enabled {
			inReplyTo := workoutObjectURLForOwner(ownerNickname, workoutID)
			_ = a.federationDelivery.DeliverWorkoutCommentDeleteWithReply(viewerNickname, comment.User.Handle, comment.NoteID, inReplyTo)
		}
		if comment.NoteID != "" && comment.User.IsLocal {
			_ = a.Comments.DeleteCommentActivityID(comment.User.Nickname, comment.NoteID)
		}
		a.publishWorkoutCommentsUpdate(ownerNickname, workoutID)
	} else {
		if a.federationDelivery == nil || !config.Cfg.Federation.Enabled {
			ctx.JSON(http.StatusBadRequest, ErrorResponse{Error: "federation is not enabled"})
			return
		}
		noteID := comment.NoteID
		if noteID == "" {
			respondInternal(ctx, "comment missing note id", errors.New("missing note id"))
			return
		}
		objectID := remoteWorkoutObjectID(item.Author.Handle, item.Author.Nickname, workoutID)
		if deliverErr := a.federationDelivery.DeliverWorkoutCommentDeleteWithReply(viewerNickname, item.Author.Handle, noteID, objectID); deliverErr != nil {
			respondInternal(ctx, "failed to deliver comment delete", deliverErr)
			return
		}
		_ = a.Comments.DeleteCommentActivityID(viewerNickname, noteID)
		if err := a.Comments.PutFederated(viewerNickname, item.Author.Handle, workoutID, &updated); err != nil {
			respondInternal(ctx, "failed to update federated comment cache", err)
			return
		}
	}

	ctx.JSON(http.StatusOK, WorkoutCommentDeleteResponse{Count: updated.CommentsNum})
}

func commentNoteURL(authorNickname, commentID string) string {
	domain := config.Cfg.Federation.Domain
	if domain == "" {
		domain = "localhost"
	}
	return "https://" + domain + "/users/" + authorNickname + "/notes/" + commentID
}

func workoutObjectURLForOwner(ownerNickname, workoutID string) string {
	domain := config.Cfg.Federation.Domain
	if domain == "" {
		domain = "localhost"
	}
	return "https://" + domain + "/users/" + ownerNickname + "/workouts/" + workoutID
}

func (a *App) loadWorkoutComments(viewerNickname, ownerNickname, workoutID string) (*workouts.WorkoutComments, *workouts.FeedWorkout, error) {
	if comments, err := a.Comments.GetLocal(ownerNickname, workoutID); err == nil {
		if _, localErr := a.Workouts.Get(ownerNickname, workoutID); localErr == nil {
			return comments, nil, nil
		}
	}
	if a.Federation.Inbox() == nil {
		return nil, nil, workouts.ErrWorkoutNotFound
	}
	item, err := a.Federation.Inbox().Get(viewerNickname, ownerNickname, workoutID)
	if err != nil {
		return nil, nil, err
	}
	cached, err := a.Comments.GetFederated(viewerNickname, item.Author.Handle, workoutID)
	if err != nil {
		return nil, nil, err
	}
	if cached.CommentsNum > 0 || len(cached.Comments) > 0 {
		return cached, item, nil
	}
	snapshot := workouts.NormalizeWorkoutComments(&workouts.WorkoutComments{
		CommentsNum: item.Workout.CommentsCount,
		Comments:    item.Workout.Comments,
	})
	return &snapshot, item, nil
}

func (a *App) toWorkoutCommentsResponse(viewerNickname, viewerHandle, ownerNickname, ownerHandle string, comments *workouts.WorkoutComments) WorkoutCommentsResponse {
	norm := workouts.NormalizeWorkoutComments(comments)
	out := make([]WorkoutCommentResponse, 0, len(norm.Comments))
	for _, c := range norm.Comments {
		out = append(out, a.toWorkoutCommentResponse(viewerNickname, viewerHandle, ownerNickname, ownerHandle, c))
	}
	return WorkoutCommentsResponse{Count: norm.CommentsNum, Comments: out}
}

func (a *App) toWorkoutCommentResponse(viewerNickname, viewerHandle, ownerNickname, ownerHandle string, c workouts.WorkoutComment) WorkoutCommentResponse {
	isLocal := a.likeUserIsLocal(c.User)
	hasAvatar, avatarURL := a.likeAvatarFields(viewerNickname, c.User)
	if avatarURL == "" && c.User.AvatarURL != "" {
		hasAvatar = true
		avatarURL = c.User.AvatarURL
	}
	return WorkoutCommentResponse{
		ID: c.ID,
		User: WorkoutCommentUserResponse{
			Handle:    c.User.Handle,
			Nickname:  c.User.Nickname,
			Name:      c.User.Name,
			IsLocal:   isLocal,
			HasAvatar: hasAvatar,
			AvatarURL: avatarURL,
		},
		Datetime:  c.Datetime.UTC().Format(time.RFC3339),
		Text:      c.Text,
		CanDelete: workouts.CanDeleteComment(viewerHandle, ownerNickname, &c, ownerHandle),
	}
}

func (a *App) publishWorkoutCommentsUpdate(ownerNickname, workoutID string) {
	workout, err := a.Workouts.Get(ownerNickname, workoutID)
	if err != nil {
		return
	}
	comments, err := a.Comments.GetLocal(ownerNickname, workoutID)
	if err != nil {
		return
	}
	workout.CommentsCount = comments.CommentsNum
	workout.Comments = comments.Comments
	if likes, err := a.Likes.GetLocal(ownerNickname, workoutID); err == nil {
		workout.LikesCount = likes.Likes
		workout.LikedUsers = likes.Users
	}
	a.publishUpdatedWorkout(ownerNickname, workout)
}

func (a *App) commentsCountForWorkout(viewerNickname string, item *workouts.FeedWorkout) int {
	if item == nil {
		return 0
	}
	ownerNickname := item.Owner
	handle := a.ownerHandleForLikeOwner(ownerNickname, item)
	comments := workouts.NormalizeWorkoutComments(&workouts.WorkoutComments{
		CommentsNum: item.Workout.CommentsCount,
		Comments:    item.Workout.Comments,
	})
	if item.Author.IsLocal {
		if loaded, err := a.Comments.GetLocal(ownerNickname, item.ID); err == nil {
			comments = *loaded
		}
	} else if loaded, err := a.Comments.GetFederated(viewerNickname, handle, item.ID); err == nil && (loaded.CommentsNum > 0 || len(loaded.Comments) > 0) {
		comments = *loaded
	}
	return comments.CommentsNum
}

func (a *App) applyCommentsSummary(viewerNickname string, item *workouts.FeedWorkout, resp *WorkoutResponse) {
	if item == nil || resp == nil {
		return
	}
	resp.CommentsCount = a.commentsCountForWorkout(viewerNickname, item)
}

func (a *App) applyCommentsSummaryToLocalWorkout(viewerNickname, ownerNickname string, workout *workouts.Workout, resp *WorkoutResponse) {
	if workout == nil || resp == nil {
		return
	}
	item := &workouts.FeedWorkout{
		Workout: *workout,
		Owner:   ownerNickname,
		Author: workouts.FeedAuthor{
			Nickname: ownerNickname,
			Handle:   a.Social.LocalHandle(ownerNickname),
			IsLocal:  true,
		},
	}
	a.applyCommentsSummary(viewerNickname, item, resp)
}
