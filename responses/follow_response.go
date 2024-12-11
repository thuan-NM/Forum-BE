package responses

import (
	"Forum_BE/models"
	"time"
)

// FollowResponse định nghĩa cấu trúc dữ liệu trả về cho Follow
type FollowResponse struct {
	ID          uint   `json:"id"`
	FollowerID  uint   `json:"follower_id"`
	FollowingID uint   `json:"following_id"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
}

// ToFollowResponse chuyển đổi từ model Follow sang FollowResponse
func ToFollowResponse(follow *models.Follow) FollowResponse {
	return FollowResponse{
		ID:          follow.ID,
		FollowerID:  follow.FollowerID,
		FollowingID: follow.FollowingID,
		CreatedAt:   follow.CreatedAt.Format(time.RFC3339),
		UpdatedAt:   follow.UpdatedAt.Format(time.RFC3339),
	}
}
