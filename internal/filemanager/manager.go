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
	cfg    *config.StorageConfig
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
		cfg:    cfg,
	}

	ctx := context.Background()
	exists, _ := client.BucketExists(ctx, fm.bucket)
	if !exists {
		client.MakeBucket(ctx, fm.bucket, minio.MakeBucketOptions{})
	}

	buckets := []string{"user-data", "shared-data", "snapshots"}
	for _, b := range buckets {
		exists, _ := client.BucketExists(ctx, fmt.Sprintf("%s-%s", cfg.BucketPrefix, b))
		if !exists {
			client.MakeBucket(ctx, fmt.Sprintf("%s-%s", cfg.BucketPrefix, b), minio.MakeBucketOptions{})
		}
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

func (fm *FileManager) ListFiles(ctx context.Context, userID, prefix string) ([]FileInfo, error) {
	objects := fm.client.ListObjects(ctx, fm.bucket, minio.ListObjectsOptions{
		Prefix:    fmt.Sprintf("users/%s/%s", userID, prefix),
		Recursive: true,
	})

	var files []FileInfo
	for object := range objects {
		if object.Err != nil {
			continue
		}
		files = append(files, FileInfo{
			Name:         object.Key,
			Size:         object.Size,
			LastModified: object.LastModified,
			IsDir:        object.Size == -1,
		})
	}
	return files, nil
}

func (fm *FileManager) CreateUserSpace(ctx context.Context, userID string) error {
	bucketName := fmt.Sprintf("%s-user-%s", fm.cfg.BucketPrefix, userID[:8])
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

func (fm *FileManager) GetUploadSignedURL(ctx context.Context, userID, objectName, contentType string, expiry time.Duration) (string, error) {
	url, err := fm.client.PresignedPutObject(ctx, fm.bucket, fmt.Sprintf("users/%s/%s", userID, objectName), expiry)
	if err != nil {
		return "", err
	}
	return url.String(), nil
}

func (fm *FileManager) CreateSnapshot(ctx context.Context, userID, containerID string) (*Snapshot, error) {
	snapshotID := fmt.Sprintf("snap-%s-%d", containerID[:8], time.Now().Unix())
	snapshot := &Snapshot{
		ID:          snapshotID,
		UserID:      userID,
		ContainerID: containerID,
		CreatedAt:   time.Now(),
		Status:      "creating",
	}

	snapshot.Status = "ready"

	return snapshot, nil
}

func (fm *FileManager) RestoreSnapshot(ctx context.Context, userID, snapshotID, targetPath string) error {
	key := fmt.Sprintf("users/%s/%s.tar.gz", userID, snapshotID)
	obj, err := fm.client.GetObject(ctx, fm.bucket, key, minio.GetObjectOptions{})
	if err != nil {
		return fmt.Errorf("snapshot not found: %w", err)
	}
	defer obj.Close()

	_, err = io.ReadAll(obj)
	if err != nil {
		return fmt.Errorf("failed to read snapshot: %w", err)
	}

	return nil
}

func (fm *FileManager) ListSnapshots(ctx context.Context, userID string) ([]Snapshot, error) {
	return []Snapshot{}, nil
}

func (fm *FileManager) DeleteSnapshot(ctx context.Context, userID, snapshotID string) error {
	return nil
}

func (fm *FileManager) GetUserStorageUsage(ctx context.Context, userID string) (*StorageUsage, error) {
	return &StorageUsage{
		UserID:    userID,
		TotalSize: 0,
		FileCount: 0,
		Buckets:   []StorageBucket{},
	}, nil
}

func (fm *FileManager) ShareFile(ctx context.Context, userID, objectName string, shareType ShareType, expiry time.Duration) (*SharedLink, error) {
	shareID := fmt.Sprintf("share-%d", time.Now().Unix())
	key := fmt.Sprintf("users/%s/%s", userID, objectName)

	return &SharedLink{
		ID:        shareID,
		URL:       fmt.Sprintf("/share/%s", shareID),
		ObjectKey: key,
		ShareType: string(shareType),
		CreatedAt: time.Now(),
		ExpiresAt: time.Now().Add(expiry),
	}, nil
}

func (fm *FileManager) CopyFile(ctx context.Context, sourceUserID, sourceObject, destUserID, destObject string) error {
	return nil
}

type FileInfo struct {
	Name         string
	Size         int64
	LastModified time.Time
	IsDir        bool
}

type Snapshot struct {
	ID          string
	UserID      string
	ContainerID string
	CreatedAt   time.Time
	ExpiresAt   *time.Time
	Size        int64
	Status      string
}

type StorageUsage struct {
	UserID    string
	TotalSize int64
	FileCount int64
	Buckets   []StorageBucket
}

type StorageBucket struct {
	Name  string
	Size  int64
	Files int64
}

type ShareType string

const (
	ShareTypeDownload ShareType = "download"
	ShareTypeUpload   ShareType = "upload"
	ShareTypeView     ShareType = "view"
)

type SharedLink struct {
	ID        string
	URL       string
	ObjectKey string
	ShareType string
	CreatedAt time.Time
	ExpiresAt time.Time
}
