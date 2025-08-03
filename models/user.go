package models

import (
	"gorm.io/gorm"
	"time"
)

type Status string

const (
	StatusActive   Status = "active"
	StatusInactive Status = "inactive"
	StatusBanned   Status = "banned"
)

type User struct {
	ID             uint           `gorm:"primaryKey" json:"id"`
	Username       string         `gorm:"unique;not null;index;size:50" json:"username"`
	Email          string         `gorm:"unique;not null;index;size:100" json:"email"`
	Password       string         `gorm:"not null" json:"-"`
	Role           Role           `gorm:"type:enum('root','admin','employee','user');default:'user'" json:"role"`
	Avatar         *string        `json:"avatar,omitempty"`
	Bio            *string        `gorm:"type:text" json:"bio,omitempty"`
	Status         Status         `gorm:"type:enum('active','inactive','banned');default:'inactive'" json:"status"`
	Location       *string        `json:"location,omitempty"`
	FullName       string         `gorm:"not null" json:"fullName"`
	Reputation     uint           `gorm:"default:0;index" json:"reputation"`
	FollowersCount uint           `gorm:"default:0" json:"followers_count"`
	FollowingCount uint           `gorm:"default:0" json:"following_count"`
	LastLogin      *time.Time     `json:"last_login,omitempty"`
	CreatedAt      time.Time      `json:"created_at"`
	UpdatedAt      time.Time      `json:"updated_at"`
	DeletedAt      gorm.DeletedAt `gorm:"index" json:"-"`
	EmailVerified  bool           `json:"email_verified"`
	PostCount      int64          `gorm:"-" json:"postCount,omitempty"`
	AnswerCount    int64          `gorm:"-" json:"answerCount,omitempty"`
	QuestionCount  int64          `gorm:"-" json:"questionCount,omitempty"`

	Posts            []Post         `gorm:"foreignKey:UserID" json:"posts,omitempty"`
	Questions        []Question     `gorm:"foreignKey:UserID" json:"questions,omitempty"`
	Answers          []Answer       `gorm:"foreignKey:UserID" json:"answers,omitempty"`
	Reactions        []Reaction     `gorm:"foreignKey:UserID" json:"reactions,omitempty"` // Thêm mối quan hệ với Reaction
	Comments         []Comment      `gorm:"foreignKey:UserID" json:"comments,omitempty"`
	Notifications    []Notification `gorm:"foreignKey:UserID" json:"notifications,omitempty"`
	SentMessages     []Message      `gorm:"foreignKey:FromUserID" json:"sent_messages,omitempty"`
	ReceivedMessages []Message      `gorm:"foreignKey:ToUserID" json:"received_messages,omitempty"`
	Following        []UserFollow   `json:"following,omitempty" gorm:"foreignKey:UserID"`
	Followers        []UserFollow   `json:"followers,omitempty" gorm:"foreignKey:FollowedUserID"`
}
