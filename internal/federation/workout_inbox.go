package federation

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/solargate/grom/internal/avatars"
	"github.com/solargate/grom/internal/data"
	"github.com/solargate/grom/internal/maprender"
	"github.com/solargate/grom/internal/storage/blob"
	"github.com/solargate/grom/internal/storage/keys"
	"github.com/solargate/grom/internal/tracks"
	"github.com/solargate/grom/internal/workouts"
	"gopkg.in/yaml.v3"
)

type WorkoutInboxStore struct {
	dataDir         string
	blobs           blob.Store
	speedCharts     workouts.SpeedChartStore
	heartRateCharts workouts.HeartRateChartStore
	client          *http.Client
}

func NewWorkoutInboxStore(dataDir string, blobs blob.Store, speedCharts workouts.SpeedChartStore, heartRateCharts workouts.HeartRateChartStore) *WorkoutInboxStore {
	return &WorkoutInboxStore{
		dataDir:         dataDir,
		blobs:           blobs,
		speedCharts:     speedCharts,
		heartRateCharts: heartRateCharts,
	}
}

func (s *WorkoutInboxStore) SetHTTPClient(client *http.Client) {
	s.client = client
}

func (s *WorkoutInboxStore) ownerDir(viewerNickname, handle string) string {
	return filepath.Join(s.inboxDir(viewerNickname), OwnerKeyFromHandle(handle))
}

func (s *WorkoutInboxStore) EnsureAuthor(viewerNickname, handle, nickname, name, remoteAvatarURL string, refresh bool) error {
	dir := s.ownerDir(viewerNickname, handle)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return err
	}
	ownerKey := OwnerKeyFromHandle(handle)
	return mergeAuthorMetaWithRefresh(dir, handle, nickname, authorActor(name, remoteAvatarURL), s.client, refresh, s.blobs, viewerNickname, ownerKey)
}

func (s *WorkoutInboxStore) AuthorAvatarFields(viewerNickname, handle string) (bool, string) {
	ownerKey := OwnerKeyFromHandle(handle)
	ownerDir := s.ownerDir(viewerNickname, handle)
	if _, err := os.Stat(ownerDir); err != nil {
		return false, ""
	}
	meta, err := readAuthorMeta(ownerDir)
	if err != nil {
		return false, ""
	}
	return authorAvatarFields(s.blobs, viewerNickname, ownerKey, handle, meta)
}

func (s *WorkoutInboxStore) inboxDir(viewerNickname string) string {
	return filepath.Join(data.UserDir(s.dataDir, viewerNickname), "federation", "inbox", "workouts")
}

func (s *WorkoutInboxStore) findOwnerDir(viewerNickname, ownerNickname string) (string, string, error) {
	root := s.inboxDir(viewerNickname)
	entries, err := os.ReadDir(root)
	if err != nil {
		if os.IsNotExist(err) {
			return "", "", workouts.ErrWorkoutNotFound
		}
		return "", "", err
	}
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		if strings.EqualFold(ownerNicknameFromDir(entry.Name()), ownerNickname) {
			return filepath.Join(root, entry.Name()), entry.Name(), nil
		}
	}
	return "", "", workouts.ErrWorkoutNotFound
}

func ownerNicknameFromDir(dirName string) string {
	handle := strings.ReplaceAll(dirName, "_", "@")
	if idx := strings.Index(handle, "@"); idx > 0 {
		return handle[:idx]
	}
	return handle
}

func ownerHandleFromDir(dirName string) string {
	return strings.ReplaceAll(dirName, "_", "@")
}

