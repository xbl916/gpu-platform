package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type GPUServer struct {
	ID            uuid.UUID      `json:"id" gorm:"type:uuid;primaryKey"`
	Hostname      string         `json:"hostname" gorm:"size:255;uniqueIndex"`
	IPAddress     string         `json:"ipAddress" gorm:"size:45"`
	Status        ServerStatus   `json:"status" gorm:"size:20;default:'offline'"`
	Specs         ServerSpecs    `json:"specs" gorm:"type:jsonb"`
	GPUPoolID     *uuid.UUID     `json:"gpuPoolId" gorm:"type:uuid;index"`
	LastHeartbeat time.Time      `json:"lastHeartbeat"`
	CreatedAt     time.Time      `json:"createdAt"`
	UpdatedAt     time.Time      `json:"updatedAt"`
	DeletedAt     gorm.DeletedAt `json:"-" gorm:"index"`

	GPUPool    *GPUPool            `json:"gpuPool,omitempty" gorm:"foreignKey:GPUPoolID"`
	GPUDevices []GPUDevice         `json:"gpuDevices,omitempty" gorm:"foreignKey:ServerID"`
	Containers []ContainerInstance `json:"containers,omitempty" gorm:"foreignKey:GPUServerID"`
}

type ServerStatus string

const (
	ServerStatusOnline   ServerStatus = "online"
	ServerStatusOffline  ServerStatus = "offline"
	ServerStatusMaintain ServerStatus = "maintain"
	ServerStatusError    ServerStatus = "error"
)

type ServerSpecs struct {
	CPUCores      int    `json:"cpuCores"`
	MemoryMB      int    `json:"memoryMb"`
	StorageGB     int    `json:"storageGb"`
	OSVersion     string `json:"osVersion"`
	DockerVersion string `json:"dockerVersion"`
	NVIDIADriver  string `json:"nvidiaDriver"`
}

func (s *GPUServer) BeforeCreate(tx *gorm.DB) error {
	if s.ID == uuid.Nil {
		s.ID = uuid.New()
	}
	return nil
}

func (s *GPUServer) IsOnline() bool {
	return s.Status == ServerStatusOnline
}

func (s *GPUServer) AvailableGPUCount() int {
	count := 0
	for _, device := range s.GPUDevices {
		if device.Status == GPUStatusAvailable {
			count++
		}
	}
	return count
}
