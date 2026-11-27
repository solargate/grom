package v1

import (
	"errors"
	"log/slog"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/solargate/grom/internal/auth"
)

// deleteAllWorkouts godoc
// @Summary      Delete all of the current user's workouts
// @Description  Permanently delete the authenticated user's saved workouts and their attachments after password confirmation. JWT only (PAT rejected).
// @Tags         workouts
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        body  body  deleteAccountRequest  true  "Current password"
// @Success      204
// @Failure      400  {object}  ErrorResponse  "Invalid request"
// @Failure      401  {object}  ErrorResponse  "Unauthorized"
// @Failure      403  {object}  ErrorResponse  "Invalid password"
// @Failure      500  {object}  ErrorResponse  "Internal server error; some workouts may have been deleted"
// @Router       /workouts [delete]
func (a *App) deleteAllWorkouts(ctx *gin.Context) {
	userID, err := a.currentUserID(ctx)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, ErrorResponse{Error: "user not found"})
		return
	}

	var req deleteAccountRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid request"})
		return
	}
	if strings.TrimSpace(req.Password) == "" {
		ctx.JSON(http.StatusBadRequest, ErrorResponse{Error: "password is required"})
		return
	}
	unlock := a.lockUserDeletion(userID)
	defer unlock()
	user, err := a.Users.FindByID(userID)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, ErrorResponse{Error: "user not found"})
		return
	}
	if !auth.CheckPassword(user.PasswordHash, req.Password) {
		slog.Warn("workout history deletion rejected", "reason", "bad_password", "user_id", userID)
		ctx.JSON(http.StatusForbidden, ErrorResponse{Error: "invalid password"})
		return
	}

	deleted, deleteErr := a.Workouts.DeleteAll(user.Nickname)
	a.scheduleWorkoutDeleteDelivery(user.Nickname, deleted)
	// Even a partial deletion changes these derived values. Recalculate on an
	// empty history too, so a retry can repair a previous recalculation failure.
	profileErr := a.RefreshLastSportType(user.Nickname, userID)
	equipmentErr := a.EquipmentDistance.RecalculateAll(user.Nickname)
	if err := errors.Join(deleteErr, profileErr, equipmentErr); err != nil {
		respondInternal(ctx, "failed to delete workout data", err)
		return
	}
	slog.Info("workout history deleted", "user_id", userID, "workout_count", len(deleted))
	ctx.Status(http.StatusNoContent)
}

func (a *App) scheduleWorkoutDeleteDelivery(nickname string, workoutIDs []string) {
	if a.federationDelivery == nil || len(workoutIDs) == 0 {
		return
	}
	a.workoutDeleteDeliveryWG.Go(func() {
		inboxes, err := a.Federation.Followers().ListInboxes(nickname)
		if err != nil {
			slog.Error("federation list inboxes for workout history deletion failed", "user", nickname, "err", err)
			return
		}
		for _, id := range workoutIDs {
			_ = a.federationDelivery.DeliverWorkoutDelete(nickname, id, inboxes)
		}
	})
}
