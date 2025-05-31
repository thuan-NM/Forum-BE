package models

import (
	"gorm.io/gorm"
	"time"
)

type Permission struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	Role      Role           `gorm:"type:ENUM('root','admin','employee','user');not null;index" json:"role"`
	Resource  string         `gorm:"not null;index" json:"resource"`
	Action    string         `gorm:"not null;index" json:"action"`
	Allowed   bool           `gorm:"not null" json:"allowed"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	// Composite index
	_ struct{} `gorm:"uniqueIndex:idx_role_resource_action,where:role != '' AND resource != '' AND action != ''"`
}
