package filemanager

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"gpu-platform/internal/config"
)

type FileManager struct {
	client *minio.Client
	bucket string
}

func NewFileManager(cfg *config.StorageConfig) (*FileManager, error) {
	client, err := minio.New(cfg.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.AccessKey, cfg.SecretKey, ""),
		Secure: cfg.UseSSL,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create client: %w", err)
	}

	fm := &FileManager{
		client: client,
		bucket: fmt.Sprintf("%s-files", cfg.BucketPrefix),
	}

	ctx := context.Background()
	exists, _ := client.BucketExists(ctx, fm.bucket)
	if !exists {
		client.MakeBucket(ctx, fm.bucket, minio.MakeBucketOptions{})
	}

	return fm, nil
}

func (fm *FileManager) UploadFile(ctx context.Context, userID, objectName string, data []byte, contentType string) error {
	reader := io.NopCloser(bytes.NewReader(data))
	_, err := fm.client.PutObject(ctx, fm.bucket, fmt.Sprintf("users/%s/%s", userID, objectName), reader, int64(len(data)), minio.PutObjectOptions{
		ContentType: contentType,
	})
	return err
}

func (fm *FileManager) DownloadFile(ctx context.Context, userID, objectName string) ([]byte, error) {
	obj, err := fm.client.GetObject(ctx, fm.bucket, fmt.Sprintf("users/%s/%s", userID, objectName), minio.GetObjectOptions{})
	if err != nil {
		return nil, err
	}
	return io.ReadAll(obj)
}

func (fm *FileManager) DeleteFile(ctx context.Context, userID, objectName string) error {
	return fm.client.RemoveObject(ctx, fm.bucket, fmt.Sprintf("users/%s/%s", userID, objectName), minio.RemoveObjectOptions{})
}

func (fm *FileManager) ListFiles(ctx context.Context, userID, prefix string) ([]string, error) {
	objects := fm.client.ListObjects(ctx, fm.bucket, minio.ListObjectsOptions{
		Prefix: fmt.Sprintf("users/%s/%s", userID, prefix),
	})

	var files []string
	for object := range objects {
		files = append(files, object.Key)
	}
	return files, nil
}

func (fm *FileManager) CreateUserSpace(ctx context.Context, userID string) error {
	bucketName := fmt.Sprintf("user-%s", userID[:8])
	exists, _ := fm.client.BucketExists(ctx, bucketName)
	if !exists {
		return fm.client.MakeBucket(ctx, bucketName, minio.MakeBucketOptions{})
	}
	return nil
}

func (fm *FileManager) GetSignedURL(ctx context.Context, userID, objectName string, expiry time.Duration) (string, error) {
	url, err := fm.client.PresignedGetObject(ctx, fm.bucket, fmt.Sprintf("users/%s/%s", userID, objectName), expiry, nil)
	if err != nil {
		return "", err
	}
	return url.String(), nil
}
