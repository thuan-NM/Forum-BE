package responses

import (
	"Forum_BE/models"
	"time"
)

type PostResponse struct {
	PostID    uint      `json:"post_id"`
	Content   string    `json:"content"`
	UserID    uint      `json:"user_id"`
	GroupID   uint      `json:"group_id"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func ToPostResponse(post *models.Post) PostResponse {
	return PostResponse{
		PostID:    post.PostID,
		Content:   post.Content,
		UserID:    post.UserID,
		Status:    string(post.Status),
		CreatedAt: post.CreatedAt,
		UpdatedAt: post.UpdatedAt,
	}
}
