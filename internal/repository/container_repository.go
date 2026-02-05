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
	ErrContainerNotFound = errors.New("container instance not found")
)

type ContainerRepository struct {
	db *gorm.DB
}

func NewContainerRepository(db *gorm.DB) *ContainerRepository {
	return &ContainerRepository{db: db}
}

func (r *ContainerRepository) Create(ctx context.Context, container *models.ContainerInstance) error {
	return r.db.WithContext(ctx).Create(container).Error
}

func (r *ContainerRepository) FindByID(ctx context.Context, id uuid.UUID) (*models.ContainerInstance, error) {
	var container models.ContainerInstance
	err := r.db.WithContext(ctx).
		Preload("User").
		Preload("Project").
		Preload("Template").
		Preload("GPUServer").
		First(&container, "id = ?", id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrContainerNotFound
	}
	if err != nil {
		return nil, err
	}
	return &container, nil
}

func (r *ContainerRepository) FindByUserID(ctx context.Context, userID uuid.UUID, limit, offset int) ([]models.ContainerInstance, int64, error) {
	var containers []models.ContainerInstance
	var total int64

	query := r.db.WithContext(ctx).Model(&models.ContainerInstance{}).Where("user_id = ?", userID)
	query.Count(&total)

	err := query.
		Preload("Project").
		Preload("Template").
		Preload("GPUServer").
		Limit(limit).
		Offset(offset).
		Order("created_at DESC").
		Find(&containers).Error

	if err != nil {
		return nil, 0, err
	}

	return containers, total, nil
}

func (r *ContainerRepository) FindByProjectID(ctx context.Context, projectID uuid.UUID, limit, offset int) ([]models.ContainerInstance, int64, error) {
	var containers []models.ContainerInstance
	var total int64

	query := r.db.WithContext(ctx).Model(&models.ContainerInstance{}).Where("project_id = ?", projectID)
	query.Count(&total)

	err := query.
		Preload("User").
		Preload("Template").
		Preload("GPUServer").
		Limit(limit).
		Offset(offset).
		Order("created_at DESC").
		Find(&containers).Error

	if err != nil {
		return nil, 0, err
	}

	return containers, total, nil
}

func (r *ContainerRepository) Update(ctx context.Context, container *models.ContainerInstance) error {
	return r.db.WithContext(ctx).Save(container).Error
}

func (r *ContainerRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status models.InstanceStatus) error {
	updates := map[string]interface{}{
		"status":     string(status),
		"updated_at": time.Now(),
	}

	if status == models.InstanceStatusRunning {
		updates["started_at"] = time.Now()
	} else if status == models.InstanceStatusStopped {
		now := time.Now()
		updates["stopped_at"] = &now
	}

	return r.db.WithContext(ctx).Model(&models.ContainerInstance{}).
		Where("id = ?", id).
		Updates(updates).Error
}

func (r *ContainerRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&models.ContainerInstance{}, "id = ?", id).Error
}

func (r *ContainerRepository) List(ctx context.Context, limit, offset int, status *string) ([]models.ContainerInstance, int64, error) {
	var containers []models.ContainerInstance
	var total int64

	query := r.db.WithContext(ctx).Model(&models.ContainerInstance{})
	if status != nil {
		query = query.Where("status = ?", *status)
	}

	query.Count(&total)

	err := query.
		Preload("User").
		Preload("Project").
		Preload("Template").
		Preload("GPUServer").
		Limit(limit).
		Offset(offset).
		Order("created_at DESC").
		Find(&containers).Error

	if err != nil {
		return nil, 0, err
	}

	return containers, total, nil
}

func (r *ContainerRepository) ListByGPUServer(ctx context.Context, serverID uuid.UUID) ([]models.ContainerInstance, error) {
	var containers []models.ContainerInstance
	err := r.db.WithContext(ctx).
		Preload("User").
		Preload("Project").
		Where("gpu_server_id = ? AND status IN ?", serverID,
			[]string{string(models.InstanceStatusRunning), string(models.InstanceStatusPending)}).
		Find(&containers).Error
	return containers, err
}

func (r *ContainerRepository) CountByUserID(ctx context.Context, userID uuid.UUID) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&models.ContainerInstance{}).
		Where("user_id = ? AND status NOT IN ?", userID,
			[]string{string(models.InstanceStatusDeleted), string(models.InstanceStatusError)}).
		Count(&count).Error
	return count, err
}

func (r *ContainerRepository) SaveMetrics(ctx context.Context, metrics *models.ContainerMetrics) error {
	return r.db.WithContext(ctx).Create(metrics).Error
}

func (r *ContainerRepository) GetMetrics(ctx context.Context, instanceID uuid.UUID, since time.Time) ([]models.ContainerMetrics, error) {
	var metrics []models.ContainerMetrics
	err := r.db.WithContext(ctx).
		Where("instance_id = ? AND recorded_at > ?", instanceID, since).
		Order("recorded_at DESC").
		Find(&metrics).Error
	return metrics, err
}
