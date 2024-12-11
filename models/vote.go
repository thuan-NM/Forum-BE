package models

import (
	"fmt"
	"gorm.io/gorm"
	"time"
)

type VoteType string

const (
	VoteUp   VoteType = "upvote"
	VoteDown VoteType = "downvote"
)

type Vote struct {
	ID          uint           `gorm:"primaryKey" json:"id"`
	UserID      uint           `gorm:"not null;index" json:"user_id"`
	VotableType string         `gorm:"type:ENUM('question','answer','comment');not null" json:"votable_type"`
	VotableID   uint           `gorm:"not null;index" json:"votable_id"`
	VoteType    VoteType       `gorm:"type:ENUM('upvote','downvote');not null" json:"vote_type"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`

	// Relationships
	User User `json:"user,omitempty" gorm:"foreignKey:UserID"`
}

func (v *Vote) BeforeCreate(tx *gorm.DB) (err error) {
	if v.VotableType != "question" && v.VotableType != "answer" && v.VotableType != "comment" {
		return fmt.Errorf("invalid VotableType value")
	}
	if v.VotableID == 0 {
		return fmt.Errorf("VotableID cannot be 0")
	}
	return nil
}
