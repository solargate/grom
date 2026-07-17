package blob_test

import (
	"bytes"
	"context"
	"testing"

	"github.com/solargate/grom/internal/storage/blob"
	blobfs "github.com/solargate/grom/internal/storage/blob/fs"
)

func TestPutBytesAndReadAll(t *testing.T) {
	store := blobfs.NewStore(t.TempDir())
	ctx := context.Background()
	payload := []byte("hello")

	if _, err := blob.PutBytes(ctx, store, "test/key.bin", payload, blob.PutOptions{}); err != nil {
		t.Fatal(err)
	}

	got, err := blob.ReadAll(ctx, store, "test/key.bin")
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got, payload) {
		t.Fatalf("got %q", got)
	}
}
