package repository

import (
	"context"
	"errors"

	"gpu-platform/internal/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

var (
	ErrProjectNotFound = errors.New("project not found")
)

type ProjectRepository struct {
	db *gorm.DB
}

func NewProjectRepository(db *gorm.DB) *ProjectRepository {
	return &ProjectRepository{db: db}
}

func (r *ProjectRepository) Create(ctx context.Context, project *models.Project) error {
	return r.db.WithContext(ctx).Create(project).Error
}

func (r *ProjectRepository) FindByID(ctx context.Context, id uuid.UUID) (*models.Project, error) {
	var project models.Project
	err := r.db.WithContext(ctx).Preload("Members").First(&project, "id = ?", id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrProjectNotFound
	}
	if err != nil {
		return nil, err
	}
	return &project, nil
}

func (r *ProjectRepository) FindByName(ctx context.Context, name string) (*models.Project, error) {
	var project models.Project
	err := r.db.WithContext(ctx).First(&project, "name = ?", name).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrProjectNotFound
	}
	if err != nil {
		return nil, err
	}
	return &project, nil
}

func (r *ProjectRepository) Update(ctx context.Context, project *models.Project) error {
	return r.db.WithContext(ctx).Save(project).Error
}

func (r *ProjectRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&models.Project{}, "id = ?", id).Error
}

func (r *ProjectRepository) List(ctx context.Context, limit, offset int) ([]models.Project, int64, error) {
	var projects []models.Project
	var total int64

	r.db.WithContext(ctx).Model(&models.Project{}).Count(&total)

	err := r.db.WithContext(ctx).
		Preload("Members").
		Limit(limit).
		Offset(offset).
		Order("created_at DESC").
		Find(&projects).Error

	if err != nil {
		return nil, 0, err
	}

	return projects, total, nil
}

func (r *ProjectRepository) AddMember(ctx context.Context, member *models.ProjectMember) error {
	return r.db.WithContext(ctx).Create(member).Error
}

func (r *ProjectRepository) RemoveMember(ctx context.Context, projectID, userID uuid.UUID) error {
	return r.db.WithContext(ctx).
		Where("project_id = ? AND user_id = ?", projectID, userID).
		Delete(&models.ProjectMember{}).Error
}

func (r *ProjectRepository) ListMembers(ctx context.Context, projectID uuid.UUID) ([]models.ProjectMember, error) {
	var members []models.ProjectMember
	err := r.db.WithContext(ctx).
		Preload("User").
		Where("project_id = ?", projectID).
		Find(&members).Error
	return members, err
}

func (r *ProjectRepository) IsMember(ctx context.Context, projectID, userID uuid.UUID) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&models.ProjectMember{}).
		Where("project_id = ? AND user_id = ?", projectID, userID).
		Count(&count).Error
	return count > 0, err
}

func (r *ProjectRepository) GetUserRole(ctx context.Context, projectID, userID uuid.UUID) (string, error) {
	var member models.ProjectMember
	err := r.db.WithContext(ctx).
		Where("project_id = ? AND user_id = ?", projectID, userID).
		First(&member).Error
	if err != nil {
		return "", err
	}
	return member.Role, nil
}

func (r *ProjectRepository) GetMemberCount(ctx context.Context, projectID uuid.UUID) (int, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&models.ProjectMember{}).
		Where("project_id = ?", projectID).
		Count(&count).Error
	return int(count), err
}

func (r *ProjectRepository) ListByUserID(ctx context.Context, userID uuid.UUID, limit, offset int) ([]models.Project, int64, error) {
	var projects []models.Project
	var total int64

	subQuery := r.db.WithContext(ctx).
		Model(&models.ProjectMember{}).
		Select("project_id").
		Where("user_id = ?", userID)

	err := r.db.WithContext(ctx).
		Model(&models.Project{}).
		Where("id IN (?)", subQuery).
		Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	err = r.db.WithContext(ctx).
		Preload("Members").
		Where("id IN (?)", subQuery).
		Limit(limit).
		Offset(offset).
		Order("created_at DESC").
		Find(&projects).Error

	return projects, total, err
}

func (r *ProjectRepository) GetContainersByProjectID(ctx context.Context, projectID uuid.UUID, limit, offset int) ([]models.ContainerInstance, int64, error) {
	var containers []models.ContainerInstance
	var total int64

	err := r.db.WithContext(ctx).
		Model(&models.ContainerInstance{}).
		Where("project_id = ?", projectID).
		Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	err = r.db.WithContext(ctx).
		Where("project_id = ?", projectID).
		Limit(limit).
		Offset(offset).
		Order("created_at DESC").
		Find(&containers).Error

	return containers, total, err
}
