package responses

import (
	"Forum_BE/models"
	"time"
)

// VoteResponse định nghĩa cấu trúc dữ liệu trả về cho Vote
type VoteResponse struct {
	ID          uint   `json:"id"`
	UserID      uint   `json:"user_id"`
	VotableType string `json:"votable_type"`
	VotableID   uint   `json:"votable_id"`
	VoteType    string `json:"vote_type"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
}

// ToVoteResponse chuyển đổi từ model Vote sang VoteResponse
func ToVoteResponse(vote *models.Vote) VoteResponse {
	return VoteResponse{
		ID:          vote.ID,
		UserID:      vote.UserID,
		VotableType: vote.VotableType,
		VotableID:   vote.VotableID,
		VoteType:    string(vote.VoteType),
		CreatedAt:   vote.CreatedAt.Format(time.RFC3339),
		UpdatedAt:   vote.UpdatedAt.Format(time.RFC3339),
	}
}
