package v1

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/solargate/grom/internal/equipment"
	"github.com/solargate/grom/internal/equipment/distance"
	"github.com/solargate/grom/internal/users"
	"github.com/solargate/grom/internal/workouts"
)

type EquipmentResponse struct {
	ID        string   `json:"id" example:"550e8400-e29b-41d4-a716-446655440000"`
	Type      string   `json:"type" example:"bike"`
	Name      string   `json:"name" example:"Gravel bike"`
	BikeType  string   `json:"bike_type,omitempty" example:"gravel"`
	WaterType string   `json:"water_type,omitempty" example:""`
	Brand     string   `json:"brand,omitempty" example:"Canyon"`
	Model     string   `json:"model,omitempty" example:"Grizl"`
	WeightKg  *float64 `json:"weight_kg,omitempty" example:"9.5"`
	Notes     string   `json:"notes,omitempty" example:"Race setup"`
	Distance  float64  `json:"distance" example:"12500"`
}

type CreateEquipmentRequest struct {
	Type      string   `json:"type" binding:"required" example:"bike"`
	Name      string   `json:"name" binding:"required" example:"Gravel bike"`
	BikeType  string   `json:"bike_type,omitempty" example:"gravel"`
	WaterType string   `json:"water_type,omitempty" example:""`
	Brand     string   `json:"brand,omitempty" example:"Canyon"`
	Model     string   `json:"model,omitempty" example:"Grizl"`
	WeightKg  *float64 `json:"weight_kg,omitempty" example:"9.5"`
	Notes     string   `json:"notes,omitempty" example:"Race setup"`
}

type UpdateEquipmentRequest struct {
	Type      string   `json:"type" binding:"required" example:"bike"`
	Name      string   `json:"name" binding:"required" example:"Gravel bike"`
	BikeType  string   `json:"bike_type,omitempty" example:"gravel"`
	WaterType string   `json:"water_type,omitempty" example:""`
	Brand     string   `json:"brand,omitempty" example:"Canyon"`
	Model     string   `json:"model,omitempty" example:"Grizl"`
	WeightKg  *float64 `json:"weight_kg,omitempty" example:"9.5"`
	Notes     string   `json:"notes,omitempty" example:"Race setup"`
}

func toEquipmentResponse(item *equipment.Equipment) EquipmentResponse {
	return EquipmentResponse{
		ID:        item.ID,
		Type:      item.Type,
		Name:      item.Name,
		BikeType:  item.BikeType,
		WaterType: item.WaterType,
		Brand:     item.Brand,
		Model:     item.Model,
		WeightKg:  item.WeightKg,
		Notes:     item.Notes,
		Distance:  item.Distance,
	}
}

func handleEquipmentError(ctx *gin.Context, err error) {
	switch {
	case errors.Is(err, equipment.ErrInvalidEquipment):
		ctx.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
	case errors.Is(err, equipment.ErrEquipmentNotFound):
		ctx.JSON(http.StatusNotFound, ErrorResponse{Error: err.Error()})
	default:
		respondInternal(ctx, "failed to process equipment", err)
	}
}

// listEquipment godoc
// @Summary      List equipment
// @Description  Return equipment catalog for the authenticated user
// @Tags         equipment
// @Produce      json
// @Security     BearerAuth
// @Success      200  {array}   EquipmentResponse
// @Failure      401  {object}  ErrorResponse
// @Failure      500  {object}  ErrorResponse
// @Router       /equipment [get]
func (a *App) listEquipment(ctx *gin.Context) {
	
	nickname, err := a.currentUserNickname(ctx)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, ErrorResponse{Error: "user not found"})
		return
	}

	items, err := a.Equipment.List(nickname)
	if err != nil {
		respondInternal(ctx, "failed to list equipment", err)
		return
	}

	response := make([]EquipmentResponse, 0, len(items))
	for i := range items {
		response = append(response, toEquipmentResponse(&items[i]))
	}
	ctx.JSON(http.StatusOK, response)
}

// createEquipment godoc
// @Summary      Create equipment
// @Description  Add new equipment item for the authenticated user
// @Tags         equipment
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        body  body  CreateEquipmentRequest  true  "Equipment data"
// @Success      201   {object}  EquipmentResponse
// @Failure      400   {object}  ErrorResponse
// @Failure      401   {object}  ErrorResponse
// @Failure      500   {object}  ErrorResponse
// @Router       /equipment [post]
func (a *App) createEquipment(ctx *gin.Context) {
	
	nickname, err := a.currentUserNickname(ctx)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, ErrorResponse{Error: "user not found"})
		return
	}

	var req CreateEquipmentRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	created, err := a.Equipment.Create(nickname, &equipment.Equipment{
		Type:      req.Type,
		Name:      req.Name,
		BikeType:  req.BikeType,
		WaterType: req.WaterType,
		Brand:     req.Brand,
		Model:     req.Model,
		WeightKg:  req.WeightKg,
		Notes:     req.Notes,
	})
	if err != nil {
		handleEquipmentError(ctx, err)
		return
	}

	ctx.JSON(http.StatusCreated, toEquipmentResponse(created))
}

