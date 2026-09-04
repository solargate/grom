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
	"github.com/gin-gonic/gin/binding"
	"github.com/solargate/grom/internal/auth"
	"github.com/solargate/grom/internal/config"
	"github.com/solargate/grom/internal/social"
	"github.com/solargate/grom/internal/tracks"
	"github.com/solargate/grom/internal/workouts"
)

type ExternalIDRequest struct {
	Name string `json:"name" example:"device-import"`
	ID   string `json:"id" example:"CYCLING 2026.07.30 16.26 Strava.csv"`
}

type CreateWorkoutRequest struct {
	Name                 string             `json:"name" binding:"required" example:"Morning run"`
	Description          string             `json:"description" example:"Easy session"`
	SportType            string             `json:"sport_type" binding:"required" example:"Run"`
	StartDate            string             `json:"start_date" binding:"required" example:"2026-07-05T14:30:00+03:00"`
	DurationSeconds      int                `json:"duration_seconds" example:"3600"`
	DurationTotalSeconds int                `json:"duration_total_seconds,omitempty" example:"3900"`
	Distance             float64            `json:"distance" example:"5200"`
	SpeedMaxKmh          *float64           `json:"speed_max_kmh,omitempty" example:"32.5"`
	SpeedAvgKmh          *float64           `json:"speed_avg_kmh,omitempty" example:"18.2"`
	EquipmentIDs         []string           `json:"equipment_ids" example:"550e8400-e29b-41d4-a716-446655440000"`
	ExternalID           *ExternalIDRequest `json:"external_id,omitempty"`
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
	ExternalIDName       string                `form:"external_id_name"`
	ExternalIDID         string                `form:"external_id_id"`
	Track                *multipart.FileHeader `form:"track"`
}

type ExternalIDExistsResponse struct {
	Exists bool `json:"exists" example:"true"`
}

type ParseTrackResponse struct {
	Name                 string   `json:"name,omitempty" example:"Morning run"`
	SportType            string   `json:"sport_type,omitempty" example:"Run"`
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
	SpeedMaxKmh          *float64               `json:"speed_max_kmh,omitempty" example:"32.4"`
	SpeedAvgKmh          *float64               `json:"speed_avg_kmh,omitempty" example:"17.5"`
	ElevationGain        *float64               `json:"elevation_gain,omitempty" example:"77"`
	HeartRateMax         *float64               `json:"heart_rate_max,omitempty" example:"187"`
	HeartRateAvg         *float64               `json:"heart_rate_avg,omitempty" example:"130"`
	StepsTotal           *int                   `json:"steps_total,omitempty" example:"2583"`
	Calories             *float64               `json:"calories,omitempty" example:"415"`
	Track                string                 `json:"track,omitempty" example:"track.gpx"`
	ExternalID           *ExternalIDRequest     `json:"external_id,omitempty"`
	Equipment            []WorkoutEquipmentItem `json:"equipment,omitempty"`
	HasMapPreview        bool                   `json:"has_map_preview" example:"true"`
	HasMedia             bool                   `json:"has_media" example:"true"`
	MediaFiles           []string               `json:"media_files,omitempty"`
	LikesCount           int                    `json:"likes_count" example:"5"`
	LikedByMe            bool                   `json:"liked_by_me" example:"false"`
	CanLike              bool                   `json:"can_like" example:"true"`
	CommentsCount        int                    `json:"comments_count" example:"2"`
	Author               *WorkoutAuthorResponse `json:"author,omitempty"`
}

// WorkoutSpeedSampleResponse is one point of the per-workout speed series.
type WorkoutSpeedSampleResponse struct {
	T         string  `json:"t" example:"2026-07-05T14:30:01Z"`
	SpeedKmh  float64 `json:"speed_kmh" example:"18.4"`
	DistanceM float64 `json:"distance_m" example:"12.5"`
}

