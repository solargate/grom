package v1

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/solargate/grom/internal/config"
)

// statusCheck godoc
// @Summary      Server status
// @Description  Get server status
// @Tags         server-info
// @Produce      json
// @Success      200
// @Router       /status [get]
func (a *App) checkStatus(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, gin.H{
		"message": "OK",
	})
}

// getServerInfo godoc
// @Summary      Server info
// @Description  Get server info
// @Tags         server-info
// @Produce      json
// @Success      200
// @Router       /server-info [get]
func (a *App) getServerInfo(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, gin.H{
		"name":               config.Cfg.Server.Name,
		"federation_enabled": config.Cfg.Federation.Enabled,
	})
}
