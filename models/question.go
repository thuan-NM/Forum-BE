package models

import (
	"encoding/json"
	"gorm.io/gorm"
	"time"
)

type QuestionStatus string

const (
	StatusApproved QuestionStatus = "approved"
	StatusPending  QuestionStatus = "pending"
	StatusRejected QuestionStatus = "rejected"
)

type Question struct {
	ID             uint            `gorm:"primaryKey" json:"id"`
	Title          string          `gorm:"not null;index" json:"title"`
	Description    string          `gorm:"type:text" json:"description,omitempty"`
	UserID         uint            `gorm:"not null;index" json:"user_id"`
	ViewCount      uint            `gorm:"default:0" json:"view_count" gorm:"index"`
	ReportCount    int             `gorm:"default:0" json:"report_count"`
	Status         QuestionStatus  `gorm:"type:ENUM('approved','pending','rejected');default:'pending'" json:"status"`
	Metadata       json.RawMessage `gorm:"type:json" json:"metadata,omitempty"`
	HasEditHistory bool            `gorm:"default:false" json:"has_edit_history"`
	CreatedAt      time.Time       `json:"created_at"`
	UpdatedAt      time.Time       `json:"updated_at"`
	DeletedAt      gorm.DeletedAt  `gorm:"index" json:"-"`

	User          User           `json:"user,omitempty" gorm:"foreignKey:UserID"`
	Answers       []Answer       `json:"answers,omitempty" gorm:"foreignKey:QuestionID"`
	Votes         []Vote         `json:"votes,omitempty" gorm:"polymorphic:Votable;"`
	Topics        []Topic        `json:"topics,omitempty" gorm:"many2many:question_topics;"`
	Follows       []Follow       `json:"follows,omitempty" gorm:"foreignKey:QuestionID;references:ID;constraint:OnDelete:CASCADE;"`
	Attachments   []Attachment   `json:"attachments,omitempty" gorm:"polymorphic:Entity;"`
	Reports       []Report       `json:"reports,omitempty" gorm:"polymorphic:Entity;"`
	Notifications []Notification `json:"notifications,omitempty" gorm:"polymorphic:Entity;"`
}
