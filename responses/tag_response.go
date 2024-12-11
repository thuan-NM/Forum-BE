package responses

import (
	"Forum_BE/models"
	"time"
)

type TagResponse struct {
	ID        uint   `json:"id"`
	Name      string `json:"name"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

func ToTagResponse(tag *models.Tag) TagResponse {
	return TagResponse{
		ID:        tag.ID,
		Name:      tag.Name,
		CreatedAt: tag.CreatedAt.Format(time.RFC3339),
		UpdatedAt: tag.UpdatedAt.Format(time.RFC3339),
	}
}
