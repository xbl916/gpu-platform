package services

import (
	"context"
	"errors"

	"gpu-platform/internal/config"
	"gpu-platform/internal/models"
	"gpu-platform/internal/repository"

	"github.com/google/uuid"
)

var (
	ErrQuotaGPUExceeded      = errors.New("GPU quota exceeded")
	ErrQuotaCPUExceeded      = errors.New("CPU quota exceeded")
	ErrQuotaMemoryExceeded   = errors.New("memory quota exceeded")
	ErrQuotaStorageExceeded  = errors.New("storage quota exceeded")
	ErrQuotaInstanceExceeded = errors.New("maximum instances quota exceeded")
)

type QuotaManager struct {
	cfg *config.Config
}

func NewQuotaManager() *QuotaManager {
	return &QuotaManager{}
}

func (q *QuotaManager) CheckQuota(ctx context.Context, req *AllocationRequest) error {
	return nil
}

func (q *QuotaManager) ReserveQuota(ctx context.Context, req *AllocationRequest) error {
	return nil
}

func (q *QuotaManager) ReleaseQuota(ctx context.Context, req *AllocationRequest) error {
	return nil
}

func (q *QuotaManager) GetUserQuota(ctx context.Context, userID uuid.UUID) (*QuotaInfo, error) {
	return &QuotaInfo{
		MaxGPUCount:   q.cfg.GPU.MaxGPUPerInstance,
		MaxCPUCores:   q.cfg.GPU.MaxCPUCores,
		MaxMemoryMB:   q.cfg.GPU.MaxMemoryMB,
		MaxStorageGB:  q.cfg.GPU.MaxStorageGB,
		MaxInstances:  q.cfg.Container.MaxInstancesPerUser,
		UsedGPUCount:  0,
		UsedCPUCores:  0,
		UsedMemoryMB:  0,
		UsedStorageGB: 0,
		UsedInstances: 0,
	}, nil
}

type QuotaInfo struct {
	MaxGPUCount   int
	MaxCPUCores   int
	MaxMemoryMB   int
	MaxStorageGB  int
	MaxInstances  int
	UsedGPUCount  int
	UsedCPUCores  int
	UsedMemoryMB  int
	UsedStorageGB int
	UsedInstances int
}

func (q *QuotaInfo) CanAllocateGPU(count int) bool {
	return q.UsedGPUCount+count <= q.MaxGPUCount
}

func (q *QuotaInfo) CanAllocateCPU(cores int) bool {
	return q.UsedCPUCores+cores <= q.MaxCPUCores
}

func (q *QuotaInfo) CanAllocateMemory(memoryMB int) bool {
	return q.UsedMemoryMB+memoryMB <= q.MaxMemoryMB
}

func (q *QuotaInfo) CanAllocateStorage(storageGB int) bool {
	return q.UsedStorageGB+storageGB <= q.MaxStorageGB
}

func (q *QuotaInfo) CanCreateInstance() bool {
	return q.UsedInstances < q.MaxInstances
}

func (q *QuotaInfo) RemainingGPU() int {
	if q.MaxGPUCount <= 0 {
		return 0
	}
	return q.MaxGPUCount - q.UsedGPUCount
}

func (q *QuotaInfo) RemainingCPU() int {
	if q.MaxCPUCores <= 0 {
		return 0
	}
	return q.MaxCPUCores - q.UsedCPUCores
}

func (q *QuotaInfo) RemainingMemory() int {
	if q.MaxMemoryMB <= 0 {
		return 0
	}
	return q.MaxMemoryMB - q.UsedMemoryMB
}

func (q *QuotaInfo) RemainingStorage() int {
	if q.MaxStorageGB <= 0 {
		return 0
	}
	return q.MaxStorageGB - q.UsedStorageGB
}

func (q *QuotaInfo) RemainingInstances() int {
	if q.MaxInstances <= 0 {
		return 0
	}
	return q.MaxInstances - q.UsedInstances
}

type QuotaChecker struct {
	cfg *config.Config
}

func NewQuotaChecker(cfg *config.Config) *QuotaChecker {
	return &QuotaChecker{cfg: cfg}
}

func (c *QuotaChecker) Check(userQuota *QuotaInfo, resources *models.ContainerResources) error {
	if !userQuota.CanAllocateGPU(resources.GPUCount) {
		return ErrQuotaGPUExceeded
	}
	if !userQuota.CanAllocateCPU(resources.CPUCores) {
		return ErrQuotaCPUExceeded
	}
	if !userQuota.CanAllocateMemory(resources.MemoryMB) {
		return ErrQuotaMemoryExceeded
	}
	if !userQuota.CanAllocateStorage(resources.StorageGB) {
		return ErrQuotaStorageExceeded
	}
	if !userQuota.CanCreateInstance() {
		return ErrQuotaInstanceExceeded
	}
	return nil
}

func (c *QuotaChecker) CalculateUsage(resources *models.ContainerResources) {
	_ = resources
}

type QuotaRepository struct {
	userRepo      *repository.UserRepository
	projectRepo   *repository.ProjectRepository
	containerRepo *repository.ContainerRepository
}

func NewQuotaRepository(
	userRepo *repository.UserRepository,
	projectRepo *repository.ProjectRepository,
	containerRepo *repository.ContainerRepository,
) *QuotaRepository {
	return &QuotaRepository{
		userRepo:      userRepo,
		projectRepo:   projectRepo,
		containerRepo: containerRepo,
	}
}

func (r *QuotaRepository) GetProjectQuota(ctx context.Context, projectID uuid.UUID) (*models.QuotaConfig, error) {
	project, err := r.projectRepo.FindByID(ctx, projectID)
	if err != nil {
		return nil, err
	}
	return &project.QuotaConfig, nil
}

func (r *QuotaRepository) GetUserUsage(ctx context.Context, userID uuid.UUID) (*QuotaInfo, error) {
	containers, _, err := r.containerRepo.FindByUserID(ctx, userID, 1000, 0)
	if err != nil {
		return nil, err
	}

	usage := &QuotaInfo{}

	for _, container := range containers {
		if container.Status == models.InstanceStatusRunning || container.Status == models.InstanceStatusPending {
			usage.UsedGPUCount += container.Resources.GPUCount
			usage.UsedCPUCores += container.Resources.CPUCores
			usage.UsedMemoryMB += container.Resources.MemoryMB
			usage.UsedStorageGB += container.Resources.StorageGB
			usage.UsedInstances++
		}
	}

	return usage, nil
}
