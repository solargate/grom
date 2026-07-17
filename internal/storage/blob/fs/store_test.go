package fs_test

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/solargate/grom/internal/storage/blob"
	blobfs "github.com/solargate/grom/internal/storage/blob/fs"
	"github.com/solargate/grom/internal/storage/keys"
)

func TestBlobStorePutGetDelete(t *testing.T) {
	dir := t.TempDir()
	store := blobfs.NewStore(dir)
	ctx := context.Background()

	key := keys.UserAvatar("athlete")
	payload := []byte("avatar-bytes")

	if _, err := blob.PutBytes(ctx, store, key, payload, blob.PutOptions{ContentType: "image/webp"}); err != nil {
		t.Fatal(err)
	}

	got, err := blob.ReadAll(ctx, store, key)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got, payload) {
		t.Fatalf("payload = %q, want %q", got, payload)
	}

	ok, err := store.Exists(ctx, key)
	if err != nil || !ok {
		t.Fatalf("Exists() = %v, %v", ok, err)
	}

	if err := store.Delete(ctx, key); err != nil {
		t.Fatal(err)
	}
	ok, err = store.Exists(ctx, key)
	if err != nil || ok {
		t.Fatalf("after delete Exists() = %v, %v", ok, err)
	}
}

func TestBlobStoreRejectsInvalidKey(t *testing.T) {
	store := blobfs.NewStore(t.TempDir())
	_, err := blob.PutBytes(context.Background(), store, "../escape", []byte("x"), blob.PutOptions{})
	if err == nil {
		t.Fatal("expected error for invalid key")
	}
}

func TestBlobStoreCreatesNestedPath(t *testing.T) {
	dir := t.TempDir()
	store := blobfs.NewStore(dir)
	key := keys.WorkoutTrack("athlete", "20260708T100000Z-1", "track.gpx")

	if _, err := blob.PutBytes(context.Background(), store, key, []byte("gpx"), blob.PutOptions{}); err != nil {
		t.Fatal(err)
	}

	path := filepath.Join(dir, filepath.FromSlash(key))
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("expected file at %q: %v", path, err)
	}
}
