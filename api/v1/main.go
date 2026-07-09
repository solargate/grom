package v1

import (
	"github.com/gin-gonic/gin"

	_ "github.com/solargate/travka/api/docs"
	"github.com/solargate/travka/internal/auth"
	"github.com/solargate/travka/internal/config"
	"github.com/solargate/travka/internal/server"
	"github.com/solargate/travka/internal/web"

	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// @title          Travka API
// @version        1.0
// @description    Travka API documentation server
// @contact.name   Alexander Cheryomukhin
// @contact.email  solarwind.palm@gmail.com
// @license.name  Apache 2.0
// @license.url   http://www.apache.org/licenses/LICENSE-2.0.html
// @host      localhost:8080
// @BasePath  /api/v1
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and JWT token.
func RunRouter() {
	if err := initUserStore(); err != nil {
		panic(err)
	}

	router := gin.Default()
	router.MaxMultipartMemory = 20 << 20

	if config.Cfg.Federation.Enabled {
		RegisterFederationRoutes(router, userStore)
	}

	apiV1 := router.Group("/api/v1")
	{
		apiV1.GET("/status", checkStatus)
		apiV1.GET("/server_info", getServerInfo)

		authGroup := apiV1.Group("/auth")
		authGroup.POST("/register", register)
		authGroup.POST("/login", login)
		authGroup.GET("/me", auth.AuthRequired(), getMe)

		apiV1.GET("/users/search", auth.AuthRequired(), searchUsers)

		socialGroup := apiV1.Group("/social", auth.AuthRequired())
		socialGroup.POST("/follow", followUser)
		socialGroup.DELETE("/follow/:id", unfollowUser)
		socialGroup.GET("/following", listFollowing)

		workoutGroup := apiV1.Group("/workouts", auth.AuthRequired())
		workoutGroup.POST("", createWorkout)
		workoutGroup.POST("/parse-track", parseTrack)
		workoutGroup.GET("/:id/track", getWorkoutTrack)
		workoutGroup.GET("/:id/map-preview", getWorkoutMapPreview)
		workoutGroup.GET("", listWorkouts)
	}

	swaggerHandler := ginSwagger.WrapHandler(swaggerFiles.Handler)
	serveSwaggerUI := func(c *gin.Context) {
		c.Request.URL.Path = "/api/docs/index.html"
		c.Request.RequestURI = "/api/docs/index.html"
		swaggerHandler(c)
	}
	router.GET("/api/docs", serveSwaggerUI)
	router.GET("/api/docs/*any", func(c *gin.Context) {
		switch c.Param("any") {
		case "", "/":
			serveSwaggerUI(c)
		default:
			swaggerHandler(c)
		}
	})

	web.RegisterRoutes(router)

	if err := server.Run(router); err != nil {
		panic(err)
	}
}
