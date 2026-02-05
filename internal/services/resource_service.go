package services

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"gpu-platform/internal/config"
	"gpu-platform/internal/models"
	"gpu-platform/internal/repository"

	"github.com/google/uuid"
)

var (
	ErrNoAvailableGPU     = errors.New("no available GPU resources")
	ErrGPUServerOffline   = errors.New("GPU server is offline")
	ErrInvalidResourceReq = errors.New("invalid resource request")
	ErrQuotaExceeded      = errors.New("resource quota exceeded")
	ErrGPUNotFound        = errors.New("GPU device not found")
)

type ResourceService struct {
	cfg           *config.Config
	gpuPoolRepo   *repository.GPUPoolRepository
	gpuServerRepo *repository.GPUServerRepository
	containerRepo *repository.ContainerRepository
	quotaManager  *QuotaManager
	scheduler     *GPUScheduler
	mu            sync.RWMutex
	resourceCache map[uuid.UUID]*CachedServer
}

type CachedServer struct {
	Server       *models.GPUServer
	LastUpdate   time.Time
	AvailableGPU int
}

type AllocationRequest struct {
	GPUCount  int
	GPUModel  string
	CPUCores  int
	MemoryMB  int
	StorageGB int
	ProjectID *uuid.UUID
	UserID    uuid.UUID
	Priority  int
}

type AvailabilityInfo struct {
	TotalServers      int
	OnlineServers     int
	TotalGPUCount     int
	AvailableGPUCount int
	GPUByModel        map[string]int
}

func NewResourceService(
	cfg *config.Config,
	gpuPoolRepo *repository.GPUPoolRepository,
	gpuServerRepo *repository.GPUServerRepository,
	containerRepo *repository.ContainerRepository,
) *ResourceService {
	svc := &ResourceService{
		cfg:           cfg,
		gpuPoolRepo:   gpuPoolRepo,
		gpuServerRepo: gpuServerRepo,
		containerRepo: containerRepo,
		quotaManager:  NewQuotaManager(),
		scheduler:     NewGPUScheduler(),
		resourceCache: make(map[uuid.UUID]*CachedServer),
	}

	go svc.cacheRefreshLoop()

	return svc
}

func (s *ResourceService) ListGPUPools(ctx context.Context, limit, offset int) ([]models.GPUPool, int64, error) {
	return s.gpuPoolRepo.List(ctx, limit, offset)
}

func (s *ResourceService) GetGPUPool(ctx context.Context, id uuid.UUID) (*models.GPUPool, error) {
	return s.gpuPoolRepo.FindByID(ctx, id)
}

func (s *ResourceService) CreateGPUPool(ctx context.Context, pool *models.GPUPool) error {
	return s.gpuPoolRepo.Create(ctx, pool)
}

func (s *ResourceService) UpdateGPUPool(ctx context.Context, pool *models.GPUPool) error {
	return s.gpuPoolRepo.Update(ctx, pool)
}

func (s *ResourceService) DeleteGPUPool(ctx context.Context, id uuid.UUID) error {
	return s.gpuPoolRepo.Delete(ctx, id)
}

func (s *ResourceService) ListGPUServers(ctx context.Context, limit, offset int, poolID *uuid.UUID) ([]models.GPUServer, int64, error) {
	return s.gpuServerRepo.List(ctx, limit, offset, poolID)
}

func (s *ResourceService) GetGPUServer(ctx context.Context, id uuid.UUID) (*models.GPUServer, error) {
	return s.gpuServerRepo.FindByID(ctx, id)
}

func (s *ResourceService) CreateGPUServer(ctx context.Context, server *models.GPUServer) error {
	return s.gpuServerRepo.Create(ctx, server)
}

func (s *ResourceService) UpdateGPUServer(ctx context.Context, server *models.GPUServer) error {
	return s.gpuServerRepo.Update(ctx, server)
}

func (s *ResourceService) DeleteGPUServer(ctx context.Context, id uuid.UUID) error {
	return s.gpuServerRepo.Delete(ctx, id)
}

func (s *ResourceService) UpdateServerHeartbeat(ctx context.Context, id uuid.UUID) error {
	return s.gpuServerRepo.UpdateHeartbeat(ctx, id)
}

func (s *ResourceService) UpdateServerStatus(ctx context.Context, id uuid.UUID, status models.ServerStatus) error {
	return s.gpuServerRepo.UpdateStatus(ctx, id, status)
}

func (s *ResourceService) RegisterGPUDevice(ctx context.Context, serverID uuid.UUID, device *models.GPUDevice) error {
	return s.gpuServerRepo.AddGPUDevice(ctx, serverID, device)
}

func (s *ResourceService) UpdateGPUDevice(ctx context.Context, device *models.GPUDevice) error {
	return s.gpuServerRepo.UpdateGPUDevice(ctx, device)
}

