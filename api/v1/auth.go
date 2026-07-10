package v1

import (
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/solargate/travka/internal/config"
	"github.com/solargate/travka/internal/auth"
	"github.com/solargate/travka/internal/users"
)

var userStore *users.Store

func initUserStore() error {
	if userStore != nil {
		return nil
	}
	store, err := users.NewStore(config.Cfg.Data.ResolvedDir)
	if err != nil {
		return err
	}
	userStore = store
	return nil
}

type RegisterRequest struct {
	Nickname string `json:"nickname" binding:"required" example:"solarwind"`
	Name     string `json:"name" example:"Alexander Cheryomukhin"`
	Email    string `json:"email" binding:"required,email" example:"solarwind.palm@gmail.com"`
	Password string `json:"password" binding:"required,min=8" example:"secret123"`
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email" example:"solarwind.palm@gmail.com"`
	Password string `json:"password" binding:"required" example:"secret123"`
}

type UserResponse struct {
	ID                   string              `json:"id" example:"550e8400-e29b-41d4-a716-446655440000"`
	Nickname             string              `json:"nickname" example:"solarwind"`
	Name                 string              `json:"name" example:"Alexander Cheryomukhin"`
	Email                string              `json:"email" example:"solarwind.palm@gmail.com"`
	HasAvatar            bool                `json:"has_avatar" example:"true"`
	AvatarURL            string              `json:"avatar_url,omitempty" example:"/api/v1/users/solarwind/avatar"`
	LastEquipmentBySport map[string][]string `json:"last_equipment_by_sport,omitempty"`
}

type LoginResponse struct {
	Token     string       `json:"token" example:"eyJhbGciOiJIUzI1NiIs..."`
	ExpiresAt string       `json:"expires_at" example:"2026-07-05T14:30:00Z"`
	User      UserResponse `json:"user"`
}

type ErrorResponse struct {
	Error string `json:"error" example:"email already registered"`
}

func toUserResponse(user *users.User) UserResponse {
	hasAvatar, avatarURL := localAvatarFieldsForUser(user.Nickname)
	return UserResponse{
		ID:                   user.ID,
		Nickname:             user.Nickname,
		Name:                 user.Name,
		Email:                user.Email,
		HasAvatar:            hasAvatar,
		AvatarURL:            avatarURL,
		LastEquipmentBySport: user.LastEquipmentBySport,
	}
}

// register godoc
// @Summary      Register new user
// @Description  Create a new user account
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        body  body  RegisterRequest  true  "Registration data"
// @Success      201   {object}  UserResponse
// @Failure      400   {object}  ErrorResponse
// @Failure      409   {object}  ErrorResponse
// @Router       /auth/register [post]
func register(ctx *gin.Context) {
	var req RegisterRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	user, err := userStore.Create(req.Nickname, req.Name, req.Email, req.Password)
	if err != nil {
		if errors.Is(err, users.ErrInvalidNickname) {
			ctx.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
			return
		}
		if errors.Is(err, users.ErrEmailTaken) || errors.Is(err, users.ErrNicknameTaken) {
			ctx.JSON(http.StatusConflict, ErrorResponse{Error: err.Error()})
			return
		}
		ctx.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to create user"})
		return
	}

	ctx.JSON(http.StatusCreated, toUserResponse(user))
}

// login godoc
// @Summary      Login user
// @Description  Authenticate user with email and password
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        body  body  LoginRequest  true  "Login credentials"
// @Success      200   {object}  LoginResponse
// @Failure      400   {object}  ErrorResponse
// @Failure      401   {object}  ErrorResponse
// @Router       /auth/login [post]
func login(ctx *gin.Context) {
	var req LoginRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	user, err := userStore.FindByEmail(req.Email)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, ErrorResponse{Error: "invalid email or password"})
		return
	}

	if !auth.CheckPassword(user.PasswordHash, req.Password) {
		ctx.JSON(http.StatusUnauthorized, ErrorResponse{Error: "invalid email or password"})
		return
	}

	token, expiresAt, err := auth.GenerateToken(user.ID, user.Email)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to generate token"})
		return
	}

	ctx.JSON(http.StatusOK, LoginResponse{
		Token:     token,
		ExpiresAt: expiresAt.UTC().Format(time.RFC3339),
		User:      toUserResponse(user),
	})
}

// getMe godoc
// @Summary      Get current user
// @Description  Return profile of authenticated user
// @Tags         auth
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  UserResponse
// @Failure      401  {object}  ErrorResponse
// @Router       /auth/me [get]
func getMe(ctx *gin.Context) {
	userID, _ := ctx.Get(auth.ContextUserIDKey)
	id, ok := userID.(string)
	if !ok || id == "" {
		ctx.JSON(http.StatusUnauthorized, ErrorResponse{Error: "invalid token"})
		return
	}

	user, err := userStore.FindByID(id)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, ErrorResponse{Error: "user not found"})
		return
	}

	ctx.JSON(http.StatusOK, toUserResponse(user))
}

type UpdateProfileRequest struct {
	Name string `json:"name" example:"Alexander Cheryomukhin"`
}

// updateMe godoc
// @Summary      Update current user profile
// @Description  Update profile fields for the authenticated user
// @Tags         auth
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        body  body  UpdateProfileRequest  true  "Profile data"
// @Success      200   {object}  UserResponse
// @Failure      400   {object}  ErrorResponse
// @Failure      401   {object}  ErrorResponse
// @Router       /auth/me [patch]
func updateMe(ctx *gin.Context) {
	userID, _ := ctx.Get(auth.ContextUserIDKey)
	id, ok := userID.(string)
	if !ok || id == "" {
		ctx.JSON(http.StatusUnauthorized, ErrorResponse{Error: "invalid token"})
		return
	}

	var req UpdateProfileRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	user, err := userStore.UpdateProfile(id, req.Name)
	if err != nil {
		if errors.Is(err, users.ErrUserNotFound) {
			ctx.JSON(http.StatusUnauthorized, ErrorResponse{Error: "user not found"})
			return
		}
		ctx.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to update profile"})
		return
	}

	ctx.JSON(http.StatusOK, toUserResponse(user))
}
