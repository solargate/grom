package storage

import (
	"context"

	"github.com/solargate/grom/internal/equipment"
	"github.com/solargate/grom/internal/federation"
	"github.com/solargate/grom/internal/social"
	"github.com/solargate/grom/internal/storage/blob"
	"github.com/solargate/grom/internal/users"
	"github.com/solargate/grom/internal/workouts"
)

type Backend interface {
	Users() users.Repository
	Workouts() *workouts.Service
	Likes() workouts.LikesRepository
	Equipment() equipment.Repository
	Social() social.Repository
	Federation() federation.Storage
	Blobs() blob.Store

	Close() error
	Ping(ctx context.Context) error
}
