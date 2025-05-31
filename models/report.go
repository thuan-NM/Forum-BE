package models

import (
	"encoding/json"
	"gorm.io/gorm"
	"time"
)

type Report struct {
	ID            uint            `gorm:"primaryKey" json:"id"`
	UserID        uint            `gorm:"not null;index" json:"user_id"` // người report
	EntityType    string          `gorm:"type:varchar(50);index" json:"entity_type"`
	EntityID      uint            `gorm:"not null;index" json:"entity_id"`
	Reason        string          `gorm:"type:text" json:"reason"`
	Status        string          `gorm:"type:ENUM('pending','approved','rejected','closed');default:'pending'" json:"status"`
	ModeratorID   *uint           `json:"moderator_id,omitempty"`
	ModeratorNote string          `gorm:"type:text" json:"moderator_note,omitempty"`
	Result        string          `gorm:"type:text" json:"result,omitempty"`
	Metadata      json.RawMessage `gorm:"type:json" json:"metadata,omitempty"`
	CreatedAt     time.Time       `json:"created_at"`
	ProcessedAt   *time.Time      `json:"processed_at,omitempty"`
	DeletedAt     gorm.DeletedAt  `gorm:"index" json:"-"`

	// Relationships
	User User `gorm:"foreignKey:UserID" json:"user,omitempty"`
}
