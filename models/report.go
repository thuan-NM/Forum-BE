package models

import (
	"time"

	"gorm.io/gorm"
)

type ReportStatus string

const (
	PendingStatus   ReportStatus = "pending"
	ResolvedStatus  ReportStatus = "resolved"
	DismissedStatus ReportStatus = "dismissed"
)

type Report struct {
	ID             string         `gorm:"primaryKey;type:varchar(255)" json:"id"`
	Reason         string         `gorm:"type:varchar(255);not null" json:"reason"`
	Details        string         `gorm:"type:text" json:"details,omitempty"`
	ReporterID     uint           `gorm:"not null;index" json:"reporter_id"`
	ContentType    string         `gorm:"type:ENUM('post', 'comment', 'user', 'question', 'answer');not null" json:"content_type"`
	ContentID      string         `gorm:"type:varchar(255);not null" json:"content_id"`
	ContentPreview string         `gorm:"type:text;not null" json:"content_preview"`
	Status         ReportStatus   `gorm:"type:ENUM('pending', 'resolved', 'dismissed');default:'pending'" json:"status"`
	ResolvedByID   *uint          `gorm:"index" json:"resolved_by_id,omitempty"`
	CreatedAt      time.Time      `json:"created_at"`
	UpdatedAt      time.Time      `json:"updated_at"`
	DeletedAt      gorm.DeletedAt `gorm:"index" json:"-"`

	Reporter   User  `gorm:"foreignKey:ReporterID" json:"reporter,omitempty"`
	ResolvedBy *User `gorm:"foreignKey:ResolvedByID" json:"resolved_by,omitempty"`
}
