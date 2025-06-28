package responses

import (
	"Forum_BE/models"
	"time"
)

type QuestionResponse struct {
	ID                uint         `json:"id"`
	Title             string       `json:"title"`
	Description       string       `json:"description,omitempty"` // Uses Description field
	AnswerCount       int          `json:"answersCount"`
	ReactionCount     int          `json:"reactionsCount"`
	LastFollowed      string       `json:"lastFollowed"`
	FollowCount       int          `json:"followsCount"`
	Topic             models.Topic `json:"topic"`
	Status            string       `json:"status"`
	InteractionStatus string       `json:"interactionStatus"`
	Author            models.User  `json:"author"`
	CreatedAt         string       `json:"createdAt"`
	UpdatedAt         string       `json:"updatedAt"`
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
		ID:                question.ID,
		Title:             question.Title,
		Description:       question.Description,
		Author:            question.User,
		AnswerCount:       len(question.Answers),
		ReactionCount:     len(question.Reactions),
		LastFollowed:      lastFollowed,
		FollowCount:       len(question.Follows),
		Topic:             question.Topic,
		Status:            string(question.Status),
		InteractionStatus: string(question.InteractionStatus),
		CreatedAt:         question.CreatedAt.Format(time.RFC3339),
		UpdatedAt:         question.UpdatedAt.Format(time.RFC3339),
	}
}
