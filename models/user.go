package models

import (
	"gorm.io/gorm"
	"time"
)

type User struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	Username  string         `gorm:"unique;not null;index" json:"username"`
	Email     string         `gorm:"unique;not null;index" json:"email"`
	Password  string         `gorm:"not null" json:"-"`
	Role      Role           `gorm:"type:ENUM('root','admin','employee','user');default:'user'" json:"role"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	// Relationships
	Questions []Question `json:"questions,omitempty" gorm:"foreignKey:UserID"`
	Answers   []Answer   `json:"answers,omitempty" gorm:"foreignKey:UserID"`
	Comments  []Comment  `json:"comments,omitempty" gorm:"foreignKey:UserID"`
	Votes     []Vote     `json:"votes,omitempty" gorm:"foreignKey:UserID"`
	Followers []Follow   `json:"followers,omitempty" gorm:"foreignKey:FollowingID"`
	Following []Follow   `json:"following,omitempty" gorm:"foreignKey:FollowerID"`
}
