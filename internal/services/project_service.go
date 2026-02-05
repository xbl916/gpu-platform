package services

import (
	"context"
	"errors"
	"time"

	"gpu-platform/internal/models"
	"gpu-platform/internal/repository"

	"github.com/google/uuid"
)

var (
	ErrProjectNotFound     = errors.New("project not found")
	ErrProjectAccessDenied = errors.New("access denied to project")
	ErrAlreadyMember       = errors.New("already a member of project")
	ErrNotMember           = errors.New("not a member of project")
)

type ProjectService struct {
	projectRepo   *repository.ProjectRepository
	userRepo      *repository.UserRepository
	containerRepo *repository.ContainerRepository
	templateRepo  *repository.TemplateRepository
}

type CreateProjectRequest struct {
	Name        string              `json:"name" binding:"required,min=3,max=100"`
	Description string              `json:"description" binding:"max=1000"`
	QuotaConfig *models.QuotaConfig `json:"quotaConfig"`
}

type UpdateProjectRequest struct {
	Name        *string             `json:"name"`
	Description *string             `json:"description"`
	QuotaConfig *models.QuotaConfig `json:"quotaConfig"`
	Status      *string             `json:"status"`
}

type ProjectResponse struct {
	ID          string              `json:"id"`
	Name        string              `json:"name"`
	Description string              `json:"description"`
	QuotaConfig *models.QuotaConfig `json:"quotaConfig"`
	Status      string              `json:"status"`
	MemberCount int                 `json:"memberCount"`
	CreatedAt   time.Time           `json:"createdAt"`
}

type MemberResponse struct {
	ID       string    `json:"id"`
	UserID   string    `json:"userId"`
	Email    string    `json:"email"`
	Name     string    `json:"name"`
	Role     string    `json:"role"`
	JoinedAt time.Time `json:"joinedAt"`
}

type InviteMemberRequest struct {
	Email string `json:"email" binding:"required,email"`
	Role  string `json:"role" binding:"required"`
}

func NewProjectService(
	projectRepo *repository.ProjectRepository,
	userRepo *repository.UserRepository,
	containerRepo *repository.ContainerRepository,
	templateRepo *repository.TemplateRepository,
) *ProjectService {
	return &ProjectService{
		projectRepo:   projectRepo,
		userRepo:      userRepo,
		containerRepo: containerRepo,
		templateRepo:  templateRepo,
	}
}

func (s *ProjectService) CreateProject(ctx context.Context, ownerID uuid.UUID, req *CreateProjectRequest) (*ProjectResponse, error) {
	project := &models.Project{
		Name:        req.Name,
		Description: req.Description,
		Status:      string(models.ProjectStatusActive),
	}

	if req.QuotaConfig != nil {
		project.QuotaConfig = *req.QuotaConfig
	} else {
		project.QuotaConfig = models.QuotaConfig{
			MaxGPUCount:       8,
			MaxGPUInstances:   10,
			MaxCPUCores:       64,
			MaxTotalMemoryMB:  131072,
			MaxTotalStorageGB: 1000,
		}
	}

	if err := s.projectRepo.Create(ctx, project); err != nil {
		return nil, err
	}

	member := &models.ProjectMember{
		ProjectID: project.ID,
		UserID:    ownerID,
		Role:      "owner",
	}
	s.projectRepo.AddMember(ctx, member)

	return s.toProjectResponse(project, 1), nil
}

func (s *ProjectService) GetProject(ctx context.Context, projectID uuid.UUID, userID uuid.UUID) (*ProjectResponse, error) {
	project, err := s.projectRepo.FindByID(ctx, projectID)
	if err != nil {
		return nil, err
	}

	isMember, err := s.IsMember(ctx, projectID, userID)
	if err != nil {
		return nil, err
	}

	if !isMember && project.OwnerID != userID.String() {
		return nil, ErrProjectAccessDenied
	}

	memberCount, _ := s.projectRepo.GetMemberCount(ctx, projectID)

	return s.toProjectResponse(project, memberCount), nil
}

func (s *ProjectService) ListProjects(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*ProjectResponse, int64, error) {
	projects, total, err := s.projectRepo.ListByUserID(ctx, userID, limit, offset)
	if err != nil {
		return nil, 0, err
	}

	responses := make([]*ProjectResponse, len(projects))
	for i, p := range projects {
		memberCount, _ := s.projectRepo.GetMemberCount(ctx, p.ID)
		responses[i] = s.toProjectResponse(&p, memberCount)
	}

	return responses, total, nil
}

