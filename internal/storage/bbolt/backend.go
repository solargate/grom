package bbolt

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	bolt "go.etcd.io/bbolt"

	"github.com/solargate/grom/internal/equipment"
	"github.com/solargate/grom/internal/federation"
	"github.com/solargate/grom/internal/social"
	"github.com/solargate/grom/internal/storage/blob"
	blobfs "github.com/solargate/grom/internal/storage/blob/fs"
	"github.com/solargate/grom/internal/users"
	"github.com/solargate/grom/internal/workouts"
)

type Backend struct {
	db          *bolt.DB
	location    string
	users       *UsersStore
	workoutRepo *WorkoutsStore
	workouts    *workouts.Service
	equipment   *EquipmentStore
	social      *SocialStore
	fed         federation.Storage
	blobs       blob.Store
}

// Open opens a bbolt metadata database and filesystem blob store under location.
func Open(dbPath, location string) (*Backend, error) {
	if location == "" {
		return nil, fmt.Errorf("storage location is required")
	}
	if dbPath == "" {
		return nil, fmt.Errorf("bbolt path is required")
	}
	if err := os.MkdirAll(filepath.Dir(dbPath), 0700); err != nil {
		return nil, fmt.Errorf("create bbolt parent dir: %w", err)
	}
	if err := os.MkdirAll(location, 0700); err != nil {
		return nil, fmt.Errorf("create storage location: %w", err)
	}

	db, err := bolt.Open(dbPath, 0600, &bolt.Options{Timeout: 0})
	if err != nil {
		return nil, fmt.Errorf("open bbolt: %w", err)
	}
	if err := ensureBuckets(db); err != nil {
		_ = db.Close()
		return nil, err
	}

	blobStore := blobfs.NewStore(location)
	userStore := NewUsersStore(db, location)
	socialStore := NewSocialStore(db)
	equipmentStore := NewEquipmentStore(db)
	workoutRepo := NewWorkoutsStore(db, location)
	workoutSvc := workouts.NewService(workoutRepo, blobStore)
	workoutSvc.SetSpeedSidecarFormat(workouts.SpeedSidecarJSON)
	workoutSvc.SetEquipmentCatalog(equipmentStore)

	followersStore := NewFederationFollowersStore(db)
	inboxStore := NewInboxStore(db, blobStore)

	return &Backend{
		db:          db,
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

func ensureBuckets(db *bolt.DB) error {
	return db.Update(func(tx *bolt.Tx) error {
		for _, name := range allBuckets {
			if _, err := tx.CreateBucketIfNotExists(name); err != nil {
				return fmt.Errorf("create bucket %q: %w", name, err)
			}
		}
		meta := tx.Bucket(bucketMeta)
		if meta.Get([]byte("schema_version")) == nil {
			if err := meta.Put([]byte("schema_version"), []byte(schemaVersion)); err != nil {
				return err
			}
		}
		return nil
	})
}

func (b *Backend) Users() users.Repository         { return b.users }
func (b *Backend) Workouts() *workouts.Service     { return b.workouts }
func (b *Backend) WorkoutsRepo() *WorkoutsStore    { return b.workoutRepo }
func (b *Backend) Equipment() equipment.Repository { return b.equipment }
func (b *Backend) Social() social.Repository       { return b.social }
func (b *Backend) Federation() federation.Storage  { return b.fed }
func (b *Backend) Blobs() blob.Store               { return b.blobs }
func (b *Backend) DB() *bolt.DB                    { return b.db }
func (b *Backend) Location() string                { return b.location }

func (b *Backend) Close() error {
	if b.db == nil {
		return nil
	}
	return b.db.Close()
}

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
	return b.db.View(func(tx *bolt.Tx) error {
		if tx.Bucket(bucketMeta) == nil {
			return fmt.Errorf("bbolt meta bucket missing")
		}
		return nil
	})
}
