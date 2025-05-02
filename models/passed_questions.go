package models

import (
	"time"
)

type PassedQuestion struct {
	ID         uint `gorm:"primaryKey"`
	UserID     uint `gorm:"index"`
	QuestionID uint `gorm:"index"`
	CreatedAt  time.Time
}
