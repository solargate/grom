package v1

import (
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"mime/multipart"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/solargate/grom/internal/auth"
	"github.com/solargate/grom/internal/config"
	"github.com/solargate/grom/internal/social"
	"github.com/solargate/grom/internal/tracks"
	"github.com/solargate/grom/internal/workouts"
)

type CreateWorkoutRequest struct {
	Name                 string   `json:"name" binding:"required" example:"Morning run"`
	Description          string   `json:"description" example:"Easy session"`
	SportType            string   `json:"sport_type" binding:"required" example:"Run"`
	StartDate            string   `json:"start_date" binding:"required" example:"2026-07-05T14:30:00+03:00"`
	DurationSeconds      int      `json:"duration_seconds" example:"3600"`
	DurationTotalSeconds int      `json:"duration_total_seconds,omitempty" example:"3900"`
	Distance             float64  `json:"distance" example:"5200"`
	SpeedMaxKmh          *float64 `json:"speed_max_kmh,omitempty" example:"32.5"`
	SpeedAvgKmh          *float64 `json:"speed_avg_kmh,omitempty" example:"18.2"`
	EquipmentIDs         []string `json:"equipment_ids,omitempty" example:"550e8400-e29b-41d4-a716-446655440000"`
}

type CreateWorkoutForm struct {
	Name                 string                `form:"name" binding:"required"`
	Description          string                `form:"description"`
	SportType            string                `form:"sport_type" binding:"required"`
	StartDate            string                `form:"start_date" binding:"required"`
	DurationSeconds      int                   `form:"duration_seconds"`
	DurationTotalSeconds int                   `form:"duration_total_seconds"`
	Distance             float64               `form:"distance"`
	SpeedMaxKmh          string                `form:"speed_max_kmh"`
	SpeedAvgKmh          string                `form:"speed_avg_kmh"`
	EquipmentIDs         string                `form:"equipment_ids"`
	Track                *multipart.FileHeader `form:"track"`
}

type ParseTrackResponse struct {
	StartDate            string   `json:"start_date,omitempty" example:"2026-07-05T14:30:00+03:00"`
	Device               string   `json:"device,omitempty" example:"Garmin Edge 530"`
	DurationSeconds      int      `json:"duration_seconds,omitempty" example:"3600"`
	DurationTotalSeconds int      `json:"duration_total_seconds,omitempty" example:"3900"`
	Distance             float64  `json:"distance,omitempty" example:"5200"`
	HasGPS               bool     `json:"has_gps" example:"true"`
	SpeedMaxKmh          *float64 `json:"speed_max_kmh,omitempty" example:"32.4"`
	SpeedAvgKmh          *float64 `json:"speed_avg_kmh,omitempty" example:"17.5"`
	ElevationGain        *float64 `json:"elevation_gain,omitempty" example:"77"`
	ElevationLoss        *float64 `json:"elevation_loss,omitempty" example:"90"`
	ElevationLow         *float64 `json:"elevation_low,omitempty" example:"130.6"`
	ElevationHigh        *float64 `json:"elevation_high,omitempty" example:"183.6"`
	GradeMax             *float64 `json:"grade_max,omitempty" example:"8.5"`
	GradeAvg             *float64 `json:"grade_avg,omitempty" example:"2.1"`
	CadenceMax           *float64 `json:"cadence_max,omitempty" example:"116"`
	CadenceAvg           *float64 `json:"cadence_avg,omitempty" example:"34"`
	HeartRateMax         *float64 `json:"heart_rate_max,omitempty" example:"187"`
	HeartRateAvg         *float64 `json:"heart_rate_avg,omitempty" example:"130"`
	WattsMax             *float64 `json:"watts_max,omitempty" example:"350"`
	WattsAvg             *float64 `json:"watts_avg,omitempty" example:"180"`
	Calories             *float64 `json:"calories,omitempty" example:"415"`
	TemperatureMax       *float64 `json:"temperature_max,omitempty" example:"20"`
	TemperatureAvg       *float64 `json:"temperature_avg,omitempty" example:"18"`
	StepsTotal           *int     `json:"steps_total,omitempty" example:"2583"`
	CyclesTotal          *int     `json:"cycles_total,omitempty" example:"1200"`
	SetsTotal            *int     `json:"sets_total,omitempty" example:"12"`
	RepsTotal            *int     `json:"reps_total,omitempty" example:"120"`
	TempAvgKmm           *string  `json:"temp_avg_kmm,omitempty" example:"12:22"`
}

