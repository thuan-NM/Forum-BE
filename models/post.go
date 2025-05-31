package models

import (
	"encoding/json"
	"time"

	"gorm.io/gorm"
)

type PostStatus string

const (
	Approved PostStatus = "approved"
	Pending  PostStatus = "pending"
	Rejected PostStatus = "rejected"
)

type Post struct {
	ID            uint            `gorm:"primaryKey" json:"id"`
	Title         string          `gorm:"type:varchar(255);index" json:"title,omitempty"`
	Content       string          `gorm:"type:text" json:"content"`
	UserID        uint            `gorm:"not null;index" json:"user_id"`
	QuestionID    *uint           `json:"question_id,omitempty" gorm:"index"`
	ReplyToUserID *uint           `json:"reply_to_user_id,omitempty" gorm:"index"`
	Upvotes       int             `gorm:"default:0" json:"upvotes"`
	Downvotes     int             `gorm:"default:0" json:"downvotes"`
	ReportCount   int             `gorm:"default:0" json:"report_count"`
	Status        PostStatus      `gorm:"type:ENUM('approved','pending','rejected');default:'pending'" json:"status"`
	Metadata      json.RawMessage `gorm:"type:json" json:"metadata,omitempty"`
	CreatedAt     time.Time       `json:"created_at"`
	UpdatedAt     time.Time       `json:"updated_at"`
	DeletedAt     gorm.DeletedAt  `gorm:"index" json:"-"`

	// Relationships
	User          User           `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Attachments   []Attachment   `json:"attachments,omitempty" gorm:"polymorphic:Entity;"`
	Reports       []Report       `json:"reports,omitempty" gorm:"polymorphic:Entity;"`
	Notifications []Notification `json:"notifications,omitempty" gorm:"polymorphic:Entity;"`
}
