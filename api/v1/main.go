package v1

import (
	"strconv"

	"github.com/gin-gonic/gin"

	_ "github.com/solargate/travka/api/docs"
	"github.com/solargate/travka/config"

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
func RunRouter() {
	//gin.SetMode(gin.ReleaseMode)
	router := gin.Default()

	apiV1 := router.Group("/api/v1")
	{
		apiV1.GET("/status", checkStatus)
		apiV1.GET("/server_info", getServerInfo)
	}

	router.GET("/api/docs/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	router.Run(":" + strconv.Itoa(config.Cfg.Server.Port))
}
