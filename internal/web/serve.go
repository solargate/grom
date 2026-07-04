package web

import (
	"io/fs"
	"mime"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
)

func RegisterRoutes(router *gin.Engine) {
	sub, err := fs.Sub(Assets, "dist")
	if err != nil {
		panic(err)
	}

	serveUI := func(c *gin.Context) {
		if strings.HasPrefix(c.Request.URL.Path, "/api/") {
			c.Status(http.StatusNotFound)
			return
		}

		path := strings.TrimPrefix(c.Request.URL.Path, "/")
		if path == "" || strings.HasSuffix(path, "/") {
			serveFile(c, sub, "index.html")
			return
		}

		if _, err := sub.Open(path); err != nil {
			serveFile(c, sub, "index.html")
			return
		}

		serveFile(c, sub, path)
	}

	router.GET("/", serveUI)
	router.NoRoute(serveUI)
}

func serveFile(c *gin.Context, fsys fs.FS, name string) {
	data, err := fs.ReadFile(fsys, name)
	if err != nil {
		c.Status(http.StatusNotFound)
		return
	}

	contentType := mime.TypeByExtension(filepath.Ext(name))
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	c.Data(http.StatusOK, contentType, data)
}