type WorkoutAuthorResponse struct {
	Nickname  string `json:"nickname" example:"bob"`
	Name      string `json:"name" example:"Bob"`
	Handle    string `json:"handle" example:"bob@grom.example"`
	IsLocal   bool   `json:"is_local" example:"true"`
	HasAvatar bool   `json:"has_avatar" example:"true"`
	AvatarURL string `json:"avatar_url,omitempty" example:"/api/v1/users/bob/avatar"`
}

// WorkoutListResponse is a cursor-paginated workout list.
type WorkoutListResponse struct {
	Items      []WorkoutResponse `json:"items"`
	NextCursor string            `json:"next_cursor,omitempty"`
	HasMore    bool              `json:"has_more"`
}

type WorkoutResponse struct {
	ID                   string                 `json:"id" example:"38472901"`
	Owner                string                 `json:"owner,omitempty" example:"solarwind"`
	Name                 string                 `json:"name" example:"Morning run"`
	Description          string                 `json:"description,omitempty" example:"Easy session"`
	SportType            string                 `json:"sport_type" example:"Run"`
	StartDate            string                 `json:"start_date" example:"2026-07-05T14:30:00+03:00"`
	Device               string                 `json:"device,omitempty" example:"Grom"`
	DurationSeconds      int                    `json:"duration_seconds" example:"3600"`
	DurationTotalSeconds int                    `json:"duration_total_seconds,omitempty" example:"3900"`
	Distance             float64                `json:"distance" example:"5200"`
	TempAvgKmm           *string                `json:"temp_avg_kmm,omitempty" example:"12:22"`
	SpeedAvgKmh          *float64               `json:"speed_avg_kmh,omitempty" example:"17.5"`
	ElevationGain        *float64               `json:"elevation_gain,omitempty" example:"77"`
	HeartRateAvg         *float64               `json:"heart_rate_avg,omitempty" example:"130"`
	StepsTotal           *int                   `json:"steps_total,omitempty" example:"2583"`
	Calories             *float64               `json:"calories,omitempty" example:"415"`
	Track                string                 `json:"track,omitempty" example:"track.gpx"`
	Equipment            []WorkoutEquipmentItem `json:"equipment,omitempty"`
	HasMapPreview        bool                   `json:"has_map_preview" example:"true"`
	HasMedia             bool                   `json:"has_media" example:"true"`
	MediaFiles           []string               `json:"media_files,omitempty"`
	Author               *WorkoutAuthorResponse `json:"author,omitempty"`
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
		ID:                   workout.ID,
		Name:                 workout.Name,
		Description:          workout.Description,
		SportType:            workout.SportType,
		StartDate:            workout.StartDate.Format(time.RFC3339),
		Device:               workout.Device,
		DurationSeconds:      workout.DurationSeconds,
		DurationTotalSeconds: workout.DurationTotalSeconds,
		Distance:             workout.Distance,
		TempAvgKmm:           workout.TempAvgKmm,
		SpeedAvgKmh:          workout.SpeedAvgKmh,
		ElevationGain:        workout.ElevationGain,
		HeartRateAvg:         workout.HeartRateAvg,
		StepsTotal:           workout.StepsTotal,
		Calories:             workout.Calories,
		Track:                workout.Track,
		Equipment:            equipment,
		HasMapPreview:        workout.HasMapPreview,
		HasMedia:             workout.HasMedia,
		MediaFiles:           workout.MediaFiles,
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
	meta := data.Metadata()
	return ParseTrackResponse{
		StartDate:            meta.StartDate,
		Device:               meta.Device,
		DurationSeconds:      meta.DurationSeconds,
		DurationTotalSeconds: meta.DurationTotalSeconds,
		Distance:             meta.Distance,
		HasGPS:               meta.HasGPS,
		SpeedMaxKmh:          meta.SpeedMaxKmh,
		SpeedAvgKmh:          meta.SpeedAvgKmh,
		ElevationGain:        meta.ElevationGain,
		ElevationLoss:        meta.ElevationLoss,
		ElevationLow:         meta.ElevationLow,
		ElevationHigh:        meta.ElevationHigh,
		GradeMax:             meta.GradeMax,
		GradeAvg:             meta.GradeAvg,
		CadenceMax:           meta.CadenceMax,
		CadenceAvg:           meta.CadenceAvg,
		HeartRateMax:         meta.HeartRateMax,
		HeartRateAvg:         meta.HeartRateAvg,
		WattsMax:             meta.WattsMax,
		WattsAvg:             meta.WattsAvg,
		Calories:             meta.Calories,
		TemperatureMax:       meta.TemperatureMax,
		TemperatureAvg:       meta.TemperatureAvg,
		StepsTotal:           meta.StepsTotal,
		CyclesTotal:          meta.CyclesTotal,
		SetsTotal:            meta.SetsTotal,
		RepsTotal:            meta.RepsTotal,
		TempAvgKmm:           meta.TempAvgKmm,
	}
}

