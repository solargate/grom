package v1

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/solargate/grom/internal/config"
	"github.com/solargate/grom/internal/equipment"
	"github.com/solargate/grom/internal/users"
	"github.com/solargate/grom/internal/workouts"
)

var equipmentStore *equipment.Store

func initEquipmentStore() {
	if equipmentStore == nil {
		equipmentStore = equipment.NewStore(config.Cfg.Data.ResolvedDir)
	}
}

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
	}
}

func handleEquipmentError(ctx *gin.Context, err error) {
	switch {
	case errors.Is(err, equipment.ErrInvalidEquipment):
		ctx.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
	case errors.Is(err, equipment.ErrEquipmentNotFound):
		ctx.JSON(http.StatusNotFound, ErrorResponse{Error: err.Error()})
	default:
		ctx.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to process equipment"})
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
func listEquipment(ctx *gin.Context) {
	initEquipmentStore()

	nickname, err := currentUserNickname(ctx)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, ErrorResponse{Error: "user not found"})
		return
	}

	items, err := equipmentStore.List(nickname)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to list equipment"})
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
func createEquipment(ctx *gin.Context) {
	initEquipmentStore()

	nickname, err := currentUserNickname(ctx)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, ErrorResponse{Error: "user not found"})
		return
	}

	var req CreateEquipmentRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	created, err := equipmentStore.Create(nickname, &equipment.Equipment{
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
func updateEquipment(ctx *gin.Context) {
	initEquipmentStore()

	nickname, err := currentUserNickname(ctx)
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

	updated, err := equipmentStore.Update(nickname, &equipment.Equipment{
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
func deleteEquipment(ctx *gin.Context) {
	initEquipmentStore()
	initWorkoutStore()

	nickname, err := currentUserNickname(ctx)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, ErrorResponse{Error: "user not found"})
		return
	}

	userID, err := currentUserID(ctx)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, ErrorResponse{Error: "invalid token"})
		return
	}

	id := ctx.Param("id")
	if _, err := equipmentStore.FindByID(nickname, id); err != nil {
		handleEquipmentError(ctx, err)
		return
	}

	if err := workoutStore.RemoveEquipmentFromAll(nickname, id); err != nil {
		ctx.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to update workouts"})
		return
	}

	if err := userStore.RemoveEquipmentFromLastSets(userID, id); err != nil {
		if errors.Is(err, users.ErrUserNotFound) {
			ctx.JSON(http.StatusUnauthorized, ErrorResponse{Error: "user not found"})
			return
		}
		ctx.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to update user preferences"})
		return
	}

	if err := equipmentStore.Delete(nickname, id); err != nil {
		handleEquipmentError(ctx, err)
		return
	}

	ctx.Status(http.StatusNoContent)
}

func resolveWorkoutEquipment(nickname string, equipmentIDs []string) ([]workouts.WorkoutEquipment, error) {
	if len(equipmentIDs) == 0 {
		return nil, nil
	}

	initEquipmentStore()
	items, err := equipmentStore.FindByIDs(nickname, equipmentIDs)
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

func saveLastEquipmentForSport(userID, sportType string, equipmentIDs []string) {
	if userStore == nil {
		if err := initUserStore(); err != nil {
			return
		}
	}
	_ = userStore.SetLastEquipmentForSport(userID, sportType, equipmentIDs)
}
