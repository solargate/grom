package file

import (
	"context"
	"fmt"
	"os"

	"github.com/solargate/grom/internal/equipment"
	"github.com/solargate/grom/internal/federation"
	"github.com/solargate/grom/internal/social"
	"github.com/solargate/grom/internal/storage/blob"
	"github.com/solargate/grom/internal/users"
	"github.com/solargate/grom/internal/workouts"
)

type Backend struct {
	location    string
	users       users.Repository
	workoutRepo *WorkoutsStore
	workouts    *workouts.Service
	equipment   equipment.Repository
	social      social.Repository
	fed         federation.Storage
	blobs       blob.Store
}

func Open(location string) (*Backend, error) {
	if location == "" {
		return nil, fmt.Errorf("storage location is required")
	}

	blobStore := NewBlobStore(location)

	userStore, err := NewUsersStore(location)
	if err != nil {
		return nil, fmt.Errorf("open users store: %w", err)
	}

	socialStore, err := NewSocialStore(location)
	if err != nil {
		return nil, fmt.Errorf("open social store: %w", err)
	}

	followersStore := NewFederationFollowersStore(location)
	inboxStore := federation.NewWorkoutInboxStore(location, blobStore)

	workoutRepo := NewWorkoutsStore(location)
	equipmentStore := NewEquipmentStore(location)
	workoutSvc := workouts.NewService(workoutRepo, blobStore)
	workoutSvc.SetEquipmentCatalog(equipmentStore)

	return &Backend{
		location:    location,
		users:       userStore,
		workoutRepo: workoutRepo,
		workouts:    workoutSvc,
		equipment:   equipmentStore,
		social:      socialStore,
		fed:         federation.NewStorage(followersStore, inboxStore),
		blobs:       blobStore,
	}, nil
}

func (b *Backend) Users() users.Repository         { return b.users }
func (b *Backend) Workouts() *workouts.Service     { return b.workouts }
func (b *Backend) WorkoutsRepo() *WorkoutsStore    { return b.workoutRepo }
func (b *Backend) Equipment() equipment.Repository { return b.equipment }
func (b *Backend) Social() social.Repository       { return b.social }
func (b *Backend) Federation() federation.Storage  { return b.fed }
func (b *Backend) Blobs() blob.Store               { return b.blobs }

func (b *Backend) Close() error { return nil }

func (b *Backend) Ping(ctx context.Context) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}
	info, err := os.Stat(b.location)
	if err != nil {
		return fmt.Errorf("stat storage location: %w", err)
	}
	if !info.IsDir() {
		return fmt.Errorf("storage location %q is not a directory", b.location)
	}
	return nil
}

func (b *Backend) Location() string { return b.location }
