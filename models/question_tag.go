package models

import (
	"time"

	"gorm.io/gorm"
)

type QuestionTag struct {
	QuestionID uint           `gorm:"primaryKey"`
	TagID      uint           `gorm:"primaryKey"`
	CreatedAt  time.Time      `json:"created_at"`
	DeletedAt  gorm.DeletedAt `gorm:"index" json:"-"`
}
