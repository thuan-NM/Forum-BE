package responses

import (
	"Forum_BE/models"
	"time"
)

type QuestionResponse struct {
	ID        uint              `json:"id"`
	Title     string            `json:"title"`
	Content   string            `json:"content"`
	UserID    uint              `json:"user_id"`
	GroupID   uint              `json:"group_id"`
	Status    string            `json:"status"`
	CreatedAt string            `json:"created_at"`
	UpdatedAt string            `json:"updated_at"`
	Answers   []AnswerResponse  `json:"answers,omitempty"`
	Comments  []CommentResponse `json:"comments,omitempty"`
	Tags      []TagResponse     `json:"tags,omitempty"`
	VoteCount int64             `json:"vote_count,omitempty"`
}

func ToQuestionResponse(question *models.Question, voteCount int64) QuestionResponse {
	var answers []AnswerResponse
	for _, answer := range question.Answers {
		answers = append(answers, ToAnswerResponse(&answer, voteCount))
	}

	var comments []CommentResponse
	for _, comment := range question.Comments {
		comments = append(comments, ToCommentResponse(&comment, voteCount))
	}

	var tags []TagResponse
	for _, tag := range question.Tags {
		tags = append(tags, ToTagResponse(&tag))
	}

	return QuestionResponse{
		ID:        question.ID,
		Title:     question.Title,
		Content:   question.Content,
		UserID:    question.UserID,
		GroupID:   question.GroupID,
		Status:    string(question.Status),
		CreatedAt: question.CreatedAt.Format(time.RFC3339),
		UpdatedAt: question.UpdatedAt.Format(time.RFC3339),
		Answers:   answers,
		Comments:  comments,
		Tags:      tags,
		VoteCount: voteCount,
	}
}
