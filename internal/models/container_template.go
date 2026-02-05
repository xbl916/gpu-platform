package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ContainerTemplate struct {
	ID          uuid.UUID      `json:"id" gorm:"type:uuid;primaryKey"`
	Name        string         `json:"name" gorm:"size:100;index"`
	Description string         `json:"description" gorm:"size:1000"`
	UserID      uuid.UUID      `json:"userId" gorm:"type:uuid;index"`
	ProjectID   *uuid.UUID     `json:"projectId" gorm:"type:uuid;index"`
	Config      TemplateConfig `json:"config" gorm:"type:jsonb"`
	DockerImage string         `json:"dockerImage" gorm:"size:500"`
	Dockerfile  string         `json:"dockerfile" gorm:"type:text"`
	Version     int            `json:"version" gorm:"default:1"`
	Status      string         `json:"status" gorm:"size:20;default:'draft'"`
	IsPublic    bool           `json:"isPublic" gorm:"default:false"`
	UsageCount  int            `json:"usageCount" gorm:"default:0"`
	CreatedAt   time.Time      `json:"createdAt"`
	UpdatedAt   time.Time      `json:"updatedAt"`
	DeletedAt   gorm.DeletedAt `json:"-" gorm:"index"`

	User      *User               `json:"user,omitempty" gorm:"foreignKey:UserID"`
	Project   *Project            `json:"project,omitempty" gorm:"foreignKey:ProjectID"`
	Instances []ContainerInstance `json:"instances,omitempty" gorm:"foreignKey:TemplateID"`
}

type TemplateStatus string

const (
	TemplateStatusDraft      TemplateStatus = "draft"
	TemplateStatusBuilding   TemplateStatus = "building"
	TemplateStatusReady      TemplateStatus = "ready"
	TemplateStatusFailed     TemplateStatus = "failed"
	TemplateStatusDeprecated TemplateStatus = "deprecated"
)

type TemplateConfig struct {
	BaseImage      string            `json:"baseImage"`
	CUDAVersion    string            `json:"cudaVersion"`
	CUDNNVersion   string            `json:"cudnnVersion"`
	PythonVersion  string            `json:"pythonVersion"`
	Packages       []string          `json:"packages"`
	PipPackages    []string          `json:"pipPackages"`
	CondaPackages  []string          `json:"condaPackages"`
	Environment    map[string]string `json:"environment"`
	Entrypoint     string            `json:"entrypoint"`
	WorkingDir     string            `json:"workingDir"`
	Labels         map[string]string `json:"labels"`
	PreInstallCmd  string            `json:"preInstallCmd"`
	PostInstallCmd string            `json:"postInstallCmd"`
}

func (t *ContainerTemplate) BeforeCreate(tx *gorm.DB) error {
	if t.ID == uuid.Nil {
		t.ID = uuid.New()
	}
	return nil
}

func (t *ContainerTemplate) IsReady() bool {
	return t.Status == string(TemplateStatusReady)
}
