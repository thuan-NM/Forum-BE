package responses

import (
	"Forum_BE/models"
	"time"
)

type AnswerResponse struct {
	ID         uint              `json:"id"`
	Content    string            `json:"content"`
	UserID     uint              `json:"user_id"`
	QuestionID uint              `json:"question_id"`
	CreatedAt  string            `json:"created_at"`
	UpdatedAt  string            `json:"updated_at"`
	Comments   []CommentResponse `json:"comments,omitempty"`
	Votes      []VoteResponse    `json:"votes,omitempty"`
	VoteCount  int64             `json:"vote_count,omitempty"`
}

func ToAnswerResponse(answer *models.Answer, voteCount int64) AnswerResponse {
	var comments []CommentResponse
	for _, comment := range answer.Comments {
		comments = append(comments, ToCommentResponse(&comment, voteCount))
	}

	var votes []VoteResponse
	for _, vote := range answer.Votes {
		votes = append(votes, ToVoteResponse(&vote))
	}

	return AnswerResponse{
		ID:         answer.ID,
		Content:    answer.Content,
		UserID:     answer.UserID,
		QuestionID: answer.QuestionID,
		CreatedAt:  answer.CreatedAt.Format(time.RFC3339),
		UpdatedAt:  answer.UpdatedAt.Format(time.RFC3339),
		Comments:   comments,
		Votes:      votes,
		VoteCount:  voteCount,
	}
}
