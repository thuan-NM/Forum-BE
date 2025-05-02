package models

import (
	"time"

	"gorm.io/gorm"
)

type QuestionStatus string

const (
	StatusApproved QuestionStatus = "approved"
	StatusPending  QuestionStatus = "pending"
	StatusRejected QuestionStatus = "rejected"
)

type Question struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	Title     string         `gorm:"not null;index" json:"title"`
	UserID    uint           `gorm:"not null;index" json:"user_id"`
	Status    QuestionStatus `gorm:"type:ENUM('approved','pending','rejected');default:'pending'" json:"status"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	// Relationships
	User     User      `json:"user,omitempty"`
	Answers  []Answer  `json:"answers,omitempty"`
	Comments []Comment `json:"comments,omitempty"`
	Votes    []Vote    `gorm:"foreignKey:VotableID;references:ID;constraint:OnDelete:CASCADE;" json:"votes,omitempty"`
	Tags     []Tag     `gorm:"many2many:question_tags;" json:"tags,omitempty"`
	Follows  []Follow  `gorm:"foreignKey:QuestionID;references:ID;constraint:OnDelete:CASCADE;" json:"follows,omitempty"`
}
