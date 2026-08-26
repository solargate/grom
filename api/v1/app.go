package v1

import (
	"log/slog"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/solargate/grom/internal/auth"
	"github.com/solargate/grom/internal/auth/captcha"
	"github.com/solargate/grom/internal/auth/pat"
	"github.com/solargate/grom/internal/auth/reset"
	"github.com/solargate/grom/internal/config"
	"github.com/solargate/grom/internal/equipment"
	"github.com/solargate/grom/internal/equipment/distance"
	"github.com/solargate/grom/internal/federation"
	"github.com/solargate/grom/internal/integrations/strava"
	"github.com/solargate/grom/internal/mailer"
	"github.com/solargate/grom/internal/social"
	"github.com/solargate/grom/internal/storage"
	"github.com/solargate/grom/internal/storage/blob"
	"github.com/solargate/grom/internal/users"
	"github.com/solargate/grom/internal/workouts"
)

type App struct {
	Backend           storage.Backend
	Users             users.Repository
	Workouts          *workouts.Service
	Likes             workouts.LikesRepository
	Comments          workouts.CommentsRepository
	Equipment         equipment.Repository
	EquipmentDistance *distance.Service
	Social            *social.Service
	Federation        federation.Storage
	Blobs             blob.Store
	Mailer            mailer.Mailer
	PasswordReset     *reset.Service
	Captcha           *captcha.Service
	PAT               *pat.Service
	Location          string
	TempDir           string

	federationOnce      sync.Once
	federationDelivery  *federation.Delivery
	federationInboxProc *federation.InboxProcessor
	federationKeyResolver federation.KeyResolver
	federationHTTPKeys    *federation.HTTPKeyResolver

	stravaOnce sync.Once
	stravaJobs *strava.JobManager

	profileRefreshWG sync.WaitGroup
	// profileSportsMu serializes TouchUsedSportType with RefreshLastSportType so an
	// async prune cannot drop a sport type that was touched after a stale List.
	profileSportsMu sync.Mutex
}

func NewApp() (*App, error) {
	backend, err := storage.Open(config.Cfg.Storage)
	if err != nil {
		return nil, err
	}

	socialSvc := social.NewService(backend.Users(), backend.Social(), backend.Blobs())
	workoutSvc := backend.Workouts()

	mail, err := mailer.New(config.Cfg.Mailer)
	if err != nil {
		_ = backend.Close()
		return nil, err
	}

	var passwordReset *reset.Service
	if config.Cfg.PasswordResetEnabled() {
		passwordReset = reset.NewService(
			backend.Users(),
			backend.ResetTokens(),
			mail,
			reset.Config{
				PublicBaseURL: config.Cfg.Auth.Reset.PublicBaseURL,
				TokenTTL:      time.Duration(config.Cfg.Auth.Reset.TokenTTLMinutes) * time.Minute,
				ServerName:    config.Cfg.Server.Name,
				Enabled:       true,
			},
		)
	}

	captchaSvc := captcha.NewService(captcha.Config{
		Enabled:    config.Cfg.CaptchaEnabled(),
		HMACSecret: config.Cfg.CaptchaHMACSecret(),
		Cost:       config.Cfg.Auth.Captcha.Cost,
		Expires:    time.Duration(config.Cfg.Auth.Captcha.ExpiresSeconds) * time.Second,
	})

	app := &App{
		Backend:           backend,
		Users:             backend.Users(),
		Workouts:          workoutSvc,
		Likes:             backend.Likes(),
		Comments:          backend.Comments(),
		Equipment:         backend.Equipment(),
		EquipmentDistance: distance.NewService(backend.Equipment(), workoutSvc),
		Social:            socialSvc,
		Federation:        backend.Federation(),
		Blobs:             backend.Blobs(),
		Mailer:            mail,
		PasswordReset:     passwordReset,
		Captcha:           captchaSvc,
		PAT:               pat.NewService(backend.PAT()),
		Location:          config.Cfg.Storage.ResolvedLocation,
		TempDir:           config.Cfg.Storage.ResolvedTempDir,
	}

	socialSvc.SetInboundFollowers(federation.NewInboundFollowersAdapter(app.Federation.Followers()))

	if config.Cfg.Federation.Enabled {
		delivery, err := federation.NewDelivery(app.Users, socialSvc, app.Blobs)
		if err != nil {
			_ = backend.Close()
			return nil, err
		}
		socialSvc.SetDelivery(delivery)
		app.federationDelivery = delivery
		app.Federation.Inbox().SetHTTPClient(delivery.Client())
		app.federationHTTPKeys = federation.NewHTTPKeyResolver(delivery.Client(), app.Blobs)
		app.federationKeyResolver = app.federationHTTPKeys
		app.federationInboxProc = federation.NewInboxProcessor(
			app.Users,
			socialSvc,
			delivery,
			app.Federation.Inbox(),
			app.Federation.Followers(),
		)
		app.federationInboxProc.SetLikes(app.Likes, app.publishWorkoutLikesUpdate)
		app.federationInboxProc.SetComments(app.Comments, app.publishWorkoutCommentsUpdate)
		slog.Info("federation enabled",
			"domain", config.Cfg.Federation.Domain,
			"auto_accept_follows", config.Cfg.Federation.AutoAcceptFollows,
			"authorized_fetch", config.Cfg.AuthorizedFetchEnabled(),
		)
	}

	return app, nil
}

