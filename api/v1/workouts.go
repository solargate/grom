package v1

import (
	"encoding/json"
	"errors"
	"io"
	"mime/multipart"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/solargate/grom/internal/auth"
	"github.com/solargate/grom/internal/config"
	"github.com/solargate/grom/internal/social"
	"github.com/solargate/grom/internal/tracks"
	"github.com/solargate/grom/internal/workouts"
)

var workoutStore *workouts.Store

func initWorkoutStore() {
	if workoutStore == nil {
		workoutStore = workouts.NewStore(config.Cfg.Data.ResolvedDir)
	}
}

type CreateWorkoutRequest struct {
	Name            string   `json:"name" binding:"required" example:"Morning run"`
	Description     string   `json:"description" example:"Easy session"`
	SportType       string   `json:"sport_type" binding:"required" example:"Run"`
	StartDate       string   `json:"start_date" binding:"required" example:"2026-07-05T14:30:00+03:00"`
	DurationSeconds int      `json:"duration_seconds" example:"3600"`
	Distance        float64  `json:"distance" example:"5200"`
	EquipmentIDs    []string `json:"equipment_ids,omitempty" example:"550e8400-e29b-41d4-a716-446655440000"`
}

type CreateWorkoutForm struct {
	Name            string                `form:"name" binding:"required"`
	Description     string                `form:"description"`
	SportType       string                `form:"sport_type" binding:"required"`
	StartDate       string                `form:"start_date" binding:"required"`
	DurationSeconds int                   `form:"duration_seconds"`
	Distance        float64               `form:"distance"`
	EquipmentIDs    string                `form:"equipment_ids"`
	Track           *multipart.FileHeader `form:"track"`
}

type ParseTrackResponse struct {
	StartDate       string  `json:"start_date,omitempty" example:"2026-07-05T14:30:00+03:00"`
	Device          string  `json:"device,omitempty" example:"Garmin Edge 530"`
	DurationSeconds int     `json:"duration_seconds,omitempty" example:"3600"`
	Distance        float64 `json:"distance,omitempty" example:"5200"`
	HasGPS          bool    `json:"has_gps" example:"true"`
}

type WorkoutAuthorResponse struct {
	Nickname  string `json:"nickname" example:"bob"`
	Name      string `json:"name" example:"Bob"`
	Handle    string `json:"handle" example:"bob@grom.example"`
	IsLocal   bool   `json:"is_local" example:"true"`
	HasAvatar bool   `json:"has_avatar" example:"true"`
	AvatarURL string `json:"avatar_url,omitempty" example:"/api/v1/users/bob/avatar"`
}

type WorkoutResponse struct {
	ID              string                 `json:"id" example:"38472901"`
	Owner           string                 `json:"owner,omitempty" example:"solarwind"`
	Name            string                 `json:"name" example:"Morning run"`
	Description     string                 `json:"description,omitempty" example:"Easy session"`
	SportType       string                 `json:"sport_type" example:"Run"`
	StartDate       string                 `json:"start_date" example:"2026-07-05T14:30:00+03:00"`
	Device          string                 `json:"device,omitempty" example:"Grom"`
	DurationSeconds int                    `json:"duration_seconds" example:"3600"`
	Distance        float64                `json:"distance" example:"5200"`
	Track           string                 `json:"track,omitempty" example:"track.gpx"`
	Equipment       []WorkoutEquipmentItem `json:"equipment,omitempty"`
	HasMapPreview   bool                   `json:"has_map_preview" example:"true"`
	HasMedia        bool                   `json:"has_media" example:"true"`
	MediaFiles      []string               `json:"media_files,omitempty"`
	Author          *WorkoutAuthorResponse `json:"author,omitempty"`
}

type WorkoutEquipmentItem struct {
	ID   string `json:"id" example:"550e8400-e29b-41d4-a716-446655440000"`
	Name string `json:"name" example:"Gravel bike"`
	Type string `json:"type,omitempty" example:"bike"`
}

