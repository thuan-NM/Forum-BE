package responses

import (
	"Forum_BE/models"
	"time"
)

type ReactionResponse struct {
	ID        uint            `json:"id"`
	UserID    uint            `json:"user_id"`
	PostID    *uint           `json:"post_id,omitempty"`
	CommentID *uint           `json:"comment_id,omitempty"`
	AnswerID  *uint           `json:"answer_id,omitempty"`
	CreatedAt string          `json:"created_at"`
	UpdatedAt string          `json:"updated_at"`
	User      models.User     `json:"user,omitempty"`
	Post      *models.Post    `json:"post,omitempty"`
	Comment   *models.Comment `json:"comment,omitempty"`
	Answer    *models.Answer  `json:"answer,omitempty"`
}

func ToReactionResponse(reaction *models.Reaction) ReactionResponse {
	return ReactionResponse{
		ID:        reaction.ID,
		UserID:    reaction.UserID,
		PostID:    reaction.PostID,
		CommentID: reaction.CommentID,
		AnswerID:  reaction.AnswerID,
		CreatedAt: reaction.CreatedAt.Format(time.RFC3339),
		UpdatedAt: reaction.UpdatedAt.Format(time.RFC3339),
		Answer:    reaction.Answer,
	}
}
