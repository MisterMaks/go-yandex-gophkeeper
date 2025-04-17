package db

import (
	"context"
	"io"

	"github.com/minio/minio-go/v7"
)

const BucketName = "go-yandex-gophkeeper-bucket"

// MinioStorage is S3 storage for large data.
type MinioStorage struct {
	client *minio.Client
}

func NewMinioStorage(client *minio.Client) *MinioStorage {
	return &MinioStorage{client: client}
}

func (m *MinioStorage) GetObject(ctx context.Context, objectName string) (*minio.Object, error) {
	object, err := m.client.GetObject(ctx, BucketName, objectName, minio.GetObjectOptions{})
	return object, err
}

func (m *MinioStorage) PutObject(ctx context.Context, objectName string, reader io.Reader, objectSize int64) (minio.UploadInfo, error) {
	info, err := m.client.PutObject(ctx, BucketName, objectName, reader, objectSize, minio.PutObjectOptions{})
	return info, err
}

func (m *MinioStorage) RemoveObject(ctx context.Context, objectName string) error {
	err := m.client.RemoveObject(ctx, BucketName, objectName, minio.RemoveObjectOptions{})
	return err
}
