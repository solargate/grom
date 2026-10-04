package migrate

import (
	"fmt"

	"github.com/solargate/grom/internal/notifications"
	"github.com/solargate/grom/internal/storage"
	storebbolt "github.com/solargate/grom/internal/storage/bbolt"
	"github.com/solargate/grom/internal/storage/file"
)

func copyPushSubscriptions(src, dst storage.Backend) (int, error) {
	subs, err := listPushSubscriptions(src)
	if err != nil {
		return 0, err
	}
	for _, sub := range subs {
		if err := importPushSubscription(dst, sub); err != nil {
			return 0, fmt.Errorf("import push subscription %s/%s: %w", sub.UserID, sub.InstallationID, err)
		}
	}
	return len(subs), nil
}

func countPushSubscriptions(backend storage.Backend) (int, error) {
	subs, err := listPushSubscriptions(backend)
	if err != nil {
		return 0, err
	}
	return len(subs), nil
}

func listPushSubscriptions(backend storage.Backend) ([]notifications.PushSubscription, error) {
	switch b := backend.(type) {
	case *file.Backend:
		return b.PushSubscriptions().(*file.PushSubscriptionStore).ListAll()
	case *storebbolt.Backend:
		return b.PushSubscriptions().(*storebbolt.PushSubscriptionStore).ListAll()
	default:
		return nil, fmt.Errorf("unsupported backend type %T", backend)
	}
}

func importPushSubscription(dst storage.Backend, sub notifications.PushSubscription) error {
	switch b := dst.(type) {
	case *file.Backend:
		return b.PushSubscriptions().(*file.PushSubscriptionStore).Import(sub)
	case *storebbolt.Backend:
		return b.PushSubscriptions().(*storebbolt.PushSubscriptionStore).PutExisting(sub)
	default:
		return fmt.Errorf("unsupported backend type %T", dst)
	}
}
