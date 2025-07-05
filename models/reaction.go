package models

import (
	"fmt"
	"gorm.io/gorm"
	"time"
)

type Reaction struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	UserID    uint           `gorm:"not null;index" json:"user_id"`
	PostID    *uint          `json:"post_id,omitempty" gorm:"index"`
	CommentID *uint          `json:"comment_id,omitempty" gorm:"index"`
	AnswerID  *uint          `json:"answer_id,omitempty" gorm:"index"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	User    User     `json:"user,omitempty" gorm:"foreignKey:UserID"`
	Post    *Post    `json:"post,omitempty" gorm:"foreignKey:PostID;references:ID"`
	Comment *Comment `json:"comment,omitempty" gorm:"foreignKey:CommentID;references:ID"`
	Answer  *Answer  `json:"answer,omitempty" gorm:"foreignKey:AnswerID;references:ID"`
}

func (r *Reaction) BeforeCreate(tx *gorm.DB) (err error) {
	count := 0
	if r.PostID != nil {
		count++
	}
	if r.CommentID != nil {
		count++
	}
	if r.AnswerID != nil {
		count++
	}
	if count != 1 {
		return fmt.Errorf("exactly one of post_id, comment_id, or answer_id must be provided")
	}
	return nil
}
