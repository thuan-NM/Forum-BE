package responses

import (
	"Forum_BE/models"
	"time"
)

type TagResponse struct {
	ID           uint   `json:"id"`
	Name         string `json:"name"`
	Description  string `json:"description"`
	PostsCount   int    `json:"postsCount"`
	AnswersCount int    `json:"answersCount"`
	CreatedAt    string `json:"createdAt"`
	UpdatedAt    string `json:"updatedAt"`
}

func ToTagResponse(tag *models.Tag) TagResponse {
	return TagResponse{
		ID:           tag.ID,
		Name:         tag.Name,
		Description:  tag.Description,
		PostsCount:   len(tag.Posts),
		AnswersCount: len(tag.Answers),
		CreatedAt:    tag.CreatedAt.Format(time.RFC3339),
		UpdatedAt:    tag.UpdatedAt.Format(time.RFC3339),
	}
}
