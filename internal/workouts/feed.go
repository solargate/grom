package workouts

import (
	"container/heap"
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
	ListPage(nickname string, cursor *Cursor, limit int) ([]Workout, bool, error)
}

type FederatedWorkoutSource interface {
	ListFederated(viewerNickname string) ([]FeedWorkout, error)
	ListFederatedPage(viewerNickname string, cursor *Cursor, limit int) ([]FeedWorkout, bool, error)
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

func (f *FeedService) viewerAuthor(viewerNickname, viewerName string) FeedAuthor {
	viewerHasAvatar, viewerAvatarURL := avatars.FieldsStore(f.blobs, viewerNickname)
	return FeedAuthor{
		Nickname:  viewerNickname,
		Name:      viewerName,
		Handle:    f.localHandle(viewerNickname),
		IsLocal:   true,
		HasAvatar: viewerHasAvatar,
		AvatarURL: viewerAvatarURL,
	}
}

func (f *FeedService) ListOwn(viewerNickname, viewerName string) ([]FeedWorkout, error) {
	own, err := f.store.List(viewerNickname)
	if err != nil {
		return nil, err
	}
	viewerAuthor := f.viewerAuthor(viewerNickname, viewerName)
	result := make([]FeedWorkout, len(own))
	for i := range own {
		result[i] = FeedWorkout{
			Workout: own[i],
			Author:  viewerAuthor,
			Owner:   viewerNickname,
		}
	}
	sort.Slice(result, func(i, j int) bool {
		return FeedNewer(result[i].StartDate, result[i].ID, result[j].StartDate, result[j].ID)
	})
	return result, nil
}

// ListOwnPage returns a cursor page of the viewer's own workouts.
func (f *FeedService) ListOwnPage(viewerNickname, viewerName string, cursor *Cursor, limit int) (Page, error) {
	limit = ClampLimit(limit)
	own, hasMore, err := f.store.ListPage(viewerNickname, cursor, limit)
	if err != nil {
		return Page{}, err
	}
	viewerAuthor := f.viewerAuthor(viewerNickname, viewerName)
	items := make([]FeedWorkout, len(own))
	for i := range own {
		items[i] = FeedWorkout{
			Workout: own[i],
			Author:  viewerAuthor,
			Owner:   viewerNickname,
		}
	}
	return buildPage(items, hasMore), nil
}

func (f *FeedService) ListFeed(viewerNickname, viewerName string, followedLocal []FeedAuthor) ([]FeedWorkout, error) {
	var (
		items  []FeedWorkout
		cursor *Cursor
	)
	for {
		page, err := f.ListFeedPage(viewerNickname, viewerName, followedLocal, cursor, MaxPageLimit)
		if err != nil {
			return nil, err
		}
		items = append(items, page.Items...)
		if !page.HasMore {
			return items, nil
		}
		cursor, err = DecodeCursor(page.NextCursor)
		if err != nil {
			return nil, err
		}
	}
}

// ListFeedPage merges own, followed local, and federated sources with keyset pagination.
func (f *FeedService) ListFeedPage(viewerNickname, viewerName string, followedLocal []FeedAuthor, cursor *Cursor, limit int) (Page, error) {
	limit = ClampLimit(limit)
	viewerAuthor := f.viewerAuthor(viewerNickname, viewerName)

	batches := make([]feedBatch, 0, 2+len(followedLocal))

	own, ownMore, err := f.store.ListPage(viewerNickname, cursor, limit)
	if err != nil {
		return Page{}, err
	}
	ownFeed := make([]FeedWorkout, len(own))
	for i := range own {
		ownFeed[i] = FeedWorkout{
			Workout: own[i],
			Author:  viewerAuthor,
			Owner:   viewerNickname,
		}
	}
	batches = append(batches, feedBatch{items: ownFeed, hasMore: ownMore})

	for _, author := range followedLocal {
		list, more, err := f.store.ListPage(author.Nickname, cursor, limit)
		if err != nil {
			return Page{}, err
		}
		feedItems := make([]FeedWorkout, len(list))
		for i := range list {
			feedItems[i] = FeedWorkout{
				Workout: list[i],
				Author:  author,
				Owner:   author.Nickname,
			}
		}
		batches = append(batches, feedBatch{items: feedItems, hasMore: more})
	}

	if f.federated != nil {
		remote, more, err := f.federated.ListFederatedPage(viewerNickname, cursor, limit)
		if err != nil {
			return Page{}, err
		}
		batches = append(batches, feedBatch{items: remote, hasMore: more})
	}

	merged, hasMore := mergeFeedBatches(batches, limit)
	return buildPage(merged, hasMore), nil
}

func buildPage(items []FeedWorkout, hasMore bool) Page {
	page := Page{Items: items, HasMore: hasMore}
	if hasMore && len(items) > 0 {
		page.NextCursor = CursorFromWorkout(items[len(items)-1].Workout).Encode()
	}
	return page
}

type feedBatch struct {
	items   []FeedWorkout
	hasMore bool
}

type feedHeapItem struct {
	workout FeedWorkout
	batch   int
	index   int
}

type feedHeap []feedHeapItem

func (h feedHeap) Len() int { return len(h) }
func (h feedHeap) Less(i, j int) bool {
	return FeedNewer(h[i].workout.StartDate, h[i].workout.ID, h[j].workout.StartDate, h[j].workout.ID)
}
func (h feedHeap) Swap(i, j int) { h[i], h[j] = h[j], h[i] }
func (h *feedHeap) Push(x any)   { *h = append(*h, x.(feedHeapItem)) }
func (h *feedHeap) Pop() any {
	old := *h
	n := len(old)
	item := old[n-1]
	*h = old[:n-1]
	return item
}

func mergeFeedBatches(batches []feedBatch, limit int) ([]FeedWorkout, bool) {
	h := &feedHeap{}
	heap.Init(h)
	for bi, b := range batches {
		if len(b.items) == 0 {
			continue
		}
		heap.Push(h, feedHeapItem{workout: b.items[0], batch: bi, index: 0})
	}

	result := make([]FeedWorkout, 0, limit)
	for h.Len() > 0 && len(result) < limit {
		item := heap.Pop(h).(feedHeapItem)
		result = append(result, item.workout)
		next := item.index + 1
		if next < len(batches[item.batch].items) {
			heap.Push(h, feedHeapItem{
				workout: batches[item.batch].items[next],
				batch:   item.batch,
				index:   next,
			})
		}
	}

	if len(result) < limit {
		return result, false
	}
	if h.Len() > 0 {
		return result, true
	}
	for _, b := range batches {
		if b.hasMore {
			return result, true
		}
	}
	return result, false
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
