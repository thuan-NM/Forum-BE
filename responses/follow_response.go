package responses

import (
	"Forum_BE/models"
	"time"
)

// FollowResponse định nghĩa cấu trúc dữ liệu trả về cho Follow
type FollowResponse struct {
	UserID     uint   `json:"id"`
	QuestionID uint   `json:"question_id"`
	CreatedAt  string `json:"created_at"`
	DeletedAt  string `json:"deleted_at"`
}

// ToFollowResponse chuyển đổi từ model Follow sang FollowResponse
func ToFollowResponse(follow *models.Follow) FollowResponse {
	var deletedAt string
	if follow.DeletedAt.Valid {
		deletedAt = follow.DeletedAt.Time.Format(time.RFC3339)
	}

	return FollowResponse{
		UserID:     follow.UserID,
		QuestionID: follow.QuestionID,
		CreatedAt:  follow.CreatedAt.Format(time.RFC3339),
		DeletedAt:  deletedAt,
	}
}