func workoutFromCreateRequest(req CreateWorkoutRequest, startDate time.Time, equipment []workouts.WorkoutEquipment) *workouts.Workout {
	return &workouts.Workout{
		Name:                 req.Name,
		Description:          req.Description,
		SportType:            req.SportType,
		StartDate:            startDate,
		DurationSeconds:      req.DurationSeconds,
		DurationTotalSeconds: req.DurationTotalSeconds,
		Distance:             req.Distance,
		SpeedMaxKmh:          req.SpeedMaxKmh,
		SpeedAvgKmh:          req.SpeedAvgKmh,
		Equipment:            equipment,
	}
}

func parseOptionalFloatForm(raw string) (*float64, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil, nil
	}
	v, err := strconv.ParseFloat(raw, 64)
	if err != nil {
		return nil, err
	}
	return &v, nil
}

func (a *App) currentUserNickname(ctx *gin.Context) (string, error) {
	userID, _ := ctx.Get(auth.ContextUserIDKey)
	id, ok := userID.(string)
	if !ok || id == "" {
		return "", errors.New("invalid token")
	}

	user, err := a.Users.FindByID(id)
	if err != nil {
		return "", err
	}
	return user.Nickname, nil
}

func (a *App) workoutAccessOwners(ctx *gin.Context, viewerNickname string) ([]string, error) {
	userID, err := a.currentUserID(ctx)
	if err != nil {
		return nil, err
	}
	nicknames, err := a.Social.ActiveFollowingNicknames(userID)
	if err != nil {
		return nil, err
	}
	_ = viewerNickname
	return nicknames, nil
}

