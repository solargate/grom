package v1

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/solargate/travka/internal/avatars"
	"github.com/solargate/travka/internal/config"
	fed "github.com/solargate/travka/internal/federation"
	"github.com/solargate/travka/internal/data"
	"github.com/solargate/travka/internal/users"
)

var federationUserStore *users.Store

func RegisterFederationRoutes(router *gin.Engine, userStore *users.Store) {
	federationUserStore = userStore
	router.GET("/.well-known/webfinger", webfingerHandler(userStore))
	router.GET("/users/:nickname/avatar", publicAvatarHandler(userStore))
	router.GET("/users/:nickname", actorHandler(userStore))
	router.POST("/users/:nickname/inbox", inboxHandler(userStore))
	router.GET("/users/:nickname/outbox", outboxHandler())
	router.POST("/inbox", sharedInboxHandler())
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

func webfingerHandler(userStore *users.Store) gin.HandlerFunc {
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
		user, err := userStore.FindByNickname(parts[0])
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

func actorHandler(userStore *users.Store) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		if !strings.Contains(ctx.GetHeader("Accept"), "application/activity+json") &&
			!strings.Contains(ctx.GetHeader("Accept"), "application/ld+json") {
			ctx.Status(http.StatusNotFound)
			return
		}
		user, err := userStore.FindByNickname(ctx.Param("nickname"))
		if err != nil {
			ctx.Status(http.StatusNotFound)
			return
		}
		pubKey, keyID, err := fed.LoadOrCreateActorKey(user.Nickname)
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
		if avatars.Has(config.Cfg.Data.ResolvedDir, user.Nickname) {
			response["icon"] = gin.H{
				"type":      "Image",
				"mediaType": "image/webp",
				"url":       publicAvatarURL(user.Nickname),
			}
		}
		ctx.JSON(http.StatusOK, response)
	}
}

func publicAvatarHandler(userStore *users.Store) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		nickname := strings.TrimSpace(ctx.Param("nickname"))
		if nickname == "" {
			ctx.Status(http.StatusNotFound)
			return
		}
		if _, err := userStore.FindByNickname(nickname); err != nil {
			ctx.Status(http.StatusNotFound)
			return
		}
		if !avatars.Has(config.Cfg.Data.ResolvedDir, nickname) {
			ctx.Status(http.StatusNotFound)
			return
		}
		path := data.UserAvatarPath(config.Cfg.Data.ResolvedDir, nickname)
		ctx.Header("Cache-Control", "public, max-age=86400")
		ctx.Header("Content-Type", "image/webp")
		ctx.File(path)
	}
}

func inboxHandler(userStore *users.Store) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		if err := initFederation(); err != nil {
			ctx.Status(http.StatusInternalServerError)
			return
		}
		nickname := ctx.Param("nickname")
		if federationInboxProc != nil {
			if err := federationInboxProc.Handle(nickname, ctx.Request.Body); err != nil {
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

func sharedInboxHandler() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		ctx.Status(http.StatusAccepted)
	}
}

func outboxHandler() gin.HandlerFunc {
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
