package responses

import (
	"Forum_BE/models"
	"time"
)

type PostResponse struct {
	ID            uint              `json:"id"`
	Content       string            `json:"content"`
	Author        models.User       `json:"author"`
	Status        string            `json:"status"`
	Comments      []CommentResponse `json:"comments,omitempty"`
	ReactionCount int               `json:"reactionsCount"`
	CreatedAt     string            `json:"createdAt"`
	UpdatedAt     string            `json:"updatedAt"`
	Tags          []models.Tag      `json:"tags,omitempty"`
}

func ToPostResponse(post *models.Post) PostResponse {
	var comments []CommentResponse
	for _, comment := range post.Comments {
		comments = append(comments, ToCommentResponse(&comment))
	}

	return PostResponse{
		ID:            post.ID,
		Content:       post.Content,
		Author:        post.User,
		Status:        string(post.Status),
		ReactionCount: len(post.Reactions),
		Comments:      comments,
		CreatedAt:     post.CreatedAt.Format(time.RFC3339),
		UpdatedAt:     post.UpdatedAt.Format(time.RFC3339),
		Tags:          post.Tags,
	}
}
