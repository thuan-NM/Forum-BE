package models

import (
	"gorm.io/gorm"
	"time"
)

type TopicFollow struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	TopicID   uint           `gorm:"not null" json:"topic_id"`
	UserID    uint           `gorm:"not null" json:"user_id"`
	CreatedAt time.Time      `json:"created_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	Topic *Topic `gorm:"foreignKey:TopicID;references:ID" json:"topic,omitempty"`
	User  *User  `gorm:"foreignKey:UserID;references:ID" json:"user,omitempty"`
}
