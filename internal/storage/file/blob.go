package file

import (
	blobfs "github.com/solargate/grom/internal/storage/blob/fs"
)

type BlobStore = blobfs.Store

func NewBlobStore(root string) *BlobStore {
	return blobfs.NewStore(root)
}
