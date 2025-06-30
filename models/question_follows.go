package models

import (
	"gorm.io/gorm"
	"time"
)

type QuestionFollow struct {
	ID         uint           `gorm:"primaryKey" json:"id"`
	QuestionID uint           `gorm:"not null" json:"question_id"`
	UserID     uint           `gorm:"not null" json:"user_id"`
	CreatedAt  time.Time      `json:"created_at"`
	DeletedAt  gorm.DeletedAt `gorm:"index" json:"-"`

	Question *Question `gorm:"foreignKey:QuestionID;references:ID" json:"question,omitempty"`
	User     *User     `gorm:"foreignKey:UserID;references:ID" json:"user,omitempty"`
}
