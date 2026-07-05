package v1

import (
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/solargate/travka/internal/auth"
	"github.com/solargate/travka/internal/config"
	"github.com/solargate/travka/internal/workouts"
)

var workoutStore *workouts.Store

func initWorkoutStore() {
	if workoutStore == nil {
		workoutStore = workouts.NewStore(config.Cfg.Data.ResolvedDir)
	}
}

type CreateWorkoutRequest struct {
	Name            string  `json:"name" binding:"required" example:"Morning run"`
	Description     string  `json:"description" example:"Easy session"`
	SportType       string  `json:"sport_type" binding:"required" example:"Run"`
	StartDate       string  `json:"start_date" binding:"required" example:"2026-07-05T14:30:00+03:00"`
	DurationSeconds int     `json:"duration_seconds" example:"3600"`
	Distance        float64 `json:"distance" example:"5200"`
}

type WorkoutResponse struct {
	ID              string  `json:"id" example:"a866c734-9a31-45ab-9dd4-e4d0fd12e4fd"`
	Name            string  `json:"name" example:"Morning run"`
	Description     string  `json:"description,omitempty" example:"Easy session"`
	SportType       string  `json:"sport_type" example:"Run"`
	StartDate       string  `json:"start_date" example:"2026-07-05T14:30:00+03:00"`
	DurationSeconds int     `json:"duration_seconds" example:"3600"`
	Distance        float64 `json:"distance" example:"5200"`
}

func toWorkoutResponse(workout *workouts.Workout) WorkoutResponse {
	return WorkoutResponse{
		ID:              workout.ID,
		Name:            workout.Name,
		Description:     workout.Description,
		SportType:       workout.SportType,
		StartDate:       workout.StartDate.Format(time.RFC3339),
		DurationSeconds: workout.DurationSeconds,
		Distance:        workout.Distance,
	}
}

func currentUserNickname(ctx *gin.Context) (string, error) {
	userID, _ := ctx.Get(auth.ContextUserIDKey)
	id, ok := userID.(string)
	if !ok || id == "" {
		return "", errors.New("invalid token")
	}

	user, err := userStore.FindByID(id)
	if err != nil {
		return "", err
	}
	return user.Nickname, nil
}

// createWorkout godoc
// @Summary      Create workout
// @Description  Create a manual workout for the authenticated user
// @Tags         workouts
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        body  body  CreateWorkoutRequest  true  "Workout data"
// @Success      201   {object}  WorkoutResponse
// @Failure      400   {object}  ErrorResponse
// @Failure      401   {object}  ErrorResponse
// @Failure      500   {object}  ErrorResponse
// @Router       /workouts [post]
func createWorkout(ctx *gin.Context) {
	initWorkoutStore()

	nickname, err := currentUserNickname(ctx)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, ErrorResponse{Error: "user not found"})
		return
	}

	var req CreateWorkoutRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	startDate, err := time.Parse(time.RFC3339, req.StartDate)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid start_date format, expected RFC3339"})
		return
	}

	workout, err := workoutStore.Create(nickname, &workouts.Workout{
		Name:            req.Name,
		Description:     req.Description,
		SportType:       req.SportType,
		StartDate:       startDate,
		DurationSeconds: req.DurationSeconds,
		Distance:        req.Distance,
	})
	if err != nil {
		if errors.Is(err, workouts.ErrInvalidSportType) {
			ctx.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
			return
		}
		if errors.Is(err, workouts.ErrInvalidWorkout) {
			ctx.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
			return
		}
		ctx.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to create workout"})
		return
	}

	ctx.JSON(http.StatusCreated, toWorkoutResponse(workout))
}

// listWorkouts godoc
// @Summary      List workouts
// @Description  Return workouts for the authenticated user sorted by start date descending
// @Tags         workouts
// @Produce      json
// @Security     BearerAuth
// @Success      200  {array}   WorkoutResponse
// @Failure      401  {object}  ErrorResponse
// @Failure      500  {object}  ErrorResponse
// @Router       /workouts [get]
func listWorkouts(ctx *gin.Context) {
	initWorkoutStore()

	nickname, err := currentUserNickname(ctx)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, ErrorResponse{Error: "user not found"})
		return
	}

	items, err := workoutStore.List(nickname)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to list workouts"})
		return
	}

	response := make([]WorkoutResponse, 0, len(items))
	for i := range items {
		response = append(response, toWorkoutResponse(&items[i]))
	}

	ctx.JSON(http.StatusOK, response)
}
