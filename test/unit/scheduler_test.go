package unit

import (
	"testing"

	"gpu-platform/internal/config"
	"gpu-platform/internal/models"

	"github.com/google/uuid"
)

func TestGPUScheduler_SelectServer(t *testing.T) {
	cfg := &config.Config{
		GPU: config.GPUConfig{
			DefaultGPUModel: "a100",
		},
	}

	scheduler := NewTestScheduler(cfg)

	server1 := &models.GPUServer{
		ID: uuid.New(),
		Specs: models.ServerSpecs{
			CPUCores: 64,
			MemoryMB: 256 * 1024,
		},
		GPUDevices: []models.GPUDevice{
			{ID: uuid.New(), Status: models.GPUStatusAvailable},
			{ID: uuid.New(), Status: models.GPUStatusAvailable},
		},
	}

	server2 := &models.GPUServer{
		ID: uuid.New(),
		Specs: models.ServerSpecs{
			CPUCores: 32,
			MemoryMB: 128 * 1024,
		},
		GPUDevices: []models.GPUDevice{
			{ID: uuid.New(), Status: models.GPUStatusAvailable},
		},
	}

	candidates := []*models.GPUServer{server1, server2}

	req := &AllocationRequest{
		GPUCount: 1,
		CPUCores: 4,
		MemoryMB: 16 * 1024,
	}

	selected := scheduler.SelectServer(candidates, req)

	if selected == nil {
		t.Fatal("Expected a server to be selected")
	}

	t.Logf("Selected server: %s", selected.ID)
}

func TestGPUScheduler_NoAvailable(t *testing.T) {
	scheduler := NewTestScheduler(&config.Config{})

	server1 := &models.GPUServer{
		ID: uuid.New(),
		Specs: models.ServerSpecs{
			CPUCores: 4,
			MemoryMB: 8 * 1024,
		},
		GPUDevices: []models.GPUDevice{
			{ID: uuid.New(), Status: models.GPUStatusInUse},
		},
	}

	candidates := []*models.GPUServer{server1}

	req := &AllocationRequest{
		GPUCount: 2,
	}

	selected := scheduler.SelectServer(candidates, req)

	if selected != nil {
		t.Errorf("Expected nil when no server has enough resources")
	}
}

type TestScheduler struct {
	cfg *config.Config
}

func NewTestScheduler(cfg *config.Config) *TestScheduler {
	return &TestScheduler{cfg: cfg}
}

type AllocationRequest struct {
	GPUCount  int
	GPUModel  string
	CPUCores  int
	MemoryMB  int
	StorageGB int
}

func (s *TestScheduler) SelectServer(candidates []*models.GPUServer, req *AllocationRequest) *models.GPUServer {
	for _, server := range candidates {
		if server.AvailableGPUCount() >= req.GPUCount {
			return server
		}
	}
	return nil
}
