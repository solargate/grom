package v1

import (
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/solargate/grom/internal/avatars"
	"github.com/solargate/grom/internal/config"
	fed "github.com/solargate/grom/internal/federation"
	"github.com/solargate/grom/internal/federation/httpsig"
)

func (a *App) registerFederationRoutes(router *gin.Engine) {
	router.GET("/.well-known/webfinger", a.webfingerHandler())
	router.GET("/users/:nickname/avatar", a.publicAvatarHandler())
	router.GET("/actor", a.instanceActorHandler())
	router.GET("/users/:nickname", a.actorHandler())
	router.POST("/users/:nickname/inbox", a.inboxHandler())
	router.GET("/users/:nickname/outbox", a.outboxHandler())
	router.POST("/inbox", a.sharedInboxHandler())
}

func publicDomain() string {
	if config.Cfg.Federation.Domain != "" {
		return config.Cfg.Federation.Domain
	}
	return "localhost"
}

func actorURL(nickname string) string {
	return fmt.Sprintf("https://%s/users/%s", publicDomain(), nickname)
}

func (a *App) webfingerHandler() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		resource := ctx.Query("resource")
		if !strings.HasPrefix(resource, "acct:") {
			ctx.Status(http.StatusNotFound)
			return
		}
		handle := strings.TrimPrefix(resource, "acct:")
		parts := strings.Split(handle, "@")
		if len(parts) != 2 || !strings.EqualFold(parts[1], publicDomain()) {
			ctx.Status(http.StatusNotFound)
			return
		}
		user, err := a.Users.FindByNickname(parts[0])
		if err != nil {
			ctx.Status(http.StatusNotFound)
			return
		}
		ctx.JSON(http.StatusOK, gin.H{
			"subject": resource,
			"links": []gin.H{
				{
					"rel":  "self",
					"type": "application/activity+json",
					"href": actorURL(user.Nickname),
				},
			},
		})
	}
}

func (a *App) instanceActorHandler() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		if !wantsActivityJSON(ctx) {
			ctx.Status(http.StatusNotFound)
			return
		}
		ak, err := fed.LoadOrCreateInstanceActorKey(a.Blobs)
		if err != nil {
			logInternalError(ctx, "failed to load instance actor key", err)
			ctx.Status(http.StatusInternalServerError)
			return
		}
		id := fmt.Sprintf("https://%s/actor", publicDomain())
		ctx.Header("Content-Type", "application/activity+json")
		ctx.JSON(http.StatusOK, gin.H{
			"@context": []string{
				"https://www.w3.org/ns/activitystreams",
				"https://w3id.org/security/v1",
			},
			"id":                id,
			"type":              "Application",
			"preferredUsername": "actor",
			"name":              config.Cfg.Server.Name,
			"inbox":             fmt.Sprintf("https://%s/inbox", publicDomain()),
			"outbox":            id + "/outbox",
			"url":               id,
			"publicKey": gin.H{
				"id":           ak.KeyID,
				"owner":        id,
				"publicKeyPem": ak.PubPEM,
			},
			"endpoints": gin.H{
				"sharedInbox": fmt.Sprintf("https://%s/inbox", publicDomain()),
			},
		})
	}
}

func (a *App) actorHandler() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		if !wantsActivityJSON(ctx) {
			ctx.Status(http.StatusNotFound)
			return
		}
		if err := a.requireAuthorizedFetch(ctx); err != nil {
			ctx.Status(http.StatusUnauthorized)
			return
		}
		user, err := a.Users.FindByNickname(ctx.Param("nickname"))
		if err != nil {
			ctx.Status(http.StatusNotFound)
			return
		}
		pubKey, keyID, err := fed.LoadOrCreateActorKey(a.Blobs, user.Nickname)
		if err != nil {
			logInternalError(ctx, "failed to load actor key", err)
			ctx.Status(http.StatusInternalServerError)
			return
		}
		response := gin.H{
			"@context": []string{
				"https://www.w3.org/ns/activitystreams",
				"https://w3id.org/security/v1",
			},
			"id":                actorURL(user.Nickname),
			"type":              "Person",
			"preferredUsername": user.Nickname,
			"name":              user.Name,
			"inbox":             actorURL(user.Nickname) + "/inbox",
			"outbox":            actorURL(user.Nickname) + "/outbox",
			"followers":         actorURL(user.Nickname) + "/followers",
			"following":         actorURL(user.Nickname) + "/following",
			"publicKey": gin.H{
				"id":           keyID,
				"owner":        actorURL(user.Nickname),
				"publicKeyPem": pubKey,
			},
			"endpoints": gin.H{
				"sharedInbox": fmt.Sprintf("https://%s/inbox", publicDomain()),
			},
		}
		if avatars.HasStore(a.Blobs, user.Nickname) {
			response["icon"] = gin.H{
				"type":      "Image",
				"mediaType": "image/webp",
				"url":       a.publicAvatarURL(user.Nickname),
			}
		}
		ctx.Header("Vary", "Signature, Accept")
		ctx.Header("Content-Type", "application/activity+json")
		ctx.JSON(http.StatusOK, response)
	}
}

