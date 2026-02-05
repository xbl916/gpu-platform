package monitor

import (
	"context"
	"fmt"
	"time"

	"gpu-platform/internal/config"
	"gpu-platform/internal/models"
	"gpu-platform/internal/repository"
)

type MetricsCollector struct {
	cfg           *config.Config
	containerRepo *repository.ContainerRepository
	gpuServerRepo *repository.GPUServerRepository
}

type PlatformMetrics struct{}

func NewMetricsCollector(cfg *config.Config, containerRepo *repository.ContainerRepository, gpuServerRepo *repository.GPUServerRepository) *MetricsCollector {
	return &MetricsCollector{
		cfg:           cfg,
		containerRepo: containerRepo,
		gpuServerRepo: gpuServerRepo,
	}
}

func (m *MetricsCollector) Collect(ctx context.Context) error {
	m.collectGPUStats(ctx)
	m.collectContainerStats(ctx)
	return nil
}

func (m *MetricsCollector) collectGPUStats(ctx context.Context) {
	servers, err := m.gpuServerRepo.ListOnline(ctx)
	if err != nil {
		return
	}

	for _, server := range servers {
		for _, device := range server.GPUDevices {
			if device.Temperature > 85 {
				fmt.Printf("Warning: GPU %s temperature is %d°C\n", device.ID, device.Temperature)
			}
		}
	}
}

func (m *MetricsCollector) collectContainerStats(ctx context.Context) {
	containers, _, err := m.containerRepo.List(ctx, 1000, 0, nil)
	if err != nil {
		return
	}

	running := 0
	for _, container := range containers {
		if container.Status == models.InstanceStatusRunning {
			running++
		}
	}
	fmt.Printf("Running containers: %d\n", running)
}

type AlertManager struct {
	cfg      *config.Config
	rules    []AlertRule
	notifier *Notifier
}

type AlertRule struct {
	ID        string
	Name      string
	Condition string
	Threshold float64
	Labels    map[string]string
	Severity  string
	Duration  time.Duration
}

type Alert struct {
	ID       string            `json:"id"`
	RuleName string            `json:"ruleName"`
	Severity string            `json:"severity"`
	Labels   map[string]string `json:"labels"`
	StartsAt time.Time         `json:"startsAt"`
	EndsAt   *time.Time        `json:"endsAt"`
	Status   string            `json:"status"`
	Message  string            `json:"message"`
}

type Notifier struct {
	cfg *config.Config
}

func NewAlertManager(cfg *config.Config) *AlertManager {
	return &AlertManager{
		cfg: cfg,
		rules: []AlertRule{
			{ID: "gpu-high-temp", Name: "GPUHighTemperature", Condition: "temperature > 85", Threshold: 85, Labels: map[string]string{"component": "gpu"}, Severity: "warning", Duration: 5 * time.Minute},
			{ID: "gpu-critical-temp", Name: "GPUCriticalTemperature", Condition: "temperature > 90", Threshold: 90, Labels: map[string]string{"component": "gpu"}, Severity: "critical", Duration: 1 * time.Minute},
			{ID: "container-down", Name: "ContainerDown", Condition: "status == error", Threshold: 0, Labels: map[string]string{"component": "container"}, Severity: "critical", Duration: 5 * time.Minute},
		},
		notifier: &Notifier{cfg: cfg},
	}
}

func (a *AlertManager) SendAlert(alert Alert) error {
	fmt.Printf("Alert: [%s] %s - %s\n", alert.Severity, alert.RuleName, alert.Message)
	return nil
}

func (a *AlertManager) GetActiveAlerts(ctx context.Context) ([]Alert, error) {
	return []Alert{
		{ID: "alert-1", RuleName: "GPUHighTemperature", Severity: "warning", Labels: map[string]string{"gpu": "gpu-01"}, StartsAt: time.Now().Add(-30 * time.Minute), Status: "firing", Message: "GPU temperature is 86°C"},
	}, nil
}

func (a *AlertManager) AcknowledgeAlert(alertID string, userID string) error {
	return nil
}

