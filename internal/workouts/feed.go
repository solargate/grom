package workouts

import (
	"sort"
	"time"

	"github.com/solargate/grom/internal/avatars"
	"github.com/solargate/grom/internal/storage/blob"
)

type FeedAuthor struct {
	Nickname  string `json:"nickname"`
	Name      string `json:"name"`
	Handle    string `json:"handle"`
	IsLocal   bool   `json:"is_local"`
	HasAvatar bool   `json:"has_avatar"`
	AvatarURL string `json:"avatar_url,omitempty"`
}

type FeedWorkout struct {
	Workout
	Author FeedAuthor `json:"author"`
	Owner  string     `json:"owner"`
}

type FeedService struct {
	store     workoutLister
	blobs     blob.Store
	domain    string
	federated FederatedWorkoutSource
}

type workoutLister interface {
	List(nickname string) ([]Workout, error)
}

type FederatedWorkoutSource interface {
	ListFederated(viewerNickname string) ([]FeedWorkout, error)
}

func NewFeedService(store workoutLister, blobs blob.Store, domain string) *FeedService {
	if domain == "" {
		domain = "localhost"
	}
	return &FeedService{store: store, blobs: blobs, domain: domain}
}

func (f *FeedService) localHandle(nickname string) string {
	return nickname + "@" + f.domain
}

func (f *FeedService) SetFederatedSource(src FederatedWorkoutSource) {
	f.federated = src
}

func (f *FeedService) ListOwn(viewerNickname, viewerName string) ([]FeedWorkout, error) {
	own, err := f.store.List(viewerNickname)
	if err != nil {
		return nil, err
	}
	viewerHasAvatar, viewerAvatarURL := avatars.FieldsStore(f.blobs, viewerNickname)
	viewerAuthor := FeedAuthor{
		Nickname:  viewerNickname,
		Name:      viewerName,
		Handle:    f.localHandle(viewerNickname),
		IsLocal:   true,
		HasAvatar: viewerHasAvatar,
		AvatarURL: viewerAvatarURL,
	}
	result := make([]FeedWorkout, len(own))
	for i := range own {
		result[i] = FeedWorkout{
			Workout: own[i],
			Author:  viewerAuthor,
			Owner:   viewerNickname,
		}
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].StartDate.After(result[j].StartDate)
	})
	return result, nil
}

func (f *FeedService) ListFeed(viewerNickname, viewerName string, followedLocal []FeedAuthor) ([]FeedWorkout, error) {
	type tagged struct {
		workout FeedWorkout
	}

	items := make([]tagged, 0)

	own, err := f.store.List(viewerNickname)
	if err != nil {
		return nil, err
	}
	viewerHasAvatar, viewerAvatarURL := avatars.FieldsStore(f.blobs, viewerNickname)
	viewerAuthor := FeedAuthor{
		Nickname:  viewerNickname,
		Name:      viewerName,
		Handle:    f.localHandle(viewerNickname),
		IsLocal:   true,
		HasAvatar: viewerHasAvatar,
		AvatarURL: viewerAvatarURL,
	}
	for i := range own {
		items = append(items, tagged{
			workout: FeedWorkout{
				Workout: own[i],
				Author:  viewerAuthor,
				Owner:   viewerNickname,
			},
		})
	}

	for _, author := range followedLocal {
		workouts, err := f.store.List(author.Nickname)
		if err != nil {
			return nil, err
		}
		for i := range workouts {
			items = append(items, tagged{
				workout: FeedWorkout{
					Workout: workouts[i],
					Author:  author,
					Owner:   author.Nickname,
				},
			})
		}
	}

	if f.federated != nil {
		remote, err := f.federated.ListFederated(viewerNickname)
		if err != nil {
			return nil, err
		}
		for i := range remote {
			items = append(items, tagged{workout: remote[i]})
		}
	}

	sort.Slice(items, func(i, j int) bool {
		return items[i].workout.StartDate.After(items[j].workout.StartDate)
	})

	result := make([]FeedWorkout, len(items))
	for i := range items {
		result[i] = items[i].workout
	}
	return result, nil
}

func (f *FeedService) CanAccessWorkout(viewerNickname string, followedLocal []string, ownerNickname string) bool {
	if viewerNickname == ownerNickname {
		return true
	}
	for _, nick := range followedLocal {
		if nick == ownerNickname {
			return true
		}
	}
	return false
}

func FeedAuthorFromFollow(nickname, name, handle string, isLocal bool, hasAvatar bool, avatarURL string) FeedAuthor {
	return FeedAuthor{
		Nickname:  nickname,
		Name:      name,
		Handle:    handle,
		IsLocal:   isLocal,
		HasAvatar: hasAvatar,
		AvatarURL: avatarURL,
	}
}

// FeedWorkoutStartTime helper for tests.
func FeedWorkoutStartTime(w FeedWorkout) time.Time {
	return w.StartDate
}
