package models

import (
	"gorm.io/gorm"
	"time"
)

type Answer struct {
	ID         uint           `gorm:"primaryKey" json:"id"`
	Content    string         `gorm:"type:text" json:"content"`
	UserID     uint           `gorm:"not null;index" json:"user_id"`
	QuestionID uint           `gorm:"not null;index" json:"question_id"`
	CreatedAt  time.Time      `json:"created_at"`
	UpdatedAt  time.Time      `json:"updated_at"`
	DeletedAt  gorm.DeletedAt `gorm:"index" json:"-"`

	// Relationships
	User     User      `json:"user,omitempty" gorm:"foreignKey:UserID"`
	Question Question  `json:"question,omitempty" gorm:"foreignKey:QuestionID"`
	Comments []Comment `json:"comments,omitempty" gorm:"foreignKey:AnswerID"`
	Votes    []Vote    `json:"votes,omitempty" gorm:"polymorphic:Votable;"`
}
