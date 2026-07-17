package v1

import (
	"errors"
	"io"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/solargate/grom/internal/avatars"
)

// uploadMyAvatar godoc
// @Summary      Upload current user avatar
// @Description  Upload a square avatar image (webp, png, or jpeg)
// @Tags         auth
// @Accept       multipart/form-data
// @Produce      json
// @Security     BearerAuth
// @Param        avatar  formData  file  true  "Avatar image"
// @Success      200  {object}  UserResponse
// @Failure      400  {object}  ErrorResponse
// @Failure      401  {object}  ErrorResponse
// @Router       /auth/me/avatar [put]
func (a *App) uploadMyAvatar(ctx *gin.Context) {
	userID, err := a.currentUserID(ctx)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, ErrorResponse{Error: "invalid token"})
		return
	}

	user, err := a.Users.FindByID(userID)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, ErrorResponse{Error: "user not found"})
		return
	}

	raw, err := readAvatarUpload(ctx)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	if err := avatars.SaveStore(a.Blobs, user.Nickname, raw); err != nil {
		switch {
		case errors.Is(err, avatars.ErrInvalidAvatar),
			errors.Is(err, avatars.ErrAvatarTooLarge),
			errors.Is(err, avatars.ErrAvatarNotSquare):
			ctx.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		default:
			ctx.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to save avatar"})
		}
		return
	}

	ctx.JSON(http.StatusOK, a.toUserResponse(user))
}

// deleteMyAvatar godoc
// @Summary      Delete current user avatar
// @Description  Remove avatar for the authenticated user
// @Tags         auth
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  UserResponse
// @Failure      401  {object}  ErrorResponse
// @Failure      404  {object}  ErrorResponse
// @Router       /auth/me/avatar [delete]
func (a *App) deleteMyAvatar(ctx *gin.Context) {
	userID, err := a.currentUserID(ctx)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, ErrorResponse{Error: "invalid token"})
		return
	}

	user, err := a.Users.FindByID(userID)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, ErrorResponse{Error: "user not found"})
		return
	}

	if err := avatars.DeleteStore(a.Blobs, user.Nickname); err != nil {
		if errors.Is(err, avatars.ErrAvatarNotFound) {
			ctx.JSON(http.StatusNotFound, ErrorResponse{Error: "avatar not found"})
			return
		}
		ctx.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to delete avatar"})
		return
	}

	ctx.JSON(http.StatusOK, a.toUserResponse(user))
}

// getUserAvatar godoc
// @Summary      Get user avatar
// @Description  Return avatar image for a local user
// @Tags         users
// @Produce      image/webp
// @Security     BearerAuth
// @Param        nickname  path  string  true  "User nickname"
// @Success      200  {file}  binary
// @Failure      401  {object}  ErrorResponse
// @Failure      404  {object}  ErrorResponse
// @Router       /users/{nickname}/avatar [get]
func (a *App) getUserAvatar(ctx *gin.Context) {
	if _, err := a.currentUserNickname(ctx); err != nil {
		ctx.JSON(http.StatusUnauthorized, ErrorResponse{Error: "user not found"})
		return
	}

	nickname := strings.TrimSpace(ctx.Param("nickname"))
	if nickname == "" {
		ctx.JSON(http.StatusNotFound, ErrorResponse{Error: "avatar not found"})
		return
	}

	if _, err := a.Users.FindByNickname(nickname); err != nil {
		ctx.JSON(http.StatusNotFound, ErrorResponse{Error: "avatar not found"})
		return
	}

	if !avatars.HasStore(a.Blobs, nickname) {
		ctx.JSON(http.StatusNotFound, ErrorResponse{Error: "avatar not found"})
		return
	}

	data, err := avatars.LoadStore(a.Blobs, nickname)
	if err != nil {
		ctx.JSON(http.StatusNotFound, ErrorResponse{Error: "avatar not found"})
		return
	}
	ctx.Header("Cache-Control", "public, max-age=86400")
	ctx.Header("Content-Type", "image/webp")
	ctx.Data(http.StatusOK, "image/webp", data)
}

// getFederatedAuthorAvatar godoc
// @Summary      Get federated author avatar
// @Description  Return cached avatar for a remote author in the viewer's federation inbox
// @Tags         federation
// @Produce      image/webp
// @Security     BearerAuth
// @Param        ownerKey  path  string  true  "Encoded remote author handle"
// @Success      200  {file}  binary
// @Failure      401  {object}  ErrorResponse
// @Failure      404  {object}  ErrorResponse
// @Router       /federation/authors/{ownerKey}/avatar [get]
func (a *App) getFederatedAuthorAvatar(ctx *gin.Context) {
	nickname, err := a.currentUserNickname(ctx)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, ErrorResponse{Error: "user not found"})
		return
	}

	ownerKey := strings.TrimSpace(ctx.Param("ownerKey"))
	data, err := a.Federation.Inbox().Avatar(nickname, ownerKey)
	if err != nil {
		if errors.Is(err, avatars.ErrAvatarNotFound) {
			ctx.JSON(http.StatusNotFound, ErrorResponse{Error: "avatar not found"})
			return
		}
		ctx.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to load avatar"})
		return
	}

	ctx.Header("Cache-Control", "public, max-age=86400")
	ctx.Header("Content-Type", "image/webp")
	ctx.Data(http.StatusOK, "image/webp", data)
}

func readAvatarUpload(ctx *gin.Context) ([]byte, error) {
	file, err := ctx.FormFile("avatar")
	if err != nil {
		raw, readErr := io.ReadAll(io.LimitReader(ctx.Request.Body, avatars.MaxUploadBytes+1))
		if readErr != nil {
			return nil, errors.New("avatar file is required")
		}
		if len(raw) == 0 {
			return nil, errors.New("avatar file is required")
		}
		if len(raw) > avatars.MaxUploadBytes {
			return nil, avatars.ErrAvatarTooLarge
		}
		return raw, nil
	}

	opened, err := file.Open()
	if err != nil {
		return nil, err
	}
	defer opened.Close()

	raw, err := io.ReadAll(io.LimitReader(opened, avatars.MaxUploadBytes+1))
	if err != nil {
		return nil, err
	}
	if len(raw) == 0 {
		return nil, errors.New("avatar file is required")
	}
	if len(raw) > avatars.MaxUploadBytes {
		return nil, avatars.ErrAvatarTooLarge
	}
	return raw, nil
}