func (s *WorkoutInboxStore) Save(viewerNickname, ownerHandle string, workout *workouts.Workout, trackData []byte, mediaFiles []workouts.MediaFileInput, actor map[string]any) error {
	dir := s.ownerDir(viewerNickname, ownerHandle)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return err
	}

	ownerKey := OwnerKeyFromHandle(ownerHandle)
	nickname := ownerNicknameFromDir(ownerKey)
	if err := mergeAuthorMeta(dir, ownerHandle, nickname, actor, s.client, s.blobs, viewerNickname, ownerKey); err != nil {
		return err
	}

	ctx := context.Background()

	if workout.Track != "" && len(trackData) > 0 {
		if err := s.writeFederatedTrack(ctx, viewerNickname, ownerKey, workout, trackData); err != nil {
			return err
		}
	}

	if len(mediaFiles) > 0 {
		savedNames, err := workouts.SaveFederatedMedia(s.blobs, viewerNickname, ownerKey, workout.ID, mediaFiles)
		if err != nil {
			return err
		}
		workout.MediaFiles = savedNames
		workout.HasMedia = len(savedNames) > 0
	}

	return s.writeFederatedWorkoutYAML(dir, workout)
}

// Replace stores a full federated workout snapshot. Track and media are replaced from the
// provided payload: empty track clears track artifacts; mediaFiles is the complete new set.
func (s *WorkoutInboxStore) Replace(viewerNickname, ownerHandle string, workout *workouts.Workout, trackData []byte, mediaFiles []workouts.MediaFileInput, actor map[string]any) error {
	dir := s.ownerDir(viewerNickname, ownerHandle)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return err
	}

	ownerKey := OwnerKeyFromHandle(ownerHandle)
	nickname := ownerNicknameFromDir(ownerKey)
	if err := mergeAuthorMeta(dir, ownerHandle, nickname, actor, s.client, s.blobs, viewerNickname, ownerKey); err != nil {
		return err
	}

	ctx := context.Background()
	previous, _ := s.readWorkout(viewerNickname, dir, ownerKey, workout.ID)

	if workout.Track != "" && len(trackData) > 0 {
		if previous != nil && previous.Track != "" && previous.Track != workout.Track {
			_ = s.blobs.Delete(ctx, keys.FederatedInboxTrack(viewerNickname, ownerKey, workout.ID, previous.Track))
		}
		if err := s.writeFederatedTrack(ctx, viewerNickname, ownerKey, workout, trackData); err != nil {
			return err
		}
	} else if workout.Track == "" {
		if previous != nil && previous.Track != "" {
			_ = s.blobs.Delete(ctx, keys.FederatedInboxTrack(viewerNickname, ownerKey, workout.ID, previous.Track))
		}
		_ = s.blobs.Delete(ctx, keys.FederatedInboxMapPreview(viewerNickname, ownerKey, workout.ID))
		if s.speedCharts != nil {
			_ = s.speedCharts.DeleteFederated(ctx, viewerNickname, ownerKey, workout.ID)
		}
		if s.heartRateCharts != nil {
			_ = s.heartRateCharts.DeleteFederated(ctx, viewerNickname, ownerKey, workout.ID)
		}
		workout.HasMapPreview = false
	} else if previous != nil {
		// Keep existing track blobs; preserve preview flag from stored artifacts.
		if exists, err := s.blobs.Exists(ctx, keys.FederatedInboxMapPreview(viewerNickname, ownerKey, workout.ID)); err == nil && exists {
			workout.HasMapPreview = true
		}
	}

	var previousMedia []string
	if previous != nil {
		previousMedia = previous.MediaFiles
	}
	savedNames, err := workouts.ReplaceFederatedMedia(s.blobs, viewerNickname, ownerKey, workout.ID, previousMedia, mediaFiles)
	if err != nil {
		return err
	}
	workout.MediaFiles = savedNames
	workout.HasMedia = len(savedNames) > 0

	return s.writeFederatedWorkoutYAML(dir, workout)
}

