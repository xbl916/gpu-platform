package monitor

import (
	"context"
	"fmt"
	"time"

	"gpu-platform/internal/config"
	"gpu-platform/internal/models"
	"gpu-platform/internal/repository"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

type MetricsCollector struct {
	cfg           *config.Config
	containerRepo *repository.ContainerRepository
	gpuServerRepo *repository.GPUServerRepository

	metrics *PlatformMetrics
}

type PlatformMetrics struct {
	TotalUsers        prometheus.Counter
	ActiveUsers       prometheus.Gauge
	TotalContainers   prometheus.Counter
	RunningContainers prometheus.Gauge
	TotalGPUs         prometheus.Gauge
	AvailableGPUs     prometheus.Gauge
	CPUUsage          prometheus.Gauge
	MemoryUsage       prometheus.Gauge
}

func NewMetricsCollector(cfg *config.Config, containerRepo *repository.ContainerRepository, gpuServerRepo *repository.GPUServerRepository) *MetricsCollector {
	return &MetricsCollector{
		cfg:           cfg,
		containerRepo: containerRepo,
		gpuServerRepo: gpuServerRepo,
		metrics: &PlatformMetrics{
			TotalUsers: promauto.NewCounter(prometheus.CounterOpts{
				Name: "gpu_platform_total_users",
				Help: "Total number of registered users",
			}),
			ActiveUsers: promauto.NewGauge(prometheus.GaugeOpts{
				Name: "gpu_platform_active_users",
				Help: "Number of active users",
			}),
			TotalContainers: promauto.NewCounter(prometheus.CounterOpts{
				Name: "gpu_platform_total_containers",
				Help: "Total number of created containers",
			}),
			RunningContainers: promauto.NewGauge(prometheus.GaugeOpts{
				Name: "gpu_platform_running_containers",
				Help: "Number of currently running containers",
			}),
			TotalGPUs: promauto.NewGauge(prometheus.GaugeOpts{
				Name: "gpu_platform_total_gpus",
				Help: "Total number of GPUs",
			}),
			AvailableGPUs: promauto.NewGauge(prometheus.GaugeOpts{
				Name: "gpu_platform_available_gpus",
				Help: "Number of available GPUs",
			}),
			CPUUsage: promauto.NewGauge(prometheus.GaugeOpts{
				Name: "gpu_platform_cpu_usage",
				Help: "Average CPU usage across all nodes",
			}),
			MemoryUsage: promauto.NewGauge(prometheus.GaugeOpts{
				Name: "gpu_platform_memory_usage",
				Help: "Average memory usage across all nodes",
			}),
		},
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

	total := 0
	available := 0

	for _, server := range servers {
		for _, device := range server.GPUDevices {
			total++
			if device.Status == models.GPUStatusAvailable {
				available++
			}
		}
	}

	m.metrics.TotalGPUs.Set(float64(total))
	m.metrics.AvailableGPUs.Set(float64(available))
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

	m.metrics.RunningContainers.Set(float64(running))
}

type AlertManager struct {
	cfg      *config.Config
	rules    []AlertRule
	notifier *Notifier
}

type AlertRule struct {
	Name      string
	Condition string
	Threshold int
	Labels    map[string]string
	Severity  string
}

type Alert struct {
	RuleName string
	Severity string
	Labels   map[string]string
	StartsAt time.Time
	EndsAt   *time.Time
	Status   string
}

type Notifier struct {
	cfg *config.Config
}

func NewAlertManager(cfg *config.Config) *AlertManager {
	return &AlertManager{
		cfg: cfg,
		rules: []AlertRule{
			{
				Name:      "GPUHighTemperature",
				Condition: "temperature > 85",
				Threshold: 85,
				Labels:    map[string]string{"component": "gpu"},
				Severity:  "warning",
			},
			{
				Name:      "GPUCriticalTemperature",
				Condition: "temperature > 90",
				Threshold: 90,
				Labels:    map[string]string{"component": "gpu"},
				Severity:  "critical",
			},
			{
				Name:      "GPUMemoryHigh",
				Condition: "memory_usage > 90",
				Threshold: 90,
				Labels:    map[string]string{"component": "gpu"},
				Severity:  "warning",
			},
			{
				Name:      "ContainerDown",
				Condition: "status == error",
				Threshold: 0,
				Labels:    map[string]string{"component": "container"},
				Severity:  "critical",
			},
		},
		notifier: &Notifier{cfg: cfg},
	}
}

func (a *AlertManager) Evaluate(device *models.GPUDevice) []Alert {
	var alerts []Alert

	if device.Temperature > a.rules[0].Threshold {
		alert := Alert{
			RuleName: a.rules[0].Name,
			Severity: a.rules[0].Severity,
			Labels:   a.rules[0].Labels,
			StartsAt: time.Now(),
			Status:   "firing",
		}
		alerts = append(alerts, alert)
	}

	return alerts
}

func (a *AlertManager) SendAlert(alert Alert) error {
	fmt.Printf("Alert: %s - Severity: %s\n", alert.RuleName, alert.Severity)
	return nil
}

type DashboardService struct {
	cfg       *config.Config
	collector *MetricsCollector
}

type DashboardData struct {
	Overview     OverviewStats     `json:"overview"`
	GPUResources []GPUResourceStat `json:"gpuResources"`
	Containers   ContainerStats    `json:"containers"`
	RecentAlerts []Alert           `json:"recentAlerts"`
}

type OverviewStats struct {
	TotalUsers        int `json:"totalUsers"`
	ActiveUsers       int `json:"activeUsers"`
	TotalContainers   int `json:"totalContainers"`
	RunningContainers int `json:"runningContainers"`
	TotalGPUs         int `json:"totalGpus"`
	AvailableGPUs     int `json:"availableGpus"`
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

func NewDashboardService(cfg *config.Config, collector *MetricsCollector) *DashboardService {
	return &DashboardService{
		cfg:       cfg,
		collector: collector,
	}
}

func (s *DashboardService) GetDashboard(ctx context.Context) (*DashboardData, error) {
	return &DashboardData{
		Overview: OverviewStats{
			TotalUsers:        100,
			ActiveUsers:       45,
			TotalContainers:   200,
			RunningContainers: 120,
			TotalGPUs:         200,
			AvailableGPUs:     80,
		},
		GPUResources: []GPUResourceStat{
			{PoolName: "A100 Pool", TotalGPUs: 100, AvailableGPUs: 40, Utilization: 0.6},
			{PoolName: "RTX3090 Pool", TotalGPUs: 50, AvailableGPUs: 20, Utilization: 0.6},
			{PoolName: "V100 Pool", TotalGPUs: 50, AvailableGPUs: 20, Utilization: 0.6},
		},
		Containers: ContainerStats{
			Total:   200,
			Running: 120,
			Pending: 30,
			Stopped: 40,
			Failed:  10,
		},
		RecentAlerts: []Alert{
			{
				RuleName: "GPUHighTemperature",
				Severity: "warning",
				Labels:   map[string]string{"gpu": "gpu-01", "server": "server-01"},
				StartsAt: time.Now().Add(-30 * time.Minute),
				Status:   "firing",
			},
			{
				RuleName: "GPUMemoryHigh",
				Severity: "warning",
				Labels:   map[string]string{"gpu": "gpu-02", "server": "server-02"},
				StartsAt: time.Now().Add(-15 * time.Minute),
				Status:   "firing",
			},
		},
	}, nil
}