func (a *App) resolveWorkoutOwner(ctx *gin.Context, viewerNickname string) (ownerNickname, workoutID string, err error) {
	workoutID = ctx.Param("id")
	ownerNickname = strings.TrimSpace(ctx.Query("owner"))
	if ownerNickname == "" {
		ownerNickname = viewerNickname
	}
	followed, err := a.workoutAccessOwners(ctx, viewerNickname)
	if err != nil {
		return "", "", err
	}
	feedSvc := workouts.NewFeedService(a.Workouts, a.Blobs, config.Cfg.Federation.Domain)
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
		respondInternal(ctx, "failed to create workout", err)
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
func (a *App) createWorkout(ctx *gin.Context) {
		nickname, err := a.currentUserNickname(ctx)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, ErrorResponse{Error: "user not found"})
		return
	}

	contentType := ctx.GetHeader("Content-Type")
	if strings.HasPrefix(contentType, "multipart/form-data") {
		a.createWorkoutMultipart(ctx, nickname)
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

	equipmentItems, err := a.resolveWorkoutEquipment(nickname, req.EquipmentIDs)
	if err != nil {
		respondInternal(ctx, "failed to resolve equipment", err)
		return
	}

	workout, err := a.Workouts.Create(nickname, workoutFromCreateRequest(req, startDate, equipmentItems))
	if err != nil {
		handleCreateWorkoutError(ctx, err)
		return
	}

	if userID, err := a.currentUserID(ctx); err == nil {
		a.saveLastEquipmentForSport(userID, req.SportType, req.EquipmentIDs)
	}

	a.publishCreatedWorkout(nickname, workout)

	a.scheduleEquipmentDistanceRecalc(nickname, workout.Equipment)

	slog.Info("workout created",
		"user", nickname,
		"workout_id", workout.ID,
		"sport_type", workout.SportType,
		"has_track", workout.Track != "",
	)
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

func (a *App) attachWorkoutPhotos(nickname string, workout *workouts.Workout, photos []workouts.MediaFileInput) (*workouts.Workout, error) {
	if len(photos) == 0 {
		return workout, nil
	}
	return a.Workouts.AddMedia(nickname, workout, photos)
}

func (a *App) publishCreatedWorkout(nickname string, workout *workouts.Workout) {
	a.publishWorkoutActivity(nickname, workout, false)
}

func (a *App) publishUpdatedWorkout(nickname string, workout *workouts.Workout) {
	a.publishWorkoutActivity(nickname, workout, true)
}

func (a *App) publishWorkoutActivity(nickname string, workout *workouts.Workout, update bool) {
	if a.federationDelivery == nil {
		return
	}
	inboxes, err := a.Federation.Followers().ListInboxes(nickname)
	if err != nil {
		slog.Error("federation list inboxes failed",
			"user", nickname,
			"workout_id", workout.ID,
			"err", err,
		)
		return
	}
	if len(inboxes) == 0 {
		return
	}
	var trackData []byte
	if workout.Track != "" {
		var trackErr error
		trackData, _, _, trackErr = a.Workouts.TrackFile(nickname, workout.ID)
		if trackErr != nil {
			slog.Warn("federation track read failed",
				"user", nickname,
				"workout_id", workout.ID,
				"err", trackErr,
			)
		}
	}
	var mediaFiles []workouts.MediaFileInput
	if workout.HasMedia || len(workout.MediaFiles) > 0 {
		var mediaErr error
		mediaFiles, mediaErr = a.Workouts.ReadMediaPayload(nickname, workout.ID)
		if mediaErr != nil {
			slog.Warn("federation media read failed",
				"user", nickname,
				"workout_id", workout.ID,
				"err", mediaErr,
			)
		}
	}
	if update {
		_ = a.federationDelivery.DeliverWorkoutUpdate(nickname, workout, inboxes, trackData, mediaFiles)
		return
	}
	_ = a.federationDelivery.DeliverWorkout(nickname, workout, inboxes, trackData, mediaFiles)
}

func (a *App) publishDeletedWorkout(nickname, workoutID string) {
	if a.federationDelivery == nil {
		return
	}
	inboxes, err := a.Federation.Followers().ListInboxes(nickname)
	if err != nil {
		slog.Error("federation list inboxes failed",
			"user", nickname,
			"workout_id", workoutID,
			"err", err,
		)
		return
	}
	if len(inboxes) == 0 {
		return
	}
	_ = a.federationDelivery.DeliverWorkoutDelete(nickname, workoutID, inboxes)
}

func (a *App) createWorkoutMultipart(ctx *gin.Context, nickname string) {
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
	equipmentItems, err := a.resolveWorkoutEquipment(nickname, equipmentIDs)
	if err != nil {
		respondInternal(ctx, "failed to resolve equipment", err)
		return
	}

	speedMaxKmh, err := parseOptionalFloatForm(form.SpeedMaxKmh)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid speed_max_kmh"})
		return
	}
	speedAvgKmh, err := parseOptionalFloatForm(form.SpeedAvgKmh)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid speed_avg_kmh"})
		return
	}

	workout := &workouts.Workout{
		Name:                 form.Name,
		Description:          form.Description,
		SportType:            form.SportType,
		StartDate:            startDate,
		DurationSeconds:      form.DurationSeconds,
		DurationTotalSeconds: form.DurationTotalSeconds,
		Distance:             form.Distance,
		SpeedMaxKmh:          speedMaxKmh,
		SpeedAvgKmh:          speedAvgKmh,
		Equipment:            equipmentItems,
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

	created, err := a.Workouts.CreateWithTrack(nickname, workout, trackInput)
	if err != nil {
		handleCreateWorkoutError(ctx, err)
		return
	}

	photos, err := readWorkoutPhotos(ctx)
	if err != nil {
		handleCreateWorkoutError(ctx, err)
		return
	}
	created, err = a.attachWorkoutPhotos(nickname, created, photos)
	if err != nil {
		handleCreateWorkoutError(ctx, err)
		return
	}

	if userID, err := a.currentUserID(ctx); err == nil {
		a.saveLastEquipmentForSport(userID, form.SportType, equipmentIDs)
	}

	a.publishCreatedWorkout(nickname, created)
	a.scheduleEquipmentDistanceRecalc(nickname, created.Equipment)
	slog.Info("workout created",
		"user", nickname,
		"workout_id", created.ID,
		"sport_type", created.SportType,
		"has_track", created.Track != "",
		"has_media", created.HasMedia,
	)
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
func (a *App) parseTrack(ctx *gin.Context) {
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
func (a *App) getWorkoutTrack(ctx *gin.Context) {
		nickname, err := a.currentUserNickname(ctx)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, ErrorResponse{Error: "user not found"})
		return
	}

	owner, workoutID, err := a.resolveWorkoutOwner(ctx, nickname)
	if err != nil {
		if errors.Is(err, workouts.ErrWorkoutNotFound) {
			ctx.JSON(http.StatusNotFound, ErrorResponse{Error: "track not found"})
			return
		}
		respondInternal(ctx, "failed to resolve workout", err)
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

	data, storageName, workoutName, err := a.Workouts.TrackFile(owner, workoutID)
	if err != nil {
		if errors.Is(err, workouts.ErrWorkoutNotFound) {
						if a.Federation.Inbox() != nil {
				data, storageName, workoutName, err = a.Federation.Inbox().TrackFile(nickname, owner, workoutID)
			}
		}
	}
	if err != nil {
		if errors.Is(err, workouts.ErrWorkoutNotFound) {
			ctx.JSON(http.StatusNotFound, ErrorResponse{Error: "track not found"})
			return
		}
		respondInternal(ctx, "failed to load track", err)
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
			respondInternal(ctx, "failed to export track", err)
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
func (a *App) getWorkoutMapPreview(ctx *gin.Context) {
	nickname, err := a.currentUserNickname(ctx)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, ErrorResponse{Error: "user not found"})
		return
	}

	owner, workoutID, err := a.resolveWorkoutOwner(ctx, nickname)
	if err != nil {
		if errors.Is(err, workouts.ErrWorkoutNotFound) {
			ctx.JSON(http.StatusNotFound, ErrorResponse{Error: "map preview not found"})
			return
		}
		respondInternal(ctx, "failed to resolve workout", err)
		return
	}

	data, err := a.Workouts.MapPreview(owner, workoutID)
	if err != nil && errors.Is(err, workouts.ErrWorkoutNotFound) {
		data, err = a.Federation.Inbox().MapPreview(nickname, owner, workoutID)
	}
	if err != nil {
		if errors.Is(err, workouts.ErrWorkoutNotFound) {
			ctx.JSON(http.StatusNotFound, ErrorResponse{Error: "map preview not found"})
			return
		}
		respondInternal(ctx, "failed to load map preview", err)
		return
	}

	ctx.Header("Cache-Control", "public, max-age=31536000, immutable")
	ctx.Header("Content-Type", "image/webp")
	ctx.Data(http.StatusOK, "image/webp", data)
}

// getWorkout godoc
// @Summary      Get workout
// @Description  Return a single workout by ID. Use owner query for followed users' workouts (same as track/media).
// @Tags         workouts
// @Produce      json
// @Security     BearerAuth
// @Param        id     path   string  true   "Workout ID"
// @Param        owner  query  string  false  "Workout owner nickname (required for followed users' workouts)"
// @Success      200  {object}  WorkoutResponse
// @Failure      401  {object}  ErrorResponse
// @Failure      404  {object}  ErrorResponse
// @Failure      500  {object}  ErrorResponse
// @Router       /workouts/{id} [get]
func (a *App) getWorkout(ctx *gin.Context) {
	nickname, err := a.currentUserNickname(ctx)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, ErrorResponse{Error: "user not found"})
		return
	}

	owner, workoutID, err := a.resolveWorkoutOwner(ctx, nickname)
	if err != nil {
		if errors.Is(err, workouts.ErrWorkoutNotFound) {
			ctx.JSON(http.StatusNotFound, ErrorResponse{Error: "workout not found"})
			return
		}
		respondInternal(ctx, "failed to resolve workout", err)
		return
	}

	workout, err := a.Workouts.Get(owner, workoutID)
	if err == nil {
		ctx.JSON(http.StatusOK, a.toLocalWorkoutResponse(owner, workout))
		return
	}
	if !errors.Is(err, workouts.ErrWorkoutNotFound) {
		respondInternal(ctx, "failed to load workout", err)
		return
	}

	if a.Federation.Inbox() == nil {
		ctx.JSON(http.StatusNotFound, ErrorResponse{Error: "workout not found"})
		return
	}
	item, err := a.Federation.Inbox().Get(nickname, owner, workoutID)
	if err != nil {
		if errors.Is(err, workouts.ErrWorkoutNotFound) {
			ctx.JSON(http.StatusNotFound, ErrorResponse{Error: "workout not found"})
			return
		}
		respondInternal(ctx, "failed to load workout", err)
		return
	}
	ctx.JSON(http.StatusOK, toFeedWorkoutResponse(item))
}

