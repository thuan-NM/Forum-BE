package models

import (
	"time"

	"gorm.io/gorm"
)

type PostStatus string

const (
	Approved PostStatus = "approved"
	Pending  PostStatus = "pending"
	Rejected PostStatus = "rejected"
)

type Post struct {
	PostID    uint           `gorm:"primaryKey"`
	Content   string         `gorm:"type:text" json:"content"`
	UserID    uint           `gorm:"not null;index" json:"user_id"`
	GroupID   uint           `gorm:"not null;index" json:"group_id"`
	Status    PostStatus     `gorm:"type:ENUM('approved','pending','rejected');default:'pending'" json:"status"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	User  User  `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Group Group `gorm:"foreignKey:GroupID" json:"group,omitempty"`
}
