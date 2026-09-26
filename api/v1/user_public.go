package v1

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/solargate/grom/internal/auth"
	"github.com/solargate/grom/internal/config"
	"github.com/solargate/grom/internal/federation"
	"github.com/solargate/grom/internal/social"
	"github.com/solargate/grom/internal/workouts"
)

type ViewerFollowResponse struct {
	ID     string `json:"id" example:"550e8400-e29b-41d4-a716-446655440000"`
	Status string `json:"status" example:"active"`
}

type UserPublicProfileResponse struct {
	Nickname     string                `json:"nickname" example:"bob"`
	Name         string                `json:"name" example:"Bob"`
	Handle       string                `json:"handle" example:"bob@grom.example"`
	IsLocal      bool                  `json:"is_local" example:"true"`
	HasAvatar    bool                  `json:"has_avatar" example:"true"`
	AvatarURL    string                `json:"avatar_url,omitempty" example:"/api/v1/users/bob/avatar"`
	ViewerFollow *ViewerFollowResponse `json:"viewer_follow,omitempty"`
}

func (a *App) viewerFollowResponse(viewerID, targetHandle string) *ViewerFollowResponse {
	follow, err := a.Social.FindFollowToHandle(viewerID, targetHandle)
	if err != nil || follow == nil {
		return nil
	}
	return &ViewerFollowResponse{ID: follow.ID, Status: follow.Status}
}

// getUserPublic godoc
// @Summary      Get public user profile
// @Description  Return public identity for a local or remote user by nickname or handle. Includes whether the viewer follows them.
// @Tags         users
// @Produce      json
// @Security     BearerAuth
// @Param        handle  path  string  true  "Nickname or handle (URL-encoded)"
// @Success      200  {object}  UserPublicProfileResponse
// @Failure      400  {object}  ErrorResponse  "Invalid handle"
// @Failure      401  {object}  ErrorResponse  "Unauthorized"
// @Failure      404  {object}  ErrorResponse  "User not found"
// @Router       /users/{handle} [get]
func (a *App) getUserPublic(ctx *gin.Context) {
	viewerID, err := a.currentUserID(ctx)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, ErrorResponse{Error: "invalid token"})
		return
	}

	raw := strings.TrimSpace(ctx.Param("handle"))
	parsed, err := a.Social.ParseHandle(raw)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	if parsed.IsLocal {
		user, err := a.Users.FindByNickname(parsed.Nickname)
		if err != nil {
			ctx.JSON(http.StatusNotFound, ErrorResponse{Error: "user not found"})
			return
		}
		hasAvatar, avatarURL := a.localAvatarFieldsForUser(user.Nickname)
		ctx.JSON(http.StatusOK, UserPublicProfileResponse{
			Nickname:     user.Nickname,
			Name:         user.Name,
			Handle:       a.Social.LocalHandle(user.Nickname),
			IsLocal:      true,
			HasAvatar:    hasAvatar,
			AvatarURL:    avatarURL,
			ViewerFollow: a.viewerFollowResponse(viewerID, a.Social.LocalHandle(user.Nickname)),
		})
		return
	}

	if !config.Cfg.Federation.Enabled {
		ctx.JSON(http.StatusBadRequest, ErrorResponse{Error: social.ErrRemoteNotReady.Error()})
		return
	}
	results, err := a.Social.SearchLocal(parsed.Handle, "")
	if err != nil {
		handleSocialError(ctx, err)
		return
	}
	if len(results) == 0 {
		ctx.JSON(http.StatusNotFound, ErrorResponse{Error: "user not found"})
		return
	}
	r := results[0]
	ctx.JSON(http.StatusOK, UserPublicProfileResponse{
		Nickname:     r.Nickname,
		Name:         r.Name,
		Handle:       r.Handle,
		IsLocal:      false,
		HasAvatar:    r.HasAvatar,
		AvatarURL:    r.AvatarURL,
		ViewerFollow: a.viewerFollowResponse(viewerID, r.Handle),
	})
}

// listUserFollowing godoc
// @Summary      List a user's following
// @Description  Return users followed by the given local user. Remote targets return an empty list or a best-effort ActivityPub following collection when federation is enabled.
// @Tags         users
// @Produce      json
// @Security     BearerAuth
// @Param        handle  path  string  true  "Nickname or handle (URL-encoded)"
// @Success      200  {array}  FollowResponse
// @Failure      400  {object}  ErrorResponse
// @Failure      401  {object}  ErrorResponse
// @Failure      404  {object}  ErrorResponse
// @Router       /users/{handle}/following [get]
func (a *App) listUserFollowing(ctx *gin.Context) {
	viewerID, err := a.currentUserID(ctx)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, ErrorResponse{Error: "invalid token"})
		return
	}
	viewer, err := a.Users.FindByID(viewerID)
	if err != nil {
		respondInternal(ctx, "user not found", err)
		return
	}

	raw := strings.TrimSpace(ctx.Param("handle"))
	parsed, err := a.Social.ParseHandle(raw)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	if parsed.IsLocal {
		user, err := a.Users.FindByNickname(parsed.Nickname)
		if err != nil {
			ctx.JSON(http.StatusNotFound, ErrorResponse{Error: "user not found"})
			return
		}
		follows, err := a.Social.ListFollowingForUserID(user.ID)
		if err != nil {
			respondInternal(ctx, "failed to list following", err)
			return
		}
		response := make([]FollowResponse, 0, len(follows))
		for i := range follows {
			response = append(response, a.toFollowResponse(&follows[i], viewer.Nickname))
		}
		ctx.JSON(http.StatusOK, response)
		return
	}

	items, err := a.fetchRemoteFollowing(parsed)
	if err != nil {
		ctx.JSON(http.StatusBadGateway, ErrorResponse{Error: "failed to load following"})
		return
	}
	ctx.JSON(http.StatusOK, items)
}