// WorkoutSpeedResponse is the speed series for a workout detail chart.
type WorkoutSpeedResponse struct {
	Samples     []WorkoutSpeedSampleResponse `json:"samples"`
	SpeedMaxKmh *float64                     `json:"speed_max_kmh,omitempty" example:"32.4"`
	SpeedAvgKmh *float64                     `json:"speed_avg_kmh,omitempty" example:"17.5"`
}

// WorkoutHeartRateSampleResponse is one point of the per-workout heart-rate series.
type WorkoutHeartRateSampleResponse struct {
	T            string   `json:"t" example:"2026-07-05T14:30:01Z"`
	HeartRateBPM float64  `json:"heart_rate_bpm" example:"142"`
	DistanceM    *float64 `json:"distance_m,omitempty" example:"12.5"`
}

// WorkoutHeartRateResponse is the heart-rate series for a workout detail chart.
type WorkoutHeartRateResponse struct {
	Samples      []WorkoutHeartRateSampleResponse `json:"samples"`
	HeartRateMax *float64                         `json:"heart_rate_max,omitempty" example:"187"`
	HeartRateAvg *float64                         `json:"heart_rate_avg,omitempty" example:"130"`
	HasGPS       bool                             `json:"has_gps" example:"true"`
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
	var externalID *ExternalIDRequest
	if workout.ExternalID != nil {
		externalID = &ExternalIDRequest{
			Name: workout.ExternalID.Name,
			ID:   workout.ExternalID.ID,
		}
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
		SpeedMaxKmh:          workout.SpeedMaxKmh,
		SpeedAvgKmh:          workout.SpeedAvgKmh,
		ElevationGain:        workout.ElevationGain,
		HeartRateMax:         workout.HeartRateMax,
		HeartRateAvg:         workout.HeartRateAvg,
		StepsTotal:           workout.StepsTotal,
		Calories:             workout.Calories,
		Track:                workout.Track,
		ExternalID:           externalID,
		Equipment:            equipment,
		HasMapPreview:        workout.HasMapPreview,
		HasMedia:             workout.HasMedia,
		MediaFiles:           workout.MediaFiles,
		LikesCount:           workout.LikesCount,
		CommentsCount:        workout.CommentsCount,
	}
}

func toWorkoutSpeedResponse(workout *workouts.Workout, samples []workouts.SpeedSample) WorkoutSpeedResponse {
	outSamples := make([]WorkoutSpeedSampleResponse, 0, len(samples))
	for _, s := range samples {
		outSamples = append(outSamples, WorkoutSpeedSampleResponse{
			T:         s.Time.UTC().Format(time.RFC3339),
			SpeedKmh:  s.SpeedKmh,
			DistanceM: s.DistanceM,
		})
	}
	return WorkoutSpeedResponse{
		Samples:     outSamples,
		SpeedMaxKmh: workout.SpeedMaxKmh,
		SpeedAvgKmh: workout.SpeedAvgKmh,
	}
}

func toWorkoutHeartRateResponse(workout *workouts.Workout, samples []workouts.HeartRateSample) WorkoutHeartRateResponse {
	outSamples := make([]WorkoutHeartRateSampleResponse, 0, len(samples))
	hasGPS := false
	for _, s := range samples {
		if s.DistanceM != nil {
			hasGPS = true
		}
		outSamples = append(outSamples, WorkoutHeartRateSampleResponse{
			T:            s.Time.UTC().Format(time.RFC3339),
			HeartRateBPM: s.BPM,
			DistanceM:    s.DistanceM,
		})
	}
	return WorkoutHeartRateResponse{
		Samples:      outSamples,
		HeartRateMax: workout.HeartRateMax,
		HeartRateAvg: workout.HeartRateAvg,
		HasGPS:       hasGPS,
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
		Name:                 meta.Name,
		SportType:            meta.SportType,
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
		ExternalID:           externalIDFromRequest(req.ExternalID),
	}
}

func externalIDFromRequest(req *ExternalIDRequest) *workouts.ExternalID {
	if req == nil {
		return nil
	}
	name := strings.TrimSpace(req.Name)
	id := strings.TrimSpace(req.ID)
	if name == "" || id == "" {
		return nil
	}
	return &workouts.ExternalID{Name: name, ID: id}
}

