package v1

import (
	"log/slog"
	"net/http"
	"runtime/debug"

	"github.com/gin-gonic/gin"
	sloggin "github.com/samber/slog-gin"

	"github.com/solargate/grom/api/docs"
	"github.com/solargate/grom/internal/config"
	"github.com/solargate/grom/internal/logging"
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
// @BasePath  /api/v1
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and a JWT session token or personal access token (`grom_pat_...`).
func RunRouter() {
	app, err := NewApp()
	if err != nil {
		slog.Error("failed to initialize application", "err", err)
		panic(err)
	}

	defaultLevel := slog.LevelInfo
	if lvl, err := logging.ParseLevel(config.Cfg.Logging.Level); err == nil && lvl <= slog.LevelDebug {
		defaultLevel = slog.LevelDebug
	}

	router := gin.New()
	router.Use(sloggin.NewWithConfig(slog.Default(), sloggin.Config{
		DefaultLevel:     defaultLevel,
		ClientErrorLevel: slog.LevelWarn,
		ServerErrorLevel: slog.LevelError,
		WithRequestID:    true,
		Filters: []sloggin.Filter{
			sloggin.IgnorePathSuffix(
				".js", ".css", ".map", ".ico", ".png", ".jpg", ".jpeg", ".gif", ".webp",
				".woff", ".woff2", ".ttf", ".svg",
			),
		},
	}))
	router.Use(gin.CustomRecovery(func(c *gin.Context, recovered any) {
		attrs := []any{
			"panic", recovered,
			"method", c.Request.Method,
			"path", c.Request.URL.Path,
			"stack", string(debug.Stack()),
		}
		if rid := sloggin.GetRequestID(c); rid != "" {
			attrs = append(attrs, "request_id", rid)
		}
		slog.Error("http panic recovered", attrs...)
		c.AbortWithStatus(http.StatusInternalServerError)
	}))
	router.MaxMultipartMemory = 128 << 20

	app.RegisterRoutes(router)

	RegisterAPIDocs(router)

	web.RegisterRoutes(router)

	slog.Info("HTTP server listening",
		"tls_mode", config.Cfg.Server.TLS.Mode,
		"http_port", config.Cfg.Server.Port,
		"https_port", config.Cfg.Server.TLS.Port,
	)
	if err := server.Run(router); err != nil {
		slog.Error("HTTP server stopped", "err", err)
		panic(err)
	}
}

// RegisterAPIDocs mounts Swagger UI under /api/docs with a slash redirect.
// Host is left empty so Try it out uses the browser's current origin (localhost, IP, or domain).
func RegisterAPIDocs(router *gin.Engine) {
	docs.SwaggerInfo.Host = ""
	swaggerHandler := ginSwagger.WrapHandler(swaggerFiles.Handler)
	serveSwaggerUI := func(c *gin.Context) {
		c.Request.URL.Path = "/api/docs/index.html"
		c.Request.RequestURI = "/api/docs/index.html"
		swaggerHandler(c)
	}
	// Redirect so the browser URL ends with "/", otherwise relative Swagger UI
	// asset paths (./swagger-ui.css etc.) resolve under /api/ and the page is blank.
	router.GET("/api/docs", func(c *gin.Context) {
		c.Redirect(http.StatusMovedPermanently, "/api/docs/")
	})
	router.GET("/api/docs/*any", func(c *gin.Context) {
		switch c.Param("any") {
		case "", "/":
			serveSwaggerUI(c)
		default:
			swaggerHandler(c)
		}
	})
}
