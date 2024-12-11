package models

import (
	"gorm.io/gorm"
	"time"
)

type Tag struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	Name      string         `gorm:"unique;not null;index" json:"name"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	// Relationships
	Questions []Question `json:"questions,omitempty" gorm:"many2many:question_tags;"`
	Answers   []Answer   `json:"answers,omitempty" gorm:"many2many:answer_tags;"`
}
