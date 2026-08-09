package v1

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/solargate/grom/internal/auth/captcha"
)

// getCaptchaChallenge godoc
// @Summary      Get ALTCHA challenge
// @Description  Returns a proof-of-work challenge when auth.captcha.enabled is true
// @Tags         captcha
// @Produce      json
// @Success      200
// @Failure      404  {object}  ErrorResponse  "Captcha is disabled"
// @Failure      429  {object}  ErrorResponse  "Rate limit exceeded"
// @Router       /captcha/challenge [get]
func (a *App) getCaptchaChallenge(ctx *gin.Context) {
	if a.Captcha == nil || !a.Captcha.Enabled() {
		ctx.JSON(http.StatusNotFound, ErrorResponse{Error: "captcha is disabled"})
		return
	}
	challenge, err := a.Captcha.CreateChallenge(ctx.ClientIP())
	if err != nil {
		if errors.Is(err, captcha.ErrRateLimited) {
			writeRateLimited(ctx, 0)
			return
		}
		if errors.Is(err, captcha.ErrDisabled) {
			ctx.JSON(http.StatusNotFound, ErrorResponse{Error: "captcha is disabled"})
			return
		}
		respondInternal(ctx, "failed to create captcha challenge", err)
		return
	}
	ctx.JSON(http.StatusOK, challenge)
}

// requireCaptcha verifies the ALTCHA payload when captcha is enabled.
// Returns false when the handler should stop (response already written).
func (a *App) requireCaptcha(ctx *gin.Context, payload string) bool {
	if a.Captcha == nil || !a.Captcha.Enabled() {
		return true
	}
	err := a.Captcha.Verify(payload)
	if err == nil {
		return true
	}
	switch {
	case errors.Is(err, captcha.ErrMissing):
		ctx.JSON(http.StatusBadRequest, ErrorResponse{Error: "captcha is required"})
	case errors.Is(err, captcha.ErrExpired):
		ctx.JSON(http.StatusBadRequest, ErrorResponse{Error: "captcha expired"})
	case errors.Is(err, captcha.ErrReplay):
		ctx.JSON(http.StatusBadRequest, ErrorResponse{Error: "captcha already used"})
	default:
		ctx.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid captcha"})
	}
	return false
}
