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
	UsedSportTypes       []string            `json:"used_sport_types,omitempty"`
}

func toProfileResponse(profile *users.Profile) ProfileResponse {
	if profile == nil {
		return ProfileResponse{}
	}
	return ProfileResponse{
		LastSportType:        profile.LastSportType,
		LastEquipmentBySport: profile.LastEquipmentBySport,
		UsedSportTypes:       profile.UsedSportTypes,
	}
}

// getProfile godoc
// @Summary      Get current user profile preferences
// @Description  Returns UI/service preferences (last sport type, last equipment by sport, used sport types ordered by most recent create/update). Not part of public user identity.
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

func (a *App) rememberUsedSportType(userID, sportType string) {
	a.profileSportsMu.Lock()
	defer a.profileSportsMu.Unlock()
	if err := a.Users.TouchUsedSportType(userID, sportType); err != nil {
		slog.Warn("failed to remember used sport type",
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

func (a *App) touchUsedSportFromWorkout(userID string, workout *workouts.Workout) {
	if workout == nil {
		return
	}
	a.rememberUsedSportType(userID, workout.SportType)
}

// RefreshLastSportType rescans the user's workouts and stores the sport of the newest one.
// It also prunes used_sport_types entries that no longer appear on any workout.
// Serialized with rememberUsedSportType so prune cannot apply a List snapshot taken
// before a concurrent Touch.
func (a *App) RefreshLastSportType(nickname, userID string) error {
	a.profileSportsMu.Lock()
	defer a.profileSportsMu.Unlock()

	items, err := a.Workouts.List(nickname)
	if err != nil {
		return err
	}
	if err := a.Users.SetLastSportType(userID, workouts.NewestSportType(items)); err != nil {
		return err
	}
	return a.Users.PruneUsedSportTypes(userID, workouts.UniqueSportTypes(items))
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
