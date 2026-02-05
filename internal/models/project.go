package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Project struct {
	ID          uuid.UUID      `json:"id" gorm:"type:uuid;primaryKey"`
	Name        string         `json:"name" gorm:"size:100;uniqueIndex:idx_project_name"`
	Description string         `json:"description" gorm:"size:1000"`
	OwnerID     string         `json:"ownerId" gorm:"size:36"`
	QuotaConfig QuotaConfig    `json:"quotaConfig" gorm:"type:jsonb"`
	Status      string         `json:"status" gorm:"size:20;default:'active'"`
	CreatedAt   time.Time      `json:"createdAt"`
	UpdatedAt   time.Time      `json:"updatedAt"`
	DeletedAt   gorm.DeletedAt `json:"-" gorm:"index"`

	Members    []ProjectMember     `json:"members,omitempty" gorm:"foreignKey:ProjectID"`
	Containers []ContainerInstance `json:"containers,omitempty" gorm:"foreignKey:ProjectID"`
	Templates  []ContainerTemplate `json:"templates,omitempty" gorm:"foreignKey:ProjectID"`
}

type ProjectStatus string

const (
	ProjectStatusActive   ProjectStatus = "active"
	ProjectStatusInactive ProjectStatus = "inactive"
	ProjectStatusArchived ProjectStatus = "archived"
)

type QuotaConfig struct {
	MaxGPUCount       int `json:"maxGpuCount"`
	MaxGPUInstances   int `json:"maxGpuInstances"`
	MaxTotalMemoryMB  int `json:"maxTotalMemoryMb"`
	MaxTotalStorageGB int `json:"maxTotalStorageGb"`
	MaxCPUCores       int `json:"maxCpuCores"`
}

func (p *Project) BeforeCreate(tx *gorm.DB) error {
	if p.ID == uuid.Nil {
		p.ID = uuid.New()
	}
	return nil
}
