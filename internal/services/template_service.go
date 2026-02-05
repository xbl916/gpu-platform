package services

import (
	"context"

	"gpu-platform/internal/repository"
)

type TemplateService struct {
	templateRepo *repository.TemplateRepository
}

func NewTemplateService(templateRepo *repository.TemplateRepository) *TemplateService {
	return &TemplateService{templateRepo: templateRepo}
}

func (s *TemplateService) ListTemplates(ctx context.Context, limit, offset int) error {
	return nil
}

func (s *TemplateService) GetTemplate(ctx context.Context, id string) error {
	return nil
}
