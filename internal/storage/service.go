package storage

import (
	"bytes"
	"context"
	"fmt"
	"time"

	"gpu-platform/internal/config"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type StorageService struct {
	client *minio.Client
	bucket string
	cfg    *config.StorageConfig
}

func NewStorageService(cfg *config.StorageConfig) (*StorageService, error) {
	var client *minio.Client
	var err error

	if cfg.UseSSL {
		client, err = minio.New(cfg.Endpoint, &minio.Options{
			Creds:  credentials.NewStaticV4(cfg.AccessKey, cfg.SecretKey, ""),
			Secure: true,
		})
	} else {
		client, err = minio.New(cfg.Endpoint, &minio.Options{
			Creds:  credentials.NewStaticV4(cfg.AccessKey, cfg.SecretKey, ""),
			Secure: false,
		})
	}

	if err != nil {
		return nil, fmt.Errorf("failed to create MinIO client: %w", err)
	}

	svc := &StorageService{
		client: client,
		bucket: fmt.Sprintf("%s-data", cfg.BucketPrefix),
		cfg:    cfg,
	}

	if err := svc.ensureBucket(); err != nil {
		return nil, err
	}

	return svc, nil
}

func (s *StorageService) ensureBucket() error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	exists, err := s.client.BucketExists(ctx, s.bucket)
	if err != nil {
		return fmt.Errorf("failed to check bucket: %w", err)
	}

	if !exists {
		err = s.client.MakeBucket(ctx, s.bucket, minio.MakeBucketOptions{})
		if err != nil {
			return fmt.Errorf("failed to create bucket: %w", err)
		}
	}

	return nil
}

func (s *StorageService) UploadFile(ctx context.Context, userID, objectName string, data []byte, contentType string) error {
	reader := bytes.NewReader(data)
	_, err := s.client.PutObject(ctx, s.bucket, fmt.Sprintf("users/%s/%s", userID, objectName), reader, int64(len(data)), minio.PutObjectOptions{
		ContentType: contentType,
	})
	return err
}

func (s *StorageService) DownloadFile(ctx context.Context, userID, objectName string) ([]byte, error) {
	obj, err := s.client.GetObject(ctx, s.bucket, fmt.Sprintf("users/%s/%s", userID, objectName), minio.GetObjectOptions{})
	if err != nil {
		return nil, err
	}
	defer obj.Close()

	return []byte{}, nil
}

func (s *StorageService) DeleteFile(ctx context.Context, userID, objectName string) error {
	return s.client.RemoveObject(ctx, s.bucket, fmt.Sprintf("users/%s/%s", userID, objectName), minio.RemoveObjectOptions{})
}

func (s *StorageService) ListFiles(ctx context.Context, userID, prefix string) ([]string, error) {
	objects := s.client.ListObjects(ctx, s.bucket, minio.ListObjectsOptions{
		Prefix:    fmt.Sprintf("users/%s/%s", userID, prefix),
		Recursive: true,
	})

	var files []string
	for object := range objects {
		if object.Err != nil {
			return nil, object.Err
		}
		files = append(files, object.Key)
	}

	return files, nil
}

func (s *StorageService) CreateUserBucket(ctx context.Context, userID string) error {
	bucketName := fmt.Sprintf("user-%s", userID[:8])

	exists, err := s.client.BucketExists(ctx, bucketName)
	if err != nil {
		return err
	}

	if !exists {
		return s.client.MakeBucket(ctx, bucketName, minio.MakeBucketOptions{})
	}

	return nil
}

func (s *StorageService) GetSignedURL(ctx context.Context, userID, objectName string, expiry time.Duration) (string, error) {
	url, err := s.client.PresignedGetObject(ctx, s.bucket, fmt.Sprintf("users/%s/%s", userID, objectName), expiry, nil)
	if err != nil {
		return "", err
	}
	return url.String(), nil
}
