package models

import (
	"encoding/json"
	"gorm.io/gorm"
	"time"
)

type Activity struct {
	ID         uint            `gorm:"primaryKey" json:"id"`
	UserID     uint            `gorm:"index" json:"user_id"`
	Action     string          `gorm:"type:varchar(50)" json:"action"`
	EntityType string          `gorm:"type:varchar(50);index" json:"entity_type"`
	EntityID   uint            `gorm:"index" json:"entity_id"`
	Message    string          `gorm:"type:text" json:"message"`
	Metadata   json.RawMessage `gorm:"type:json" json:"metadata,omitempty"`
	CreatedAt  time.Time       `json:"created_at" gorm:"index"`
	DeletedAt  gorm.DeletedAt  `gorm:"index" json:"-"`

	// Relationships
	User User `gorm:"foreignKey:UserID" json:"user,omitempty"`
}