func (a *App) toLocalWorkoutResponse(ownerNickname string, workout *workouts.Workout) WorkoutResponse {
	name := ownerNickname
	if user, err := a.Users.FindByNickname(ownerNickname); err == nil && user != nil {
		name = user.Name
	}
	domain := config.Cfg.Federation.Domain
	if domain == "" {
		domain = "localhost"
	}
	hasAvatar, avatarURL := a.localAvatarFieldsForUser(ownerNickname)
	item := workouts.FeedWorkout{
		Workout: *workout,
		Owner:   ownerNickname,
		Author: workouts.FeedAuthor{
			Nickname:  ownerNickname,
			Name:      name,
			Handle:    ownerNickname + "@" + domain,
			IsLocal:   true,
			HasAvatar: hasAvatar,
			AvatarURL: avatarURL,
		},
	}
	return toFeedWorkoutResponse(&item)
}

func (a *App) getWorkoutMediaPreview(ctx *gin.Context) {
	nickname, err := a.currentUserNickname(ctx)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, ErrorResponse{Error: "user not found"})
		return
	}

	owner, workoutID, err := a.resolveWorkoutOwner(ctx, nickname)
	if err != nil {
		if errors.Is(err, workouts.ErrWorkoutNotFound) {
			ctx.JSON(http.StatusNotFound, ErrorResponse{Error: "photo not found"})
			return
		}
		respondInternal(ctx, "failed to resolve workout", err)
		return
	}

	filename := ctx.Param("filename")
	data, err := a.Workouts.MediaPreview(owner, workoutID, filename)
	if err != nil && (errors.Is(err, workouts.ErrPhotoNotFound) || errors.Is(err, workouts.ErrWorkoutNotFound)) {
		data, err = a.Federation.Inbox().MediaPreview(nickname, owner, workoutID, filename)
	}
	if err != nil {
		if errors.Is(err, workouts.ErrPhotoNotFound) || errors.Is(err, workouts.ErrWorkoutNotFound) {
			ctx.JSON(http.StatusNotFound, ErrorResponse{Error: "photo not found"})
			return
		}
		respondInternal(ctx, "failed to load photo preview", err)
		return
	}

	ctx.Header("Cache-Control", "public, max-age=31536000, immutable")
	ctx.Header("Content-Type", "image/webp")
	ctx.Data(http.StatusOK, "image/webp", data)
}

