package v1

import (
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/solargate/grom/internal/integrations/strava"
)

const maxStravaArchiveSize = 2 << 30 // 2 GiB

type StravaImportResultResponse struct {
	Imported     int `json:"imported" example:"100"`
	Skipped      int `json:"skipped" example:"2"`
	ParseSkipped int `json:"parse_skipped" example:"0"`
	MediaMissing int `json:"media_missing" example:"5"`
	Errors       int `json:"errors" example:"0"`
}

type StravaImportStatusResponse struct {
	Active         bool                        `json:"active" example:"true"`
	Phase          string                      `json:"phase" example:"importing"`
	UploadProgress float64                     `json:"upload_progress" example:"1"`
	ImportProgress float64                     `json:"import_progress" example:"0.42"`
	ImportCurrent  int                         `json:"import_current" example:"412"`
	ImportTotal    int                         `json:"import_total" example:"982"`
	Message        string                      `json:"message,omitempty" example:""`
	Result         *StravaImportResultResponse `json:"result,omitempty"`
}

// importStravaArchive godoc
// @Summary      Import Strava bulk export archive
// @Description  Upload a Strava ZIP archive and start background import
// @Tags         integrations
// @Accept       mpfd
// @Accept       application/zip
// @Produce      json
// @Security     BearerAuth
// @Param        archive  formData  file  false  "Strava bulk export ZIP (multipart)"
// @Success      202  {object}  StravaImportStatusResponse
// @Failure      400  {object}  ErrorResponse  "Invalid or too large archive"
// @Failure      401  {object}  ErrorResponse  "Unauthorized"
// @Failure      409  {object}  ErrorResponse  "Import already in progress"
// @Failure      500  {object}  ErrorResponse  "Internal server error"
// @Router       /integrations/strava/import [post]
func (a *App) importStravaArchive(ctx *gin.Context) {
	jobs := a.stravaJobManager()

	userID, err := a.currentUserID(ctx)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, ErrorResponse{Error: "user not found"})
		return
	}
	nickname, err := a.currentUserNickname(ctx)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, ErrorResponse{Error: "user not found"})
		return
	}

	job, err := jobs.BeginUpload(userID, nickname)
	if err != nil {
		ctx.JSON(http.StatusConflict, ErrorResponse{Error: err.Error()})
		return
	}

	reader, expectedSize, err := stravaArchiveReader(ctx)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}
	if expectedSize > maxStravaArchiveSize {
		ctx.JSON(http.StatusBadRequest, ErrorResponse{Error: "archive file too large"})
		return
	}

	if _, err := jobs.SaveArchive(job, reader, expectedSize); err != nil {
		ctx.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	jobs.StartImport(userID)
	ctx.JSON(http.StatusAccepted, toStravaImportStatus(jobs.Status(userID)))
}

func stravaArchiveReader(ctx *gin.Context) (io.Reader, int64, error) {
	contentType := strings.ToLower(strings.TrimSpace(ctx.GetHeader("Content-Type")))
	if strings.HasPrefix(contentType, "application/zip") ||
		strings.HasPrefix(contentType, "application/octet-stream") {
		expectedSize := ctx.Request.ContentLength
		if expectedSize < 0 {
			expectedSize = 0
		}
		return io.LimitReader(ctx.Request.Body, maxStravaArchiveSize+1), expectedSize, nil
	}

	file, err := ctx.FormFile("archive")
	if err != nil {
		return nil, 0, err
	}
	if file.Size <= 0 {
		return nil, 0, fmt.Errorf("archive file is empty")
	}

	opened, err := file.Open()
	if err != nil {
		return nil, 0, err
	}

	return struct {
		io.Reader
		io.Closer
	}{
		Reader: io.LimitReader(opened, maxStravaArchiveSize+1),
		Closer: opened,
	}, file.Size, nil
}

// getStravaImportStatus godoc
// @Summary      Get Strava import status
// @Description  Returns current upload/import progress for the authenticated user
// @Tags         integrations
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  StravaImportStatusResponse
// @Failure      401  {object}  ErrorResponse  "Unauthorized"
// @Router       /integrations/strava/import/status [get]
func (a *App) getStravaImportStatus(ctx *gin.Context) {
	jobs := a.stravaJobManager()

	userID, err := a.currentUserID(ctx)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, ErrorResponse{Error: "user not found"})
		return
	}

	status := jobs.Status(userID)
	ctx.JSON(http.StatusOK, toStravaImportStatus(status))
}

func toStravaImportStatus(status strava.JobStatus) StravaImportStatusResponse {
	resp := StravaImportStatusResponse{
		Active:         status.Active,
		Phase:          status.Phase,
		UploadProgress: status.UploadProgress,
		ImportProgress: status.ImportProgress,
		ImportCurrent:  status.ImportCurrent,
		ImportTotal:    status.ImportTotal,
		Message:        status.Message,
	}
	if status.Result != nil {
		resp.Result = &StravaImportResultResponse{
			Imported:     status.Result.Imported,
			Skipped:      status.Result.Skipped,
			ParseSkipped: status.Result.ParseSkipped,
			MediaMissing: status.Result.MediaMissing,
			Errors:       status.Result.Errors,
		}
	}
	return resp
}
