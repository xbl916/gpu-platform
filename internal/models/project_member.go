package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ProjectMember struct {
	ID        uuid.UUID      `json:"id" gorm:"type:uuid;primaryKey"`
	UserID    uuid.UUID      `json:"userId" gorm:"type:uuid;index"`
	ProjectID uuid.UUID      `json:"projectId" gorm:"type:uuid;index"`
	Role      string         `json:"role" gorm:"size:20;default:'member'"`
	JoinedAt  time.Time      `json:"joinedAt"`
	CreatedAt time.Time      `json:"createdAt"`
	UpdatedAt time.Time      `json:"updatedAt"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`

	User    *User    `json:"user,omitempty" gorm:"foreignKey:UserID"`
	Project *Project `json:"project,omitempty" gorm:"foreignKey:ProjectID"`
}

type MemberRole string

const (
	MemberRoleOwner  MemberRole = "owner"
	MemberRoleAdmin  MemberRole = "admin"
	MemberRoleMember MemberRole = "member"
	MemberRoleGuest  MemberRole = "guest"
)

func (pm *ProjectMember) BeforeCreate(tx *gorm.DB) error {
	if pm.ID == uuid.Nil {
		pm.ID = uuid.New()
	}
	return nil
}
