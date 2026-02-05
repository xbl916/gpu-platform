package services

import (
	"context"
	"errors"
	"fmt"
	"time"

	"gpu-platform/internal/config"
	"gpu-platform/internal/k8s"
	"gpu-platform/internal/models"
	"gpu-platform/internal/repository"

	"github.com/google/uuid"
)

var (
	ErrContainerNotFound     = errors.New("container instance not found")
	ErrContainerCreateFailed = errors.New("failed to create container instance")
	ErrContainerStartFailed  = errors.New("failed to start container instance")
	ErrContainerStopFailed   = errors.New("failed to stop container instance")
	ErrContainerDeleteFailed = errors.New("failed to delete container instance")
	ErrInvalidContainerReq   = errors.New("invalid container request")
	ErrContainerNotRunning   = errors.New("container instance is not running")
)

type ContainerService struct {
	cfg           *config.Config
	containerRepo *repository.ContainerRepository
	resourceSvc   *ResourceService
	k8sClient     *k8s.K8sClient
}

type CreateContainerRequest struct {
	Name        string
	UserID      uuid.UUID
	ProjectID   *uuid.UUID
	TemplateID  *uuid.UUID
	Resources   models.ContainerResources
	Environment map[string]string
	StartCmd    string
}

type ContainerResponse struct {
	Instance *models.ContainerInstance
	WebURL   string
	SSHInfo  SSHInfo
}

type SSHInfo struct {
	Host string
	Port int
	User string
}

func NewContainerService(
	cfg *config.Config,
	containerRepo *repository.ContainerRepository,
	resourceSvc *ResourceService,
	k8sClient *k8s.K8sClient,
) *ContainerService {
	return &ContainerService{
		cfg:           cfg,
		containerRepo: containerRepo,
		resourceSvc:   resourceSvc,
		k8sClient:     k8sClient,
	}
}

func (s *ContainerService) CreateContainer(ctx context.Context, req *CreateContainerRequest) (*ContainerResponse, error) {
	if err := s.validateCreateRequest(req); err != nil {
		return nil, err
	}

	allocationReq := &AllocationRequest{
		GPUCount:  req.Resources.GPUCount,
		GPUModel:  req.Resources.GPUModel,
		CPUCores:  req.Resources.CPUCores,
		MemoryMB:  req.Resources.MemoryMB,
		StorageGB: req.Resources.StorageGB,
		ProjectID: req.ProjectID,
		UserID:    req.UserID,
		Priority:  0,
	}

	server, _, err := s.resourceSvc.AllocateResources(ctx, allocationReq)
	if err != nil {
		return nil, fmt.Errorf("failed to allocate resources: %w", err)
	}

	instance := &models.ContainerInstance{
		Name:        req.Name,
		UserID:      req.UserID,
		ProjectID:   req.ProjectID,
		TemplateID:  req.TemplateID,
		GPUServerID: &server.ID,
		Status:      models.InstanceStatusPending,
		Resources:   req.Resources,
		SSHPort:     2222,
		WebPort:     8888,
	}

	if req.Resources.Image == "" {
		instance.Resources.Image = s.cfg.Container.DefaultImage
	}

	err = s.containerRepo.Create(ctx, instance)
	if err != nil {
		return nil, fmt.Errorf("failed to create container record: %w", err)
	}

	if s.k8sClient != nil {
		go s.createAndStartContainer(context.Background(), instance, server)
	}

	return &ContainerResponse{
		Instance: instance,
		WebURL:   fmt.Sprintf("http://%s:%d", server.Hostname, instance.WebPort),
		SSHInfo: SSHInfo{
			Host: server.Hostname,
			Port: instance.SSHPort,
			User: "root",
		},
	}, nil
}

func (s *ContainerService) createAndStartContainer(ctx context.Context, instance *models.ContainerInstance, server *models.GPUServer) {
	s.containerRepo.UpdateStatus(ctx, instance.ID, models.InstanceStatusCreating)

	podName := fmt.Sprintf("gpu-instance-%s", instance.ID.String()[:8])
	resources := k8s.ResourceRequirements{
		CPU:    instance.Resources.CPUCores,
		Memory: instance.Resources.MemoryMB,
	}

	podInfo, err := s.k8sClient.CreatePod(ctx, podName, instance.Resources.Image, instance.Resources.GPUCount, resources)
	if err != nil {
		s.containerRepo.UpdateStatus(ctx, instance.ID, models.InstanceStatusError)
		return
	}

	err = s.containerRepo.Update(ctx, &models.ContainerInstance{
		ID:        instance.ID,
		PodName:   podInfo.Name,
		IPAddress: podInfo.IP,
		Status:    models.InstanceStatusCreating,
	})
	if err != nil {
		s.containerRepo.UpdateStatus(ctx, instance.ID, models.InstanceStatusError)
		return
	}

	err = s.k8sClient.WaitForReady(ctx, podInfo.Name, 5*time.Minute)
	if err != nil {
		s.containerRepo.UpdateStatus(ctx, instance.ID, models.InstanceStatusError)
		return
	}

	s.containerRepo.UpdateStatus(ctx, instance.ID, models.InstanceStatusRunning)
}

func (s *ContainerService) GetContainer(ctx context.Context, id uuid.UUID) (*models.ContainerInstance, error) {
	return s.containerRepo.FindByID(ctx, id)
}

