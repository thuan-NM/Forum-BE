package models

import (
	"gorm.io/gorm"
	"time"
)

type Follow struct {
	ID          uint           `gorm:"primaryKey" json:"id"`
	FollowerID  uint           `gorm:"not null;index" json:"follower_id"`
	FollowingID uint           `gorm:"not null;index" json:"following_id"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`

	// Relationships
	Follower  User `json:"follower,omitempty" gorm:"foreignKey:FollowerID"`
	Following User `json:"following,omitempty" gorm:"foreignKey:FollowingID"`
}
