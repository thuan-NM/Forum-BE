package models

import (
	"time"

	"gorm.io/gorm"
)

type Follow struct {
	UserID     uint           `gorm:"primaryKey" json:"user_id"`
	QuestionID uint           `gorm:"primaryKey" json:"question_id"`
	CreatedAt  time.Time      `json:"created_at"`
	DeletedAt  gorm.DeletedAt `gorm:"index" json:"-"`

	User     User     `gorm:"foreignKey:UserID;references:ID;constraint:OnDelete:CASCADE;" json:"user,omitempty"`
	Question Question `gorm:"foreignKey:QuestionID;references:ID;constraint:OnDelete:CASCADE;" json:"question,omitempty"`
}
