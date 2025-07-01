package models

import (
	"time"

	"gorm.io/gorm"
)

type Follow struct {
	ID             uint           `gorm:"primaryKey" json:"id"`
	UserID         uint           `gorm:"not null;index" json:"user_id"`
	FollowableID   uint           `gorm:"not null;index:idx_followable" json:"followable_id"`
	FollowableType string         `gorm:"not null;index:idx_followable" json:"followable_type"` // e.g., "User", "Question", "Topic"
	CreatedAt      time.Time      `json:"created_at" gorm:"index"`
	DeletedAt      gorm.DeletedAt `gorm:"index" json:"-"`

	User         User     `gorm:"foreignKey:UserID;references:ID;constraint:OnDelete:CASCADE;" json:"user,omitempty"`
	Question     Question `gorm:"foreignKey:FollowableID;references:ID" json:"question,omitempty"`
	FollowedUser User     `gorm:"foreignKey:FollowableID;references:ID" json:"followed_user,omitempty"`
	Topic        Topic    `gorm:"foreignKey:FollowableID;references:ID" json:"topic,omitempty"`
}
