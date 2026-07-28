package v1

import (
	"log/slog"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/solargate/grom/internal/auth"
	"github.com/solargate/grom/internal/config"
	"github.com/solargate/grom/internal/equipment"
	"github.com/solargate/grom/internal/equipment/distance"
	"github.com/solargate/grom/internal/federation"
	"github.com/solargate/grom/internal/integrations/strava"
	"github.com/solargate/grom/internal/social"
	"github.com/solargate/grom/internal/storage"
	"github.com/solargate/grom/internal/storage/blob"
	"github.com/solargate/grom/internal/users"
	"github.com/solargate/grom/internal/workouts"
)

type App struct {
	Backend  storage.Backend
	Users    users.Repository
	Workouts *workouts.Service
	Equipment equipment.Repository
	EquipmentDistance *distance.Service
	Social   *social.Service
	Federation federation.Storage
	Blobs    blob.Store
	Location string
	TempDir  string

	federationOnce      sync.Once
	federationDelivery  *federation.Delivery
	federationInboxProc *federation.InboxProcessor

	stravaOnce sync.Once
	stravaJobs *strava.JobManager
}

func NewApp() (*App, error) {
	backend, err := storage.Open(config.Cfg.Storage)
	if err != nil {
		return nil, err
	}

	socialSvc := social.NewService(backend.Users(), backend.Social(), backend.Blobs())
	workoutSvc := backend.Workouts()
	app := &App{
		Backend:    backend,
		Users:      backend.Users(),
		Workouts:   workoutSvc,
		Equipment:  backend.Equipment(),
		EquipmentDistance: distance.NewService(backend.Equipment(), workoutSvc),
		Social:     socialSvc,
		Federation: backend.Federation(),
		Blobs:      backend.Blobs(),
		Location:   config.Cfg.Storage.ResolvedLocation,
		TempDir:    config.Cfg.Storage.ResolvedTempDir,
	}

	socialSvc.SetInboundFollowers(federation.NewInboundFollowersAdapter(app.Federation.Followers()))

	if config.Cfg.Federation.Enabled {
		delivery, err := federation.NewDelivery(app.Users, socialSvc)
		if err != nil {
			_ = backend.Close()
			return nil, err
		}
		socialSvc.SetDelivery(delivery)
		app.federationDelivery = delivery
		app.Federation.Inbox().SetHTTPClient(delivery.Client())
		app.federationInboxProc = federation.NewInboxProcessor(
			app.Users,
			socialSvc,
			delivery,
			app.Federation.Inbox(),
			app.Federation.Followers(),
		)
		slog.Info("federation enabled",
			"domain", config.Cfg.Federation.Domain,
			"auto_accept_follows", config.Cfg.Federation.AutoAcceptFollows,
		)
	}

	return app, nil
}

func (a *App) Close() error {
	if a.Backend == nil {
		return nil
	}
	return a.Backend.Close()
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

		authGroup := apiV1.Group("/auth")
		authGroup.POST("/register", a.register)
		authGroup.POST("/login", a.login)
		authGroup.GET("/me", auth.AuthRequired(), a.getMe)
		authGroup.PATCH("/me", auth.AuthRequired(), a.updateMe)
		authGroup.PUT("/me/avatar", auth.AuthRequired(), a.uploadMyAvatar)
		authGroup.DELETE("/me/avatar", auth.AuthRequired(), a.deleteMyAvatar)

		apiV1.GET("/users/search", auth.AuthRequired(), a.searchUsers)
		apiV1.GET("/users/:nickname/avatar", auth.AuthRequired(), a.getUserAvatar)
		apiV1.GET("/federation/authors/:ownerKey/avatar", auth.AuthRequired(), a.getFederatedAuthorAvatar)

		socialGroup := apiV1.Group("/social", auth.AuthRequired())
		socialGroup.POST("/follow", a.followUser)
		socialGroup.DELETE("/follow/:id", a.unfollowUser)
		socialGroup.GET("/following", a.listFollowing)
		socialGroup.GET("/followers", a.listFollowers)

		workoutGroup := apiV1.Group("/workouts", auth.AuthRequired())
		workoutGroup.POST("", a.createWorkout)
		workoutGroup.POST("/parse-track", a.parseTrack)
		workoutGroup.GET("/:id/track", a.getWorkoutTrack)
		workoutGroup.GET("/:id/speed", a.getWorkoutSpeed)
		workoutGroup.GET("/:id/map-preview", a.getWorkoutMapPreview)
		workoutGroup.GET("/:id/media/:filename/preview", a.getWorkoutMediaPreview)
		workoutGroup.GET("/:id/media/:filename", a.getWorkoutMediaOriginal)
		workoutGroup.GET("/:id", a.getWorkout)
		workoutGroup.PUT("/:id", a.updateWorkout)
		workoutGroup.DELETE("/:id", a.deleteWorkout)
		workoutGroup.GET("", a.listWorkouts)

		equipmentGroup := apiV1.Group("/equipment", auth.AuthRequired())
		equipmentGroup.GET("", a.listEquipment)
		equipmentGroup.POST("", a.createEquipment)
		equipmentGroup.PUT("/:id", a.updateEquipment)
		equipmentGroup.DELETE("/:id", a.deleteEquipment)

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