func (a *App) getWorkoutMediaOriginal(ctx *gin.Context) {
	nickname, err := a.currentUserNickname(ctx)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, ErrorResponse{Error: "user not found"})
		return
	}

	owner, workoutID, err := a.resolveWorkoutOwner(ctx, nickname)
	if err != nil {
		if errors.Is(err, workouts.ErrWorkoutNotFound) {
			ctx.JSON(http.StatusNotFound, ErrorResponse{Error: "photo not found"})
			return
		}
		respondInternal(ctx, "failed to resolve workout", err)
		return
	}

	filename := ctx.Param("filename")
	data, contentType, err := a.Workouts.MediaOriginal(owner, workoutID, filename)
	if err != nil && (errors.Is(err, workouts.ErrPhotoNotFound) || errors.Is(err, workouts.ErrWorkoutNotFound)) {
		data, contentType, err = a.Federation.Inbox().MediaOriginal(nickname, owner, workoutID, filename)
	}
	if err != nil {
		if errors.Is(err, workouts.ErrPhotoNotFound) || errors.Is(err, workouts.ErrWorkoutNotFound) {
			ctx.JSON(http.StatusNotFound, ErrorResponse{Error: "photo not found"})
			return
		}
		respondInternal(ctx, "failed to load photo", err)
		return
	}

	ctx.Header("Cache-Control", "public, max-age=31536000, immutable")
	ctx.Header("Content-Type", contentType)
	ctx.Data(http.StatusOK, contentType, data)
}

