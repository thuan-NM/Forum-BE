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
	ID             uint           `gorm:"primaryKey" json:"id"`
	Name           string         `gorm:"not null;uniqueIndex;size:255" json:"name"`
	Description    string         `json:"description"`
	CreatedAt      time.Time      `json:"created_at"`
	UpdatedAt      time.Time      `json:"updated_at"`
	DeletedAt      gorm.DeletedAt `gorm:"index" json:"-"`
	FollowersCount int            `gorm:"default:0" json:"followers_count"`

	Questions []Question `json:"questions,omitempty" gorm:"foreignKey:TopicID"`
	Followers []User     `gorm:"many2many:user_topics;" json:"followers,omitempty"`
}
