package v1

import (
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/solargate/grom/internal/notifications"
)

type PushSubscriptionKeys struct {
	P256dh string `json:"p256dh" binding:"required" example:"BNcRdba..."`
	Auth   string `json:"auth" binding:"required" example:"tBHItJI5svbpez7..."`
}

type RegisterPushSubscriptionRequest struct {
	InstallationID string               `json:"installation_id" binding:"required" example:"550e8400-e29b-41d4-a716-446655440000"`
	Endpoint       string               `json:"endpoint" binding:"required" example:"https://fcm.googleapis.com/fcm/send/..."`
	Keys           PushSubscriptionKeys `json:"keys" binding:"required"`
	Platform       string               `json:"platform,omitempty" example:"android"`
	DeviceName     string               `json:"device_name,omitempty" example:"Pixel 8"`
}

type PushSubscriptionResponse struct {
	InstallationID string `json:"installation_id" example:"550e8400-e29b-41d4-a716-446655440000"`
	Platform       string `json:"platform,omitempty" example:"android"`
	DeviceName     string `json:"device_name,omitempty" example:"Pixel 8"`
	UpdatedAt      string `json:"updated_at" example:"2026-10-02T12:00:00Z"`
}

// registerPushSubscription godoc
// @Summary      Register push subscription
// @Description  Upsert a Web Push subscription for the authenticated user's device (JWT only).
// @Tags         notifications
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        body  body      RegisterPushSubscriptionRequest  true  "Subscription"
// @Success      200   {object}  PushSubscriptionResponse
// @Failure      400   {object}  ErrorResponse  "Invalid subscription"
// @Failure      401   {object}  ErrorResponse  "Unauthorized"
// @Failure      500   {object}  ErrorResponse  "Internal server error"
// @Router       /notifications/push [post]
func (a *App) registerPushSubscription(ctx *gin.Context) {
	userID, err := a.currentUserID(ctx)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, ErrorResponse{Error: "user not found"})
		return
	}
	var req RegisterPushSubscriptionRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid request"})
		return
	}
	installationID := strings.TrimSpace(req.InstallationID)
	endpoint := strings.TrimSpace(req.Endpoint)
	p256dh := strings.TrimSpace(req.Keys.P256dh)
	authKey := strings.TrimSpace(req.Keys.Auth)
	if installationID == "" || endpoint == "" || p256dh == "" || authKey == "" {
		ctx.JSON(http.StatusBadRequest, ErrorResponse{Error: notifications.ErrInvalidSubscription.Error()})
		return
	}
	if a.PushSubscriptions == nil {
		respondInternal(ctx, "push subscriptions unavailable", errors.New("nil repository"))
		return
	}
	now := time.Now().UTC()
	sub := notifications.PushSubscription{
		UserID:         userID,
		InstallationID: installationID,
		Endpoint:       endpoint,
		P256dh:         p256dh,
		Auth:           authKey,
		Platform:       strings.TrimSpace(req.Platform),
		DeviceName:     strings.TrimSpace(req.DeviceName),
		UpdatedAt:      now,
	}
	if err := a.PushSubscriptions.Upsert(sub); err != nil {
		respondInternal(ctx, "failed to store push subscription", err)
		return
	}
	ctx.JSON(http.StatusOK, PushSubscriptionResponse{
		InstallationID: sub.InstallationID,
		Platform:       sub.Platform,
		DeviceName:     sub.DeviceName,
		UpdatedAt:      sub.UpdatedAt.Format(time.RFC3339),
	})
}

// deletePushSubscription godoc
// @Summary      Unregister push subscription
// @Description  Remove a Web Push subscription for the authenticated user's device (JWT only).
// @Tags         notifications
// @Produce      json
// @Security     BearerAuth
// @Param        installationId  path  string  true  "Installation ID"
// @Success      204
// @Failure      401  {object}  ErrorResponse  "Unauthorized"
// @Failure      404  {object}  ErrorResponse  "Subscription not found"
// @Failure      500  {object}  ErrorResponse  "Internal server error"
// @Router       /notifications/push/{installationId} [delete]
func (a *App) deletePushSubscription(ctx *gin.Context) {
	userID, err := a.currentUserID(ctx)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, ErrorResponse{Error: "user not found"})
		return
	}
	installationID := strings.TrimSpace(ctx.Param("installationId"))
	if installationID == "" {
		ctx.JSON(http.StatusBadRequest, ErrorResponse{Error: "installation id required"})
		return
	}
	if a.PushSubscriptions == nil {
		respondInternal(ctx, "push subscriptions unavailable", errors.New("nil repository"))
		return
	}
	if err := a.PushSubscriptions.DeleteByUserAndInstallation(userID, installationID); err != nil {
		if errors.Is(err, notifications.ErrNotFound) {
			ctx.JSON(http.StatusNotFound, ErrorResponse{Error: "subscription not found"})
			return
		}
		respondInternal(ctx, "failed to delete push subscription", err)
		return
	}
	ctx.Status(http.StatusNoContent)
}
