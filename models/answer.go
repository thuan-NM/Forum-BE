package models

import (
	"encoding/json"
	"gorm.io/gorm"
	"time"
)

type Answer struct {
	ID             uint            `gorm:"primaryKey" json:"id"`
	Content        string          `gorm:"type:text" json:"content"`
	UserID         uint            `gorm:"not null;index" json:"user_id"`
	QuestionID     uint            `gorm:"not null;index" json:"question_id"`
	Status         string          `gorm:"type:ENUM('approved','pending','rejected');default:'pending'" json:"status"`
	Accepted       bool            `gorm:"default:false" json:"accepted"`
	RootCommentID  *uint           `json:"root_comment_id,omitempty" gorm:"index"`
	Metadata       json.RawMessage `gorm:"type:json" json:"metadata,omitempty"`
	HasEditHistory bool            `gorm:"default:false" json:"has_edit_history"`
	CreatedAt      time.Time       `json:"created_at"`
	UpdatedAt      time.Time       `json:"updated_at"`
	DeletedAt      gorm.DeletedAt  `gorm:"index" json:"-"`
	PlainContent   string          `gorm:"type:text"`

	// Relationships
	User        User         `json:"user,omitempty" gorm:"foreignKey:UserID"`
	Question    Question     `json:"question,omitempty" gorm:"foreignKey:QuestionID"`
	Comments    []Comment    `json:"comments,omitempty" gorm:"foreignKey:AnswerID"`
	Votes       []Vote       `json:"votes,omitempty" gorm:"polymorphic:Votable;"`
	Attachments []Attachment `json:"attachments,omitempty" gorm:"polymorphic:Entity;"`
	Reports     []Report     `json:"reports,omitempty" gorm:"polymorphic:Entity;"`
}
