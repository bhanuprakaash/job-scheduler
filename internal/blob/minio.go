package blob

import (
	"context"
	"fmt"
	"io"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type MinioBlob struct {
	MinioClient *minio.Client
	Id          string
	Secret      string
	Endpoint    string
	Bucket      string
}

func NewMinioBlob(id, secret, endpoint, bucket string, useSsl bool) (*MinioBlob, error) {
	minioClient, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(id, secret, ""),
		Secure: useSsl,
	})

	if err != nil {
		return nil, fmt.Errorf("error initializing the minio client: %w", err)
	}

	exists, err := minioClient.BucketExists(context.Background(), bucket)
	if err != nil {
		return nil, err
	}

	if !exists {
		err = minioClient.MakeBucket(context.Background(), bucket, minio.MakeBucketOptions{})
		if err != nil {
			return nil, fmt.Errorf("failed to create a bucker: %w", err)
		}
	}

	return &MinioBlob{
		MinioClient: minioClient,
		Id:          id,
		Secret:      secret,
		Endpoint:    endpoint,
		Bucket:      bucket,
	}, nil

}

func (m *MinioBlob) Upload(ctx context.Context, data io.Reader, size int64, outputPath string, contentType string) error {
	_, err := m.MinioClient.PutObject(
		ctx,
		m.Bucket,
		outputPath,
		data,
		size,
		minio.PutObjectOptions{
			ContentType: contentType,
		},
	)

	if err != nil {
		return fmt.Errorf("failed to upload to minio: %w", err)
	}

	return nil
}