// listWorkouts godoc
// @Summary      List workouts
// @Description  Return a cursor page of workouts for the authenticated user sorted by start date descending (id descending tie-breaker). Use scope=feed for the full feed (default) or scope=own for only the viewer's workouts. Default limit is 20 (max 100).
// @Tags         workouts
// @Produce      json
// @Security     BearerAuth
// @Param        scope   query  string  false  "feed (default) or own"
// @Param        limit   query  int     false  "page size (default 20, max 100)"
// @Param        cursor  query  string  false  "opaque cursor from previous page next_cursor"
// @Success      200  {object}  WorkoutListResponse
// @Failure      400  {object}  ErrorResponse
// @Failure      401  {object}  ErrorResponse
// @Failure      500  {object}  ErrorResponse
// @Router       /workouts [get]
func (a *App) listWorkouts(ctx *gin.Context) {
	nickname, err := a.currentUserNickname(ctx)
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

	limit := workouts.DefaultPageLimit
	if raw := strings.TrimSpace(ctx.Query("limit")); raw != "" {
		parsed, err := strconv.Atoi(raw)
		if err != nil || parsed < 1 {
			ctx.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid limit"})
			return
		}
		limit = workouts.ClampLimit(parsed)
	}

	var cursor *workouts.Cursor
	if raw := strings.TrimSpace(ctx.Query("cursor")); raw != "" {
		cursor, err = workouts.DecodeCursor(raw)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid cursor"})
			return
		}
	}

	userID, err := a.currentUserID(ctx)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, ErrorResponse{Error: "invalid token"})
		return
	}

	viewer, err := a.Users.FindByID(userID)
	if err != nil {
		respondInternal(ctx, "user not found", err)
		return
	}

	feedSvc := a.newFeedService()
	var page workouts.Page

	if scope == "own" {
		page, err = feedSvc.ListOwnPage(nickname, viewer.Name, cursor, limit)
	} else {
		var follows []social.Follow
		follows, err = a.Social.ListFollowing(userID)
		if err != nil {
			respondInternal(ctx, "failed to list following", err)
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
			hasAvatar, avatarURL := a.localAvatarFieldsForUser(follows[i].TargetNickname)
			followedAuthors = append(followedAuthors, workouts.FeedAuthor{
				Nickname:  follows[i].TargetNickname,
				Name:      follows[i].TargetName,
				Handle:    follows[i].TargetHandle,
				IsLocal:   true,
				HasAvatar: hasAvatar,
				AvatarURL: avatarURL,
			})
		}

		page, err = feedSvc.ListFeedPage(nickname, viewer.Name, followedAuthors, cursor, limit)
	}
	if err != nil {
		respondInternal(ctx, "failed to list workouts", err)
		return
	}

	response := WorkoutListResponse{
		Items:      make([]WorkoutResponse, 0, len(page.Items)),
		NextCursor: page.NextCursor,
		HasMore:    page.HasMore,
	}
	for i := range page.Items {
		response.Items = append(response.Items, toFeedWorkoutResponse(&page.Items[i]))
	}

	ctx.JSON(http.StatusOK, response)
}