func toWorkoutResponse(workout *workouts.Workout) WorkoutResponse {
	equipment := make([]WorkoutEquipmentItem, 0, len(workout.Equipment))
	for _, item := range workout.Equipment {
		equipment = append(equipment, WorkoutEquipmentItem{
			ID:   item.ID,
			Name: item.Name,
			Type: item.Type,
		})
	}
	return WorkoutResponse{
		ID:              workout.ID,
		Name:            workout.Name,
		Description:     workout.Description,
		SportType:       workout.SportType,
		StartDate:       workout.StartDate.Format(time.RFC3339),
		Device:          workout.Device,
		DurationSeconds: workout.DurationSeconds,
		Distance:        workout.Distance,
		Track:           workout.Track,
		Equipment:       equipment,
		HasMapPreview:   workout.HasMapPreview,
		HasMedia:        workout.HasMedia,
		MediaFiles:      workout.MediaFiles,
	}
}

func toFeedWorkoutResponse(item *workouts.FeedWorkout) WorkoutResponse {
	resp := toWorkoutResponse(&item.Workout)
	resp.Owner = item.Owner
	resp.Author = &WorkoutAuthorResponse{
		Nickname:  item.Author.Nickname,
		Name:      item.Author.Name,
		Handle:    item.Author.Handle,
		IsLocal:   item.Author.IsLocal,
		HasAvatar: item.Author.HasAvatar,
		AvatarURL: item.Author.AvatarURL,
	}
	return resp
}

func toParseTrackResponse(data *tracks.Data) ParseTrackResponse {
	resp := ParseTrackResponse{
		HasGPS: data.HasGPS(),
	}
	if data.StartTime != nil {
		resp.StartDate = data.StartTime.Format(time.RFC3339)
	}
	if data.Device != nil {
		resp.Device = *data.Device
	}
	if data.DurationSeconds != nil {
		resp.DurationSeconds = *data.DurationSeconds
	}
	if data.DistanceMeters != nil {
		resp.Distance = *data.DistanceMeters
	}
	return resp
}

func currentUserNickname(ctx *gin.Context) (string, error) {
	if userStore == nil {
		if err := initUserStore(); err != nil {
			return "", err
		}
	}
	userID, _ := ctx.Get(auth.ContextUserIDKey)
	id, ok := userID.(string)
	if !ok || id == "" {
		return "", errors.New("invalid token")
	}

	user, err := userStore.FindByID(id)
	if err != nil {
		return "", err
	}
	return user.Nickname, nil
}

func workoutAccessOwners(ctx *gin.Context, viewerNickname string) ([]string, error) {
	if err := initSocialService(); err != nil {
		return nil, err
	}
	userID, err := currentUserID(ctx)
	if err != nil {
		return nil, err
	}
	nicknames, err := socialService.ActiveFollowingNicknames(userID)
	if err != nil {
		return nil, err
	}
	_ = viewerNickname
	return nicknames, nil
}

func resolveWorkoutOwner(ctx *gin.Context, viewerNickname string) (ownerNickname, workoutID string, err error) {
	workoutID = ctx.Param("id")
	ownerNickname = strings.TrimSpace(ctx.Query("owner"))
	if ownerNickname == "" {
		ownerNickname = viewerNickname
	}
	followed, err := workoutAccessOwners(ctx, viewerNickname)
	if err != nil {
		return "", "", err
	}
	feedSvc := workouts.NewFeedService(workoutStore, config.Cfg.Federation.Domain)
	if !feedSvc.CanAccessWorkout(viewerNickname, followed, ownerNickname) {
		return "", "", workouts.ErrWorkoutNotFound
	}
	return ownerNickname, workoutID, nil
}

func readTrackFile(file *multipart.FileHeader) ([]byte, error) {
	opened, err := file.Open()
	if err != nil {
		return nil, err
	}
	defer opened.Close()

	data, err := io.ReadAll(io.LimitReader(opened, tracks.MaxTrackSizeBytes+1))
	if err != nil {
		return nil, err
	}
	if len(data) > tracks.MaxTrackSizeBytes {
		return nil, tracks.ErrTrackTooLarge
	}
	return data, nil
}

