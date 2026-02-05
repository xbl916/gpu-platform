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

	server, devices, err := s.resourceSvc.AllocateResources(ctx, allocationReq)
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
		s.resourceSvc.ReleaseResources(server.ID, req.Resources.GPUCount)
		return nil, fmt.Errorf("failed to create container record: %w", err)
	}

	go s.createAndStartContainer(context.Background(), instance, server, devices)

	return &ContainerResponse{
		Instance: instance,
		WebURL:   fmt.Sprintf("https://%s:%d", server.IPAddress, instance.WebPort),
		SSHInfo: SSHInfo{
			Host: server.IPAddress,
			Port: instance.SSHPort,
			User: "root",
		},
	}, nil
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

	if instance.Status != models.InstanceStatusStopped && instance.Status != models.InstanceStatusPending {
		return fmt.Errorf("cannot start container in status: %s", instance.Status)
	}

	err = s.containerRepo.UpdateStatus(ctx, id, models.InstanceStatusCreating)
	if err != nil {
		return err
	}

	go s.startContainerAsync(context.Background(), instance)

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

	err = s.containerRepo.UpdateStatus(ctx, id, models.InstanceStatusStopping)
	if err != nil {
		return err
	}

	go s.stopContainerAsync(context.Background(), instance)

	return nil
}

func (s *ContainerService) RestartContainer(ctx context.Context, id uuid.UUID) error {
	err := s.StopContainer(ctx, id)
	if err != nil {
		return err
	}

	time.Sleep(5 * time.Second)

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

	err = s.containerRepo.UpdateStatus(ctx, id, models.InstanceStatusDeleted)
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

	if s.k8sClient == nil {
		return []string{"Container logs not available"}, nil
	}

	return s.k8sClient.GetPodLogs(ctx, instance.PodName, s.cfg.Kubernetes.Namespace, tailLines)
}

func (s *ContainerService) ExecCommand(ctx context.Context, id uuid.UUID, command []string) ([]string, error) {
	instance, err := s.containerRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if instance.Status != models.InstanceStatusRunning {
		return nil, ErrContainerNotRunning
	}

	if s.k8sClient == nil {
		return []string{"Command execution not available"}, nil
	}

	return s.k8sClient.ExecCommand(ctx, instance.PodName, s.cfg.Kubernetes.Namespace, command)
}

func (s *ContainerService) GetContainerMetrics(ctx context.Context, id uuid.UUID) (*models.ContainerMetrics, error) {
	instance, err := s.containerRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if s.k8sClient == nil {
		return &models.ContainerMetrics{}, nil
	}

	k8sMetrics, err := s.k8sClient.GetPodMetrics(ctx, instance.PodName, s.cfg.Kubernetes.Namespace)
	if err != nil {
		return &models.ContainerMetrics{}, nil
	}

	return &models.ContainerMetrics{
		CPUUsage:      k8sMetrics.CPUUsage,
		MemoryUsageMB: k8sMetrics.MemoryUsageMB,
	}, nil
}

