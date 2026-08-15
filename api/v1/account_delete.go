package v1

import (
	"log/slog"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/solargate/grom/internal/auth"
	"github.com/solargate/grom/internal/config"
)

type deleteAccountRequest struct {
	Password string `json:"password" binding:"required" example:"secret123"`
}

// deleteMe godoc
// @Summary      Delete current account
// @Description  Permanently delete the authenticated user's account and all related data after password confirmation. JWT only (PAT rejected). Federation Delete Person is delivered best-effort before local wipe.
// @Tags         auth
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        body  body  deleteAccountRequest  true  "Current password"
// @Success      204
// @Failure      400  {object}  ErrorResponse  "Invalid request"
// @Failure      401  {object}  ErrorResponse  "Unauthorized"
// @Failure      403  {object}  ErrorResponse  "Invalid password"
// @Failure      500  {object}  ErrorResponse  "Internal server error"
// @Router       /auth/me [delete]
func (a *App) deleteMe(ctx *gin.Context) {
	userID, _ := ctx.Get(auth.ContextUserIDKey)
	id, ok := userID.(string)
	if !ok || id == "" {
		ctx.JSON(http.StatusUnauthorized, ErrorResponse{Error: "invalid token"})
		return
	}

	var req deleteAccountRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}
	if strings.TrimSpace(req.Password) == "" {
		ctx.JSON(http.StatusBadRequest, ErrorResponse{Error: "password is required"})
		return
	}

	user, err := a.Users.FindByID(id)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, ErrorResponse{Error: "user not found"})
		return
	}
	if !auth.CheckPassword(user.PasswordHash, req.Password) {
		slog.Warn("account_delete_failed", "reason", "bad_password", "user_id", user.ID)
		ctx.JSON(http.StatusForbidden, ErrorResponse{Error: "invalid password"})
		return
	}

	localHandle := a.Social.LocalHandle(user.Nickname)
	a.deliverAccountDelete(user.Nickname, id)

	if err := a.Backend.PurgeUser(user.ID, user.Nickname, localHandle); err != nil {
		respondInternal(ctx, "failed to delete account", err)
		return
	}

	slog.Info("account deleted", "user", user.Nickname, "user_id", user.ID)
	ctx.Status(http.StatusNoContent)
}

func (a *App) deliverAccountDelete(nickname, userID string) {
	if a.federationDelivery == nil || !config.Cfg.Federation.Enabled {
		return
	}
	inboxes := a.collectAccountDeleteInboxes(nickname, userID)
	if len(inboxes) == 0 {
		return
	}
	a.federationDelivery.DeliverActorDelete(nickname, inboxes)
}

func (a *App) collectAccountDeleteInboxes(nickname, userID string) []string {
	seen := map[string]struct{}{}
	var inboxes []string
	add := func(inbox string) {
		inbox = strings.TrimSpace(inbox)
		if inbox == "" {
			return
		}
		if _, ok := seen[inbox]; ok {
			return
		}
		seen[inbox] = struct{}{}
		inboxes = append(inboxes, inbox)
	}

	if a.Federation != nil {
		listed, err := a.Federation.Followers().ListInboxes(nickname)
		if err != nil {
			slog.Error("federation list inboxes for account delete failed",
				"user", nickname,
				"err", err,
			)
		} else {
			for _, inbox := range listed {
				add(inbox)
			}
		}
	}

	if a.Social != nil {
		follows, err := a.Social.ListFollowing(userID)
		if err != nil {
			slog.Error("list following for account delete failed",
				"user", nickname,
				"err", err,
			)
		} else {
			for _, f := range follows {
				if f.TargetIsLocal {
					continue
				}
				if f.TargetActorURI != "" {
					add(strings.TrimSuffix(f.TargetActorURI, "/") + "/inbox")
				}
			}
		}
	}

	return inboxes
}