func handleCreateWorkoutError(ctx *gin.Context, err error) {
	switch {
	case errors.Is(err, workouts.ErrInvalidSportType),
		errors.Is(err, workouts.ErrInvalidWorkout),
		errors.Is(err, tracks.ErrInvalidFormat),
		errors.Is(err, tracks.ErrTrackTooLarge),
		errors.Is(err, tracks.ErrEmptyTrack),
		errors.Is(err, tracks.ErrInvalidTrack),
		errors.Is(err, workouts.ErrInvalidPhoto),
		errors.Is(err, workouts.ErrPhotoTooLarge),
		errors.Is(err, workouts.ErrTooManyPhotos):
		ctx.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
	case errors.Is(err, workouts.ErrWorkoutExists):
		ctx.JSON(http.StatusConflict, ErrorResponse{Error: err.Error()})
	default:
		ctx.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to create workout"})
	}
}

// createWorkout godoc
// @Summary      Create workout
// @Description  Create a manual workout for the authenticated user
// @Tags         workouts
// @Accept       json
// @Accept       mpfd
// @Produce      json
// @Security     BearerAuth
// @Param        body  body  CreateWorkoutRequest  false  "Workout data (JSON)"
// @Param        name  formData  string  false  "Workout name (multipart)"
// @Param        description  formData  string  false  "Description (multipart)"
// @Param        sport_type  formData  string  false  "Sport type (multipart)"
// @Param        start_date  formData  string  false  "Start date RFC3339 (multipart)"
// @Param        duration_seconds  formData  int  false  "Duration seconds (multipart)"
// @Param        distance  formData  number  false  "Distance meters (multipart)"
// @Param        track  formData  file  false  "Track file FIT or GPX (multipart)"
// @Success      201   {object}  WorkoutResponse
// @Failure      400   {object}  ErrorResponse
// @Failure      401   {object}  ErrorResponse
// @Failure      500   {object}  ErrorResponse
// @Router       /workouts [post]
func createWorkout(ctx *gin.Context) {
	initWorkoutStore()

	nickname, err := currentUserNickname(ctx)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, ErrorResponse{Error: "user not found"})
		return
	}

	contentType := ctx.GetHeader("Content-Type")
	if strings.HasPrefix(contentType, "multipart/form-data") {
		createWorkoutMultipart(ctx, nickname)
		return
	}

	var req CreateWorkoutRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	startDate, err := time.Parse(time.RFC3339, req.StartDate)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid start_date format, expected RFC3339"})
		return
	}

	equipmentItems, err := resolveWorkoutEquipment(nickname, req.EquipmentIDs)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to resolve equipment"})
		return
	}

	workout, err := workoutStore.Create(nickname, &workouts.Workout{
		Name:            req.Name,
		Description:     req.Description,
		SportType:       req.SportType,
		StartDate:       startDate,
		DurationSeconds: req.DurationSeconds,
		Distance:        req.Distance,
		Equipment:       equipmentItems,
	})
	if err != nil {
		handleCreateWorkoutError(ctx, err)
		return
	}

	if userID, err := currentUserID(ctx); err == nil {
		saveLastEquipmentForSport(userID, req.SportType, req.EquipmentIDs)
	}

	publishCreatedWorkout(nickname, workout)

	ctx.JSON(http.StatusCreated, toWorkoutResponse(workout))
}

func readPhotoFile(file *multipart.FileHeader) ([]byte, error) {
	opened, err := file.Open()
	if err != nil {
		return nil, err
	}
	defer opened.Close()

	data, err := io.ReadAll(io.LimitReader(opened, workouts.MaxPhotoBytes+1))
	if err != nil {
		return nil, err
	}
	if len(data) > workouts.MaxPhotoBytes {
		return nil, workouts.ErrPhotoTooLarge
	}
	return data, nil
}