func (s *ContainerService) validateCreateRequest(req *CreateContainerRequest) error {
	if req.Name == "" || len(req.Name) < 3 || len(req.Name) > 63 {
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

func (s *ContainerService) createAndStartContainer(ctx context.Context, instance *models.ContainerInstance, server *models.GPUServer, devices []models.GPUDevice) {
	s.containerRepo.UpdateStatus(ctx, instance.ID, models.InstanceStatusCreating)

	podConfig := s.buildPodConfig(instance, server, devices)
	podName, err := s.k8sClient.CreatePod(ctx, podConfig)
	if err != nil {
		s.containerRepo.UpdateStatus(ctx, instance.ID, models.InstanceStatusError)
		return
	}

	instance.PodName = podName
	instance.ContainerID = podName
	s.containerRepo.Update(ctx, instance)

	err = s.k8sClient.WaitForPodReady(ctx, podName, s.cfg.Kubernetes.Namespace, 5*time.Minute)
	if err != nil {
		s.containerRepo.UpdateStatus(ctx, instance.ID, models.InstanceStatusError)
		return
	}

	s.containerRepo.UpdateStatus(ctx, instance.ID, models.InstanceStatusRunning)
}

func (s *ContainerService) startContainerAsync(ctx context.Context, instance *models.ContainerInstance) {
	s.containerRepo.UpdateStatus(ctx, instance.ID, models.InstanceStatusRunning)
}

func (s *ContainerService) stopContainerAsync(ctx context.Context, instance *models.ContainerInstance) {
	err := s.k8sClient.StopPod(ctx, instance.PodName, s.cfg.Kubernetes.Namespace)
	if err != nil {
		s.containerRepo.UpdateStatus(ctx, instance.ID, models.InstanceStatusError)
		return
	}

	now := time.Now()
	instance.StoppedAt = &now
	s.containerRepo.UpdateStatus(ctx, instance.ID, models.InstanceStatusStopped)
}

func (s *ContainerService) buildPodConfig(instance *models.ContainerInstance, server *models.GPUServer, devices []models.GPUDevice) *k8s.PodConfig {
	gpuIndices := make([]int, len(devices))
	for i := range devices {
		gpuIndices[i] = i
	}
	instance.Resources.GPUIndices = gpuIndices

	return &k8s.PodConfig{
		Name:      fmt.Sprintf("gpu-instance-%s", instance.ID.String()[:8]),
		Namespace: s.cfg.Kubernetes.Namespace,
		Labels: map[string]string{
			"app":                      "gpu-instance",
			"gpu-platform.io/instance": instance.ID.String(),
			"gpu-platform.io/user":     instance.UserID.String(),
		},
		Annotations: map[string]string{
			"gpu-platform.io/gpu-count": fmt.Sprintf("%d", instance.Resources.GPUCount),
		},
		Container: k8s.ContainerConfig{
			Name:    "main",
			Image:   instance.Resources.Image,
			Command: []string{"/bin/bash"},
			Args:    []string{"-c", instance.Resources.Command},
			Env: []k8s.EnvVar{
				{Name: "NVIDIA_VISIBLE_DEVICES", Value: "all"},
				{Name: "NVIDIA_DRIVER_CAPABILITIES", Value: "compute,utility"},
				{Name: "CUDA_VISIBLE_DEVICES", Value: fmt.Sprintf("%v", gpuIndices)},
				{Name: "USER_ID", Value: instance.UserID.String()},
				{Name: "INSTANCE_ID", Value: instance.ID.String()},
			},
			Resources: k8s.ResourceRequirements{
				Limits: k8s.ResourceList{
					"nvidia.com/gpu": fmt.Sprintf("%d", instance.Resources.GPUCount),
					"cpu":            fmt.Sprintf("%d", instance.Resources.CPUCores),
					"memory":         fmt.Sprintf("%dMi", instance.Resources.MemoryMB),
				},
				Requests: k8s.ResourceList{
					"nvidia.com/gpu": fmt.Sprintf("%d", instance.Resources.GPUCount),
					"cpu":            fmt.Sprintf("%d", instance.Resources.CPUCores),
					"memory":         fmt.Sprintf("%dMi", instance.Resources.MemoryMB),
				},
			},
			VolumeMounts: []k8s.VolumeMount{
				{Name: "workspace", MountPath: "/workspace"},
				{Name: "data", MountPath: "/data"},
			},
			Ports: []k8s.ContainerPort{
				{Name: "ssh", ContainerPort: 22},
				{Name: "jupyter", ContainerPort: 8888},
			},
		},
		Volumes: []k8s.Volume{
			{
				Name: "workspace",
				PersistentVolumeClaim: &k8s.PersistentVolumeClaimSource{
					ClaimName: fmt.Sprintf("workspace-%s", instance.ID.String()[:8]),
					SizeGB:    s.cfg.Container.WorkspaceSizeGB,
				},
			},
			{
				Name: "data",
				PersistentVolumeClaim: &k8s.PersistentVolumeClaimSource{
					ClaimName: fmt.Sprintf("data-%s", instance.ID.String()[:8]),
					SizeGB:    s.cfg.Container.DataSizeGB,
				},
			},
		},
		NodeSelector: map[string]string{
			"kubernetes.io/hostname": server.Hostname,
		},
		Tolerations: []k8s.Toleration{
			{
				Key:      "gpu",
				Operator: "Exists",
				Effect:   "NoSchedule",
			},
		},
	}
}