func externalIDFromForm(name, id string) *workouts.ExternalID {
	return externalIDFromRequest(&ExternalIDRequest{Name: name, ID: id})
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

// parseSportTypesQuery splits a comma-separated sport_types query into an allow-set.
// Empty tokens are dropped; unknown ids are kept (they simply match nothing).
func parseSportTypesQuery(raw string) map[string]struct{} {
	parts := strings.Split(raw, ",")
	out := make(map[string]struct{}, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		out[part] = struct{}{}
	}
	return out
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
	if auth.IsPAT(ctx) {
		ownerNickname = strings.TrimSpace(ctx.Query("owner"))
		if ownerNickname != "" && ownerNickname != viewerNickname {
			return "", "", workouts.ErrWorkoutNotFound
		}
		return viewerNickname, workoutID, nil
	}
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
	case errors.Is(err, workouts.ErrWorkoutExists),
		errors.Is(err, workouts.ErrExternalIDExists):
		ctx.JSON(http.StatusConflict, ErrorResponse{Error: err.Error()})
	default:
		respondInternal(ctx, "failed to create workout", err)
	}
}

// createWorkout godoc
// @Summary      Create workout
// @Description  Create a manual workout for the authenticated user. When equipment_ids is omitted, equipment is taken from the user's profile last_equipment_by_sport for the sport_type. An explicit empty equipment_ids list means no equipment.
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
// @Param        equipment_ids  formData  string  false  "JSON array of equipment IDs; omit to use profile last_equipment_by_sport for sport_type; [] for none"
// @Param        track  formData  file  false  "Track file FIT or GPX (multipart)"
// @Success      201   {object}  WorkoutResponse
// @Failure      400   {object}  ErrorResponse  "Invalid workout data or track"
// @Failure      401   {object}  ErrorResponse  "Unauthorized"
// @Failure      500   {object}  ErrorResponse  "Internal server error"
// @Router       /workouts [post]
func (a *App) createWorkout(ctx *gin.Context) {
	nickname, err := a.currentUserNickname(ctx)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, ErrorResponse{Error: "user not found"})
		return
	}
	userID, err := a.currentUserID(ctx)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, ErrorResponse{Error: "user not found"})
		return
	}

	contentType := ctx.GetHeader("Content-Type")
	if strings.HasPrefix(contentType, "multipart/form-data") {
		a.createWorkoutMultipart(ctx, nickname, userID)
		return
	}

	body, err := io.ReadAll(ctx.Request.Body)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, ErrorResponse{Error: "failed to read request body"})
		return
	}
	req, equipmentProvided, err := decodeCreateWorkoutJSON(body)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}
	if err := binding.Validator.ValidateStruct(req); err != nil {
		ctx.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	startDate, err := time.Parse(time.RFC3339, req.StartDate)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid start_date format, expected RFC3339"})
		return
	}

	equipmentItems, resolvedIDs, err := a.resolveEquipmentForCreate(nickname, userID, req.SportType, req.EquipmentIDs, equipmentProvided)
	if err != nil {
		respondInternal(ctx, "failed to resolve equipment", err)
		return
	}

	workout, err := a.Workouts.Create(nickname, workoutFromCreateRequest(req, startDate, equipmentItems))
	if err != nil {
		handleCreateWorkoutError(ctx, err)
		return
	}

	a.saveLastEquipmentForSport(userID, req.SportType, resolvedIDs)
	a.rememberUsedSportType(userID, req.SportType)
	a.scheduleRefreshLastSportType(nickname, userID)

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

