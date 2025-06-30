package responses

import (
	"Forum_BE/models"
	"time"
)

type ReactionResponse struct {
	ID            uint        `json:"id"`
	UserID        uint        `json:"user_id"`
	ReactableID   uint        `json:"reactable_id"`
	ReactableType string      `json:"reactable_type"`
	CreatedAt     string      `json:"created_at"`
	UpdatedAt     string      `json:"updated_at"`
	User          models.User `json:"user,omitempty"`
}

func ToReactionResponse(reaction *models.Reaction) ReactionResponse {
	return ReactionResponse{
		ID:            reaction.ID,
		UserID:        reaction.UserID,
		ReactableID:   reaction.ReactableID,
		ReactableType: reaction.ReactableType,
		CreatedAt:     reaction.CreatedAt.Format(time.RFC3339),
		UpdatedAt:     reaction.UpdatedAt.Format(time.RFC3339),
		User:          reaction.User,
	}
}
