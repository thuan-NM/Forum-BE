package models

import (
	"encoding/json"
	"gorm.io/gorm"
	"time"
)

type Tag struct {
	ID             uint            `gorm:"primaryKey" json:"id"`
	Name           string          `gorm:"unique;not null;index" json:"name"`
	Description    string          `gorm:"type:text" json:"description,omitempty"`
	Status         string          `gorm:"type:ENUM('approved','pending','rejected');default:'approved'" json:"status" gorm:"index"`
	FollowersCount int             `gorm:"default:0" json:"followers_count"`
	Metadata       json.RawMessage `gorm:"type:json" json:"metadata,omitempty"`
	CreatedAt      time.Time       `json:"created_at"`
	UpdatedAt      time.Time       `json:"updated_at"`
	DeletedAt      gorm.DeletedAt  `gorm:"index" json:"-"`

	// Relationships
	Questions     []Question     `json:"questions,omitempty" gorm:"many2many:question_tags;"`
	Answers       []Answer       `json:"answers,omitempty" gorm:"many2many:answer_tags;"`
	Notifications []Notification `json:"notifications,omitempty" gorm:"polymorphic:Entity;"`
	Reports       []Report       `json:"reports,omitempty" gorm:"polymorphic:Entity;"`
	Attachments   []Attachment   `json:"attachments,omitempty" gorm:"polymorphic:Entity;"`
}