func (a *App) publicAvatarHandler() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		nickname := strings.TrimSpace(ctx.Param("nickname"))
		if nickname == "" {
			ctx.Status(http.StatusNotFound)
			return
		}
		if _, err := a.Users.FindByNickname(nickname); err != nil {
			ctx.Status(http.StatusNotFound)
			return
		}
		data, err := avatars.LoadStore(a.Blobs, nickname)
		if err != nil {
			ctx.Status(http.StatusNotFound)
			return
		}
		ctx.Header("Cache-Control", "public, max-age=86400")
		ctx.Header("Content-Type", "image/webp")
		ctx.Data(http.StatusOK, "image/webp", data)
	}
}

func (a *App) inboxHandler() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		nickname := ctx.Param("nickname")
		body, err := httpsig.ReadBody(ctx.Request)
		if err != nil {
			ctx.Status(http.StatusBadRequest)
			return
		}
		activity, err := decodeJSONObject(body)
		if err != nil {
			ctx.Status(http.StatusBadRequest)
			return
		}
		if err := a.authenticateFederationActivity(ctx, body, activity); err != nil {
			slog.Warn("federation inbox unauthorized", "nickname", nickname, "err", err)
			ctx.Status(http.StatusUnauthorized)
			return
		}
		if a.federationInboxProc != nil {
			if err := a.federationInboxProc.Handle(nickname, strings.NewReader(string(body))); err != nil {
				ctx.Status(http.StatusBadRequest)
				return
			}
		}
		ctx.Status(http.StatusAccepted)
	}
}

func (a *App) sharedInboxHandler() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		body, err := httpsig.ReadBody(ctx.Request)
		if err != nil {
			ctx.Status(http.StatusBadRequest)
			return
		}
		activity, err := decodeJSONObject(body)
		if err != nil {
			ctx.Status(http.StatusBadRequest)
			return
		}
		if err := a.authenticateFederationActivity(ctx, body, activity); err != nil {
			slog.Warn("federation shared inbox unauthorized", "err", err)
			ctx.Status(http.StatusUnauthorized)
			return
		}
		if a.federationInboxProc != nil {
			recipients := a.federationInboxProc.ResolveSharedInboxRecipients(activity)
			for _, nick := range recipients {
				if err := a.federationInboxProc.Handle(nick, strings.NewReader(string(body))); err != nil {
					slog.Warn("federation shared inbox handle failed", "nickname", nick, "err", err)
				}
			}
		}
		ctx.Status(http.StatusAccepted)
	}
}

func (a *App) outboxHandler() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		if err := a.requireAuthorizedFetch(ctx); err != nil {
			ctx.Status(http.StatusUnauthorized)
			return
		}
		ctx.Header("Vary", "Signature, Accept")
		ctx.Header("Content-Type", "application/activity+json")
		ctx.JSON(http.StatusOK, gin.H{
			"@context":     "https://www.w3.org/ns/activitystreams",
			"id":           actorURL(ctx.Param("nickname")) + "/outbox",
			"type":         "OrderedCollection",
			"totalItems":   0,
			"orderedItems": []any{},
		})
	}
}

func (a *App) publicAvatarURL(nickname string) string {
	return avatars.PublicURL(publicDomain(), nickname)
}

func wantsActivityJSON(ctx *gin.Context) bool {
	accept := ctx.GetHeader("Accept")
	return strings.Contains(accept, "application/activity+json") ||
		strings.Contains(accept, "application/ld+json")
}

func (a *App) requireAuthorizedFetch(ctx *gin.Context) error {
	if !fed.AuthorizedFetchRequired() {
		return nil
	}
	body, err := httpsig.ReadBody(ctx.Request)
	if err != nil {
		return err
	}
	if a.federationKeyResolver == nil {
		return errors.New("key resolver unavailable")
	}
	_, _, err = fed.AuthenticateRequest(ctx.Request, body, a.federationKeyResolver)
	return err
}

func (a *App) authenticateFederationActivity(ctx *gin.Context, body []byte, activity map[string]any) error {
	if a.federationKeyResolver == nil {
		return errors.New("key resolver unavailable")
	}
	return fed.AuthenticateActivity(ctx.Request, body, activity, a.federationKeyResolver)
}

func decodeJSONObject(body []byte) (map[string]any, error) {
	var activity map[string]any
	if err := json.Unmarshal(body, &activity); err != nil {
		return nil, err
	}
	if activity == nil {
		return nil, errors.New("empty activity")
	}
	return activity, nil
}
