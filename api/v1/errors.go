package v1

import (
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	sloggin "github.com/samber/slog-gin"
)

// logInternalError records an unexpected server-side failure with request context.
func logInternalError(ctx *gin.Context, msg string, err error) {
	attrs := []any{"err", err}
	if ctx != nil && ctx.Request != nil {
		attrs = append(attrs, "method", ctx.Request.Method)
		if path := ctx.FullPath(); path != "" {
			attrs = append(attrs, "path", path)
		} else {
			attrs = append(attrs, "path", ctx.Request.URL.Path)
		}
		if rid := sloggin.GetRequestID(ctx); rid != "" {
			attrs = append(attrs, "request_id", rid)
		}
	}
	slog.Error(msg, attrs...)
}

// respondInternal logs an unexpected error and returns a generic 500 JSON body.
func respondInternal(ctx *gin.Context, publicMsg string, err error) {
	logInternalError(ctx, publicMsg, err)
	ctx.JSON(http.StatusInternalServerError, ErrorResponse{Error: publicMsg})
}
