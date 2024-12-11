package responses

import (
	"Forum_BE/models"
	"time"
)

type CommentResponse struct {
	ID         uint           `json:"id"`
	Content    string         `json:"content"`
	UserID     uint           `json:"user_id"`
	QuestionID *uint          `json:"question_id,omitempty"`
	AnswerID   *uint          `json:"answer_id,omitempty"`
	CreatedAt  string         `json:"created_at"`
	UpdatedAt  string         `json:"updated_at"`
	Votes      []VoteResponse `json:"votes,omitempty"`
	VoteCount  int64          `json:"vote_count,omitempty"`
}

func ToCommentResponse(comment *models.Comment, voteCount int64) CommentResponse {
	var votes []VoteResponse
	for _, vote := range comment.Votes {
		votes = append(votes, ToVoteResponse(&vote))
	}

	return CommentResponse{
		ID:         comment.ID,
		Content:    comment.Content,
		UserID:     comment.UserID,
		QuestionID: comment.QuestionID,
		AnswerID:   comment.AnswerID,
		CreatedAt:  comment.CreatedAt.Format(time.RFC3339),
		UpdatedAt:  comment.UpdatedAt.Format(time.RFC3339),
		Votes:      votes,
		VoteCount:  voteCount,
	}
}
