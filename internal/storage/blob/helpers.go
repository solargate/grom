package blob

import (
	"bytes"
	"context"
	"io"
)

// ReadAll reads the full blob payload into memory.
func ReadAll(ctx context.Context, store Store, key string) ([]byte, error) {
	rc, _, err := store.Get(ctx, key)
	if err != nil {
		return nil, err
	}
	defer rc.Close()
	return io.ReadAll(rc)
}

// PutBytes stores a byte slice as a blob.
func PutBytes(ctx context.Context, store Store, key string, data []byte, opts PutOptions) (*Ref, error) {
	return store.Put(ctx, key, bytes.NewReader(data), opts)
}