func readWorkoutPhotos(ctx *gin.Context) ([]workouts.MediaFileInput, error) {
	form, err := ctx.MultipartForm()
	if err != nil || form == nil {
		return nil, nil
	}
	headers := form.File["photos"]
	if len(headers) == 0 {
		return nil, nil
	}
	if len(headers) > workouts.MaxPhotosPerWorkout {
		return nil, workouts.ErrTooManyPhotos
	}
	files := make([]workouts.MediaFileInput, 0, len(headers))
	for _, header := range headers {
		data, err := readPhotoFile(header)
		if err != nil {
			return nil, err
		}
		files = append(files, workouts.MediaFileInput{
			Filename: header.Filename,
			Data:     data,
		})
	}
	return files, nil
}

func attachWorkoutPhotos(nickname string, workout *workouts.Workout, photos []workouts.MediaFileInput) (*workouts.Workout, error) {
	if len(photos) == 0 {
		return workout, nil
	}
	return workoutStore.AddMedia(nickname, workout, photos)
}

func publishCreatedWorkout(nickname string, workout *workouts.Workout) {
	if err := initFederation(); err != nil || federationDelivery == nil || followersStore == nil {
		return
	}
	inboxes, err := followersStore.ListInboxes(nickname)
	if err != nil || len(inboxes) == 0 {
		return
	}
	var trackData []byte
	if workout.Track != "" {
		trackData, _, _, _ = workoutStore.TrackFile(nickname, workout.ID)
	}
	var mediaFiles []workouts.MediaFileInput
	if workout.HasMedia {
		mediaFiles, _ = workoutStore.ReadMediaPayload(nickname, workout.ID)
	}
	_ = federationDelivery.DeliverWorkout(nickname, workout, inboxes, trackData, mediaFiles)
}

func createWorkoutMultipart(ctx *gin.Context, nickname string) {
	var form CreateWorkoutForm
	if err := ctx.ShouldBind(&form); err != nil {
		ctx.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	startDate, err := time.Parse(time.RFC3339, form.StartDate)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid start_date format, expected RFC3339"})
		return
	}

	equipmentIDs, err := parseEquipmentIDsForm(form.EquipmentIDs)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid equipment_ids format"})
		return
	}
	equipmentItems, err := resolveWorkoutEquipment(nickname, equipmentIDs)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to resolve equipment"})
		return
	}

	workout := &workouts.Workout{
		Name:            form.Name,
		Description:     form.Description,
		SportType:       form.SportType,
		StartDate:       startDate,
		DurationSeconds: form.DurationSeconds,
		Distance:        form.Distance,
		Equipment:       equipmentItems,
	}

	var trackInput *workouts.TrackInput
	if form.Track != nil {
		data, err := readTrackFile(form.Track)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
			return
		}
		parsed, err := tracks.Parse(data, form.Track.Filename)
		if err != nil {
			handleCreateWorkoutError(ctx, err)
			return
		}
		trackInput = &workouts.TrackInput{
			Filename: form.Track.Filename,
			Data:     data,
			Parsed:   parsed,
		}
	}

	created, err := workoutStore.CreateWithTrack(nickname, workout, trackInput)
	if err != nil {
		handleCreateWorkoutError(ctx, err)
		return
	}

	photos, err := readWorkoutPhotos(ctx)
	if err != nil {
		handleCreateWorkoutError(ctx, err)
		return
	}
	created, err = attachWorkoutPhotos(nickname, created, photos)
	if err != nil {
		handleCreateWorkoutError(ctx, err)
		return
	}

	if userID, err := currentUserID(ctx); err == nil {
		saveLastEquipmentForSport(userID, form.SportType, equipmentIDs)
	}

	publishCreatedWorkout(nickname, created)
	ctx.JSON(http.StatusCreated, toWorkoutResponse(created))
}

func parseEquipmentIDsForm(raw string) ([]string, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil, nil
	}
	var ids []string
	if err := json.Unmarshal([]byte(raw), &ids); err != nil {
		return nil, err
	}
	return ids, nil
}

