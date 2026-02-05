package repository

import (
	"context"
	"errors"
	"time"

	"gpu-platform/internal/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

var (
	ErrGPUServerNotFound = errors.New("GPU server not found")
)

type GPUServerRepository struct {
	db *gorm.DB
}

func NewGPUServerRepository(db *gorm.DB) *GPUServerRepository {
	return &GPUServerRepository{db: db}
}

func (r *GPUServerRepository) Create(ctx context.Context, server *models.GPUServer) error {
	return r.db.WithContext(ctx).Create(server).Error
}

func (r *GPUServerRepository) FindByID(ctx context.Context, id uuid.UUID) (*models.GPUServer, error) {
	var server models.GPUServer
	err := r.db.WithContext(ctx).
		Preload("GPUDevices").
		Preload("GPUPool").
		First(&server, "id = ?", id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrGPUServerNotFound
	}
	if err != nil {
		return nil, err
	}
	return &server, nil
}

func (r *GPUServerRepository) FindByHostname(ctx context.Context, hostname string) (*models.GPUServer, error) {
	var server models.GPUServer
	err := r.db.WithContext(ctx).
		Preload("GPUDevices").
		Preload("GPUPool").
		First(&server, "hostname = ?", hostname).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrGPUServerNotFound
	}
	if err != nil {
		return nil, err
	}
	return &server, nil
}

func (r *GPUServerRepository) Update(ctx context.Context, server *models.GPUServer) error {
	return r.db.WithContext(ctx).Save(server).Error
}

func (r *GPUServerRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status models.ServerStatus) error {
	return r.db.WithContext(ctx).Model(&models.GPUServer{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"status":         string(status),
			"last_heartbeat": time.Now(),
		}).Error
}

func (r *GPUServerRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&models.GPUServer{}, "id = ?", id).Error
}

func (r *GPUServerRepository) List(ctx context.Context, limit, offset int, poolID *uuid.UUID) ([]models.GPUServer, int64, error) {
	var servers []models.GPUServer
	var total int64

	query := r.db.WithContext(ctx).Model(&models.GPUServer{})
	if poolID != nil {
		query = query.Where("gpu_pool_id = ?", poolID)
	}

	query.Count(&total)

	err := query.
		Preload("GPUDevices").
		Preload("GPUPool").
		Limit(limit).
		Offset(offset).
		Order("created_at DESC").
		Find(&servers).Error

	if err != nil {
		return nil, 0, err
	}

	return servers, total, nil
}

func (r *GPUServerRepository) ListOnline(ctx context.Context) ([]models.GPUServer, error) {
	var servers []models.GPUServer
	err := r.db.WithContext(ctx).
		Preload("GPUDevices").
		Preload("GPUPool").
		Where("status = ?", models.ServerStatusOnline).
		Find(&servers).Error
	return servers, err
}

func (r *GPUServerRepository) UpdateHeartbeat(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Model(&models.GPUServer{}).
		Where("id = ?", id).
		Update("last_heartbeat", time.Now()).Error
}

func (r *GPUServerRepository) AddGPUDevice(ctx context.Context, serverID uuid.UUID, device *models.GPUDevice) error {
	device.ServerID = serverID
	return r.db.WithContext(ctx).Create(device).Error
}

func (r *GPUServerRepository) UpdateGPUDevice(ctx context.Context, device *models.GPUDevice) error {
	return r.db.WithContext(ctx).Save(device).Error
}