type DashboardService struct {
	cfg *config.Config
}

type DashboardData struct {
	Overview     OverviewStats     `json:"overview"`
	GPUResources []GPUResourceStat `json:"gpuResources"`
	Containers   ContainerStats    `json:"containers"`
	RecentAlerts []Alert           `json:"recentAlerts"`
}

type OverviewStats struct {
	TotalUsers        int     `json:"totalUsers"`
	ActiveUsers       int     `json:"activeUsers"`
	TotalContainers   int     `json:"totalContainers"`
	RunningContainers int     `json:"runningContainers"`
	TotalGPUs         int     `json:"totalGpus"`
	AvailableGPUs     int     `json:"availableGpus"`
	CPUUsage          float64 `json:"cpuUsage"`
	MemoryUsage       float64 `json:"memoryUsage"`
}

type GPUResourceStat struct {
	PoolName      string  `json:"poolName"`
	TotalGPUs     int     `json:"totalGpus"`
	AvailableGPUs int     `json:"availableGpus"`
	Utilization   float64 `json:"utilization"`
}

type ContainerStats struct {
	Total   int `json:"total"`
	Running int `json:"running"`
	Pending int `json:"pending"`
	Stopped int `json:"stopped"`
	Failed  int `json:"failed"`
}

func NewDashboardService(cfg *config.Config) *DashboardService {
	return &DashboardService{cfg: cfg}
}

func (s *DashboardService) GetDashboard(ctx context.Context) (*DashboardData, error) {
	return &DashboardData{
		Overview: OverviewStats{TotalUsers: 100, ActiveUsers: 45, TotalContainers: 200, RunningContainers: 120, TotalGPUs: 200, AvailableGPUs: 80, CPUUsage: 45.5, MemoryUsage: 62.3},
		GPUResources: []GPUResourceStat{
			{PoolName: "A100 Pool", TotalGPUs: 100, AvailableGPUs: 40, Utilization: 0.6},
			{PoolName: "RTX3090 Pool", TotalGPUs: 50, AvailableGPUs: 20, Utilization: 0.65},
		},
		Containers:   ContainerStats{Total: 200, Running: 120, Pending: 30, Stopped: 40, Failed: 10},
		RecentAlerts: []Alert{{ID: "alert-1", RuleName: "GPUHighTemperature", Severity: "warning", Labels: map[string]string{"gpu": "gpu-01"}, StartsAt: time.Now().Add(-30 * time.Minute), Status: "firing", Message: "GPU temperature is 86°C"}},
	}, nil
}

type MetricsService struct {
	cfg *config.Config
}

func NewMetricsService(cfg *config.Config) *MetricsService {
	return &MetricsService{cfg: cfg}
}

func (s *MetricsService) GetGPUMetrics(ctx context.Context) ([]GPUMetric, error) {
	return []GPUMetric{
		{GPUName: "NVIDIA A100-SXM4-40GB", Index: 0, Temp: 65, Power: 250, Util: 45, MemoryUsed: 32 * 1024, MemoryTotal: 40 * 1024},
	}, nil
}

type GPUMetric struct {
	GPUName     string `json:"gpuName"`
	Index       int    `json:"index"`
	Temp        int    `json:"temp"`
	Power       int    `json:"power"`
	Util        int    `json:"util"`
	MemoryUsed  int64  `json:"memoryUsed"`
	MemoryTotal int64  `json:"memoryTotal"`
}

type NodeMetrics struct {
	NodeName    string  `json:"nodeName"`
	CPUUsage    float64 `json:"cpuUsage"`
	MemoryUsage float64 `json:"memoryUsage"`
	DiskUsage   float64 `json:"diskUsage"`
	DiskIOPS    int     `json:"diskIOPS"`
}

func (s *MetricsService) GetNodeMetrics(ctx context.Context) ([]NodeMetrics, error) {
	return []NodeMetrics{
		{NodeName: "gpu-node-01", CPUUsage: 45.5, MemoryUsage: 62.3, DiskUsage: 35.0, DiskIOPS: 1500},
	}, nil
}
