package v1

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/solargate/grom/internal/workouts"
)

type WorkoutLikeStateResponse struct {
	Count     int  `json:"count" example:"5"`
	LikedByMe bool `json:"liked_by_me" example:"true"`
}

type WorkoutLikeUserResponse struct {
	Handle    string `json:"handle" example:"alice@grom.example"`
	Nickname  string `json:"nickname" example:"alice"`
	Name      string `json:"name" example:"Alice"`
	IsLocal   bool   `json:"is_local" example:"true"`
	HasAvatar bool   `json:"has_avatar" example:"true"`
	AvatarURL string `json:"avatar_url,omitempty" example:"/api/v1/users/alice/avatar"`
}

type WorkoutLikesResponse struct {
	Count int                       `json:"count" example:"5"`
	Users []WorkoutLikeUserResponse `json:"users"`
}

// getWorkoutLikes godoc
// @Summary      List workout likes
// @Description  Return users who liked a workout (avatar, name, handle). Use owner query for followed users' workouts (same as get workout).
// @Tags         workouts
// @Produce      json
// @Security     BearerAuth
// @Param        id     path   string  true   "Workout ID"
// @Param        owner  query  string  false  "Workout owner nickname (required for followed users' workouts)"
// @Success      200  {object}  WorkoutLikesResponse
// @Failure      401  {object}  ErrorResponse
// @Failure      404  {object}  ErrorResponse
// @Failure      500  {object}  ErrorResponse
// @Router       /workouts/{id}/likes [get]
func (a *App) getWorkoutLikes(ctx *gin.Context) {
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
	likes, item, err := a.loadWorkoutLikes(viewerNickname, ownerNickname, workoutID)
	if err != nil {
		if errors.Is(err, workouts.ErrWorkoutNotFound) {
			ctx.JSON(http.StatusNotFound, ErrorResponse{Error: "workout not found"})
			return
		}
		respondInternal(ctx, "failed to load workout likes", err)
		return
	}
	ownerHandle := a.ownerHandleForLikeOwner(ownerNickname, item)
	users := a.toWorkoutLikeUserResponses(viewerNickname, ownerHandle, likes.Users)
	ctx.JSON(http.StatusOK, WorkoutLikesResponse{Count: likes.Likes, Users: users})
}

// likeWorkout godoc
// @Summary      Like workout
// @Description  Like another user's workout (local or federated). Cannot like your own workout. Idempotent if already liked.
// @Tags         workouts
// @Produce      json
// @Security     BearerAuth
// @Param        id     path   string  true   "Workout ID"
// @Param        owner  query  string  false  "Workout owner nickname (required for followed users' workouts)"
// @Success      200  {object}  WorkoutLikeStateResponse
// @Failure      400  {object}  ErrorResponse
// @Failure      401  {object}  ErrorResponse
// @Failure      404  {object}  ErrorResponse
// @Failure      500  {object}  ErrorResponse
// @Router       /workouts/{id}/likes [post]
func (a *App) likeWorkout(ctx *gin.Context) {
	a.mutateWorkoutLike(ctx, true)
}

// unlikeWorkout godoc
// @Summary      Unlike workout
// @Description  Remove the current user's like from a workout (local or federated). Idempotent if not liked.
// @Tags         workouts
// @Produce      json
// @Security     BearerAuth
// @Param        id     path   string  true   "Workout ID"
// @Param        owner  query  string  false  "Workout owner nickname (required for followed users' workouts)"
// @Success      200  {object}  WorkoutLikeStateResponse
// @Failure      400  {object}  ErrorResponse
// @Failure      401  {object}  ErrorResponse
// @Failure      404  {object}  ErrorResponse
// @Failure      500  {object}  ErrorResponse
// @Router       /workouts/{id}/likes [delete]
func (a *App) unlikeWorkout(ctx *gin.Context) {
	a.mutateWorkoutLike(ctx, false)
}