// parseTrack godoc
// @Summary      Parse track file
// @Description  Extract metadata from a FIT or GPX track without saving it
// @Tags         workouts
// @Accept       mpfd
// @Produce      json
// @Security     BearerAuth
// @Param        track  formData  file  true  "Track file FIT or GPX"
// @Success      200  {object}  ParseTrackResponse
// @Failure      400  {object}  ErrorResponse
// @Failure      401  {object}  ErrorResponse
// @Router       /workouts/parse-track [post]
func parseTrack(ctx *gin.Context) {
	file, err := ctx.FormFile("track")
	if err != nil {
		ctx.JSON(http.StatusBadRequest, ErrorResponse{Error: "track file is required"})
		return
	}

	data, err := readTrackFile(file)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	parsed, err := tracks.Parse(data, file.Filename)
	if err != nil {
		handleCreateWorkoutError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, toParseTrackResponse(parsed))
}

// getWorkoutTrack godoc
// @Summary      Get workout track file
// @Description  Return the workout track file. Use format=gpx to download as GPX (FIT is converted on the fly). Original format is only available for your own workouts.
// @Tags         workouts
// @Produce      application/gpx+xml
// @Produce      application/vnd.ant.fit
// @Security     BearerAuth
// @Param        id      path   string  true   "Workout ID"
// @Param        owner   query  string  false  "Workout owner nickname (required for followed users' workouts)"
// @Param        format  query  string  false  "gpx to download as GPX; omit for original file (own workouts only)"
// @Success      200  {file}  binary
// @Failure      401  {object}  ErrorResponse
// @Failure      403  {object}  ErrorResponse
// @Failure      404  {object}  ErrorResponse
// @Router       /workouts/{id}/track [get]
func getWorkoutTrack(ctx *gin.Context) {
	initWorkoutStore()

	nickname, err := currentUserNickname(ctx)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, ErrorResponse{Error: "user not found"})
		return
	}

	owner, workoutID, err := resolveWorkoutOwner(ctx, nickname)
	if err != nil {
		if errors.Is(err, workouts.ErrWorkoutNotFound) {
			ctx.JSON(http.StatusNotFound, ErrorResponse{Error: "track not found"})
			return
		}
		ctx.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to resolve workout"})
		return
	}

	format := strings.ToLower(strings.TrimSpace(ctx.Query("format")))
	wantGPX := format == "gpx"
	if format != "" && !wantGPX {
		ctx.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid format"})
		return
	}
	if !wantGPX && owner != nickname {
		ctx.JSON(http.StatusForbidden, ErrorResponse{Error: "original track download is only available for your own workouts"})
		return
	}

	data, storageName, workoutName, err := workoutStore.TrackFile(owner, workoutID)
	if err != nil {
		if errors.Is(err, workouts.ErrWorkoutNotFound) {
			_ = initFederation()
			if workoutInboxStore != nil {
				data, storageName, workoutName, err = workoutInboxStore.TrackFile(nickname, owner, workoutID)
			}
		}
	}
	if err != nil {
		if errors.Is(err, workouts.ErrWorkoutNotFound) {
			ctx.JSON(http.StatusNotFound, ErrorResponse{Error: "track not found"})
			return
		}
		ctx.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to load track"})
		return
	}

	var output []byte
	var downloadFilename string
	var contentType string

	if wantGPX {
		output, err = tracks.ExportGPX(data, storageName, workoutName)
		if err != nil {
			if errors.Is(err, tracks.ErrInvalidFormat) ||
				errors.Is(err, tracks.ErrEmptyTrack) ||
				errors.Is(err, tracks.ErrInvalidTrack) {
				ctx.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
				return
			}
			ctx.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to export track"})
			return
		}
		downloadFilename = workouts.TrackDownloadFilename(workoutName, ".gpx")
		contentType = "application/gpx+xml"
	} else {
		output = data
		downloadFilename = workouts.TrackDownloadFilename(workoutName, filepath.Ext(storageName))
		switch storageName {
		case tracks.TrackFileGPX:
			contentType = "application/gpx+xml"
		case tracks.TrackFileFIT:
			contentType = "application/vnd.ant.fit"
		default:
			contentType = "application/octet-stream"
		}
	}

	ctx.Header("Content-Disposition", workouts.ContentDispositionAttachment(downloadFilename))
	ctx.Data(http.StatusOK, contentType, output)
}

