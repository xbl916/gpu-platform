package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type User struct {
	ID           uuid.UUID      `json:"id" gorm:"type:uuid;primaryKey"`
	Email        string         `json:"email" gorm:"uniqueIndex:idx_user_email;size:255"`
	PasswordHash string         `json:"-" gorm:"size:255"`
	Name         string         `json:"name" gorm:"size:100"`
	Role         string         `json:"role" gorm:"size:20;default:'developer'"`
	Status       string         `json:"status" gorm:"size:20;default:'active'"`
	ProjectID    *uuid.UUID     `json:"projectId" gorm:"type:uuid;index"`
	LastLoginAt  *time.Time     `json:"lastLoginAt"`
	CreatedAt    time.Time      `json:"createdAt"`
	UpdatedAt    time.Time      `json:"updatedAt"`
	DeletedAt    gorm.DeletedAt `json:"-" gorm:"index"`

	Project    *Project            `json:"project,omitempty" gorm:"foreignKey:ProjectID"`
	Containers []ContainerInstance `json:"containers,omitempty" gorm:"foreignKey:UserID"`
	Templates  []ContainerTemplate `json:"templates,omitempty" gorm:"foreignKey:UserID"`
}

type UserRole string

const (
	RoleAdmin     UserRole = "admin"
	RoleDeveloper UserRole = "developer"
	RoleGuest     UserRole = "guest"
)

type UserStatus string

const (
	UserStatusActive   UserStatus = "active"
	UserStatusInactive UserStatus = "inactive"
	UserStatusLocked   UserStatus = "locked"
)

func (u *User) BeforeCreate(tx *gorm.DB) error {
	if u.ID == uuid.Nil {
		u.ID = uuid.New()
	}
	return nil
}

func (u *User) IsAdmin() bool {
	return u.Role == string(RoleAdmin)
}

func (u *User) IsActive() bool {
	return u.Status == string(UserStatusActive)
}
