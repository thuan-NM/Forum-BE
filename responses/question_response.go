package responses

import (
	"Forum_BE/models"
	"time"
)

type QuestionResponse struct {
	ID           uint   `json:"id"`
	Title        string `json:"title"`
	AnswerCount  int    `json:"answerCount"`
	LastFollowed string `json:"lastFollowed"`
	FollowCount  int    `json:"followCount"`
}

func ToQuestionResponse(question *models.Question) QuestionResponse {
	var lastFollowed string
	if len(question.Follows) > 0 {
		latest := question.Follows[0].CreatedAt
		for _, follow := range question.Follows {
			if follow.CreatedAt.After(latest) {
				latest = follow.CreatedAt
			}
		}
		lastFollowed = latest.Format(time.RFC3339)
	}

	return QuestionResponse{
		ID:           question.ID,
		Title:        question.Title,
		AnswerCount:  len(question.Answers),
		LastFollowed: lastFollowed,
		FollowCount:  len(question.Follows),
	}
}