func (s *ContainerService) ListContainers(ctx context.Context, userID uuid.UUID, limit, offset int) ([]models.ContainerInstance, int64, error) {
	return s.containerRepo.FindByUserID(ctx, userID, limit, offset)
}

func (s *ContainerService) ListAllContainers(ctx context.Context, limit, offset int, status *string) ([]models.ContainerInstance, int64, error) {
	return s.containerRepo.List(ctx, limit, offset, status)
}

func (s *ContainerService) StartContainer(ctx context.Context, id uuid.UUID) error {
	instance, err := s.containerRepo.FindByID(ctx, id)
	if err != nil {
		return err
	}

	if !s.isStopped(instance.Status) {
		return fmt.Errorf("cannot start container in status: %s", instance.Status)
	}

	err = s.containerRepo.UpdateStatus(ctx, id, models.InstanceStatusPending)
	if err != nil {
		return err
	}

	if s.k8sClient != nil && instance.PodName != "" {
		err = s.k8sClient.WaitForReady(ctx, instance.PodName, 2*time.Minute)
		if err != nil {
			s.containerRepo.UpdateStatus(ctx, id, models.InstanceStatusError)
			return err
		}

		s.containerRepo.UpdateStatus(ctx, id, models.InstanceStatusRunning)
	}

	return nil
}

func (s *ContainerService) StopContainer(ctx context.Context, id uuid.UUID) error {
	instance, err := s.containerRepo.FindByID(ctx, id)
	if err != nil {
		return err
	}

	if instance.Status != models.InstanceStatusRunning {
		return ErrContainerNotRunning
	}

	if s.k8sClient != nil && instance.PodName != "" {
		err = s.k8sClient.DeletePod(ctx, instance.PodName)
		if err != nil {
			return fmt.Errorf("failed to delete pod: %w", err)
		}
	}

	now := time.Now()
	instance.StoppedAt = &now
	err = s.containerRepo.Update(ctx, instance)
	if err != nil {
		return err
	}

	return s.containerRepo.UpdateStatus(ctx, id, models.InstanceStatusStopped)
}

func (s *ContainerService) RestartContainer(ctx context.Context, id uuid.UUID) error {
	err := s.StopContainer(ctx, id)
	if err != nil {
		return err
	}

	time.Sleep(2 * time.Second)

	return s.StartContainer(ctx, id)
}

func (s *ContainerService) DeleteContainer(ctx context.Context, id uuid.UUID) error {
	instance, err := s.containerRepo.FindByID(ctx, id)
	if err != nil {
		return err
	}

	if instance.Status == models.InstanceStatusRunning {
		return errors.New("cannot delete running container, stop it first")
	}

	if s.k8sClient != nil && instance.PodName != "" {
		s.k8sClient.DeletePod(ctx, instance.PodName)
	}

	err = s.containerRepo.Delete(ctx, id)
	if err != nil {
		return err
	}

	s.resourceSvc.ReleaseResources(*instance.GPUServerID, instance.Resources.GPUCount)

	return nil
}

func (s *ContainerService) GetContainerLogs(ctx context.Context, id uuid.UUID, tailLines int) ([]string, error) {
	instance, err := s.containerRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if s.k8sClient == nil || instance.PodName == "" {
		return []string{"Logs not available"}, nil
	}

	return s.k8sClient.GetPodLogs(ctx, instance.PodName, int64(tailLines))
}

func (s *ContainerService) ExecCommand(ctx context.Context, id uuid.UUID, command []string) ([]string, error) {
	instance, err := s.containerRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if instance.Status != models.InstanceStatusRunning {
		return nil, ErrContainerNotRunning
	}

	if s.k8sClient == nil || instance.PodName == "" {
		return []string{"Command execution not available"}, nil
	}

	return s.k8sClient.ExecCommand(ctx, instance.PodName, command)
}

func (s *ContainerService) GetContainerMetrics(ctx context.Context, id uuid.UUID) (*models.ContainerMetrics, error) {
	instance, err := s.containerRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if s.k8sClient == nil || instance.PodName == "" {
		return &models.ContainerMetrics{
			CPUUsage:      0,
			MemoryUsageMB: 0,
		}, nil
	}

	return &models.ContainerMetrics{
		CPUUsage:      0,
		MemoryUsageMB: 0,
		RecordedAt:    time.Now(),
	}, nil
}

func (s *ContainerService) validateCreateRequest(req *CreateContainerRequest) error {
	if len(req.Name) < 3 || len(req.Name) > 63 {
		return fmt.Errorf("%w: name must be 3-63 characters", ErrInvalidContainerReq)
	}
	if req.Resources.GPUCount <= 0 {
		return fmt.Errorf("%w: GPU count must be positive", ErrInvalidContainerReq)
	}
	if req.Resources.CPUCores <= 0 {
		return fmt.Errorf("%w: CPU cores must be positive", ErrInvalidContainerReq)
	}
	if req.Resources.MemoryMB <= 0 {
		return fmt.Errorf("%w: memory must be positive", ErrInvalidContainerReq)
	}
	return nil
}

func (s *ContainerService) isStopped(status models.InstanceStatus) bool {
	return status == models.InstanceStatusStopped ||
		status == models.InstanceStatusPending ||
		status == models.InstanceStatusError
}
