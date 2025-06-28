package models

import (
	"encoding/json"
	"time"

	"gorm.io/gorm"
)

type Group struct {
	ID           uint            `gorm:"primaryKey" json:"id"`
	Name         string          `gorm:"unique;not null;index" json:"name"`
	Avatar       string          `json:"avatar,omitempty"`
	Description  string          `json:"description,omitempty"`
	Rule         string          `gorm:"type:text" json:"rule,omitempty"`
	CreatorID    uint            `json:"creator_id" gorm:"index"`
	Status       string          `gorm:"type:ENUM('active','pending','blocked');default:'active'" json:"status" gorm:"index"`
	MembersCount int             `gorm:"default:0" json:"members_count"`
	Metadata     json.RawMessage `gorm:"type:json" json:"metadata,omitempty"`
	CreatedAt    time.Time       `json:"created_at"`
	UpdatedAt    time.Time       `json:"updated_at"`
	DeletedAt    gorm.DeletedAt  `gorm:"index" json:"-"`

	// Relationships
	Notifications []Notification `json:"notifications,omitempty" gorm:"polymorphic:Entity;"`
	Attachments   []Attachment   `json:"attachments,omitempty" gorm:"polymorphic:Entity;"`
	Members       []User         `json:"members,omitempty" gorm:"many2many:group_members;"`
}