func (a *App) Close() error {
	a.profileRefreshWG.Wait()
	if a.EquipmentDistance != nil {
		a.EquipmentDistance.Wait()
	}
	if a.Backend == nil {
		return nil
	}
	return a.Backend.Close()
}

// SetFederationHTTPClient replaces the client used for ActivityPub delivery and
// federated avatar fetches. Intended for tests.
func (a *App) SetFederationHTTPClient(client *http.Client) {
	if a == nil || client == nil {
		return
	}
	if a.federationDelivery != nil {
		a.federationDelivery.SetClient(client)
	}
	if a.federationHTTPKeys != nil {
		a.federationHTTPKeys.SetClient(client)
	}
	if a.Federation != nil && a.Federation.Inbox() != nil {
		a.Federation.Inbox().SetHTTPClient(client)
	}
}

// SetFederationKeyResolver replaces the key resolver used for inbound HTTP Signature verification.
// Intended for tests.
func (a *App) SetFederationKeyResolver(resolver federation.KeyResolver) {
	if a == nil || resolver == nil {
		return
	}
	a.federationKeyResolver = resolver
}

func (a *App) newFeedService() *workouts.FeedService {
	feedSvc := workouts.NewFeedService(a.Workouts, a.Blobs, config.Cfg.Federation.Domain)
	feedSvc.SetFederatedSource(federatedFeedAdapter{store: a.Federation.Inbox()})
	return feedSvc
}

func (a *App) stravaJobManager() *strava.JobManager {
	a.stravaOnce.Do(func() {
		a.stravaJobs = strava.NewJobManager(
			a.TempDir,
			a.Workouts,
			a.Equipment,
			a.publishCreatedWorkout,
			a.EquipmentDistance,
			func(userID, nickname string, workout *workouts.Workout) {
				a.touchLastEquipmentFromWorkout(userID, workout)
				a.touchUsedSportFromWorkout(userID, workout)
			},
			func(userID, nickname string) {
				a.scheduleRefreshLastSportType(nickname, userID)
			},
		)
	})
	return a.stravaJobs
}

