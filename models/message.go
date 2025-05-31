package models

import (
	"encoding/json"
	"gorm.io/gorm"
	"time"
)

type Message struct {
	ID          uint            `gorm:"primaryKey" json:"id"`
	FromUserID  uint            `gorm:"index" json:"from_user_id"`
	ToUserID    *uint           `gorm:"index" json:"to_user_id,omitempty"` // Null nếu là chat group
	GroupID     *uint           `gorm:"index" json:"group_id,omitempty"`
	Type        string          `gorm:"type:varchar(20);default:'text'" json:"type"` // text, image, file, system
	Content     string          `gorm:"type:text" json:"content"`
	IsRead      bool            `gorm:"default:false" json:"is_read"`
	Attachments json.RawMessage `gorm:"type:json" json:"attachments,omitempty"`
	Metadata    json.RawMessage `gorm:"type:json" json:"metadata,omitempty"`
	CreatedAt   time.Time       `json:"created_at" gorm:"index"`
	DeletedAt   gorm.DeletedAt  `gorm:"index" json:"-"`

	// Relationships
	FromUser User   `gorm:"foreignKey:FromUserID" json:"from_user,omitempty"`
	ToUser   *User  `gorm:"foreignKey:ToUserID" json:"to_user,omitempty"`
	Group    *Group `gorm:"foreignKey:GroupID" json:"group,omitempty"`
}
