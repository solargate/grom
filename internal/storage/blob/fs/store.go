package fs

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/solargate/grom/internal/storage/blob"
)

type Store struct {
	root string
}

func NewStore(root string) *Store {
	return &Store{root: root}
}

func (s *Store) resolve(key string) (string, error) {
	if key == "" || strings.Contains(key, "..") {
		return "", fmt.Errorf("invalid blob key %q", key)
	}
	path := filepath.Join(s.root, filepath.FromSlash(key))
	absRoot, err := filepath.Abs(s.root)
	if err != nil {
		return "", err
	}
	absPath, err := filepath.Abs(path)
	if err != nil {
		return "", err
	}
	rel, err := filepath.Rel(absRoot, absPath)
	if err != nil || strings.HasPrefix(rel, "..") {
		return "", fmt.Errorf("invalid blob key %q", key)
	}
	return absPath, nil
}

func (s *Store) Put(ctx context.Context, key string, r io.Reader, opts blob.PutOptions) (*blob.Ref, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	path, err := s.resolve(key)
	if err != nil {
		return nil, err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return nil, err
	}

	tmp := path + ".tmp"
	f, err := os.OpenFile(tmp, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0600)
	if err != nil {
		return nil, err
	}
	n, err := io.Copy(f, r)
	closeErr := f.Close()
	if err != nil {
		_ = os.Remove(tmp)
		return nil, err
	}
	if closeErr != nil {
		_ = os.Remove(tmp)
		return nil, closeErr
	}
	if err := os.Rename(tmp, path); err != nil {
		_ = os.Remove(tmp)
		return nil, err
	}

	return &blob.Ref{
		Key:         key,
		ContentType: opts.ContentType,
		Size:        n,
	}, nil
}

func (s *Store) Get(ctx context.Context, key string) (io.ReadCloser, *blob.Ref, error) {
	if err := ctx.Err(); err != nil {
		return nil, nil, err
	}

	path, err := s.resolve(key)
	if err != nil {
		return nil, nil, err
	}
	info, err := os.Stat(path)
	if err != nil {
		return nil, nil, err
	}
	f, err := os.Open(path)
	if err != nil {
		return nil, nil, err
	}
	return f, &blob.Ref{Key: key, Size: info.Size()}, nil
}

func (s *Store) Delete(ctx context.Context, key string) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	path, err := s.resolve(key)
	if err != nil {
		return err
	}
	err = os.Remove(path)
	if os.IsNotExist(err) {
		return nil
	}
	return err
}

func (s *Store) Exists(ctx context.Context, key string) (bool, error) {
	if err := ctx.Err(); err != nil {
		return false, err
	}
	path, err := s.resolve(key)
	if err != nil {
		return false, err
	}
	_, err = os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

var _ blob.Store = (*Store)(nil)