// updateEquipment godoc
// @Summary      Update equipment
// @Description  Update existing equipment item
// @Tags         equipment
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id    path  string  true  "Equipment ID"
// @Param        body  body  UpdateEquipmentRequest  true  "Equipment data"
// @Success      200   {object}  EquipmentResponse
// @Failure      400   {object}  ErrorResponse
// @Failure      401   {object}  ErrorResponse
// @Failure      404   {object}  ErrorResponse
// @Failure      500   {object}  ErrorResponse
// @Router       /equipment/{id} [put]
func (a *App) updateEquipment(ctx *gin.Context) {
	
	nickname, err := a.currentUserNickname(ctx)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, ErrorResponse{Error: "user not found"})
		return
	}

	id := ctx.Param("id")
	var req UpdateEquipmentRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	updated, err := a.Equipment.Update(nickname, &equipment.Equipment{
		ID:        id,
		Type:      req.Type,
		Name:      req.Name,
		BikeType:  req.BikeType,
		WaterType: req.WaterType,
		Brand:     req.Brand,
		Model:     req.Model,
		WeightKg:  req.WeightKg,
		Notes:     req.Notes,
	})
	if err != nil {
		handleEquipmentError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, toEquipmentResponse(updated))
}

// deleteEquipment godoc
// @Summary      Delete equipment
// @Description  Delete equipment and remove it from all workouts and saved sport presets
// @Tags         equipment
// @Produce      json
// @Security     BearerAuth
// @Param        id  path  string  true  "Equipment ID"
// @Success      204
// @Failure      401  {object}  ErrorResponse
// @Failure      404  {object}  ErrorResponse
// @Failure      500  {object}  ErrorResponse
// @Router       /equipment/{id} [delete]
func (a *App) deleteEquipment(ctx *gin.Context) {
			nickname, err := a.currentUserNickname(ctx)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, ErrorResponse{Error: "user not found"})
		return
	}

	userID, err := a.currentUserID(ctx)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, ErrorResponse{Error: "invalid token"})
		return
	}

	id := ctx.Param("id")
	if _, err := a.Equipment.FindByID(nickname, id); err != nil {
		handleEquipmentError(ctx, err)
		return
	}

	if err := a.Workouts.RemoveEquipmentFromAll(nickname, id); err != nil {
		respondInternal(ctx, "failed to update workouts", err)
		return
	}

	if err := a.Users.RemoveEquipmentFromLastSets(userID, id); err != nil {
		if errors.Is(err, users.ErrUserNotFound) {
			ctx.JSON(http.StatusUnauthorized, ErrorResponse{Error: "user not found"})
			return
		}
		respondInternal(ctx, "failed to update user preferences", err)
		return
	}

	if err := a.Equipment.Delete(nickname, id); err != nil {
		handleEquipmentError(ctx, err)
		return
	}

	ctx.Status(http.StatusNoContent)
}

func (a *App) resolveWorkoutEquipment(nickname string, equipmentIDs []string) ([]workouts.WorkoutEquipment, error) {
	if len(equipmentIDs) == 0 {
		return nil, nil
	}

	items, err := a.Equipment.FindByIDs(nickname, equipmentIDs)
	if err != nil {
		return nil, err
	}

	byID := make(map[string]equipment.Equipment, len(items))
	for i := range items {
		byID[items[i].ID] = items[i]
	}

	result := make([]workouts.WorkoutEquipment, 0, len(equipmentIDs))
	for _, id := range equipmentIDs {
		item, ok := byID[id]
		if !ok {
			continue
		}
		result = append(result, workouts.WorkoutEquipment{
			ID:   item.ID,
			Name: item.Name,
			Type: item.Type,
		})
	}
	return result, nil
}

func (a *App) saveLastEquipmentForSport(userID, sportType string, equipmentIDs []string) {
	_ = a.Users.SetLastEquipmentForSport(userID, sportType, equipmentIDs)
}

func (a *App) scheduleEquipmentDistanceRecalc(nickname string, items []workouts.WorkoutEquipment) {
	if a.EquipmentDistance == nil {
		return
	}
	a.EquipmentDistance.ScheduleRecalculateForIDs(nickname, distance.CollectWorkoutEquipmentIDs(items))
}
