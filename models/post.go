package models

import (
	"encoding/json"
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
	ID           uint            `gorm:"primaryKey" json:"id"`
	Title        string          `gorm:"type:varchar(255);index" json:"title,omitempty"`
	Content      string          `gorm:"type:text" json:"content"`
	PlainContent string          `gorm:"type:text"`
	UserID       uint            `gorm:"not null;index" json:"user_id"`
	Status       PostStatus      `gorm:"type:ENUM('approved','pending','rejected');default:'pending'" json:"status"`
	Metadata     json.RawMessage `gorm:"type:json" json:"metadata,omitempty"`
	CreatedAt    time.Time       `json:"created_at"`
	UpdatedAt    time.Time       `json:"updated_at"`
	DeletedAt    gorm.DeletedAt  `gorm:"index" json:"-"`

	User          User           `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Notifications []Notification `json:"notifications,omitempty" gorm:"polymorphic:Entity;"`
	Comments      []Comment      `json:"comments,omitempty" gorm:"foreignKey:PostID"`
	Reactions     []Reaction     `json:"reactions,omitempty" gorm:"foreignKey:PostID"`
	Tags          []Tag          `json:"tags,omitempty" gorm:"many2many:post_tags;"`
}
