package models

import (
	"encoding/json"
	"encoding/json"
	"gorm.io/gorm"
	"time"
)

type Answer struct {
	ID             uint            `gorm:"primaryKey" json:"id"`
	Content        string          `gorm:"type:text" json:"content"`
	Title          string          `gorm:"type:text" json:"title"`
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

	User        User         `json:"user,omitempty" gorm:"foreignKey:UserID"`
	Question    Question     `json:"question,omitempty" gorm:"foreignKey:QuestionID"`
	Comments    []Comment    `json:"comments,omitempty" gorm:"foreignKey:AnswerID"`
	Reactions   []Reaction   `json:"reactions,omitempty" gorm:"foreignKey:AnswerID"`
	Tags        []Tag        `json:"tags,omitempty" gorm:"many2many:answer_tags;"`
	Attachments []Attachment `json:"attachments,omitempty" gorm:"many2many:answer_attachments;"`
}