func (s *WorkoutInboxStore) writeFederatedTrack(ctx context.Context, viewerNickname, ownerKey string, workout *workouts.Workout, trackData []byte) error {
	trackKey := keys.FederatedInboxTrack(viewerNickname, ownerKey, workout.ID, workout.Track)
	if _, err := blob.PutBytes(ctx, s.blobs, trackKey, trackData, blob.PutOptions{}); err != nil {
		return err
	}

	parsed, err := tracks.Parse(trackData, workout.Track)
	if err != nil {
		slog.Warn("federated track parse failed", "workout_id", workout.ID, "err", err)
		return nil
	}

	if s.speedCharts != nil {
		if err := s.speedCharts.WriteFederated(ctx, viewerNickname, ownerKey, workout.ID, workouts.BuildSpeedChartSamples(parsed)); err != nil {
			slog.Error("federated speed chart write failed", "workout_id", workout.ID, "err", err)
		}
	}

	if s.heartRateCharts != nil {
		if err := s.heartRateCharts.WriteFederated(ctx, viewerNickname, ownerKey, workout.ID, workouts.BuildHeartRateChartSamples(parsed)); err != nil {
			slog.Error("federated heart rate chart write failed", "workout_id", workout.ID, "err", err)
		}
	}

	if parsed.HasGPS() {
		if preview, err := maprender.RenderPreview(parsed.Points); err != nil {
			slog.Error("federated map preview render failed", "workout_id", workout.ID, "err", err)
		} else if len(preview) > 0 {
			previewKey := keys.FederatedInboxMapPreview(viewerNickname, ownerKey, workout.ID)
			if _, err := blob.PutBytes(ctx, s.blobs, previewKey, preview, blob.PutOptions{ContentType: "image/webp"}); err != nil {
				slog.Error("federated map preview write failed", "workout_id", workout.ID, "err", err)
			} else {
				workout.HasMapPreview = true
			}
		}
	}
	return nil
}

func (s *WorkoutInboxStore) writeFederatedWorkoutYAML(dir string, workout *workouts.Workout) error {
	path := filepath.Join(dir, workout.ID+".yaml")
	data, err := yaml.Marshal(workout)
	if err != nil {
		return err
	}
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, data, 0600); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}

func (s *WorkoutInboxStore) Delete(viewerNickname, ownerHandle, workoutID string) error {
	dir := s.ownerDir(viewerNickname, ownerHandle)
	ownerKey := OwnerKeyFromHandle(ownerHandle)
	workout, err := s.readWorkout(viewerNickname, dir, ownerKey, workoutID)
	if err != nil {
		if errors.Is(err, workouts.ErrWorkoutNotFound) {
			return nil
		}
		return err
	}

	if err := os.Remove(filepath.Join(dir, workoutID+".yaml")); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("delete federated workout yaml: %w", err)
	}

	ctx := context.Background()
	if workout.Track != "" {
		_ = s.blobs.Delete(ctx, keys.FederatedInboxTrack(viewerNickname, ownerKey, workoutID, workout.Track))
	}
	_ = s.blobs.Delete(ctx, keys.FederatedInboxMapPreview(viewerNickname, ownerKey, workoutID))
	if s.speedCharts != nil {
		_ = s.speedCharts.DeleteFederated(ctx, viewerNickname, ownerKey, workoutID)
	}
	if s.heartRateCharts != nil {
		_ = s.heartRateCharts.DeleteFederated(ctx, viewerNickname, ownerKey, workoutID)
	}
	for _, name := range workout.MediaFiles {
		_ = s.blobs.Delete(ctx, keys.FederatedInboxMediaOriginal(viewerNickname, ownerKey, workoutID, name))
		_ = s.blobs.Delete(ctx, keys.FederatedInboxMediaPreview(viewerNickname, ownerKey, workoutID, name))
	}
	return nil
}

func (s *WorkoutInboxStore) DeleteAllForOwner(viewerNickname, ownerHandle string) error {
	if viewerNickname == "" || ownerHandle == "" {
		return nil
	}
	dir := s.ownerDir(viewerNickname, ownerHandle)
	ownerKey := OwnerKeyFromHandle(ownerHandle)

	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	for _, entry := range entries {
		name := entry.Name()
		if !strings.HasSuffix(name, ".yaml") || name == "author.yaml" {
			continue
		}
		workoutID := strings.TrimSuffix(name, ".yaml")
		if err := s.Delete(viewerNickname, ownerHandle, workoutID); err != nil {
			return err
		}
	}
	ctx := context.Background()
	_ = s.blobs.Delete(ctx, keys.FederatedInboxAvatar(viewerNickname, ownerKey))
	if err := os.RemoveAll(dir); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("remove federated owner dir: %w", err)
	}
	return nil
}

