package responses

import (
	"Forum_BE/models"
	"time"
)

type TopicResponse struct {
	ID             uint   `json:"id"`
	Name           string `json:"name"`
	Description    string `json:"description"`
	QuestionCount  int    `json:"questionsCount"`
	FollowersCount int    `json:"followersCount"`
	CreatedAt      string `json:"createdAt"`
	UpdatedAt      string `json:"updatedAt"`
}

func ToTopicResponse(topic *models.Topic) TopicResponse {
	return TopicResponse{
		ID:             topic.ID,
		Name:           topic.Name,
		Description:    topic.Description,
		CreatedAt:      topic.CreatedAt.Format(time.RFC3339),
		UpdatedAt:      topic.UpdatedAt.Format(time.RFC3339),
		QuestionCount:  len(topic.Questions),
		FollowersCount: topic.FollowersCount,
	}
}
