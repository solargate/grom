package v1

import (
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/solargate/grom/internal/auth/pat"
)

type CreatePATRequest struct {
	Name          string   `json:"name" binding:"required" example:"Backup script"`
	Scopes        []string `json:"scopes" binding:"required" example:"workouts:read,equipment:read"`
	ExpiresInDays int      `json:"expires_in_days,omitempty" example:"90"`
	NoExpiration  bool     `json:"no_expiration,omitempty" example:"false"`
}

type PATResponse struct {
	ID          string     `json:"id" example:"550e8400-e29b-41d4-a716-446655440000"`
	Name        string     `json:"name" example:"Backup script"`
	TokenPrefix string     `json:"token_prefix" example:"grom_pat_a1"`
	Scopes      []string   `json:"scopes" example:"workouts:read,equipment:read"`
	CreatedAt   time.Time  `json:"created_at" example:"2026-08-12T10:00:00Z"`
	ExpiresAt   *time.Time `json:"expires_at,omitempty" example:"2026-11-10T10:00:00Z"`
	LastUsedAt  *time.Time `json:"last_used_at,omitempty" example:"2026-08-12T15:00:00Z"`
}

type CreatePATResponse struct {
	Token string      `json:"token" example:"grom_pat_abc123secret"`
	PAT   PATResponse `json:"pat"`
}

func toPATResponse(rec pat.TokenRecord) PATResponse {
	return PATResponse{
		ID:          rec.ID,
		Name:        rec.Name,
		TokenPrefix: rec.TokenPrefix,
		Scopes:      rec.Scopes,
		CreatedAt:   rec.CreatedAt,
		ExpiresAt:   rec.ExpiresAt,
		LastUsedAt:  rec.LastUsedAt,
	}
}

// listPAT godoc
// @Summary      List personal access tokens
// @Description  Returns metadata for the authenticated user's personal access tokens (secrets are never returned)
// @Tags         pat
// @Produce      json
// @Security     BearerAuth
// @Success      200  {array}   PATResponse
// @Failure      401  {object}  ErrorResponse  "Unauthorized"
// @Failure      500  {object}  ErrorResponse  "Internal server error"
// @Router       /auth/pat [get]
func (a *App) listPAT(ctx *gin.Context) {
	userID, err := a.currentUserID(ctx)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, ErrorResponse{Error: "user not found"})
		return
	}
	records, err := a.PAT.List(userID)
	if err != nil {
		respondInternal(ctx, "failed to list tokens", err)
		return
	}
	resp := make([]PATResponse, 0, len(records))
	for _, rec := range records {
		resp = append(resp, toPATResponse(rec))
	}
	ctx.JSON(http.StatusOK, resp)
}

// createPAT godoc
// @Summary      Create personal access token
// @Description  Creates a scoped personal access token. The full token value is returned only in this response.
// @Tags         pat
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        body  body      CreatePATRequest  true  "Token options"
// @Success      201   {object}  CreatePATResponse
// @Failure      400   {object}  ErrorResponse  "Invalid name, scopes, or expiry"
// @Failure      401   {object}  ErrorResponse  "Unauthorized"
// @Failure      409   {object}  ErrorResponse  "Token limit reached"
// @Failure      500   {object}  ErrorResponse  "Internal server error"
// @Router       /auth/pat [post]
func (a *App) createPAT(ctx *gin.Context) {
	userID, err := a.currentUserID(ctx)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, ErrorResponse{Error: "user not found"})
		return
	}
	var req CreatePATRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid request"})
		return
	}

	result, err := a.PAT.Create(userID, pat.CreateInput{
		Name:          req.Name,
		Scopes:        req.Scopes,
		ExpiresInDays: req.ExpiresInDays,
		NoExpiration:  req.NoExpiration,
	})
	if err != nil {
		switch {
		case errors.Is(err, pat.ErrInvalidRequest),
			errors.Is(err, pat.ErrInvalidScope):
			ctx.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		case errors.Is(err, pat.ErrTooManyTokens):
			ctx.JSON(http.StatusConflict, ErrorResponse{Error: err.Error()})
		default:
			respondInternal(ctx, "failed to create token", err)
		}
		return
	}

	ctx.JSON(http.StatusCreated, CreatePATResponse{
		Token: result.Token,
		PAT:   toPATResponse(result.Record),
	})
}

// revokePAT godoc
// @Summary      Revoke personal access token
// @Description  Permanently revokes a personal access token by ID
// @Tags         pat
// @Produce      json
// @Security     BearerAuth
// @Param        id   path      string  true  "Token ID"
// @Success      204  "No Content"
// @Failure      401  {object}  ErrorResponse  "Unauthorized"
// @Failure      404  {object}  ErrorResponse  "Token not found"
// @Failure      500  {object}  ErrorResponse  "Internal server error"
// @Router       /auth/pat/{id} [delete]
func (a *App) revokePAT(ctx *gin.Context) {
	userID, err := a.currentUserID(ctx)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, ErrorResponse{Error: "user not found"})
		return
	}
	id := ctx.Param("id")
	if err := a.PAT.Revoke(userID, id); err != nil {
		if errors.Is(err, pat.ErrNotFound) {
			ctx.JSON(http.StatusNotFound, ErrorResponse{Error: "token not found"})
			return
		}
		if errors.Is(err, pat.ErrInvalidRequest) {
			ctx.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
			return
		}
		respondInternal(ctx, "failed to revoke token", err)
		return
	}
	ctx.Status(http.StatusNoContent)
}
