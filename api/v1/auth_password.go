package v1

import (
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/solargate/grom/internal/auth/reset"
	"github.com/solargate/grom/internal/users"
)

type forgotPasswordRequest struct {
	Email  string `json:"email" binding:"required,email" example:"solarwind.palm@gmail.com"`
	Altcha string `json:"altcha,omitempty" example:""`
}

type resetPasswordRequest struct {
	Token    string `json:"token" binding:"required" example:"abc123"`
	Password string `json:"password" binding:"required,min=8" example:"secret123"`
}

// forgotPassword godoc
// @Summary      Request password reset
// @Description  Sends a password reset email if the account exists. Always returns 204 when reset is enabled (except rate limits).
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        body  body  forgotPasswordRequest  true  "Account email"
// @Success      204
// @Failure      400   {object}  ErrorResponse  "Invalid request or captcha"
// @Failure      429   {object}  ErrorResponse  "Rate limit exceeded"
// @Failure      503   {object}  ErrorResponse  "Password reset is not configured"
// @Router       /auth/password/forgot [post]
func (a *App) forgotPassword(ctx *gin.Context) {
	if a.PasswordReset == nil || !a.PasswordReset.Enabled() {
		ctx.JSON(http.StatusServiceUnavailable, ErrorResponse{Error: "password reset is not configured"})
		return
	}

	var req forgotPasswordRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}
	if !a.requireCaptcha(ctx, req.Altcha) {
		return
	}

	email := strings.ToLower(strings.TrimSpace(req.Email))
	if ok, retry := a.PasswordReset.Limiter().AllowForgot(ctx.ClientIP(), email); !ok {
		writeRateLimited(ctx, retry)
		return
	}

	if err := a.PasswordReset.RequestReset(ctx.Request.Context(), email); err != nil {
		if errors.Is(err, reset.ErrNotConfigured) {
			ctx.JSON(http.StatusServiceUnavailable, ErrorResponse{Error: "password reset is not configured"})
			return
		}
		respondInternal(ctx, "failed to request password reset", err)
		return
	}
	ctx.Status(http.StatusNoContent)
}

// resetPassword godoc
// @Summary      Reset password with token
// @Description  Sets a new password using a one-time token from the reset email
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        body  body  resetPasswordRequest  true  "Reset token and new password"
// @Success      204
// @Failure      400   {object}  ErrorResponse  "Invalid request or reset token"
// @Failure      429   {object}  ErrorResponse  "Rate limit exceeded"
// @Failure      503   {object}  ErrorResponse  "Password reset is not configured"
// @Router       /auth/password/reset [post]
func (a *App) resetPassword(ctx *gin.Context) {
	if a.PasswordReset == nil || !a.PasswordReset.Enabled() {
		ctx.JSON(http.StatusServiceUnavailable, ErrorResponse{Error: "password reset is not configured"})
		return
	}

	var req resetPasswordRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	if ok, retry := a.PasswordReset.Limiter().AllowReset(ctx.ClientIP()); !ok {
		writeRateLimited(ctx, retry)
		return
	}

	err := a.PasswordReset.ConfirmReset(ctx.Request.Context(), req.Token, req.Password)
	if err != nil {
		switch {
		case errors.Is(err, reset.ErrNotConfigured):
			ctx.JSON(http.StatusServiceUnavailable, ErrorResponse{Error: "password reset is not configured"})
		case errors.Is(err, reset.ErrInvalidToken):
			ctx.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid or expired reset token"})
		case errors.Is(err, reset.ErrWeakPassword):
			ctx.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		case errors.Is(err, users.ErrUserNotFound):
			ctx.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid or expired reset token"})
		default:
			respondInternal(ctx, "failed to reset password", err)
		}
		return
	}
	ctx.Status(http.StatusNoContent)
}

func writeRateLimited(ctx *gin.Context, retryAfter time.Duration) {
	secs := int(retryAfter.Seconds())
	if secs < 1 {
		secs = 1
	}
	ctx.Header("Retry-After", strconv.Itoa(secs))
	ctx.JSON(http.StatusTooManyRequests, ErrorResponse{Error: "too many requests, try again later"})
}
