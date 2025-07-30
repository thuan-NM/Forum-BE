package responses

import (
	"Forum_BE/models"
	"encoding/json"
	"time"
)

type AttachmentResponse struct {
	ID           uint            `json:"id"`
	UserID       uint            `json:"user_id"`
	URL          string          `json:"url"`
	ThumbnailURL string          `json:"thumbnail_url,omitempty"`
	FileName     string          `json:"file_name"`
	FileType     string          `json:"file_type"`
	FileSize     int64           `json:"file_size"`
	Metadata     json.RawMessage `json:"metadata,omitempty"`
	CreatedAt    string          `json:"created_at"`
	UpdatedAt    string          `json:"updated_at"`
	User         models.User     `json:"user,omitempty"`
}

func ToAttachmentResponse(attachment *models.Attachment) AttachmentResponse {
	return AttachmentResponse{
		ID:           attachment.ID,
		UserID:       attachment.UserID,
		URL:          attachment.URL,
		ThumbnailURL: attachment.ThumbnailURL,
		FileName:     attachment.FileName,
		FileType:     attachment.FileType,
		FileSize:     attachment.FileSize,
		Metadata:     attachment.Metadata,
		CreatedAt:    attachment.CreatedAt.Format(time.RFC3339),
		UpdatedAt:    attachment.UpdatedAt.Format(time.RFC3339),
		User:         attachment.User,
	}
}