func (s *WorkoutInboxStore) readWorkout(viewerNickname, ownerDir, ownerKey, workoutID string) (*workouts.Workout, error) {
	path := filepath.Join(ownerDir, workoutID+".yaml")
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, workouts.ErrWorkoutNotFound
		}
		return nil, err
	}
	var workout workouts.Workout
	if err := yaml.Unmarshal(data, &workout); err != nil {
		return nil, fmt.Errorf("parse federated workout: %w", err)
	}
	ctx := context.Background()
	if exists, err := s.blobs.Exists(ctx, keys.FederatedInboxMapPreview(viewerNickname, ownerKey, workoutID)); err == nil && exists {
		workout.HasMapPreview = true
	}
	workout.HasMedia = len(workout.MediaFiles) > 0
	return &workout, nil
}

func (s *WorkoutInboxStore) TrackFile(viewerNickname, ownerNickname, workoutID string) ([]byte, string, string, error) {
	ownerDir, ownerKey, err := s.findOwnerDir(viewerNickname, ownerNickname)
	if err != nil {
		return nil, "", "", err
	}
	workout, err := s.readWorkout(viewerNickname, ownerDir, ownerKey, workoutID)
	if err != nil {
		return nil, "", "", err
	}
	if workout.Track == "" {
		return nil, "", "", workouts.ErrWorkoutNotFound
	}
	data, err := blob.ReadAll(context.Background(), s.blobs, keys.FederatedInboxTrack(viewerNickname, ownerKey, workoutID, workout.Track))
	if err != nil {
		if os.IsNotExist(err) {
			return nil, "", "", workouts.ErrWorkoutNotFound
		}
		return nil, "", "", fmt.Errorf("read federated track: %w", err)
	}
	return data, workout.Track, workout.Name, nil
}

func (s *WorkoutInboxStore) MapPreview(viewerNickname, ownerNickname, workoutID string) ([]byte, error) {
	ownerDir, ownerKey, err := s.findOwnerDir(viewerNickname, ownerNickname)
	if err != nil {
		return nil, err
	}
	if _, err := s.readWorkout(viewerNickname, ownerDir, ownerKey, workoutID); err != nil {
		return nil, err
	}
	data, err := blob.ReadAll(context.Background(), s.blobs, keys.FederatedInboxMapPreview(viewerNickname, ownerKey, workoutID))
	if err != nil {
		if os.IsNotExist(err) {
			return nil, workouts.ErrWorkoutNotFound
		}
		return nil, err
	}
	return data, nil
}

func (s *WorkoutInboxStore) MediaOriginal(viewerNickname, ownerNickname, workoutID, filename string) ([]byte, string, error) {
	ownerDir, ownerKey, err := s.findOwnerDir(viewerNickname, ownerNickname)
	if err != nil {
		return nil, "", err
	}
	if _, err := s.readWorkout(viewerNickname, ownerDir, ownerKey, workoutID); err != nil {
		return nil, "", err
	}
	safeName, err := workouts.SanitizeMediaFilename(filename)
	if err != nil {
		return nil, "", err
	}
	data, err := blob.ReadAll(context.Background(), s.blobs, keys.FederatedInboxMediaOriginal(viewerNickname, ownerKey, workoutID, safeName))
	if err != nil {
		if os.IsNotExist(err) {
			return nil, "", workouts.ErrPhotoNotFound
		}
		return nil, "", err
	}
	return data, workouts.MediaContentType(safeName), nil
}

func (s *WorkoutInboxStore) MediaPreview(viewerNickname, ownerNickname, workoutID, filename string) ([]byte, error) {
	ownerDir, ownerKey, err := s.findOwnerDir(viewerNickname, ownerNickname)
	if err != nil {
		return nil, err
	}
	if _, err := s.readWorkout(viewerNickname, ownerDir, ownerKey, workoutID); err != nil {
		return nil, err
	}
	safeName, err := workouts.SanitizeMediaFilename(filename)
	if err != nil {
		return nil, err
	}
	data, err := blob.ReadAll(context.Background(), s.blobs, keys.FederatedInboxMediaPreview(viewerNickname, ownerKey, workoutID, safeName))
	if err != nil {
		if os.IsNotExist(err) {
			return nil, workouts.ErrPhotoNotFound
		}
		return nil, err
	}
	return data, nil
}

