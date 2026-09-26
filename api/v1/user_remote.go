package v1

import (
	"net/http"
	"sort"
	"strings"

	"github.com/solargate/grom/internal/config"
	"github.com/solargate/grom/internal/federation"
	"github.com/solargate/grom/internal/maprender"
	"github.com/solargate/grom/internal/social"
	"github.com/solargate/grom/internal/tracks"
	"github.com/solargate/grom/internal/workouts"
)

func (a *App) federationHTTPClient() *http.Client {
	if a.federationDelivery != nil {
		return a.federationDelivery.Client()
	}
	return http.DefaultClient
}

func (a *App) fetchRemoteFollowing(parsed social.ParsedHandle) ([]FollowResponse, error) {
	if !config.Cfg.Federation.Enabled {
		return nil, social.ErrRemoteNotReady
	}
	items, err := federation.FetchActorCollectionItems(a.federationHTTPClient(), a.Blobs, parsed, "following", 100)
	if err != nil {
		return nil, err
	}
	out := make([]FollowResponse, 0, len(items))
	for _, u := range items {
		out = append(out, FollowResponse{
			TargetHandle:   u.Handle,
			TargetNickname: u.Nickname,
			TargetName:     u.Name,
			TargetIsLocal:  false,
			Status:         social.StatusActive,
		})
	}
	return out, nil
}

func (a *App) fetchRemoteFollowers(parsed social.ParsedHandle) ([]FollowerResponse, error) {
	if !config.Cfg.Federation.Enabled {
		return nil, social.ErrRemoteNotReady
	}
	items, err := federation.FetchActorCollectionItems(a.federationHTTPClient(), a.Blobs, parsed, "followers", 100)
	if err != nil {
		return nil, err
	}
	out := make([]FollowerResponse, 0, len(items))
	for _, u := range items {
		out = append(out, FollowerResponse{
			FollowerHandle:   u.Handle,
			FollowerNickname: u.Nickname,
			FollowerName:     u.Name,
			FollowerIsLocal:  false,
		})
	}
	return out, nil
}

func (a *App) listRemoteUserWorkouts(viewerNickname string, parsed social.ParsedHandle, cursor *workouts.Cursor, limit int) (WorkoutListResponse, error) {
	if !config.Cfg.Federation.Enabled {
		return WorkoutListResponse{}, social.ErrRemoteNotReady
	}
	fetchLimit := limit * 3
	if fetchLimit < 50 {
		fetchLimit = 50
	}
	if fetchLimit > 200 {
		fetchLimit = 200
	}
	raw, err := federation.FetchOutboxWorkouts(a.federationHTTPClient(), a.Blobs, parsed, fetchLimit)
	if err != nil {
		return WorkoutListResponse{}, err
	}
	if len(raw) == 0 {
		return WorkoutListResponse{Items: []WorkoutResponse{}, HasMore: false}, nil
	}

	sort.Slice(raw, func(i, j int) bool {
		return workouts.FeedNewer(raw[i].Workout.StartDate, raw[i].Workout.ID, raw[j].Workout.StartDate, raw[j].Workout.ID)
	})

	filtered := make([]federation.OutboxWorkout, 0, len(raw))
	for _, ow := range raw {
		if !workouts.AfterCursor(ow.Workout.StartDate, ow.Workout.ID, cursor) {
			continue
		}
		filtered = append(filtered, ow)
	}

	hasMore := len(filtered) > limit
	if hasMore {
		filtered = filtered[:limit]
	}

	authorName := ""
	remoteAvatarURL := ""
	if a.federationDelivery != nil {
		if remote, resolveErr := a.federationDelivery.ResolveRemote(parsed); resolveErr == nil {
			authorName = remote.Name
			remoteAvatarURL = remote.AvatarURL
		}
	}
	hasAvatar, avatarURL := a.remoteUserAvatarFields(
		viewerNickname,
		parsed.Handle,
		parsed.Nickname,
		authorName,
		remoteAvatarURL,
	)

	items := make([]WorkoutResponse, 0, len(filtered))
	for _, ow := range filtered {
		w := *ow.Workout
		if len(ow.Track) > 0 {
			if parsedTrack, err := tracks.Parse(ow.Track, w.Track); err == nil && parsedTrack.HasGPS() {
				w.HasMapPreview = true
			}
		}
		if len(ow.Media) > 0 {
			names := make([]string, 0, len(ow.Media))
			for _, m := range ow.Media {
				names = append(names, m.Filename)
			}
			w.MediaFiles = names
			w.HasMedia = true
		}
		feed := workouts.FeedWorkout{
			Workout: w,
			Owner:   parsed.Nickname,
			Author: workouts.FeedAuthor{
				Nickname:  parsed.Nickname,
				Name:      authorName,
				Handle:    parsed.Handle,
				IsLocal:   false,
				HasAvatar: hasAvatar,
				AvatarURL: avatarURL,
			},
		}
		resp := toFeedWorkoutResponse(&feed)
		resp.ObjectID = ow.ObjectID
		resp.CanLike = false
		resp.LikedByMe = false
		items = append(items, resp)
	}

	out := WorkoutListResponse{Items: items, HasMore: hasMore}
	if hasMore && len(filtered) > 0 {
		last := filtered[len(filtered)-1].Workout
		cur := workouts.CursorFromWorkout(*last)
		out.NextCursor = cur.Encode()
	}
	return out, nil
}

func (a *App) liveRemoteWorkout(objectID string) (*workouts.Workout, []byte, []workouts.MediaFileInput, error) {
	objectID = strings.TrimSpace(objectID)
	if objectID == "" {
		return nil, nil, nil, workouts.ErrWorkoutNotFound
	}
	workout, track, media, _, err := federation.FetchWorkoutObject(a.federationHTTPClient(), a.Blobs, objectID)
	if err != nil || workout == nil {
		return nil, nil, nil, workouts.ErrWorkoutNotFound
	}
	return workout, track, media, nil
}

func (a *App) liveMapPreview(trackName string, trackData []byte) ([]byte, error) {
	if len(trackData) == 0 {
		return nil, workouts.ErrWorkoutNotFound
	}
	parsed, err := tracks.Parse(trackData, trackName)
	if err != nil || !parsed.HasGPS() {
		return nil, workouts.ErrWorkoutNotFound
	}
	preview, err := maprender.RenderPreview(parsed.Points)
	if err != nil || len(preview) == 0 {
		return nil, workouts.ErrWorkoutNotFound
	}
	return preview, nil
}

func (a *App) liveSpeedSamples(trackName string, trackData []byte) []workouts.SpeedSample {
	parsed, err := tracks.Parse(trackData, trackName)
	if err != nil {
		return nil
	}
	return workouts.BuildSpeedChartSamples(parsed)
}

func (a *App) liveHeartRateSamples(trackName string, trackData []byte) []workouts.HeartRateSample {
	parsed, err := tracks.Parse(trackData, trackName)
	if err != nil {
		return nil
	}
	return workouts.BuildHeartRateChartSamples(parsed)
}
