package models

import (
	"gorm.io/gorm"
	"time"
)

type UserFollow struct {
	ID             uint           `gorm:"primaryKey" json:"id"`
	FollowedUserID uint           `gorm:"not null" json:"followed_user_id"`
	UserID         uint           `gorm:"not null" json:"user_id"`
	CreatedAt      time.Time      `json:"created_at"`
	DeletedAt      gorm.DeletedAt `gorm:"index" json:"-"`

	FollowedUser *User `gorm:"foreignKey:FollowedUserID;references:ID" json:"followed_user,omitempty"`
	User         *User `gorm:"foreignKey:UserID;references:ID" json:"user,omitempty"`
}
