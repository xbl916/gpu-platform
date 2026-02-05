package repository

import (
	"context"
	"errors"

	"gpu-platform/internal/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type GPUPoolRepository struct {
	db *gorm.DB
}

func NewGPUPoolRepository(db *gorm.DB) *GPUPoolRepository {
	return &GPUPoolRepository{db: db}
}

func (r *GPUPoolRepository) Create(ctx context.Context, pool *models.GPUPool) error {
	return r.db.WithContext(ctx).Create(pool).Error
}

func (r *GPUPoolRepository) FindByID(ctx context.Context, id uuid.UUID) (*models.GPUPool, error) {
	var pool models.GPUPool
	err := r.db.WithContext(ctx).Preload("Servers").First(&pool, "id = ?", id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrProjectNotFound
	}
	if err != nil {
		return nil, err
	}
	return &pool, nil
}

func (r *GPUPoolRepository) FindByName(ctx context.Context, name string) (*models.GPUPool, error) {
	var pool models.GPUPool
	err := r.db.WithContext(ctx).Preload("Servers").First(&pool, "name = ?", name).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrProjectNotFound
	}
	if err != nil {
		return nil, err
	}
	return &pool, nil
}

func (r *GPUPoolRepository) Update(ctx context.Context, pool *models.GPUPool) error {
	return r.db.WithContext(ctx).Save(pool).Error
}

func (r *GPUPoolRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&models.GPUPool{}, "id = ?", id).Error
}

func (r *GPUPoolRepository) List(ctx context.Context, limit, offset int) ([]models.GPUPool, int64, error) {
	var pools []models.GPUPool
	var total int64

	r.db.WithContext(ctx).Model(&models.GPUPool{}).Count(&total)

	err := r.db.WithContext(ctx).
		Preload("Servers").
		Limit(limit).
		Offset(offset).
		Order("created_at DESC").
		Find(&pools).Error

	if err != nil {
		return nil, 0, err
	}

	return pools, total, nil
}

func (r *GPUPoolRepository) ListActive(ctx context.Context) ([]models.GPUPool, error) {
	var pools []models.GPUPool
	err := r.db.WithContext(ctx).
		Preload("Servers").
		Where("status = ?", "active").
		Find(&pools).Error
	return pools, err
}
