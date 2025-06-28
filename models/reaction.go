package models

import (
	"gorm.io/gorm"
	"time"
)

type Reaction struct {
	ID            uint           `gorm:"primaryKey" json:"id"`
	UserID        uint           `gorm:"not null;index:idx_reactable_user,unique" json:"user_id"`
	ReactableID   uint           `gorm:"not null;index:idx_reactable_user,unique" json:"reactable_id"`
	ReactableType string         `gorm:"not null;index:idx_reactable_user,unique" json:"reactable_type"` // "Question" hoặc "Answer"
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
	DeletedAt     gorm.DeletedAt `gorm:"index" json:"-"`

	User User `json:"user,omitempty" gorm:"foreignKey:UserID"`
}
