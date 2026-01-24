package blob

import (
	"context"
	"io"
)

type Uploader interface {
	Upload(ctx context.Context, data io.Reader, size int64, outputPath string, contentType string) error
}

type Checker interface {
	Exists(ctx context.Context, path string) (bool, error)
}

type StorageClient interface {
	Uploader
	Checker
}