func (s *ResourceService) GetAvailability(ctx context.Context) (*AvailabilityInfo, error) {
	servers, err := s.gpuServerRepo.ListOnline(ctx)
	if err != nil {
		return nil, err
	}

	info := &AvailabilityInfo{
		TotalServers:  len(servers),
		OnlineServers: len(servers),
		GPUByModel:    make(map[string]int),
	}

	for _, server := range servers {
		info.TotalGPUCount += len(server.GPUDevices)
		for _, device := range server.GPUDevices {
			if device.Status == models.GPUStatusAvailable {
				info.AvailableGPUCount++
				info.GPUByModel[device.Model]++
			}
		}
	}

	return info, nil
}

func (s *ResourceService) AllocateResources(ctx context.Context, req *AllocationRequest) (*models.GPUServer, []models.GPUDevice, error) {
	if err := s.validateAllocationRequest(req); err != nil {
		return nil, nil, err
	}

	if err := s.quotaManager.CheckQuota(ctx, req); err != nil {
		return nil, nil, err
	}

	servers, err := s.gpuServerRepo.ListOnline(ctx)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to list online servers: %w", err)
	}

	candidates := s.filterCandidates(servers, req)

	if len(candidates) == 0 {
		return nil, nil, ErrNoAvailableGPU
	}

	selectedServer := s.scheduler.SelectServer(candidates, req)

	devices := s.allocateGPUDevices(selectedServer, req.GPUCount)

	err = s.quotaManager.ReserveQuota(ctx, req)
	if err != nil {
		return nil, nil, err
	}

	s.mu.Lock()
	s.resourceCache[selectedServer.ID] = &CachedServer{
		Server:       selectedServer,
		LastUpdate:   time.Now(),
		AvailableGPU: selectedServer.AvailableGPUCount() - req.GPUCount,
	}
	s.mu.Unlock()

	return selectedServer, devices, nil
}

func (s *ResourceService) ReleaseResources(serverID uuid.UUID, gpuCount int) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if cached, ok := s.resourceCache[serverID]; ok {
		cached.AvailableGPU += gpuCount
		cached.LastUpdate = time.Now()
	}
}

func (s *ResourceService) validateAllocationRequest(req *AllocationRequest) error {
	if req.GPUCount <= 0 {
		return fmt.Errorf("%w: GPU count must be positive", ErrInvalidResourceReq)
	}
	if req.GPUCount > s.cfg.GPU.MaxGPUPerInstance {
		return fmt.Errorf("%w: GPU count exceeds maximum", ErrInvalidResourceReq)
	}
	if req.CPUCores < s.cfg.GPU.DefaultCPUCores || req.CPUCores > s.cfg.GPU.MaxCPUCores {
		return fmt.Errorf("%w: CPU cores out of range", ErrInvalidResourceReq)
	}
	if req.MemoryMB < s.cfg.GPU.DefaultMemoryMB || req.MemoryMB > s.cfg.GPU.MaxMemoryMB {
		return fmt.Errorf("%w: memory out of range", ErrInvalidResourceReq)
	}
	if req.StorageGB < s.cfg.GPU.DefaultStorageGB || req.StorageGB > s.cfg.GPU.MaxStorageGB {
		return fmt.Errorf("%w: storage out of range", ErrInvalidResourceReq)
	}
	return nil
}

func (s *ResourceService) filterCandidates(servers []models.GPUServer, req *AllocationRequest) []*models.GPUServer {
	var candidates []*models.GPUServer

	for _, server := range servers {
		if server.AvailableGPUCount() < req.GPUCount {
			continue
		}

		if req.GPUModel != "" && !s.serverHasGPUModel(server, req.GPUModel) {
			continue
		}

		if server.Specs.CPUCores < req.CPUCores {
			continue
		}

		if server.Specs.MemoryMB < req.MemoryMB {
			continue
		}

		candidates = append(candidates, &server)
	}

	return candidates
}

func (s *ResourceService) serverHasGPUModel(server models.GPUServer, model string) bool {
	for _, device := range server.GPUDevices {
		if device.Model == model && device.Status == models.GPUStatusAvailable {
			return true
		}
	}
	return false
}

func (s *ResourceService) allocateGPUDevices(server *models.GPUServer, count int) []models.GPUDevice {
	var devices []models.GPUDevice

	for i := range server.GPUDevices {
		device := &server.GPUDevices[i]
		if device.Status == models.GPUStatusAvailable && len(devices) < count {
			devices = append(devices, *device)
			device.Status = models.GPUStatusReserved
		}
	}

	return devices
}

func (s *ResourceService) cacheRefreshLoop() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		s.refreshCache()
	}
}

func (s *ResourceService) refreshCache() {
	s.mu.Lock()
	defer s.mu.Unlock()

	for id, cached := range s.resourceCache {
		if time.Since(cached.LastUpdate) > 5*time.Minute {
			delete(s.resourceCache, id)
		}
	}
}
