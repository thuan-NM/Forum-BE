package models

import (
	"gorm.io/gorm"
	"time"
)

type TopicStatus string

const (
	TopicStatusApproved TopicStatus = "approved"
	TopicStatusPending  TopicStatus = "pending"
	TopicStatusRejected TopicStatus = "rejected"
)

type Topic struct {
	ID          uint           `gorm:"primaryKey" json:"id"`
	Name        string         `gorm:"not null;uniqueIndex;size:255" json:"name"`
	Description string         `json:"description"`
	Status      TopicStatus    `gorm:"type:ENUM('approved','pending','rejected');default:'pending'" json:"status"`
	CreatedBy   uint           `gorm:"not null;index" json:"created_by"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`

	// Relationships
	Questions []Question `gorm:"many2many:question_topics;" json:"questions,omitempty"`
}
