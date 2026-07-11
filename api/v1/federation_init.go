package v1

import (
	"sync"

	"github.com/solargate/grom/internal/config"
	"github.com/solargate/grom/internal/federation"
	"github.com/solargate/grom/internal/workouts"
)

var (
	federationOnce      sync.Once
	workoutInboxStore   *federation.WorkoutInboxStore
	followersStore      *federation.FollowersStore
	federationDelivery  *federation.Delivery
	federationInboxProc *federation.InboxProcessor
)

func initFederation() error {
	var initErr error
	federationOnce.Do(func() {
		if err := initSocialService(); err != nil {
			initErr = err
			return
		}
		workoutInboxStore = federation.NewWorkoutInboxStore(config.Cfg.Data.ResolvedDir)
		if err := initFollowersStore(); err != nil {
			initErr = err
			return
		}
		if config.Cfg.Federation.Enabled {
			var err error
			federationDelivery, err = federation.NewDelivery(userStore, socialService)
			if err != nil {
				initErr = err
				return
			}
			workoutInboxStore.SetHTTPClient(federationDelivery.Client())
			federationInboxProc = federation.NewInboxProcessor(userStore, socialService, federationDelivery, workoutInboxStore, followersStore)
		}
	})
	return initErr
}

type federatedFeedAdapter struct {
	store *federation.WorkoutInboxStore
}

func (a federatedFeedAdapter) ListFederated(viewerNickname string) ([]workouts.FeedWorkout, error) {
	return a.store.List(viewerNickname)
}

func newFeedService() *workouts.FeedService {
	feedSvc := workouts.NewFeedService(workoutStore, config.Cfg.Federation.Domain)
	if workoutInboxStore != nil {
		feedSvc.SetFederatedSource(federatedFeedAdapter{store: workoutInboxStore})
	}
	return feedSvc
}
