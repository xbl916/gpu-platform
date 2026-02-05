package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ContainerInstance struct {
	ID          uuid.UUID          `json:"id" gorm:"type:uuid;primaryKey"`
	Name        string             `json:"name" gorm:"size:255;index"`
	UserID      uuid.UUID          `json:"userId" gorm:"type:uuid;index"`
	ProjectID   *uuid.UUID         `json:"projectId" gorm:"type:uuid;index"`
	TemplateID  *uuid.UUID         `json:"templateId" gorm:"type:uuid"`
	GPUServerID *uuid.UUID         `json:"gpuServerId" gorm:"type:uuid;index"`
	Status      InstanceStatus     `json:"status" gorm:"size:20;default:'pending'"`
	Resources   ContainerResources `json:"resources" gorm:"type:jsonb"`
	AccessToken string             `json:"accessToken" gorm:"size:255"`
	SSHPort     int                `json:"sshPort"`
	WebPort     int                `json:"webPort"`
	ContainerID string             `json:"containerId" gorm:"size:100"`
	PodName     string             `json:"podName" gorm:"size:255"`
	IPAddress   string             `json:"ipAddress" gorm:"size:45"`
	StartedAt   *time.Time         `json:"startedAt"`
	StoppedAt   *time.Time         `json:"stoppedAt"`
	ExpiresAt   *time.Time         `json:"expiresAt"`
	CreatedAt   time.Time          `json:"createdAt"`
	UpdatedAt   time.Time          `json:"updatedAt"`
	DeletedAt   gorm.DeletedAt     `json:"-" gorm:"index"`

	User      *User              `json:"user,omitempty" gorm:"foreignKey:UserID"`
	Project   *Project           `json:"project,omitempty" gorm:"foreignKey:ProjectID"`
	Template  *ContainerTemplate `json:"template,omitempty" gorm:"foreignKey:TemplateID"`
	GPUServer *GPUServer         `json:"gpuServer,omitempty" gorm:"foreignKey:GPUServerID"`
	Metrics   *ContainerMetrics  `json:"metrics,omitempty" gorm:"foreignKey:InstanceID"`
}

type InstanceStatus string

const (
	InstanceStatusPending  InstanceStatus = "pending"
	InstanceStatusCreating InstanceStatus = "creating"
	InstanceStatusRunning  InstanceStatus = "running"
	InstanceStatusStopping InstanceStatus = "stopping"
	InstanceStatusStopped  InstanceStatus = "stopped"
	InstanceStatusDeleted  InstanceStatus = "deleted"
	InstanceStatusError    InstanceStatus = "error"
)

type ContainerResources struct {
	GPUCount    int               `json:"gpuCount"`
	GPUModel    string            `json:"gpuModel"`
	GPUIndices  []int             `json:"gpuIndices"`
	CPUCores    int               `json:"cpuCores"`
	MemoryMB    int               `json:"memoryMb"`
	StorageGB   int               `json:"storageGb"`
	Image       string            `json:"image"`
	Command     string            `json:"command"`
	Environment map[string]string `json:"environment"`
	Ports       []int             `json:"ports"`
	Volumes     []VolumeMount     `json:"volumes"`
}

type VolumeMount struct {
	Name      string `json:"name"`
	MountPath string `json:"mountPath"`
	SizeGB    int    `json:"sizeGb"`
}

type ContainerMetrics struct {
	ID              uuid.UUID `json:"id" gorm:"type:uuid;primaryKey"`
	InstanceID      uuid.UUID `json:"instanceId" gorm:"type:uuid;uniqueIndex"`
	CPUUsage        float64   `json:"cpuUsage"`
	MemoryUsageMB   int       `json:"memoryUsageMb"`
	GPUUtilization  int       `json:"gpuUtilization"`
	GPUMemoryUsedMB int       `json:"gpuMemoryUsedMb"`
	NetworkInBytes  int64     `json:"networkInBytes"`
	NetworkOutBytes int64     `json:"networkOutBytes"`
	DiskReadBytes   int64     `json:"diskReadBytes"`
	DiskWriteBytes  int64     `json:"diskWriteBytes"`
	RecordedAt      time.Time `json:"recordedAt"`
	CreatedAt       time.Time `json:"createdAt"`

	Instance *ContainerInstance `json:"instance,omitempty" gorm:"foreignKey:InstanceID"`
}

func (c *ContainerInstance) BeforeCreate(tx *gorm.DB) error {
	if c.ID == uuid.Nil {
		c.ID = uuid.New()
	}
	if c.AccessToken == "" {
		c.AccessToken = uuid.New().String()
	}
	return nil
}

func (c *ContainerInstance) IsRunning() bool {
	return c.Status == InstanceStatusRunning
}

func (c *ContainerInstance) IsActive() bool {
	return c.Status == InstanceStatusRunning || c.Status == InstanceStatusPending
}
