package models

import (
	"gorm.io/gorm"
	"time"
)

type QuestionStatus string
type InteractionStatus string

const (
	StatusApproved QuestionStatus = "approved"
	StatusPending  QuestionStatus = "pending"
	StatusRejected QuestionStatus = "rejected"
)

const (
	InteractionOpened InteractionStatus = "opened"
	InteractionSolved InteractionStatus = "solved"
	InteractionClosed InteractionStatus = "closed"
)

type Question struct {
	ID                uint              `gorm:"primaryKey" json:"id"`
	Title             string            `gorm:"not null;index" json:"title"`
	Description       string            `gorm:"type:text" json:"description,omitempty"`
	UserID            uint              `gorm:"not null;index" json:"user_id"`
	TopicID           uint              `gorm:"index" json:"topic_id"` // Liên kết với một topic duy nhất
	ReportCount       int               `gorm:"default:0" json:"report_count"`
	Status            QuestionStatus    `gorm:"type:ENUM('approved','pending','rejected');default:'pending'" json:"status"`
	InteractionStatus InteractionStatus `gorm:"type:ENUM('opened','solved','closed');default:'opened'" json:"interaction_status"`
	CreatedAt         time.Time         `json:"created_at"`
	UpdatedAt         time.Time         `json:"updated_at"`
	DeletedAt         gorm.DeletedAt    `gorm:"index" json:"-"`

	User          User             `json:"user,omitempty" gorm:"foreignKey:UserID"`
	Topic         Topic            `json:"topic,omitempty" gorm:"foreignKey:TopicID"`
	Answers       []Answer         `json:"answers,omitempty" gorm:"foreignKey:QuestionID"`
	Follows       []QuestionFollow `json:"follows,omitempty" gorm:"foreignKey:QuestionID"`
	Attachments   []Attachment     `json:"attachments,omitempty" gorm:"polymorphic:Entity;"`
	Notifications []Notification   `json:"notifications,omitempty" gorm:"polymorphic:Entity;"`
}