func (s *WorkoutInboxStore) Avatar(viewerNickname, ownerKey string) ([]byte, error) {
	ctx := context.Background()
	avatarKey := keys.FederatedInboxAvatar(viewerNickname, ownerKey)
	if exists, err := s.blobs.Exists(ctx, avatarKey); err == nil && exists {
		return blob.ReadAll(ctx, s.blobs, avatarKey)
	}

	ownerDir, err := s.ownerDirForKey(viewerNickname, ownerKey)
	if err != nil {
		return nil, err
	}

	meta, err := readAuthorMeta(ownerDir)
	if err != nil {
		return nil, err
	}
	remoteURL := effectiveRemoteAvatarURL(meta)
	if remoteURL == "" {
		return nil, avatars.ErrAvatarNotFound
	}

	if s.client == nil {
		return nil, avatars.ErrAvatarNotFound
	}
	if err := cacheRemoteAvatar(s.client, s.blobs, viewerNickname, ownerKey, remoteURL); err != nil {
		return nil, err
	}

	meta.AvatarVersion++
	meta.AvatarURL = FederatedAvatarAPIPath(meta.Handle, meta.AvatarVersion)
	if meta.RemoteAvatarURL == "" {
		meta.RemoteAvatarURL = remoteURL
	}
	_ = writeAuthorMeta(ownerDir, meta)

	return blob.ReadAll(ctx, s.blobs, avatarKey)
}

func (s *WorkoutInboxStore) ownerDirForKey(viewerNickname, ownerKey string) (string, error) {
	ownerKey = strings.TrimSpace(ownerKey)
	if ownerKey == "" {
		return "", avatars.ErrAvatarNotFound
	}
	ownerDir := filepath.Join(s.inboxDir(viewerNickname), ownerKey)
	info, err := os.Stat(ownerDir)
	if err != nil {
		if os.IsNotExist(err) {
			return "", avatars.ErrAvatarNotFound
		}
		return "", err
	}
	if !info.IsDir() {
		return "", avatars.ErrAvatarNotFound
	}
	return ownerDir, nil
}

func (s *WorkoutInboxStore) Get(viewerNickname, ownerNickname, workoutID string) (*workouts.FeedWorkout, error) {
	ownerDir, ownerKey, err := s.findOwnerDir(viewerNickname, ownerNickname)
	if err != nil {
		return nil, err
	}
	workout, err := s.readWorkout(viewerNickname, ownerDir, ownerKey, workoutID)
	if err != nil {
		return nil, err
	}
	handle := ownerHandleFromDir(ownerKey)
	nickname := ownerNicknameFromDir(ownerKey)
	meta, err := readAuthorMeta(ownerDir)
	if err != nil {
		return nil, err
	}
	if meta.Handle != "" {
		handle = meta.Handle
	}
	if meta.Nickname != "" {
		nickname = meta.Nickname
	}
	authorHasAvatar, authorAvatarURL := authorAvatarFields(s.blobs, viewerNickname, ownerKey, handle, meta)
	return &workouts.FeedWorkout{
		Workout: *workout,
		Owner:   nickname,
		Author: workouts.FeedAuthor{
			Nickname:  nickname,
			Name:      meta.Name,
			Handle:    handle,
			IsLocal:   false,
			HasAvatar: authorHasAvatar,
			AvatarURL: authorAvatarURL,
		},
	}, nil
}

func (s *WorkoutInboxStore) GetSpeedChart(viewerNickname, ownerNickname, workoutID string) (*workouts.Workout, []workouts.SpeedSample, error) {
	ownerDir, ownerKey, err := s.findOwnerDir(viewerNickname, ownerNickname)
	if err != nil {
		return nil, nil, err
	}
	workout, err := s.readWorkout(viewerNickname, ownerDir, ownerKey, workoutID)
	if err != nil {
		return nil, nil, err
	}
	if workout.Track == "" || s.speedCharts == nil {
		return workout, nil, nil
	}
	samples, err := s.speedCharts.ReadFederated(context.Background(), viewerNickname, ownerKey, workoutID)
	if err != nil {
		return nil, nil, fmt.Errorf("read federated speed chart: %w", err)
	}
	return workout, samples, nil
}