func (a *App) RegisterRoutes(router *gin.Engine) {
	if config.Cfg.Federation.Enabled {
		a.registerFederationRoutes(router)
	}

	apiV1 := router.Group("/api/v1")
	{
		apiV1.GET("/status", a.checkStatus)
		apiV1.GET("/server-info", a.getServerInfo)
		apiV1.GET("/captcha/challenge", a.getCaptchaChallenge)

		authGroup := apiV1.Group("/auth")
		authGroup.POST("/register", a.register)
		authGroup.POST("/login", a.login)
		authGroup.POST("/password/forgot", a.forgotPassword)
		authGroup.POST("/password/reset", a.resetPassword)
		authGroup.GET("/me", auth.AuthRequired(), a.getMe)
		authGroup.PATCH("/me", auth.AuthRequired(), a.updateMe)
		authGroup.DELETE("/me", auth.AuthRequired(), a.deleteMe)
		authGroup.PUT("/me/avatar", auth.AuthRequired(), a.uploadMyAvatar)
		authGroup.DELETE("/me/avatar", auth.AuthRequired(), a.deleteMyAvatar)
		authGroup.GET("/pat", auth.AuthRequired(), a.listPAT)
		authGroup.POST("/pat", auth.AuthRequired(), a.createPAT)
		authGroup.DELETE("/pat/:id", auth.AuthRequired(), a.revokePAT)

		apiV1.GET("/profile", auth.AuthRequired(), a.getProfile)

		apiV1.GET("/users/search", auth.AuthRequired(), a.searchUsers)
		apiV1.GET("/users/:nickname/avatar", auth.AuthRequired(), a.getUserAvatar)
		apiV1.GET("/federation/authors/:ownerKey/avatar", auth.AuthRequired(), a.getFederatedAuthorAvatar)

		socialGroup := apiV1.Group("/social", auth.AuthRequired())
		socialGroup.POST("/follow", a.followUser)
		socialGroup.DELETE("/follow/:id", a.unfollowUser)
		socialGroup.GET("/following", a.listFollowing)
		socialGroup.GET("/followers", a.listFollowers)

		workoutRead := auth.AuthAPI(a.PAT, pat.ScopeWorkoutsRead)
		workoutWrite := auth.AuthAPI(a.PAT, pat.ScopeWorkoutsWrite)

		workoutGroup := apiV1.Group("/workouts")
		workoutGroup.POST("", workoutWrite, a.createWorkout)
		workoutGroup.POST("/parse-track", workoutWrite, a.parseTrack)
		workoutGroup.GET("/external", workoutRead, a.checkWorkoutExternalID)
		workoutGroup.GET("/:id/track", workoutRead, a.getWorkoutTrack)
		workoutGroup.GET("/:id/speed", workoutRead, a.getWorkoutSpeed)
		workoutGroup.GET("/:id/heartrate", workoutRead, a.getWorkoutHeartRate)
		workoutGroup.GET("/:id/map-preview", workoutRead, a.getWorkoutMapPreview)
		workoutGroup.GET("/:id/media/:filename/preview", workoutRead, a.getWorkoutMediaPreview)
		workoutGroup.GET("/:id/media/:filename", workoutRead, a.getWorkoutMediaOriginal)
		workoutGroup.GET("/:id/likes", auth.AuthRequired(), a.getWorkoutLikes)
		workoutGroup.POST("/:id/likes", auth.AuthRequired(), a.likeWorkout)
		workoutGroup.DELETE("/:id/likes", auth.AuthRequired(), a.unlikeWorkout)
		workoutGroup.GET("/:id/comments", auth.AuthRequired(), a.getWorkoutComments)
		workoutGroup.POST("/:id/comments", auth.AuthRequired(), a.createWorkoutComment)
		workoutGroup.DELETE("/:id/comments/:commentId", auth.AuthRequired(), a.deleteWorkoutComment)
		workoutGroup.POST("/:id/media", workoutWrite, a.addWorkoutMedia)
		workoutGroup.DELETE("/:id/media/:filename", workoutWrite, a.deleteWorkoutMedia)
		workoutGroup.GET("/:id", workoutRead, a.getWorkout)
		workoutGroup.PUT("/:id", workoutWrite, a.updateWorkout)
		workoutGroup.DELETE("/:id", workoutWrite, a.deleteWorkout)
		workoutGroup.GET("", workoutRead, a.listWorkouts)

		equipmentRead := auth.AuthAPI(a.PAT, pat.ScopeEquipmentRead)
		equipmentWrite := auth.AuthAPI(a.PAT, pat.ScopeEquipmentWrite)

		equipmentGroup := apiV1.Group("/equipment")
		equipmentGroup.GET("", equipmentRead, a.listEquipment)
		equipmentGroup.POST("", equipmentWrite, a.createEquipment)
		equipmentGroup.PUT("/:id", equipmentWrite, a.updateEquipment)
		equipmentGroup.DELETE("/:id", equipmentWrite, a.deleteEquipment)

		integrationsGroup := apiV1.Group("/integrations", auth.AuthRequired())
		integrationsGroup.POST("/strava/import", a.importStravaArchive)
		integrationsGroup.GET("/strava/import/status", a.getStravaImportStatus)
	}
}

type federatedFeedAdapter struct {
	store federation.InboxRepository
}

func (a federatedFeedAdapter) ListFederated(viewerNickname string) ([]workouts.FeedWorkout, error) {
	return a.store.List(viewerNickname)
}

func (a federatedFeedAdapter) ListFederatedPage(viewerNickname string, cursor *workouts.Cursor, limit int) ([]workouts.FeedWorkout, bool, error) {
	return a.store.ListPage(viewerNickname, cursor, limit)
}