func (a *App) mutateWorkoutLike(ctx *gin.Context, add bool) {
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
	if ownerNickname == viewerNickname {
		ctx.JSON(http.StatusBadRequest, ErrorResponse{Error: workouts.ErrCannotLikeOwnWorkout.Error()})
		return
	}
	actor, err := a.currentLikeActor(ctx)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, ErrorResponse{Error: "user not found"})
		return
	}

	likes, item, err := a.loadWorkoutLikes(viewerNickname, ownerNickname, workoutID)
	if err != nil {
		if errors.Is(err, workouts.ErrWorkoutNotFound) {
			ctx.JSON(http.StatusNotFound, ErrorResponse{Error: "workout not found"})
			return
		}
		respondInternal(ctx, "failed to load workout likes", err)
		return
	}

	isLocal := item == nil || item.Author.IsLocal
	var updated workouts.WorkoutLikes
	if add {
		updated = workouts.AddWorkoutLikeUser(likes, actor)
	} else {
		updated = workouts.RemoveWorkoutLikeUser(likes, actor.Handle)
	}

	if isLocal {
		if err := a.Likes.PutLocal(ownerNickname, workoutID, &updated); err != nil {
			respondInternal(ctx, "failed to store workout likes", err)
			return
		}
		a.publishWorkoutLikesUpdate(ownerNickname, workoutID)
	} else {
		ownerHandle := item.Author.Handle
		objectID := remoteWorkoutObjectID(ownerHandle, item.Author.Nickname, workoutID)
		if add {
			activityID, deliverErr := a.federationDelivery.DeliverWorkoutLike(viewerNickname, ownerHandle, objectID)
			if deliverErr != nil {
				respondInternal(ctx, "failed to deliver like", deliverErr)
				return
			}
			if err := a.Likes.PutLikeActivityID(viewerNickname, objectID, activityID); err != nil {
				respondInternal(ctx, "failed to store like activity", err)
				return
			}
		} else {
			activityID, activityErr := a.Likes.GetLikeActivityID(viewerNickname, objectID)
			if activityErr != nil {
				respondInternal(ctx, "failed to read like activity", activityErr)
				return
			}
			if deliverErr := a.federationDelivery.DeliverWorkoutUndoLike(viewerNickname, ownerHandle, objectID, activityID); deliverErr != nil {
				respondInternal(ctx, "failed to deliver unlike", deliverErr)
				return
			}
			if err := a.Likes.DeleteLikeActivityID(viewerNickname, objectID); err != nil {
				respondInternal(ctx, "failed to delete like activity", err)
				return
			}
		}
		if err := a.Likes.PutFederated(viewerNickname, ownerHandle, workoutID, &updated); err != nil {
			respondInternal(ctx, "failed to update federated like cache", err)
			return
		}
	}

	ctx.JSON(http.StatusOK, WorkoutLikeStateResponse{
		Count:     updated.Likes,
		LikedByMe: add && workouts.LikesContainUser(&updated, actor.Handle),
	})
}

func (a *App) currentLikeActor(ctx *gin.Context) (workouts.WorkoutLikeUser, error) {
	userID, err := a.currentUserID(ctx)
	if err != nil {
		return workouts.WorkoutLikeUser{}, err
	}
	user, err := a.Users.FindByID(userID)
	if err != nil {
		return workouts.WorkoutLikeUser{}, err
	}
	return workouts.WorkoutLikeUser{
		Handle:   a.Social.LocalHandle(user.Nickname),
		Nickname: user.Nickname,
		Name:     user.Name,
		IsLocal:  true,
	}, nil
}

func (a *App) loadWorkoutLikes(viewerNickname, ownerNickname, workoutID string) (*workouts.WorkoutLikes, *workouts.FeedWorkout, error) {
	if likes, err := a.Likes.GetLocal(ownerNickname, workoutID); err == nil {
		if _, localErr := a.Workouts.Get(ownerNickname, workoutID); localErr == nil {
			return likes, nil, nil
		}
	}
	if a.Federation.Inbox() == nil {
		return nil, nil, workouts.ErrWorkoutNotFound
	}
	item, err := a.Federation.Inbox().Get(viewerNickname, ownerNickname, workoutID)
	if err != nil {
		return nil, nil, err
	}
	cached, err := a.Likes.GetFederated(viewerNickname, item.Author.Handle, workoutID)
	if err != nil {
		return nil, nil, err
	}
	if cached.Likes > 0 || len(cached.Users) > 0 {
		return cached, item, nil
	}
	snapshot := workouts.NormalizeWorkoutLikes(&workouts.WorkoutLikes{
		Likes: item.Workout.LikesCount,
		Users: item.Workout.LikedUsers,
	})
	return &snapshot, item, nil
}

