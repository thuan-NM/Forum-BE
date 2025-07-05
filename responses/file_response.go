package responses

import (
	"Forum_BE/models"
	"Forum_BE/utils"
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
	if file.EntityType != "" && file.EntityID != 0 {
		relatedTo = &struct {
			Type  string `json:"type"`
			ID    string `json:"id"`
			Title string `json:"title"`
		}{
			Type: file.EntityType,
			ID:   string(file.EntityID),
		}
		if file.Post != nil {
			relatedTo.Title = utils.StripHTML(file.Post.Title)
		} else if file.Answer != nil {
			relatedTo.Title = utils.StripHTML(file.Answer.Content)
		} else if file.Comment != nil {
			relatedTo.Title = utils.StripHTML(file.Comment.Content)
		}
	}

	return FileResponse{
		ID:           file.ID,
		Name:         file.FileName,
		Type:         file.FileType,
		Size:         file.FileSize,
		URL:          file.URL,
		ThumbnailURL: file.ThumbnailURL,
		UploadedBy:   file.User.Username,
		RelatedTo:    relatedTo,
		CreatedAt:    file.CreatedAt.Format(time.RFC3339),
	}
}
