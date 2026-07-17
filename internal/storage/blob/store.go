package blob

import (
	"context"
	"io"
)

type Ref struct {
	Key         string
	ContentType string
	Size        int64
}

type PutOptions struct {
	ContentType string
}

type Store interface {
	Put(ctx context.Context, key string, r io.Reader, opts PutOptions) (*Ref, error)
	Get(ctx context.Context, key string) (io.ReadCloser, *Ref, error)
	Delete(ctx context.Context, key string) error
	Exists(ctx context.Context, key string) (bool, error)
}
