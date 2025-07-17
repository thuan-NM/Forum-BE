package responses

import (
	"Forum_BE/models"
	"time"
)

type FileResponse struct {
	ID           uint   `json:"id"`
	Name         string `json:"name"`
	Type         string `json:"type"`
	Size         int64  `json:"size"`
	URL          string `json:"url"`
	ThumbnailURL string `json:"thumbnailUrl,omitempty"`
	UploadedBy   string `json:"uploadedBy"`
	RelatedTo    *struct {
		Type  string `json:"type"`
		ID    string `json:"id"`
		Title string `json:"title"`
	} `json:"relatedTo,omitempty"`
	CreatedAt string `json:"createdAt"`
}

func ToFileResponse(file *models.Attachment) FileResponse {
	var relatedTo *struct {
		Type  string `json:"type"`
		ID    string `json:"id"`
		Title string `json:"title"`
	}

	// Since EntityType and EntityID are removed, we can optionally check for related comments
	// This requires a database query to find comments that reference this file in AttachmentIDs
	// For simplicity, we can leave relatedTo as nil or implement a query in the service layer if needed
	// Here, we'll set relatedTo to nil unless you want to add a query to find associated comments

	return FileResponse{
		ID:           file.ID,
		Name:         file.FileName,
		Type:         file.FileType,
		Size:         file.FileSize,
		URL:          file.URL,
		ThumbnailURL: file.ThumbnailURL,
		UploadedBy:   file.User.Username,
		RelatedTo:    relatedTo, // Set to nil as attachments are not directly tied to entities
		CreatedAt:    file.CreatedAt.Format(time.RFC3339),
	}
}