// getWorkoutMapPreview godoc
// @Summary      Get workout map preview
// @Description  Return cached map preview image for a workout with a track
// @Tags         workouts
// @Produce      image/webp
// @Security     BearerAuth
// @Param        id  path  string  true  "Workout ID"
// @Param        owner  query  string  false  "Workout owner nickname (required for followed users' workouts)"
// @Success      200  {file}  binary
// @Failure      401  {object}  ErrorResponse
// @Failure      404  {object}  ErrorResponse
// @Router       /workouts/{id}/map-preview [get]
func getWorkoutMapPreview(ctx *gin.Context) {
	initWorkoutStore()

	nickname, err := currentUserNickname(ctx)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, ErrorResponse{Error: "user not found"})
		return
	}

	owner, workoutID, err := resolveWorkoutOwner(ctx, nickname)
	if err != nil {
		if errors.Is(err, workouts.ErrWorkoutNotFound) {
			ctx.JSON(http.StatusNotFound, ErrorResponse{Error: "map preview not found"})
			return
		}
		ctx.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to resolve workout"})
		return
	}

	path, err := workoutStore.MapPreviewPath(owner, workoutID)
	if err != nil {
		if errors.Is(err, workouts.ErrWorkoutNotFound) {
			_ = initFederation()
			if workoutInboxStore != nil {
				path, err = workoutInboxStore.MapPreviewPath(nickname, owner, workoutID)
			}
		}
	}
	if err != nil {
		if errors.Is(err, workouts.ErrWorkoutNotFound) {
			ctx.JSON(http.StatusNotFound, ErrorResponse{Error: "map preview not found"})
			return
		}
		ctx.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to load map preview"})
		return
	}

	ctx.Header("Cache-Control", "public, max-age=31536000, immutable")
	ctx.Header("Content-Type", "image/webp")
	ctx.File(path)
}

func getWorkoutMediaPreview(ctx *gin.Context) {
	initWorkoutStore()

	nickname, err := currentUserNickname(ctx)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, ErrorResponse{Error: "user not found"})
		return
	}

	owner, workoutID, err := resolveWorkoutOwner(ctx, nickname)
	if err != nil {
		if errors.Is(err, workouts.ErrWorkoutNotFound) {
			ctx.JSON(http.StatusNotFound, ErrorResponse{Error: "photo not found"})
			return
		}
		ctx.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to resolve workout"})
		return
	}

	filename := ctx.Param("filename")
	path, err := workoutStore.MediaPreviewFile(owner, workoutID, filename)
	if err != nil {
		if errors.Is(err, workouts.ErrPhotoNotFound) || errors.Is(err, workouts.ErrWorkoutNotFound) {
			_ = initFederation()
			if workoutInboxStore != nil {
				path, err = workoutInboxStore.MediaPreviewPath(nickname, owner, workoutID, filename)
			}
		}
	}
	if err != nil {
		if errors.Is(err, workouts.ErrPhotoNotFound) || errors.Is(err, workouts.ErrWorkoutNotFound) {
			ctx.JSON(http.StatusNotFound, ErrorResponse{Error: "photo not found"})
			return
		}
		ctx.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to load photo preview"})
		return
	}

	ctx.Header("Cache-Control", "public, max-age=31536000, immutable")
	ctx.Header("Content-Type", "image/webp")
	ctx.File(path)
}

