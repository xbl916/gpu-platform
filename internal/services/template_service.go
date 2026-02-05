package services

import (
	"context"

	"gpu-platform/internal/models"
	"gpu-platform/internal/repository"

	"github.com/google/uuid"
)

type TemplateService struct {
	templateRepo *repository.TemplateRepository
}

type CreateTemplateRequest struct {
	Name        string                `json:"name"`
	Description string                `json:"description"`
	DockerImage string                `json:"dockerImage"`
	Config      models.TemplateConfig `json:"config"`
	IsPublic    bool                  `json:"isPublic"`
}

type UpdateTemplateRequest struct {
	Name        string                `json:"name"`
	Description string                `json:"description"`
	Config      models.TemplateConfig `json:"config"`
	IsPublic    bool                  `json:"isPublic"`
}

type TemplateResponse struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	DockerImage string `json:"dockerImage"`
	Version     int    `json:"version"`
	IsPublic    bool   `json:"isPublic"`
	CreatedAt   string `json:"createdAt"`
}

func NewTemplateService(templateRepo *repository.TemplateRepository) *TemplateService {
	return &TemplateService{templateRepo: templateRepo}
}

func (s *TemplateService) ListTemplates(ctx context.Context, limit, offset int, userID *uuid.UUID, projectID *uuid.UUID, isPublic *bool) ([]*TemplateResponse, int64, error) {
	templates, total, err := s.templateRepo.List(ctx, limit, offset, userID, projectID, isPublic)
	if err != nil {
		return nil, 0, err
	}

	responses := make([]*TemplateResponse, len(templates))
	for i, t := range templates {
		responses[i] = &TemplateResponse{
			ID:          t.ID.String(),
			Name:        t.Name,
			Description: t.Description,
			DockerImage: t.DockerImage,
			Version:     t.Version,
			IsPublic:    t.IsPublic,
			CreatedAt:   t.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		}
	}
	return responses, total, nil
}

func (s *TemplateService) GetTemplate(ctx context.Context, id uuid.UUID) (*TemplateResponse, error) {
	template, err := s.templateRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	return &TemplateResponse{
		ID:          template.ID.String(),
		Name:        template.Name,
		Description: template.Description,
		DockerImage: template.DockerImage,
		Version:     template.Version,
		IsPublic:    template.IsPublic,
		CreatedAt:   template.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}, nil
}

func (s *TemplateService) CreateTemplate(ctx context.Context, userID uuid.UUID, req *CreateTemplateRequest) (*TemplateResponse, error) {
	template := &models.ContainerTemplate{
		Name:        req.Name,
		Description: req.Description,
		DockerImage: req.DockerImage,
		Config:      req.Config,
		IsPublic:    req.IsPublic,
		UserID:      userID,
		Version:     1,
	}

	if err := s.templateRepo.Create(ctx, template); err != nil {
		return nil, err
	}

	return &TemplateResponse{
		ID:          template.ID.String(),
		Name:        template.Name,
		Description: template.Description,
		DockerImage: template.DockerImage,
		Version:     template.Version,
		IsPublic:    template.IsPublic,
		CreatedAt:   template.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}, nil
}

func (s *TemplateService) UpdateTemplate(ctx context.Context, id uuid.UUID, req *UpdateTemplateRequest) error {
	template, err := s.templateRepo.FindByID(ctx, id)
	if err != nil {
		return err
	}

	template.Name = req.Name
	template.Description = req.Description
	template.Config = req.Config
	template.IsPublic = req.IsPublic

	return s.templateRepo.Update(ctx, template)
}

func (s *TemplateService) DeleteTemplate(ctx context.Context, id uuid.UUID) error {
	return s.templateRepo.Delete(ctx, id)
}
