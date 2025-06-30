package responses

import (
	"Forum_BE/models"
	"time"
)

type FollowResponse struct {
	UserID         uint   `json:"user_id"`
	FollowableID   uint   `json:"followable_id"`
	FollowableType string `json:"followable_type"`
	CreatedAt      string `json:"created_at"`
	DeletedAt      string `json:"deleted_at"`
}

func ToFollowResponse(follow *models.Follow) FollowResponse {
	var deletedAt string
	if follow.DeletedAt.Valid {
		deletedAt = follow.DeletedAt.Time.Format(time.RFC3339)
	}

	return FollowResponse{
		UserID:         follow.UserID,
		FollowableID:   follow.FollowableID,
		FollowableType: follow.FollowableType,
		CreatedAt:      follow.CreatedAt.Format(time.RFC3339),
		DeletedAt:      deletedAt,
	}
}
