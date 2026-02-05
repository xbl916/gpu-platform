package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type GPUPool struct {
	ID               uuid.UUID      `json:"id" gorm:"type:uuid;primaryKey"`
	Name             string         `json:"name" gorm:"size:100;uniqueIndex"`
	Description      string         `json:"description" gorm:"size:500"`
	SchedulingPolicy string         `json:"schedulingPolicy" gorm:"size:50;default:'binpack'"`
	QuotaConfig      QuotaConfig    `json:"quotaConfig" gorm:"type:jsonb"`
	Status           string         `json:"status" gorm:"size:20;default:'active'"`
	CreatedAt        time.Time      `json:"createdAt"`
	UpdatedAt        time.Time      `json:"updatedAt"`
	DeletedAt        gorm.DeletedAt `json:"-" gorm:"index"`

	Servers []GPUServer `json:"servers,omitempty" gorm:"foreignKey:GPUPoolID"`
}

type SchedulingPolicy string

const (
	PolicyBinPack  SchedulingPolicy = "binpack"
	PolicySpread   SchedulingPolicy = "spread"
	PolicyGPUCount SchedulingPolicy = "gpu_count"
	PolicyMemory   SchedulingPolicy = "memory"
)

func (p *GPUPool) BeforeCreate(tx *gorm.DB) error {
	if p.ID == uuid.Nil {
		p.ID = uuid.New()
	}
	return nil
}

func (p *GPUPool) IsActive() bool {
	return p.Status == "active"
}
