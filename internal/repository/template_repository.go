package repository

import (
	"context"
	"errors"

	"gpu-platform/internal/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

var (
	ErrTemplateNotFound = errors.New("template not found")
)

type TemplateRepository struct {
	db *gorm.DB
}

func NewTemplateRepository(db *gorm.DB) *TemplateRepository {
	return &TemplateRepository{db: db}
}

func (r *TemplateRepository) Create(ctx context.Context, template *models.ContainerTemplate) error {
	return r.db.WithContext(ctx).Create(template).Error
}

func (r *TemplateRepository) FindByID(ctx context.Context, id uuid.UUID) (*models.ContainerTemplate, error) {
	var template models.ContainerTemplate
	err := r.db.WithContext(ctx).
		Preload("User").
		Preload("Project").
		First(&template, "id = ?", id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrTemplateNotFound
	}
	if err != nil {
		return nil, err
	}
	return &template, nil
}

func (r *TemplateRepository) FindByName(ctx context.Context, name string) (*models.ContainerTemplate, error) {
	var template models.ContainerTemplate
	err := r.db.WithContext(ctx).Preload("User").First(&template, "name = ?", name).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrTemplateNotFound
	}
	if err != nil {
		return nil, err
	}
	return &template, nil
}

func (r *TemplateRepository) Update(ctx context.Context, template *models.ContainerTemplate) error {
	return r.db.WithContext(ctx).Save(template).Error
}

func (r *TemplateRepository) UpdateVersion(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Model(&models.ContainerTemplate{}).
		Where("id = ?", id).
		UpdateColumn("version", gorm.Expr("version + 1")).Error
}

func (r *TemplateRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&models.ContainerTemplate{}, "id = ?", id).Error
}

func (r *TemplateRepository) List(ctx context.Context, limit, offset int, userID *uuid.UUID, projectID *uuid.UUID, isPublic *bool) ([]models.ContainerTemplate, int64, error) {
	var templates []models.ContainerTemplate
	var total int64

	query := r.db.WithContext(ctx).Model(&models.ContainerTemplate{})

	if userID != nil {
		query = query.Where("user_id = ?", userID)
	}
	if projectID != nil {
		query = query.Where("project_id = ?", projectID)
	}
	if isPublic != nil {
		query = query.Where("is_public = ?", *isPublic)
	}

	query.Count(&total)

	err := query.
		Preload("User").
		Preload("Project").
		Limit(limit).
		Offset(offset).
		Order("created_at DESC").
		Find(&templates).Error

	if err != nil {
		return nil, 0, err
	}

	return templates, total, nil
}

func (r *TemplateRepository) ListPublic(ctx context.Context, limit, offset int) ([]models.ContainerTemplate, int64, error) {
	return r.List(ctx, limit, offset, nil, nil, boolPtr(true))
}

func (r *TemplateRepository) IncrementUsage(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Model(&models.ContainerTemplate{}).
		Where("id = ?", id).
		UpdateColumn("usage_count", gorm.Expr("usage_count + 1")).Error
}

func boolPtr(b bool) *bool {
	return &b
}
