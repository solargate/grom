package v1

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/solargate/grom/internal/avatars"
	"github.com/solargate/grom/internal/config"
	fed "github.com/solargate/grom/internal/federation"
)

func (a *App) registerFederationRoutes(router *gin.Engine) {
	router.GET("/.well-known/webfinger", a.webfingerHandler())
	router.GET("/users/:nickname/avatar", a.publicAvatarHandler())
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

func (a *App) actorHandler() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		if !strings.Contains(ctx.GetHeader("Accept"), "application/activity+json") &&
			!strings.Contains(ctx.GetHeader("Accept"), "application/ld+json") {
			ctx.Status(http.StatusNotFound)
			return
		}
		user, err := a.Users.FindByNickname(ctx.Param("nickname"))
		if err != nil {
			ctx.Status(http.StatusNotFound)
			return
		}
		pubKey, keyID, err := fed.LoadOrCreateActorKey(a.Blobs, user.Nickname)
		if err != nil {
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
		if a.federationInboxProc != nil {
			if err := a.federationInboxProc.Handle(nickname, ctx.Request.Body); err != nil {
				ctx.Status(http.StatusBadRequest)
				return
			}
		} else {
			var activity map[string]any
			_ = json.NewDecoder(ctx.Request.Body).Decode(&activity)
		}
		ctx.Status(http.StatusAccepted)
	}
}

func (a *App) sharedInboxHandler() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		ctx.Status(http.StatusAccepted)
	}
}

func (a *App) outboxHandler() gin.HandlerFunc {
	return func(ctx *gin.Context) {
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
