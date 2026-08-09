package v1

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/solargate/grom/internal/auth"
	"github.com/solargate/grom/internal/social"
)

type FollowerResponse struct {
	FollowerHandle    string `json:"follower_handle"`
	FollowerNickname  string `json:"follower_nickname"`
	FollowerName      string `json:"follower_name"`
	FollowerIsLocal   bool   `json:"follower_is_local"`
	FollowerHasAvatar bool   `json:"follower_has_avatar"`
	FollowerAvatarURL string `json:"follower_avatar_url,omitempty"`
}

func (a *App) toFollowerResponse(f social.Follower, viewerNickname string) FollowerResponse {
	if f.FollowerIsLocal {
		hasAvatar, avatarURL := a.localAvatarFieldsForUser(f.FollowerNickname)
		return FollowerResponse{
			FollowerHandle:    f.FollowerHandle,
			FollowerNickname:  f.FollowerNickname,
			FollowerName:      f.FollowerName,
			FollowerIsLocal:   f.FollowerIsLocal,
			FollowerHasAvatar: hasAvatar,
			FollowerAvatarURL: avatarURL,
		}
	}

	hasAvatar, avatarURL := a.remoteFollowerAvatarFields(viewerNickname, &f)
	return FollowerResponse{
		FollowerHandle:    f.FollowerHandle,
		FollowerNickname:  f.FollowerNickname,
		FollowerName:      f.FollowerName,
		FollowerIsLocal:   f.FollowerIsLocal,
		FollowerHasAvatar: hasAvatar,
		FollowerAvatarURL: avatarURL,
	}
}

func (a *App) currentUserID(ctx *gin.Context) (string, error) {
	userID, _ := ctx.Get(auth.ContextUserIDKey)
	id, ok := userID.(string)
	if !ok || id == "" {
		return "", errors.New("invalid token")
	}
	return id, nil
}

type FollowRequest struct {
	Handle string `json:"handle" binding:"required" example:"bob"`
}

type FollowResponse struct {
	ID              string `json:"id"`
	TargetHandle    string `json:"target_handle"`
	TargetNickname  string `json:"target_nickname"`
	TargetName      string `json:"target_name"`
	TargetIsLocal   bool   `json:"target_is_local"`
	TargetHasAvatar bool   `json:"target_has_avatar"`
	TargetAvatarURL string `json:"target_avatar_url,omitempty"`
	Status          string `json:"status"`
}

func (a *App) toFollowResponse(f *social.Follow, viewerNickname string) FollowResponse {
	if f.TargetIsLocal {
		hasAvatar, avatarURL := a.localAvatarFieldsForUser(f.TargetNickname)
		return FollowResponse{
			ID:              f.ID,
			TargetHandle:    f.TargetHandle,
			TargetNickname:  f.TargetNickname,
			TargetName:      f.TargetName,
			TargetIsLocal:   f.TargetIsLocal,
			TargetHasAvatar: hasAvatar,
			TargetAvatarURL: avatarURL,
			Status:          f.Status,
		}
	}

	hasAvatar, avatarURL := a.remoteFollowAvatarFields(viewerNickname, f)
	return FollowResponse{
		ID:              f.ID,
		TargetHandle:    f.TargetHandle,
		TargetNickname:  f.TargetNickname,
		TargetName:      f.TargetName,
		TargetIsLocal:   f.TargetIsLocal,
		TargetHasAvatar: hasAvatar,
		TargetAvatarURL: avatarURL,
		Status:          f.Status,
	}
}

// followUser godoc
// @Summary      Follow user
// @Description  Follow a local or remote user by nickname or handle
// @Tags         social
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        body  body  FollowRequest  true  "Follow target"
// @Success      201   {object}  FollowResponse
// @Failure      400   {object}  ErrorResponse  "Invalid follow target"
// @Failure      401   {object}  ErrorResponse  "Unauthorized"
// @Failure      404   {object}  ErrorResponse  "User not found"
// @Failure      409   {object}  ErrorResponse  "Already following"
// @Router       /social/follow [post]
func (a *App) followUser(ctx *gin.Context) {

	userID, err := a.currentUserID(ctx)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, ErrorResponse{Error: "invalid token"})
		return
	}

	var req FollowRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	follow, err := a.Social.Follow(userID, req.Handle)
	if err != nil {
		handleSocialError(ctx, err)
		return
	}

	viewer, err := a.Users.FindByID(userID)
	if err != nil {
		respondInternal(ctx, "user not found", err)
		return
	}
	a.cacheRemoteFollowAvatar(viewer.Nickname, follow)

	ctx.JSON(http.StatusCreated, a.toFollowResponse(follow, viewer.Nickname))
}

