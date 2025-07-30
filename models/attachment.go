package models

import (
	"encoding/json"
	"gorm.io/gorm"
	"time"
)

type Attachment struct {
	ID           uint            `gorm:"primaryKey" json:"id"`
	UserID       uint            `gorm:"index;not null" json:"user_id"`
	URL          string          `gorm:"type:text" json:"url"`
	ThumbnailURL string          `gorm:"type:text" json:"thumbnail_url,omitempty"`
	FileName     string          `gorm:"type:varchar(255)" json:"file_name"`
	FileType     string          `gorm:"type:varchar(50)" json:"file_type"` // image, document, video, audio, other
	FileSize     int64           `json:"file_size"`
	Metadata     json.RawMessage `gorm:"type:json" json:"metadata,omitempty"`
	CreatedAt    time.Time       `json:"created_at"`
	UpdatedAt    time.Time       `json:"updated_at"`
	DeletedAt    gorm.DeletedAt  `gorm:"index" json:"-"`

	// Relationships
	User User `gorm:"foreignKey:UserID" json:"user,omitempty"`
}
