package models

import (
	"fmt"
	"gorm.io/gorm"
	"time"
)

type Reaction struct {
	ID            uint           `gorm:"primaryKey" json:"id"`
	UserID        uint           `gorm:"not null;index:idx_reactable_user,unique" json:"user_id"`
	ReactableID   uint           `gorm:"not null;index:idx_reactable_user,unique" json:"reactable_id"`
	ReactableType string         `gorm:"type:varchar(50);not null;index:idx_reactable_user,unique" json:"reactable_type"` // "Post", "Comment", "Answer"
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
	DeletedAt     gorm.DeletedAt `gorm:"index" json:"-"`

	User User `json:"user,omitempty" gorm:"foreignKey:UserID"`
}

// BeforeCreate kiểm tra logic cơ bản
func (r *Reaction) BeforeCreate(tx *gorm.DB) (err error) {
	if r.ReactableType == "" {
		return fmt.Errorf("reactable_type is required")
	}
	if r.ReactableID == 0 {
		return fmt.Errorf("reactable_id is required")
	}
	if r.ReactableType != "Post" && r.ReactableType != "Comment" && r.ReactableType != "Answer" {
		return fmt.Errorf("reactable_type must be 'Post', 'Comment', or 'Answer'")
	}
	return nil
}