// unfollowUser godoc
// @Summary      Unfollow user
// @Description  Remove a follow relationship
// @Tags         social
// @Produce      json
// @Security     BearerAuth
// @Param        id  path  string  true  "Follow ID"
// @Success      204
// @Failure      401  {object}  ErrorResponse  "Unauthorized"
// @Failure      404  {object}  ErrorResponse  "Follow not found"
// @Router       /social/follow/{id} [delete]
func (a *App) unfollowUser(ctx *gin.Context) {

	userID, err := a.currentUserID(ctx)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, ErrorResponse{Error: "invalid token"})
		return
	}

	if err := a.Social.Unfollow(userID, ctx.Param("id")); err != nil {
		handleSocialError(ctx, err)
		return
	}

	ctx.Status(http.StatusNoContent)
}

// listFollowing godoc
// @Summary      List following
// @Description  Return users the authenticated user follows
// @Tags         social
// @Produce      json
// @Security     BearerAuth
// @Success      200  {array}  FollowResponse
// @Failure      401  {object}  ErrorResponse  "Unauthorized"
// @Router       /social/following [get]
func (a *App) listFollowing(ctx *gin.Context) {

	userID, err := a.currentUserID(ctx)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, ErrorResponse{Error: "invalid token"})
		return
	}

	follows, err := a.Social.ListFollowing(userID)
	if err != nil {
		respondInternal(ctx, "failed to list following", err)
		return
	}

	viewer, err := a.Users.FindByID(userID)
	if err != nil {
		respondInternal(ctx, "user not found", err)
		return
	}

	response := make([]FollowResponse, 0, len(follows))
	for i := range follows {
		response = append(response, a.toFollowResponse(&follows[i], viewer.Nickname))
	}
	ctx.JSON(http.StatusOK, response)
}

// listFollowers godoc
// @Summary      List followers
// @Description  Return users who follow the authenticated user
// @Tags         social
// @Produce      json
// @Security     BearerAuth
// @Success      200  {array}  FollowerResponse
// @Failure      401  {object}  ErrorResponse  "Unauthorized"
// @Router       /social/followers [get]
func (a *App) listFollowers(ctx *gin.Context) {

	userID, err := a.currentUserID(ctx)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, ErrorResponse{Error: "invalid token"})
		return
	}

	followers, err := a.Social.ListFollowers(userID)
	if err != nil {
		respondInternal(ctx, "failed to list followers", err)
		return
	}

	viewer, err := a.Users.FindByID(userID)
	if err != nil {
		respondInternal(ctx, "user not found", err)
		return
	}

	response := make([]FollowerResponse, 0, len(followers))
	for i := range followers {
		response = append(response, a.toFollowerResponse(followers[i], viewer.Nickname))
	}
	ctx.JSON(http.StatusOK, response)
}

func handleSocialError(ctx *gin.Context, err error) {
	switch {
	case errors.Is(err, social.ErrInvalidHandle):
		ctx.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
	case errors.Is(err, social.ErrCannotFollowSelf):
		ctx.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
	case errors.Is(err, social.ErrUserNotFound):
		ctx.JSON(http.StatusNotFound, ErrorResponse{Error: err.Error()})
	case errors.Is(err, social.ErrAlreadyFollowing):
		ctx.JSON(http.StatusConflict, ErrorResponse{Error: err.Error()})
	case errors.Is(err, social.ErrFollowNotFound):
		ctx.JSON(http.StatusNotFound, ErrorResponse{Error: err.Error()})
	case errors.Is(err, social.ErrRemoteNotReady):
		ctx.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
	default:
		respondInternal(ctx, "social operation failed", err)
	}
}
