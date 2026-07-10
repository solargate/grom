package v1

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/solargate/travka/internal/auth"
	"github.com/solargate/travka/internal/config"
	"github.com/solargate/travka/internal/federation"
	"github.com/solargate/travka/internal/social"
)

var socialService *social.Service

func initSocialService() error {
	if socialService != nil {
		return nil
	}
	if userStore == nil {
		if err := initUserStore(); err != nil {
			return err
		}
	}
	followStore, err := social.NewStore(config.Cfg.Data.ResolvedDir)
	if err != nil {
		return err
	}
	socialService = social.NewService(userStore, followStore)
	if err := initFollowersStore(); err != nil {
		return err
	}
	if config.Cfg.Federation.Enabled {
		delivery, err := federation.NewDelivery(userStore, socialService)
		if err != nil {
			return err
		}
		socialService.SetDelivery(delivery)
	}
	return nil
}

func initFollowersStore() error {
	if followersStore == nil {
		followersStore = federation.NewFollowersStore(config.Cfg.Data.ResolvedDir)
	}
	socialService.SetInboundFollowers(federation.NewInboundFollowersAdapter(followersStore))
	return nil
}

type FollowerResponse struct {
	FollowerHandle    string `json:"follower_handle"`
	FollowerNickname  string `json:"follower_nickname"`
	FollowerName      string `json:"follower_name"`
	FollowerIsLocal   bool   `json:"follower_is_local"`
	FollowerHasAvatar bool   `json:"follower_has_avatar"`
	FollowerAvatarURL string `json:"follower_avatar_url,omitempty"`
}

func toFollowerResponse(f social.Follower, viewerNickname string) FollowerResponse {
	if f.FollowerIsLocal {
		hasAvatar, avatarURL := localAvatarFieldsForUser(f.FollowerNickname)
		return FollowerResponse{
			FollowerHandle:    f.FollowerHandle,
			FollowerNickname:  f.FollowerNickname,
			FollowerName:      f.FollowerName,
			FollowerIsLocal:   f.FollowerIsLocal,
			FollowerHasAvatar: hasAvatar,
			FollowerAvatarURL: avatarURL,
		}
	}

	hasAvatar, avatarURL := remoteFollowerAvatarFields(viewerNickname, &f)
	return FollowerResponse{
		FollowerHandle:    f.FollowerHandle,
		FollowerNickname:  f.FollowerNickname,
		FollowerName:      f.FollowerName,
		FollowerIsLocal:   f.FollowerIsLocal,
		FollowerHasAvatar: hasAvatar,
		FollowerAvatarURL: avatarURL,
	}
}

func currentUserID(ctx *gin.Context) (string, error) {
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

func toFollowResponse(f *social.Follow, viewerNickname string) FollowResponse {
	if f.TargetIsLocal {
		hasAvatar, avatarURL := localAvatarFieldsForUser(f.TargetNickname)
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

	hasAvatar, avatarURL := remoteFollowAvatarFields(viewerNickname, f)
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
// @Failure      400   {object}  ErrorResponse
// @Failure      401   {object}  ErrorResponse
// @Failure      404   {object}  ErrorResponse
// @Failure      409   {object}  ErrorResponse
// @Router       /social/follow [post]
func followUser(ctx *gin.Context) {
	if err := initSocialService(); err != nil {
		ctx.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to init social service"})
		return
	}

	userID, err := currentUserID(ctx)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, ErrorResponse{Error: "invalid token"})
		return
	}

	var req FollowRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	follow, err := socialService.Follow(userID, req.Handle)
	if err != nil {
		handleSocialError(ctx, err)
		return
	}

	viewer, err := userStore.FindByID(userID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, ErrorResponse{Error: "user not found"})
		return
	}
	cacheRemoteFollowAvatar(viewer.Nickname, follow)

	ctx.JSON(http.StatusCreated, toFollowResponse(follow, viewer.Nickname))
}

// unfollowUser godoc
// @Summary      Unfollow user
// @Description  Remove a follow relationship
// @Tags         social
// @Produce      json
// @Security     BearerAuth
// @Param        id  path  string  true  "Follow ID"
// @Success      204
// @Failure      401  {object}  ErrorResponse
// @Failure      404  {object}  ErrorResponse
// @Router       /social/follow/{id} [delete]
func unfollowUser(ctx *gin.Context) {
	if err := initSocialService(); err != nil {
		ctx.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to init social service"})
		return
	}

	userID, err := currentUserID(ctx)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, ErrorResponse{Error: "invalid token"})
		return
	}

	if err := socialService.Unfollow(userID, ctx.Param("id")); err != nil {
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
// @Failure      401  {object}  ErrorResponse
// @Router       /social/following [get]
func listFollowing(ctx *gin.Context) {
	if err := initSocialService(); err != nil {
		ctx.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to init social service"})
		return
	}

	userID, err := currentUserID(ctx)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, ErrorResponse{Error: "invalid token"})
		return
	}

	follows, err := socialService.ListFollowing(userID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to list following"})
		return
	}

	viewer, err := userStore.FindByID(userID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, ErrorResponse{Error: "user not found"})
		return
	}

	response := make([]FollowResponse, 0, len(follows))
	for i := range follows {
		response = append(response, toFollowResponse(&follows[i], viewer.Nickname))
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
// @Failure      401  {object}  ErrorResponse
// @Router       /social/followers [get]
func listFollowers(ctx *gin.Context) {
	if err := initSocialService(); err != nil {
		ctx.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to init social service"})
		return
	}

	userID, err := currentUserID(ctx)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, ErrorResponse{Error: "invalid token"})
		return
	}

	followers, err := socialService.ListFollowers(userID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to list followers"})
		return
	}

	viewer, err := userStore.FindByID(userID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, ErrorResponse{Error: "user not found"})
		return
	}

	response := make([]FollowerResponse, 0, len(followers))
	for i := range followers {
		response = append(response, toFollowerResponse(followers[i], viewer.Nickname))
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
		ctx.JSON(http.StatusInternalServerError, ErrorResponse{Error: "social operation failed"})
	}
}
