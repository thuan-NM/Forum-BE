package models

import (
	"encoding/json"
	"gorm.io/gorm"
	"time"
)

type User struct {
	ID             uint            `gorm:"primaryKey" json:"id"`
	Username       string          `gorm:"unique;not null;index" json:"username"`
	Email          string          `gorm:"unique;not null;index" json:"email"`
	Password       string          `gorm:"not null" json:"-"`
	Role           Role            `gorm:"type:ENUM('root','admin','employee','user');default:'user'" json:"role"`
	Avatar         string          `json:"avatar,omitempty"`
	Bio            string          `gorm:"type:text" json:"bio,omitempty"`
	IsActive       bool            `gorm:"default:true" json:"is_active"`
	IsBanned       bool            `gorm:"default:false" json:"is_banned"`
	Location       string          `json:"location,omitempty"`
	Reputation     uint            `gorm:"default:0" json:"reputation" gorm:"index"`
	FollowersCount uint            `gorm:"default:0" json:"followers_count"`
	FollowingCount uint            `gorm:"default:0" json:"following_count"`
	LastLogin      *time.Time      `json:"last_login" gorm:"default:null"`
	Metadata       json.RawMessage `gorm:"type:json" json:"metadata,omitempty"`
	CreatedAt      time.Time       `json:"created_at"`
	UpdatedAt      time.Time       `json:"updated_at"`
	DeletedAt      gorm.DeletedAt  `gorm:"index" json:"-"`

	// Relationships
	Questions        []Question     `json:"questions,omitempty" gorm:"foreignKey:UserID"`
	Answers          []Answer       `json:"answers,omitempty" gorm:"foreignKey:UserID"`
	Comments         []Comment      `json:"comments,omitempty" gorm:"foreignKey:UserID"`
	Votes            []Vote         `json:"votes,omitempty" gorm:"foreignKey:UserID"`
	Notifications    []Notification `json:"notifications,omitempty" gorm:"foreignKey:UserID"`
	Reports          []Report       `json:"reports,omitempty" gorm:"foreignKey:UserID"`
	SentMessages     []Message      `json:"sent_messages,omitempty" gorm:"foreignKey:FromUserID"`
	ReceivedMessages []Message      `json:"received_messages,omitempty" gorm:"foreignKey:ToUserID"`
	Attachments      []Attachment   `json:"attachments,omitempty" gorm:"foreignKey:UserID"`
}