func (s *ProjectService) UpdateProject(ctx context.Context, projectID uuid.UUID, userID uuid.UUID, req *UpdateProjectRequest) (*ProjectResponse, error) {
	project, err := s.projectRepo.FindByID(ctx, projectID)
	if err != nil {
		return nil, err
	}

	if req.Name != nil {
		project.Name = *req.Name
	}
	if req.Description != nil {
		project.Description = *req.Description
	}
	if req.QuotaConfig != nil {
		project.QuotaConfig = *req.QuotaConfig
	}
	if req.Status != nil {
		project.Status = *req.Status
	}

	if err := s.projectRepo.Update(ctx, project); err != nil {
		return nil, err
	}

	memberCount, _ := s.projectRepo.GetMemberCount(ctx, projectID)

	return s.toProjectResponse(project, memberCount), nil
}

func (s *ProjectService) DeleteProject(ctx context.Context, projectID uuid.UUID, userID uuid.UUID) error {
	project, err := s.projectRepo.FindByID(ctx, projectID)
	if err != nil {
		return err
	}

	if project.OwnerID != userID.String() {
		return ErrProjectAccessDenied
	}

	return s.projectRepo.Delete(ctx, projectID)
}

func (s *ProjectService) AddMember(ctx context.Context, projectID uuid.UUID, req *InviteMemberRequest, inviterID uuid.UUID) (*MemberResponse, error) {
	user, err := s.userRepo.FindByEmail(ctx, req.Email)
	if err != nil {
		return nil, errors.New("user not found")
	}

	isMember, _ := s.IsMember(ctx, projectID, user.ID)
	if isMember {
		return nil, ErrAlreadyMember
	}

	member := &models.ProjectMember{
		ProjectID: projectID,
		UserID:    user.ID,
		Role:      req.Role,
	}

	if err := s.projectRepo.AddMember(ctx, member); err != nil {
		return nil, err
	}

	return &MemberResponse{
		ID:       member.ID.String(),
		UserID:   user.ID.String(),
		Email:    user.Email,
		Name:     user.Name,
		Role:     member.Role,
		JoinedAt: member.JoinedAt,
	}, nil
}

func (s *ProjectService) RemoveMember(ctx context.Context, projectID uuid.UUID, targetUserID uuid.UUID, requesterID uuid.UUID) error {
	project, err := s.projectRepo.FindByID(ctx, projectID)
	if err != nil {
		return err
	}

	if project.OwnerID == targetUserID.String() {
		return errors.New("cannot remove project owner")
	}

	return s.projectRepo.RemoveMember(ctx, projectID, targetUserID)
}

func (s *ProjectService) ListMembers(ctx context.Context, projectID uuid.UUID, userID uuid.UUID) ([]*MemberResponse, error) {
	members, err := s.projectRepo.ListMembers(ctx, projectID)
	if err != nil {
		return nil, err
	}

	responses := make([]*MemberResponse, len(members))
	for i, m := range members {
		user, _ := s.userRepo.FindByID(ctx, m.UserID)
		email := ""
		name := ""
		if user != nil {
			email = user.Email
			name = user.Name
		}
		responses[i] = &MemberResponse{
			ID:       m.ID.String(),
			UserID:   m.UserID.String(),
			Email:    email,
			Name:     name,
			Role:     m.Role,
			JoinedAt: m.JoinedAt,
		}
	}

	return responses, nil
}

func (s *ProjectService) IsMember(ctx context.Context, projectID uuid.UUID, userID uuid.UUID) (bool, error) {
	return s.projectRepo.IsMember(ctx, projectID, userID)
}

func (s *ProjectService) GetUserRole(ctx context.Context, projectID uuid.UUID, userID uuid.UUID) (string, error) {
	return s.projectRepo.GetUserRole(ctx, projectID, userID)
}

func (s *ProjectService) toProjectResponse(project *models.Project, memberCount int) *ProjectResponse {
	return &ProjectResponse{
		ID:          project.ID.String(),
		Name:        project.Name,
		Description: project.Description,
		QuotaConfig: &project.QuotaConfig,
		Status:      project.Status,
		MemberCount: memberCount,
		CreatedAt:   project.CreatedAt,
	}
}

type ProjectUsage struct {
	ProjectID    string `json:"projectId"`
	GPUCount     int    `json:"gpuCount"`
	GPUInstances int    `json:"gpuInstances"`
	CPUCores     int    `json:"cpuCores"`
	MemoryMB     int    `json:"memoryMb"`
	StorageGB    int    `json:"storageGb"`
}

func (s *ProjectService) GetProjectUsage(ctx context.Context, projectID uuid.UUID) (*ProjectUsage, error) {
	containers, _, err := s.containerRepo.FindByProjectID(ctx, projectID, 1000, 0)
	if err != nil {
		return nil, err
	}

	usage := &ProjectUsage{
		ProjectID: projectID.String(),
	}

	for _, c := range containers {
		if c.Status == models.InstanceStatusRunning || c.Status == models.InstanceStatusPending {
			usage.GPUCount += c.Resources.GPUCount
			usage.GPUInstances++
			usage.CPUCores += c.Resources.CPUCores
			usage.MemoryMB += c.Resources.MemoryMB
			usage.StorageGB += c.Resources.StorageGB
		}
	}

	return usage, nil
}
