package models

import (
	"gorm.io/gorm"
	"time"
)

type Tag struct {
	ID             uint           `gorm:"primaryKey" json:"id"`
	Name           string         `gorm:"unique;not null;index" json:"name"`
	Description    string         `gorm:"type:text" json:"description,omitempty"`
	FollowersCount int            `gorm:"default:0" json:"followers_count"`
	CreatedAt      time.Time      `json:"created_at"`
	UpdatedAt      time.Time      `json:"updated_at"`
	DeletedAt      gorm.DeletedAt `gorm:"index" json:"-"`

	Answers       []Answer       `json:"answers,omitempty" gorm:"many2many:answer_tags;"`
	Notifications []Notification `json:"notifications,omitempty" gorm:"polymorphic:Entity;"`
	Posts         []Post         `json:"posts,omitempty" gorm:"many2many:post_tags;"`
}
