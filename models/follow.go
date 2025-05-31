package models

import (
	"time"

	"gorm.io/gorm"
)

type Follow struct {
	ID         uint           `gorm:"primaryKey" json:"id"`
	UserID     uint           `gorm:"not null;index" json:"user_id"`
	QuestionID uint           `gorm:"not null;index" json:"question_id"`
	CreatedAt  time.Time      `json:"created_at" gorm:"index"`
	DeletedAt  gorm.DeletedAt `gorm:"index" json:"-"`

	User     User     `gorm:"foreignKey:UserID;references:ID;constraint:OnDelete:CASCADE;" json:"user,omitempty"`
	Question Question `gorm:"foreignKey:QuestionID;references:ID;constraint:OnDelete:CASCADE;" json:"question,omitempty"`
}
