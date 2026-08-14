package migrate

import (
	"fmt"

	"github.com/solargate/grom/internal/auth/pat"
	"github.com/solargate/grom/internal/storage"
	storebbolt "github.com/solargate/grom/internal/storage/bbolt"
	"github.com/solargate/grom/internal/storage/file"
)

func copyPersonalAccessTokens(src, dst storage.Backend) (int, error) {
	tokens, err := listPersonalAccessTokens(src)
	if err != nil {
		return 0, err
	}
	for _, rec := range tokens {
		if err := importPersonalAccessToken(dst, rec); err != nil {
			return 0, fmt.Errorf("import pat %s: %w", rec.ID, err)
		}
	}
	return len(tokens), nil
}

func countPersonalAccessTokens(backend storage.Backend) (int, error) {
	tokens, err := listPersonalAccessTokens(backend)
	if err != nil {
		return 0, err
	}
	return len(tokens), nil
}

func listPersonalAccessTokens(backend storage.Backend) ([]pat.TokenRecord, error) {
	switch b := backend.(type) {
	case *file.Backend:
		return b.PAT().(*file.PATStore).ListAll()
	case *storebbolt.Backend:
		return b.PAT().(*storebbolt.PATStore).ListAll()
	default:
		return nil, fmt.Errorf("unsupported backend type %T", backend)
	}
}

func importPersonalAccessToken(dst storage.Backend, record pat.TokenRecord) error {
	switch b := dst.(type) {
	case *file.Backend:
		return b.PAT().(*file.PATStore).Import(record)
	case *storebbolt.Backend:
		return b.PAT().(*storebbolt.PATStore).PutExisting(record)
	default:
		return fmt.Errorf("unsupported backend type %T", dst)
	}
}