func decodeCreateWorkoutJSON(body []byte) (CreateWorkoutRequest, bool, error) {
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(body, &raw); err != nil {
		return CreateWorkoutRequest{}, false, err
	}
	_, equipmentProvided := raw["equipment_ids"]
	var req CreateWorkoutRequest
	if err := json.Unmarshal(body, &req); err != nil {
		return CreateWorkoutRequest{}, false, err
	}
	return req, equipmentProvided, nil
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

func (a *App) createWorkoutMultipart(ctx *gin.Context, nickname, userID string) {
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

	_, equipmentProvided := ctx.GetPostForm("equipment_ids")
	equipmentIDs, err := parseEquipmentIDsForm(form.EquipmentIDs)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid equipment_ids format"})
		return
	}
	equipmentItems, resolvedIDs, err := a.resolveEquipmentForCreate(nickname, userID, form.SportType, equipmentIDs, equipmentProvided)
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
		ExternalID:           externalIDFromForm(form.ExternalIDName, form.ExternalIDID),
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

	if userID != "" {
		a.saveLastEquipmentForSport(userID, form.SportType, resolvedIDs)
		a.rememberUsedSportType(userID, form.SportType)
		a.scheduleRefreshLastSportType(nickname, userID)
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
// @Failure      400  {object}  ErrorResponse  "Track file required or invalid"
// @Failure      401  {object}  ErrorResponse  "Unauthorized"
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

// getWorkoutSpeed godoc
// @Summary      Get workout speed series
// @Description  Return the precomputed speed chart series (up to 500 points). Use owner query for followed users' workouts (same as track/media). Empty samples when no chart exists.
// @Tags         workouts
// @Produce      json
// @Security     BearerAuth
// @Param        id     path   string  true   "Workout ID"
// @Param        owner  query  string  false  "Workout owner nickname (required for followed users' workouts)"
// @Success      200  {object}  WorkoutSpeedResponse
// @Failure      401  {object}  ErrorResponse  "Unauthorized"
// @Failure      404  {object}  ErrorResponse  "Workout not found"
// @Failure      500  {object}  ErrorResponse  "Internal server error"
// @Router       /workouts/{id}/speed [get]
func (a *App) getWorkoutSpeed(ctx *gin.Context) {
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

	workout, samples, err := a.Workouts.GetSpeedChart(owner, workoutID)
	if err == nil {
		ctx.JSON(http.StatusOK, toWorkoutSpeedResponse(workout, samples))
		return
	}
	if !errors.Is(err, workouts.ErrWorkoutNotFound) {
		respondInternal(ctx, "failed to load workout speed", err)
		return
	}

	if auth.IsPAT(ctx) || a.Federation.Inbox() == nil {
		ctx.JSON(http.StatusNotFound, ErrorResponse{Error: "workout not found"})
		return
	}
	workout, samples, err = a.Federation.Inbox().GetSpeedChart(nickname, owner, workoutID)
	if err != nil {
		if errors.Is(err, workouts.ErrWorkoutNotFound) {
			ctx.JSON(http.StatusNotFound, ErrorResponse{Error: "workout not found"})
			return
		}
		respondInternal(ctx, "failed to load workout speed", err)
		return
	}
	ctx.JSON(http.StatusOK, toWorkoutSpeedResponse(workout, samples))
}

// getWorkoutHeartRate godoc
// @Summary      Get workout heart-rate series
// @Description  Return the precomputed heart-rate chart series (up to 500 points). Use owner query for followed users' workouts (same as track/media). Empty samples when no chart exists. distance_m is omitted when the track has no GPS; has_gps indicates whether the X axis should use distance.
// @Tags         workouts
// @Produce      json
// @Security     BearerAuth
// @Param        id     path   string  true   "Workout ID"
// @Param        owner  query  string  false  "Workout owner nickname (required for followed users' workouts)"
// @Success      200  {object}  WorkoutHeartRateResponse
// @Failure      401  {object}  ErrorResponse  "Unauthorized"
// @Failure      404  {object}  ErrorResponse  "Workout not found"
// @Failure      500  {object}  ErrorResponse  "Internal server error"
// @Router       /workouts/{id}/heartrate [get]
func (a *App) getWorkoutHeartRate(ctx *gin.Context) {
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

	workout, samples, err := a.Workouts.GetHeartRateChart(owner, workoutID)
	if err == nil {
		ctx.JSON(http.StatusOK, toWorkoutHeartRateResponse(workout, samples))
		return
	}
	if !errors.Is(err, workouts.ErrWorkoutNotFound) {
		respondInternal(ctx, "failed to load workout heart rate", err)
		return
	}

	if auth.IsPAT(ctx) || a.Federation.Inbox() == nil {
		ctx.JSON(http.StatusNotFound, ErrorResponse{Error: "workout not found"})
		return
	}
	workout, samples, err = a.Federation.Inbox().GetHeartRateChart(nickname, owner, workoutID)
	if err != nil {
		if errors.Is(err, workouts.ErrWorkoutNotFound) {
			ctx.JSON(http.StatusNotFound, ErrorResponse{Error: "workout not found"})
			return
		}
		respondInternal(ctx, "failed to load workout heart rate", err)
		return
	}
	ctx.JSON(http.StatusOK, toWorkoutHeartRateResponse(workout, samples))
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
// @Failure      401  {object}  ErrorResponse  "Unauthorized"
// @Failure      403  {object}  ErrorResponse  "Original track download not allowed"
// @Failure      404  {object}  ErrorResponse  "Track not found"
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
		if errors.Is(err, workouts.ErrWorkoutNotFound) && !auth.IsPAT(ctx) && a.Federation.Inbox() != nil {
			data, storageName, workoutName, err = a.Federation.Inbox().TrackFile(nickname, owner, workoutID)
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
// @Failure      401  {object}  ErrorResponse  "Unauthorized"
// @Failure      404  {object}  ErrorResponse  "Map preview not found"
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
	if err != nil && errors.Is(err, workouts.ErrWorkoutNotFound) && !auth.IsPAT(ctx) && a.Federation.Inbox() != nil {
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
// @Failure      401  {object}  ErrorResponse  "Unauthorized"
// @Failure      404  {object}  ErrorResponse  "Workout not found"
// @Failure      500  {object}  ErrorResponse  "Internal server error"
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
		resp := a.toLocalWorkoutResponse(owner, workout)
		a.applyLikesSummaryToLocalWorkout(nickname, owner, workout, &resp)
		a.applyCommentsSummaryToLocalWorkout(nickname, owner, workout, &resp)
		ctx.JSON(http.StatusOK, resp)
		return
	}
	if !errors.Is(err, workouts.ErrWorkoutNotFound) {
		respondInternal(ctx, "failed to load workout", err)
		return
	}

	if auth.IsPAT(ctx) || a.Federation.Inbox() == nil {
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
	resp := toFeedWorkoutResponse(item)
	a.applyLikesSummary(nickname, item, &resp)
	a.applyCommentsSummary(nickname, item, &resp)
	ctx.JSON(http.StatusOK, resp)
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
	if err != nil && (errors.Is(err, workouts.ErrPhotoNotFound) || errors.Is(err, workouts.ErrWorkoutNotFound)) && !auth.IsPAT(ctx) && a.Federation.Inbox() != nil {
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
	if err != nil && (errors.Is(err, workouts.ErrPhotoNotFound) || errors.Is(err, workouts.ErrWorkoutNotFound)) && !auth.IsPAT(ctx) && a.Federation.Inbox() != nil {
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

// addWorkoutMedia godoc
// @Summary      Add workout photos
// @Description  Append photos to the authenticated user's workout (original + preview). Max 20 photos per workout.
// @Tags         workouts
// @Accept       mpfd
// @Produce      json
// @Security     BearerAuth
// @Param        id      path      string  true  "Workout ID"
// @Param        photos  formData  file    true  "Photo files (repeatable)"
// @Success      200     {object}  WorkoutResponse
// @Failure      400     {object}  ErrorResponse  "Photos required or invalid"
// @Failure      401     {object}  ErrorResponse  "Unauthorized"
// @Failure      404     {object}  ErrorResponse  "Workout not found"
// @Failure      500     {object}  ErrorResponse  "Internal server error"
// @Router       /workouts/{id}/media [post]
func (a *App) addWorkoutMedia(ctx *gin.Context) {
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

	photos, err := readWorkoutPhotos(ctx)
	if err != nil {
		handleWorkoutMediaError(ctx, err)
		return
	}
	if len(photos) == 0 {
		ctx.JSON(http.StatusBadRequest, ErrorResponse{Error: "photos required"})
		return
	}

	updated, err := a.Workouts.AddMedia(nickname, workout, photos)
	if err != nil {
		handleWorkoutMediaError(ctx, err)
		return
	}

	a.publishUpdatedWorkout(nickname, updated)
	slog.Info("workout media added",
		"user", nickname,
		"workout_id", updated.ID,
		"photos", len(photos),
		"media_count", len(updated.MediaFiles),
	)
	resp := toWorkoutResponse(updated)
	resp.Owner = nickname
	ctx.JSON(http.StatusOK, resp)
}

// deleteWorkoutMedia godoc
// @Summary      Delete workout photo
// @Description  Remove one photo (original + preview) from the authenticated user's workout
// @Tags         workouts
// @Produce      json
// @Security     BearerAuth
// @Param        id        path  string  true  "Workout ID"
// @Param        filename  path  string  true  "Photo filename"
// @Success      200       {object}  WorkoutResponse
// @Failure      400       {object}  ErrorResponse  "Invalid request"
// @Failure      401       {object}  ErrorResponse  "Unauthorized"
// @Failure      404       {object}  ErrorResponse  "Workout or photo not found"
// @Failure      500       {object}  ErrorResponse  "Internal server error"
// @Router       /workouts/{id}/media/{filename} [delete]
func (a *App) deleteWorkoutMedia(ctx *gin.Context) {
	nickname, err := a.currentUserNickname(ctx)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, ErrorResponse{Error: "user not found"})
		return
	}

	workoutID := ctx.Param("id")
	filename := ctx.Param("filename")
	updated, err := a.Workouts.RemoveMedia(nickname, workoutID, filename)
	if err != nil {
		handleWorkoutMediaError(ctx, err)
		return
	}

	a.publishUpdatedWorkout(nickname, updated)
	slog.Info("workout media deleted",
		"user", nickname,
		"workout_id", updated.ID,
		"filename", filename,
		"media_count", len(updated.MediaFiles),
	)
	resp := toWorkoutResponse(updated)
	resp.Owner = nickname
	ctx.JSON(http.StatusOK, resp)
}

func handleWorkoutMediaError(ctx *gin.Context, err error) {
	switch {
	case errors.Is(err, workouts.ErrWorkoutNotFound):
		ctx.JSON(http.StatusNotFound, ErrorResponse{Error: "workout not found"})
	case errors.Is(err, workouts.ErrPhotoNotFound):
		ctx.JSON(http.StatusNotFound, ErrorResponse{Error: "photo not found"})
	case errors.Is(err, workouts.ErrInvalidPhotoName),
		errors.Is(err, workouts.ErrInvalidPhoto),
		errors.Is(err, workouts.ErrPhotoTooLarge),
		errors.Is(err, workouts.ErrTooManyPhotos),
		errors.Is(err, workouts.ErrInvalidWorkout):
		ctx.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
	default:
		respondInternal(ctx, "failed to update workout media", err)
	}
}

// checkWorkoutExternalID godoc
// @Summary      Check external workout id
// @Description  Returns whether the authenticated user already has a workout with the given external_id name+id pair
// @Tags         workouts
// @Produce      json
// @Security     BearerAuth
// @Param        name  query  string  true  "external_id.name"
// @Param        id    query  string  true  "external_id.id"
// @Success      200  {object}  ExternalIDExistsResponse
// @Failure      400  {object}  ErrorResponse  "Name and id are required"
// @Failure      401  {object}  ErrorResponse  "Unauthorized"
// @Failure      500  {object}  ErrorResponse  "Internal server error"
// @Router       /workouts/external [get]
func (a *App) checkWorkoutExternalID(ctx *gin.Context) {
	nickname, err := a.currentUserNickname(ctx)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, ErrorResponse{Error: "user not found"})
		return
	}

	name := strings.TrimSpace(ctx.Query("name"))
	id := strings.TrimSpace(ctx.Query("id"))
	if name == "" || id == "" {
		ctx.JSON(http.StatusBadRequest, ErrorResponse{Error: "name and id are required"})
		return
	}

	exists, err := a.Workouts.HasExternalID(nickname, name, id)
	if err != nil {
		respondInternal(ctx, "failed to check external_id", err)
		return
	}
	ctx.JSON(http.StatusOK, ExternalIDExistsResponse{Exists: exists})
}

// listWorkouts godoc
// @Summary      List workouts
// @Description  Return a cursor page of workouts for the authenticated user sorted by start date descending (id descending tie-breaker). Use scope=feed for the full feed (default) or scope=own for only the viewer's workouts. Optional sport_types (comma-separated) filters own workouts only; omit for all types; empty value returns an empty page. Unknown type ids are ignored. Default limit is 20 (max 100).
// @Tags         workouts
// @Produce      json
// @Security     BearerAuth
// @Param        scope        query  string  false  "feed (default) or own"
// @Param        limit        query  int     false  "page size (default 20, max 100)"
// @Param        cursor       query  string  false  "opaque cursor from previous page next_cursor"
// @Param        sport_types  query  string  false  "comma-separated sport type ids; only with scope=own"
// @Success      200  {object}  WorkoutListResponse
// @Failure      400  {object}  ErrorResponse  "Invalid scope, limit, cursor, or sport_types with feed"
// @Failure      401  {object}  ErrorResponse  "Unauthorized"
// @Failure      500  {object}  ErrorResponse  "Internal server error"
// @Router       /workouts [get]
func (a *App) listWorkouts(ctx *gin.Context) {
	nickname, err := a.currentUserNickname(ctx)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, ErrorResponse{Error: "user not found"})
		return
	}

	scope := strings.TrimSpace(ctx.Query("scope"))
	if auth.IsPAT(ctx) {
		scope = "own"
	} else if scope == "" {
		scope = "feed"
	}
	if scope != "feed" && scope != "own" {
		ctx.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid scope"})
		return
	}

	rawSportTypes, sportTypesPresent := ctx.GetQuery("sport_types")
	if sportTypesPresent && scope != "own" {
		ctx.JSON(http.StatusBadRequest, ErrorResponse{Error: "sport_types is only allowed with scope=own"})
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
		var sportTypes map[string]struct{}
		if sportTypesPresent {
			sportTypes = parseSportTypesQuery(rawSportTypes)
			if len(sportTypes) == 0 {
				ctx.JSON(http.StatusOK, WorkoutListResponse{
					Items:   []WorkoutResponse{},
					HasMore: false,
				})
				return
			}
		}
		page, err = feedSvc.ListOwnPage(nickname, viewer.Name, cursor, limit, sportTypes)
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
		resp := toFeedWorkoutResponse(&page.Items[i])
		a.applyLikesSummary(nickname, &page.Items[i], &resp)
		a.applyCommentsSummary(nickname, &page.Items[i], &resp)
		response.Items = append(response.Items, resp)
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
// @Failure      401  {object}  ErrorResponse  "Unauthorized"
// @Failure      404  {object}  ErrorResponse  "Workout not found"
// @Failure      500  {object}  ErrorResponse  "Internal server error"
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
	_ = a.Likes.DeleteLocal(nickname, workoutID)
	_ = a.Comments.DeleteLocal(nickname, workoutID)

	a.publishDeletedWorkout(nickname, workoutID)
	a.scheduleEquipmentDistanceRecalc(nickname, workout.Equipment)
	if userID, err := a.currentUserID(ctx); err == nil {
		a.scheduleRefreshLastSportType(nickname, userID)
	}
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
// @Failure      400   {object}  ErrorResponse  "Invalid workout data"
// @Failure      401   {object}  ErrorResponse  "Unauthorized"
// @Failure      404   {object}  ErrorResponse  "Workout not found"
// @Failure      500   {object}  ErrorResponse  "Internal server error"
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
		a.rememberUsedSportType(userID, req.SportType)
		a.scheduleRefreshLastSportType(nickname, userID)
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
