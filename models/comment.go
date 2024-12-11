package models

import (
	"gorm.io/gorm"
	"time"
)

type Comment struct {
	ID         uint           `gorm:"primaryKey" json:"id"`
	Content    string         `gorm:"type:text" json:"content"`
	UserID     uint           `gorm:"not null;index" json:"user_id"`
	QuestionID *uint          `json:"question_id,omitempty"`
	AnswerID   *uint          `json:"answer_id,omitempty"`
	CreatedAt  time.Time      `json:"created_at"`
	UpdatedAt  time.Time      `json:"updated_at"`
	DeletedAt  gorm.DeletedAt `gorm:"index" json:"-"`

	// Relationships
	User     User      `json:"user,omitempty" gorm:"foreignKey:UserID"`
	Question *Question `json:"question,omitempty" gorm:"foreignKey:QuestionID;references:ID"` // Foreign key `QuestionID` references `ID` in `questions`
	Answer   *Answer   `json:"answer,omitempty" gorm:"foreignKey:AnswerID;references:ID"`     // Foreign key `AnswerID` references `ID` in `answers`
	Votes    []Vote    `json:"votes,omitempty" gorm:"polymorphic:Votable;"`
}