func (a *App) ownerHandleForLikeOwner(ownerNickname string, item *workouts.FeedWorkout) string {
	if item != nil && item.Author.Handle != "" {
		return item.Author.Handle
	}
	return a.Social.LocalHandle(ownerNickname)
}

func (a *App) toWorkoutLikeUserResponses(viewerNickname, _ string, users []workouts.WorkoutLikeUser) []WorkoutLikeUserResponse {
	result := make([]WorkoutLikeUserResponse, 0, len(users))
	for _, user := range users {
		hasAvatar, avatarURL := a.likeAvatarFields(viewerNickname, user)
		if avatarURL == "" && user.AvatarURL != "" {
			hasAvatar = true
			avatarURL = user.AvatarURL
		}
		result = append(result, WorkoutLikeUserResponse{
			Handle:    user.Handle,
			Nickname:  user.Nickname,
			Name:      user.Name,
			IsLocal:   user.IsLocal,
			HasAvatar: hasAvatar,
			AvatarURL: avatarURL,
		})
	}
	return result
}

func (a *App) likeAvatarFields(viewerNickname string, user workouts.WorkoutLikeUser) (bool, string) {
	if user.IsLocal {
		return a.localAvatarFieldsForUser(user.Nickname)
	}
	if a.Federation.Inbox() != nil {
		if hasAvatar, avatarURL := a.Federation.Inbox().AuthorAvatarFields(viewerNickname, user.Handle); hasAvatar || avatarURL != "" {
			return hasAvatar, avatarURL
		}
	}
	if user.AvatarURL != "" {
		return true, user.AvatarURL
	}
	return false, ""
}

func remoteWorkoutObjectID(ownerHandle, ownerNickname, workoutID string) string {
	domain := ""
	if idx := strings.LastIndex(ownerHandle, "@"); idx >= 0 && idx < len(ownerHandle)-1 {
		domain = ownerHandle[idx+1:]
	}
	if domain == "" {
		return ""
	}
	return fmt.Sprintf("https://%s/users/%s/workouts/%s", domain, ownerNickname, workoutID)
}

func (a *App) publishWorkoutLikesUpdate(ownerNickname, workoutID string) {
	workout, err := a.Workouts.Get(ownerNickname, workoutID)
	if err != nil {
		return
	}
	likes, err := a.Likes.GetLocal(ownerNickname, workoutID)
	if err != nil {
		return
	}
	workout.LikesCount = likes.Likes
	workout.LikedUsers = likes.Users
	if comments, err := a.Comments.GetLocal(ownerNickname, workoutID); err == nil {
		workout.CommentsCount = comments.CommentsNum
		workout.Comments = comments.Comments
	}
	a.publishUpdatedWorkout(ownerNickname, workout)
}

func (a *App) likesSummaryForWorkout(viewerNickname string, item *workouts.FeedWorkout) (int, bool, bool) {
	ownerNickname := item.Owner
	canLike := ownerNickname != viewerNickname
	handle := a.ownerHandleForLikeOwner(ownerNickname, item)
	likes := workouts.NormalizeWorkoutLikes(&workouts.WorkoutLikes{
		Likes: item.Workout.LikesCount,
		Users: item.Workout.LikedUsers,
	})
	if item.Author.IsLocal {
		if loaded, err := a.Likes.GetLocal(ownerNickname, item.ID); err == nil {
			likes = *loaded
		}
	} else if loaded, err := a.Likes.GetFederated(viewerNickname, handle, item.ID); err == nil && (loaded.Likes > 0 || len(loaded.Users) > 0) {
		likes = *loaded
	}
	viewerHandle := a.Social.LocalHandle(viewerNickname)
	return likes.Likes, workouts.LikesContainUser(&likes, viewerHandle), canLike
}

func (a *App) applyLikesSummary(viewerNickname string, item *workouts.FeedWorkout, resp *WorkoutResponse) {
	if item == nil || resp == nil {
		return
	}
	resp.LikesCount, resp.LikedByMe, resp.CanLike = a.likesSummaryForWorkout(viewerNickname, item)
}

func (a *App) applyLikesSummaryToLocalWorkout(viewerNickname, ownerNickname string, workout *workouts.Workout, resp *WorkoutResponse) {
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
	a.applyLikesSummary(viewerNickname, item, resp)
}
