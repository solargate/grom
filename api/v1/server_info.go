package v1

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/solargate/travka/internal/config"
)

// statusCheck godoc
// @Summary      Server status
// @Description  Get server status
// @Tags         server_info
// @Produce      json
// @Success      200
// @Router       /status [get]
func checkStatus(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, gin.H{
		"message": "OK",
	})
}

// getServerInfo godoc
// @Summary      Server info
// @Description  Get server info
// @Tags         server_info
// @Produce      json
// @Success      200
// @Router       /server_info [get]
func getServerInfo(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, gin.H{
		"name": config.Cfg.Server.Name,
	})
}
