package v1

import (
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/solargate/grom/internal/users"
	"github.com/solargate/grom/internal/workouts"
)

type ProfileResponse struct {
	LastSportType        string              `json:"last_sport_type,omitempty"`
	LastEquipmentBySport map[string][]string `json:"last_equipment_by_sport,omitempty"`
}

func toProfileResponse(profile *users.Profile) ProfileResponse {
	if profile == nil {
		return ProfileResponse{}
	}
	return ProfileResponse{
		LastSportType:        profile.LastSportType,
		LastEquipmentBySport: profile.LastEquipmentBySport,
	}
}

// getProfile godoc
// @Summary      Get current user profile preferences
// @Description  Returns UI/service preferences (last sport type, last equipment by sport). Not part of public user identity.
// @Tags         profile
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  ProfileResponse
// @Failure      401  {object}  ErrorResponse  "Unauthorized"
// @Failure      500  {object}  ErrorResponse  "Internal server error"
// @Router       /profile [get]
func (a *App) getProfile(ctx *gin.Context) {
	userID, err := a.currentUserID(ctx)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, ErrorResponse{Error: "user not found"})
		return
	}

	profile, err := a.Users.GetProfile(userID)
	if err != nil {
		respondInternal(ctx, "failed to load profile", err)
		return
	}
	ctx.JSON(http.StatusOK, toProfileResponse(profile))
}

func (a *App) saveLastEquipmentForSport(userID, sportType string, equipmentIDs []string) {
	if err := a.Users.SetLastEquipmentForSport(userID, sportType, equipmentIDs); err != nil {
		slog.Warn("failed to save last equipment for sport",
			"user_id", userID,
			"sport_type", sportType,
			"err", err,
		)
	}
}

func (a *App) touchLastEquipmentFromWorkout(userID string, workout *workouts.Workout) {
	if workout == nil {
		return
	}
	a.saveLastEquipmentForSport(userID, workout.SportType, workoutEquipmentIDs(workout.Equipment))
}

// RefreshLastSportType rescans the user's workouts and stores the sport of the newest one.
func (a *App) RefreshLastSportType(nickname, userID string) error {
	items, err := a.Workouts.List(nickname)
	if err != nil {
		return err
	}
	return a.Users.SetLastSportType(userID, workouts.NewestSportType(items))
}

func (a *App) scheduleRefreshLastSportType(nickname, userID string) {
	a.profileRefreshWG.Add(1)
	go func() {
		defer a.profileRefreshWG.Done()
		if err := a.RefreshLastSportType(nickname, userID); err != nil {
			slog.Warn("failed to refresh last sport type",
				"user", nickname,
				"user_id", userID,
				"err", err,
			)
		}
	}()
}
