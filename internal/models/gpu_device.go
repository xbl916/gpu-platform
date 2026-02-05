package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type GPUDevice struct {
	ID            uuid.UUID      `json:"id" gorm:"type:uuid;primaryKey"`
	GPUID         string         `json:"gpuId" gorm:"size:100;index"`
	ServerID      uuid.UUID      `json:"serverId" gorm:"type:uuid;index"`
	Name          string         `json:"name" gorm:"size:100"`
	Model         string         `json:"model" gorm:"size:100"`
	MemoryMB      int            `json:"memoryMb"`
	ComputeCap    int            `json:"computeCap"`
	CUDAVersion   string         `json:"cudaVersion" gorm:"size:20"`
	DriverVersion string         `json:"driverVersion" gorm:"size:50"`
	Status        GPUStatus      `json:"status" gorm:"size:20;default:'available'"`
	Temperature   int            `json:"temperature"`
	PowerUsage    int            `json:"powerUsage"`
	MemoryUsedMB  int            `json:"memoryUsedMb"`
	Utilization   int            `json:"utilization"`
	LastSeenAt    time.Time      `json:"lastSeenAt"`
	CreatedAt     time.Time      `json:"createdAt"`
	UpdatedAt     time.Time      `json:"updatedAt"`
	DeletedAt     gorm.DeletedAt `json:"-" gorm:"index"`

	Server *GPUServer `json:"server,omitempty" gorm:"foreignKey:ServerID"`
}

type GPUStatus string

const (
	GPUStatusAvailable   GPUStatus = "available"
	GPUStatusInUse       GPUStatus = "in_use"
	GPUStatusReserved    GPUStatus = "reserved"
	GPUStatusError       GPUStatus = "error"
	GPUStatusMaintenance GPUStatus = "maintenance"
)

func (g *GPUDevice) BeforeCreate(tx *gorm.DB) error {
	if g.ID == uuid.Nil {
		g.ID = uuid.New()
	}
	return nil
}

func (g *GPUDevice) IsAvailable() bool {
	return g.Status == GPUStatusAvailable
}
