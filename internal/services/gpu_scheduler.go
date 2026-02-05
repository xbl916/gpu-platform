package services

import (
	"sort"

	"gpu-platform/internal/config"
	"gpu-platform/internal/models"
)

type GPUScheduler struct {
	cfg *config.Config
}

func NewGPUScheduler() *GPUScheduler {
	return &GPUScheduler{}
}

func (s *GPUScheduler) SelectServer(candidates []*models.GPUServer, req *AllocationRequest) *models.GPUServer {
	if len(candidates) == 0 {
		return nil
	}

	policy := models.SchedulingPolicy(s.cfg.GPU.DefaultGPUModel)
	if policy == "" {
		policy = models.PolicyBinPack
	}

	switch policy {
	case models.PolicyBinPack:
		return s.selectBinPack(candidates, req)
	case models.PolicySpread:
		return s.selectSpread(candidates, req)
	case models.PolicyGPUCount:
		return s.selectByGPUCount(candidates, req)
	case models.PolicyMemory:
		return s.selectByMemory(candidates, req)
	default:
		return s.selectBinPack(candidates, req)
	}
}

func (s *GPUScheduler) selectBinPack(candidates []*models.GPUServer, req *AllocationRequest) *models.GPUServer {
	sort.Slice(candidates, func(i, j int) bool {
		serverA := candidates[i]
		serverB := candidates[j]

		gpuA := serverA.AvailableGPUCount()
		gpuB := serverB.AvailableGPUCount()

		if gpuA != gpuB {
			return gpuA < gpuB
		}

		if serverA.Specs.MemoryMB != serverB.Specs.MemoryMB {
			return serverA.Specs.MemoryMB < serverB.Specs.MemoryMB
		}

		return serverA.Specs.CPUCores < serverB.Specs.CPUCores
	})

	for _, server := range candidates {
		if server.AvailableGPUCount() >= req.GPUCount {
			return server
		}
	}

	return candidates[0]
}

func (s *GPUScheduler) selectSpread(candidates []*models.GPUServer, req *AllocationRequest) *models.GPUServer {
	sort.Slice(candidates, func(i, j int) bool {
		serverA := candidates[i]
		serverB := candidates[j]

		gpuA := serverA.AvailableGPUCount()
		gpuB := serverB.AvailableGPUCount()

		if gpuA != gpuB {
			return gpuA > gpuB
		}

		if serverA.Specs.MemoryMB != serverB.Specs.MemoryMB {
			return serverA.Specs.MemoryMB > serverB.Specs.MemoryMB
		}

		return serverA.Specs.CPUCores > serverB.Specs.CPUCores
	})

	for _, server := range candidates {
		if server.AvailableGPUCount() >= req.GPUCount {
			return server
		}
	}

	return candidates[0]
}

func (s *GPUScheduler) selectByGPUCount(candidates []*models.GPUServer, req *AllocationRequest) *models.GPUServer {
	sort.Slice(candidates, func(i, j int) bool {
		serverA := candidates[i]
		serverB := candidates[j]

		totalGPUA := len(serverA.GPUDevices)
		totalGPUB := len(serverB.GPUDevices)

		if totalGPUA != totalGPUB {
			return totalGPUA < totalGPUB
		}

		return serverA.AvailableGPUCount() < serverB.AvailableGPUCount()
	})

	for _, server := range candidates {
		if server.AvailableGPUCount() >= req.GPUCount {
			return server
		}
	}

	return candidates[0]
}

func (s *GPUScheduler) selectByMemory(candidates []*models.GPUServer, req *AllocationRequest) *models.GPUServer {
	sort.Slice(candidates, func(i, j int) bool {
		serverA := candidates[i]
		serverB := candidates[j]

		availableMemA := serverA.Specs.MemoryMB
		availableMemB := serverB.Specs.MemoryMB

		if availableMemA != availableMemB {
			return availableMemA < availableMemB
		}

		return serverA.AvailableGPUCount() < serverB.AvailableGPUCount()
	})

	for _, server := range candidates {
		if server.AvailableGPUCount() >= req.GPUCount {
			return server
		}
	}

	return candidates[0]
}

func (s *GPUScheduler) ScoreServer(server *models.GPUServer, req *AllocationRequest) float64 {
	var score float64

	gpuUtilization := float64(len(server.GPUDevices)-server.AvailableGPUCount()) / float64(len(server.GPUDevices))
	memoryUtilization := float64(req.MemoryMB) / float64(server.Specs.MemoryMB)
	cpuUtilization := float64(req.CPUCores) / float64(server.Specs.CPUCores)

	score += (1 - gpuUtilization) * 0.4
	score += (1 - memoryUtilization) * 0.3
	score += (1 - cpuUtilization) * 0.3

	return score
}
