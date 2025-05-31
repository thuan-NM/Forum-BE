package models

import (
	"encoding/json"
	"gorm.io/gorm"
	"time"
)

type Notification struct {
	ID         uint            `gorm:"primaryKey" json:"id"`
	UserID     uint            `gorm:"index" json:"user_id"`         // ai nhận thông báo
	Type       string          `gorm:"type:varchar(50)" json:"type"` // loại: new_answer, upvote, mention, reply...
	Content    string          `gorm:"type:text" json:"content"`
	IsRead     bool            `gorm:"default:false" json:"is_read"`
	EntityType string          `gorm:"type:varchar(50);index" json:"entity_type"`
	EntityID   uint            `gorm:"index" json:"entity_id"`
	Metadata   json.RawMessage `gorm:"type:json" json:"metadata,omitempty"`
	CreatedAt  time.Time       `json:"created_at"`
	DeletedAt  gorm.DeletedAt  `gorm:"index" json:"-"`

	// Relationships
	User User `gorm:"foreignKey:UserID" json:"user,omitempty"`
}
