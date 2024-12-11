package models

import (
	"gorm.io/gorm"
	"time"
)

type Permission struct {
	ID        uint   `gorm:"primaryKey"`
	Role      Role   `gorm:"type:ENUM('root','admin','employee','user');not null;index"`
	Resource  string `gorm:"not null;index"` // Ví dụ: "user", "question", "answer", "comment", "tag"
	Action    string `gorm:"not null;index"` // Ví dụ: "create", "edit", "delete"
	Allowed   bool   `gorm:"not null"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`
}
