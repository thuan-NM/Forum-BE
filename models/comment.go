package models

import (
	"encoding/json"
	"fmt"
	"gorm.io/gorm"
	"time"
)

type Comment struct {
	ID        uint            `gorm:"primaryKey" json:"id"`
	Content   string          `gorm:"type:text" json:"content"`
	UserID    uint            `gorm:"not null;index" json:"user_id"`
	PostID    *uint           `json:"post_id,omitempty" gorm:"index"`
	AnswerID  *uint           `json:"answer_id,omitempty" gorm:"index"`
	ParentID  *uint           `json:"parent_id,omitempty" gorm:"index"`
	Status    string          `gorm:"type:ENUM('approved','pending','spam');default:'pending'" json:"status"`
	Metadata  json.RawMessage `gorm:"type:json" json:"metadata,omitempty"`
	CreatedAt time.Time       `json:"created_at"`
	UpdatedAt time.Time       `json:"updated_at"`
	DeletedAt gorm.DeletedAt  `gorm:"index" json:"-"`

	// Relationships
	User     User      `json:"user,omitempty" gorm:"foreignKey:UserID"`
	Post     *Post     `json:"post,omitempty" gorm:"foreignKey:PostID;references:ID"`
	Answer   *Answer   `json:"answer,omitempty" gorm:"foreignKey:AnswerID;references:ID"`
	Votes    []Vote    `json:"votes,omitempty" gorm:"polymorphic:Votable;"`
	Parent   *Comment  `json:"parent,omitempty" gorm:"foreignKey:ParentID"`
	Children []Comment `json:"children,omitempty" gorm:"foreignKey:ParentID"`
}

func (c *Comment) BeforeCreate(tx *gorm.DB) (err error) {
	if c.ParentID != nil && *c.ParentID == c.ID {
		return fmt.Errorf("comment cannot reference itself as parent")
	}
	return nil
}
