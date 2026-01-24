package blob

import (
	"context"
	"io"
)

type BlobUploader interface {
	Upload(ctx context.Context, data io.Reader, size int64, outputPath string, contentType string) error
}
