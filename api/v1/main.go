package v1

import (
	"github.com/gin-gonic/gin"

	_ "github.com/solargate/grom/api/docs"
	"github.com/solargate/grom/internal/server"
	"github.com/solargate/grom/internal/web"

	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// @title          Grom API
// @version        1.0
// @description    Grom API documentation server
// @contact.name   Alexander Cheryomukhin
// @contact.email  solarwind.palm@gmail.com
// @license.name   GPL-3.0
// @license.url    https://www.gnu.org/licenses/gpl-3.0.html
// @host      localhost:8080
// @BasePath  /api/v1
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and JWT token.
func RunRouter() {
	app, err := NewApp()
	if err != nil {
		panic(err)
	}

	router := gin.Default()
	router.MaxMultipartMemory = 128 << 20

	app.RegisterRoutes(router)

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