// listUserFollowers godoc
// @Summary      List a user's followers
// @Description  Return followers of the given local user. Remote targets use a best-effort ActivityPub followers collection when federation is enabled.
// @Tags         users
// @Produce      json
// @Security     BearerAuth
// @Param        handle  path  string  true  "Nickname or handle (URL-encoded)"
// @Success      200  {array}  FollowerResponse
// @Failure      400  {object}  ErrorResponse
// @Failure      401  {object}  ErrorResponse
// @Failure      404  {object}  ErrorResponse
// @Router       /users/{handle}/followers [get]
func (a *App) listUserFollowers(ctx *gin.Context) {
	viewerID, err := a.currentUserID(ctx)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, ErrorResponse{Error: "invalid token"})
		return
	}
	viewer, err := a.Users.FindByID(viewerID)
	if err != nil {
		respondInternal(ctx, "user not found", err)
		return
	}

	raw := strings.TrimSpace(ctx.Param("handle"))
	parsed, err := a.Social.ParseHandle(raw)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	if parsed.IsLocal {
		user, err := a.Users.FindByNickname(parsed.Nickname)
		if err != nil {
			ctx.JSON(http.StatusNotFound, ErrorResponse{Error: "user not found"})
			return
		}
		followers, err := a.Social.ListFollowers(user.ID)
		if err != nil {
			respondInternal(ctx, "failed to list followers", err)
			return
		}
		response := make([]FollowerResponse, 0, len(followers))
		for i := range followers {
			response = append(response, a.toFollowerResponse(followers[i], viewer.Nickname))
		}
		ctx.JSON(http.StatusOK, response)
		return
	}

	items, err := a.fetchRemoteFollowers(parsed)
	if err != nil {
		ctx.JSON(http.StatusBadGateway, ErrorResponse{Error: "failed to load followers"})
		return
	}
	ctx.JSON(http.StatusOK, items)
}

// listUserWorkouts godoc
// @Summary      List a user's workouts
// @Description  Return a cursor page of workouts for the given user. Local users are visible to any authenticated JWT. Remote Grom users are loaded from their public outbox (no inbox write). Non-Grom remotes yield an empty page.
// @Tags         users
// @Produce      json
// @Security     BearerAuth
// @Param        handle  path   string  true   "Nickname or handle (URL-encoded)"
// @Param        limit   query  int     false  "page size (default 20, max 100)"
// @Param        cursor  query  string  false  "opaque cursor from previous page next_cursor"
// @Success      200  {object}  WorkoutListResponse
// @Failure      400  {object}  ErrorResponse
// @Failure      401  {object}  ErrorResponse
// @Failure      404  {object}  ErrorResponse
// @Router       /users/{handle}/workouts [get]
func (a *App) listUserWorkouts(ctx *gin.Context) {
	if auth.IsPAT(ctx) {
		ctx.JSON(http.StatusForbidden, ErrorResponse{Error: "personal access tokens cannot list other users' workouts"})
		return
	}
	viewerNickname, err := a.currentUserNickname(ctx)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, ErrorResponse{Error: "user not found"})
		return
	}

	raw := strings.TrimSpace(ctx.Param("handle"))
	parsed, err := a.Social.ParseHandle(raw)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	limit := workouts.DefaultPageLimit
	if rawLimit := strings.TrimSpace(ctx.Query("limit")); rawLimit != "" {
		parsedLimit, err := strconv.Atoi(rawLimit)
		if err != nil || parsedLimit < 1 {
			ctx.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid limit"})
			return
		}
		limit = workouts.ClampLimit(parsedLimit)
	}

	var cursor *workouts.Cursor
	if rawCursor := strings.TrimSpace(ctx.Query("cursor")); rawCursor != "" {
		cursor, err = workouts.DecodeCursor(rawCursor)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid cursor"})
			return
		}
	}

	if parsed.IsLocal {
		user, err := a.Users.FindByNickname(parsed.Nickname)
		if err != nil {
			ctx.JSON(http.StatusNotFound, ErrorResponse{Error: "user not found"})
			return
		}
		feedSvc := a.newFeedService()
		page, err := feedSvc.ListOwnPage(user.Nickname, user.Name, cursor, limit, nil)
		if err != nil {
			respondInternal(ctx, "failed to list workouts", err)
			return
		}
		items := make([]WorkoutResponse, 0, len(page.Items))
		for i := range page.Items {
			resp := toFeedWorkoutResponse(&page.Items[i])
			a.applyLikesSummaryToLocalWorkout(viewerNickname, user.Nickname, &page.Items[i].Workout, &resp)
			a.applyCommentsSummaryToLocalWorkout(viewerNickname, user.Nickname, &page.Items[i].Workout, &resp)
			resp.ObjectID = federation.WorkoutObjectURL(user.Nickname, page.Items[i].ID)
			items = append(items, resp)
		}
		ctx.JSON(http.StatusOK, WorkoutListResponse{
			Items:      items,
			NextCursor: page.NextCursor,
			HasMore:    page.HasMore,
		})
		return
	}

	page, err := a.listRemoteUserWorkouts(viewerNickname, parsed, cursor, limit)
	if err != nil {
		if errors.Is(err, social.ErrRemoteNotReady) || errors.Is(err, social.ErrUserNotFound) {
			handleSocialError(ctx, err)
			return
		}
		ctx.JSON(http.StatusBadGateway, ErrorResponse{Error: "failed to load workouts"})
		return
	}
	ctx.JSON(http.StatusOK, page)
}