func getWorkoutMediaOriginal(ctx *gin.Context) {
	initWorkoutStore()

	nickname, err := currentUserNickname(ctx)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, ErrorResponse{Error: "user not found"})
		return
	}

	owner, workoutID, err := resolveWorkoutOwner(ctx, nickname)
	if err != nil {
		if errors.Is(err, workouts.ErrWorkoutNotFound) {
			ctx.JSON(http.StatusNotFound, ErrorResponse{Error: "photo not found"})
			return
		}
		ctx.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to resolve workout"})
		return
	}

	filename := ctx.Param("filename")
	path, err := workoutStore.MediaOriginalFile(owner, workoutID, filename)
	if err != nil {
		if errors.Is(err, workouts.ErrPhotoNotFound) || errors.Is(err, workouts.ErrWorkoutNotFound) {
			_ = initFederation()
			if workoutInboxStore != nil {
				path, err = workoutInboxStore.MediaOriginalPath(nickname, owner, workoutID, filename)
			}
		}
	}
	if err != nil {
		if errors.Is(err, workouts.ErrPhotoNotFound) || errors.Is(err, workouts.ErrWorkoutNotFound) {
			ctx.JSON(http.StatusNotFound, ErrorResponse{Error: "photo not found"})
			return
		}
		ctx.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to load photo"})
		return
	}

	ctx.Header("Cache-Control", "public, max-age=31536000, immutable")
	ctx.Header("Content-Type", workouts.MediaContentType(filename))
	ctx.File(path)
}

// listWorkouts godoc
// @Summary      List workouts
// @Description  Return workouts for the authenticated user sorted by start date descending. Use scope=feed for the full feed (default) or scope=own for only the viewer's workouts.
// @Tags         workouts
// @Produce      json
// @Security     BearerAuth
// @Param        scope  query  string  false  "feed (default) or own"
// @Success      200  {array}   WorkoutResponse
// @Failure      400  {object}  ErrorResponse
// @Failure      401  {object}  ErrorResponse
// @Failure      500  {object}  ErrorResponse
// @Router       /workouts [get]
func listWorkouts(ctx *gin.Context) {
	initWorkoutStore()

	nickname, err := currentUserNickname(ctx)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, ErrorResponse{Error: "user not found"})
		return
	}

	scope := strings.TrimSpace(ctx.Query("scope"))
	if scope == "" {
		scope = "feed"
	}
	if scope != "feed" && scope != "own" {
		ctx.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid scope"})
		return
	}

	if userStore == nil {
		if err := initUserStore(); err != nil {
			ctx.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to init user store"})
			return
		}
	}

	userID, err := currentUserID(ctx)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, ErrorResponse{Error: "invalid token"})
		return
	}

	viewer, err := userStore.FindByID(userID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, ErrorResponse{Error: "user not found"})
		return
	}

	feedSvc := newFeedService()
	var items []workouts.FeedWorkout

	if scope == "own" {
		items, err = feedSvc.ListOwn(nickname, viewer.Name)
	} else {
		if err := initSocialService(); err != nil {
			ctx.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to init social service"})
			return
		}
		_ = initFederation()

		follows, err := socialService.ListFollowing(userID)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to list following"})
			return
		}

		followedAuthors := make([]workouts.FeedAuthor, 0)
		for i := range follows {
			if follows[i].Status != social.StatusActive {
				continue
			}
			if !follows[i].TargetIsLocal {
				continue
			}
			hasAvatar, avatarURL := localAvatarFieldsForUser(follows[i].TargetNickname)
			followedAuthors = append(followedAuthors, workouts.FeedAuthor{
				Nickname:  follows[i].TargetNickname,
				Name:      follows[i].TargetName,
				Handle:    follows[i].TargetHandle,
				IsLocal:   true,
				HasAvatar: hasAvatar,
				AvatarURL: avatarURL,
			})
		}

		items, err = feedSvc.ListFeed(nickname, viewer.Name, followedAuthors)
	}
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to list workouts"})
		return
	}

	response := make([]WorkoutResponse, 0, len(items))
	for i := range items {
		response = append(response, toFeedWorkoutResponse(&items[i]))
	}

	ctx.JSON(http.StatusOK, response)
}
