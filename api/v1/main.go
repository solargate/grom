package v1

import (
	"net/http"

	"github.com/gin-gonic/gin"

	_ "github.com/solargate/trava/api/docs"

	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// statusCheck godoc
// @Summary      Server status
// @Description  Get server status
// @Produce      json
// @Success      200
// @Router       /status [get]
func checkStatus(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, gin.H{
		"message": "OK",
	})
}

// @title          Trava API
// @version        1.0
// @description    Trava API documentation server
// @contact.name   Alexander Cheryomukhin
// @contact.email  solarwind.palm@gmail.com
// @license.name  Apache 2.0
// @license.url   http://www.apache.org/licenses/LICENSE-2.0.html
// @host      localhost:8080
// @BasePath  /api/v1
func RunRouter() {
	router := gin.Default()
	router.GET("/api/v1/status", checkStatus)

	router.GET("/api/docs/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	router.Run()
}