func (s *WorkoutInboxStore) GetHeartRateChart(viewerNickname, ownerNickname, workoutID string) (*workouts.Workout, []workouts.HeartRateSample, error) {
	ownerDir, ownerKey, err := s.findOwnerDir(viewerNickname, ownerNickname)
	if err != nil {
		return nil, nil, err
	}
	workout, err := s.readWorkout(viewerNickname, ownerDir, ownerKey, workoutID)
	if err != nil {
		return nil, nil, err
	}
	if workout.Track == "" || s.heartRateCharts == nil {
		return workout, nil, nil
	}
	samples, err := s.heartRateCharts.ReadFederated(context.Background(), viewerNickname, ownerKey, workoutID)
	if err != nil {
		return nil, nil, fmt.Errorf("read federated heart rate chart: %w", err)
	}
	return workout, samples, nil
}

func (s *WorkoutInboxStore) List(viewerNickname string) ([]workouts.FeedWorkout, error) {
	root := s.inboxDir(viewerNickname)
	entries, err := os.ReadDir(root)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	items := make([]workouts.FeedWorkout, 0)
	for _, ownerEntry := range entries {
		if !ownerEntry.IsDir() {
			continue
		}
		ownerDir := filepath.Join(root, ownerEntry.Name())
		ownerKey := ownerEntry.Name()
		handle := ownerHandleFromDir(ownerKey)
		nickname := ownerNicknameFromDir(ownerKey)
		meta, err := readAuthorMeta(ownerDir)
		if err != nil {
			return nil, err
		}
		if meta.Handle != "" {
			handle = meta.Handle
		}
		if meta.Nickname != "" {
			nickname = meta.Nickname
		}
		authorName := meta.Name
		authorHasAvatar, authorAvatarURL := authorAvatarFields(s.blobs, viewerNickname, ownerKey, handle, meta)
		files, err := os.ReadDir(ownerDir)
		if err != nil {
			return nil, err
		}
		for _, fileEntry := range files {
			name := fileEntry.Name()
			if fileEntry.IsDir() || !strings.HasSuffix(name, ".yaml") {
				continue
			}
			if name == authorMetaFileName {
				continue
			}
			workoutID := strings.TrimSuffix(name, ".yaml")
			workout, err := s.readWorkout(viewerNickname, ownerDir, ownerKey, workoutID)
			if err != nil {
				return nil, err
			}
			items = append(items, workouts.FeedWorkout{
				Workout: *workout,
				Owner:   nickname,
				Author: workouts.FeedAuthor{
					Nickname:  nickname,
					Name:      authorName,
					Handle:    handle,
					IsLocal:   false,
					HasAvatar: authorHasAvatar,
					AvatarURL: authorAvatarURL,
				},
			})
		}
	}

	sort.Slice(items, func(i, j int) bool {
		return workouts.FeedNewer(items[i].StartDate, items[i].ID, items[j].StartDate, items[j].ID)
	})
	return items, nil
}

func (s *WorkoutInboxStore) ListPage(viewerNickname string, cursor *workouts.Cursor, limit int) ([]workouts.FeedWorkout, bool, error) {
	if limit <= 0 {
		limit = workouts.DefaultPageLimit
	}
	all, err := s.List(viewerNickname)
	if err != nil {
		return nil, false, err
	}
	filtered := make([]workouts.FeedWorkout, 0, limit)
	for i := range all {
		if !workouts.AfterCursor(all[i].StartDate, all[i].ID, cursor) {
			continue
		}
		filtered = append(filtered, all[i])
		if len(filtered) > limit {
			break
		}
	}
	hasMore := len(filtered) > limit
	if hasMore {
		filtered = filtered[:limit]
	}
	return filtered, hasMore, nil
}