// deleteWorkout godoc
// @Summary      Delete workout
// @Description  Permanently delete the authenticated user's workout and notify federation followers
// @Tags         workouts
// @Produce      json
// @Security     BearerAuth
// @Param        id  path  string  true  "Workout ID"
// @Success      204
// @Failure      401  {object}  ErrorResponse
// @Failure      404  {object}  ErrorResponse
// @Failure      500  {object}  ErrorResponse
// @Router       /workouts/{id} [delete]
func (a *App) deleteWorkout(ctx *gin.Context) {
		nickname, err := a.currentUserNickname(ctx)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, ErrorResponse{Error: "user not found"})
		return
	}

	workoutID := ctx.Param("id")
	workout, err := a.Workouts.Get(nickname, workoutID)
	if err != nil {
		if errors.Is(err, workouts.ErrWorkoutNotFound) {
			ctx.JSON(http.StatusNotFound, ErrorResponse{Error: "workout not found"})
			return
		}
		respondInternal(ctx, "failed to load workout", err)
		return
	}

	if err := a.Workouts.Delete(nickname, workoutID); err != nil {
		if errors.Is(err, workouts.ErrWorkoutNotFound) {
			ctx.JSON(http.StatusNotFound, ErrorResponse{Error: "workout not found"})
			return
		}
		respondInternal(ctx, "failed to delete workout", err)
		return
	}

	a.publishDeletedWorkout(nickname, workoutID)
	a.scheduleEquipmentDistanceRecalc(nickname, workout.Equipment)
	slog.Info("workout deleted", "user", nickname, "workout_id", workoutID)
	ctx.Status(http.StatusNoContent)
}

// updateWorkout godoc
// @Summary      Update workout
// @Description  Update metadata of the authenticated user's workout and notify federation followers
// @Tags         workouts
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id    path  string  true  "Workout ID"
// @Param        body  body  CreateWorkoutRequest  true  "Workout data"
// @Success      200   {object}  WorkoutResponse
// @Failure      400   {object}  ErrorResponse
// @Failure      401   {object}  ErrorResponse
// @Failure      404   {object}  ErrorResponse
// @Failure      500   {object}  ErrorResponse
// @Router       /workouts/{id} [put]
func (a *App) updateWorkout(ctx *gin.Context) {
	nickname, err := a.currentUserNickname(ctx)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, ErrorResponse{Error: "user not found"})
		return
	}

	workoutID := ctx.Param("id")
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

	equipmentItems, err := a.resolveWorkoutEquipment(nickname, req.EquipmentIDs)
	if err != nil {
		respondInternal(ctx, "failed to resolve equipment", err)
		return
	}

	updated, err := a.Workouts.Update(nickname, workoutID, workoutFromCreateRequest(req, startDate, equipmentItems))
	if err != nil {
		handleUpdateWorkoutError(ctx, err)
		return
	}

	if userID, err := a.currentUserID(ctx); err == nil {
		a.saveLastEquipmentForSport(userID, req.SportType, req.EquipmentIDs)
	}

	a.publishUpdatedWorkout(nickname, updated)
	a.scheduleEquipmentDistanceRecalc(nickname, updated.Equipment)
	slog.Info("workout updated",
		"user", nickname,
		"workout_id", updated.ID,
		"sport_type", updated.SportType,
	)
	resp := toWorkoutResponse(updated)
	resp.Owner = nickname
	ctx.JSON(http.StatusOK, resp)
}

func handleUpdateWorkoutError(ctx *gin.Context, err error) {
	switch {
	case errors.Is(err, workouts.ErrWorkoutNotFound):
		ctx.JSON(http.StatusNotFound, ErrorResponse{Error: "workout not found"})
	case errors.Is(err, workouts.ErrInvalidSportType), errors.Is(err, workouts.ErrInvalidWorkout):
		ctx.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
	case errors.Is(err, workouts.ErrWorkoutExists):
		ctx.JSON(http.StatusConflict, ErrorResponse{Error: err.Error()})
	default:
		respondInternal(ctx, "failed to update workout", err)
	}
}
