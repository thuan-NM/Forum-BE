package responses

import (
	"Forum_BE/models"
)

type TopicResponse struct {
	ID            uint   `json:"id"`
	Name          string `json:"name"`
	Description   string `json:"description"`
	Status        string `json:"status"`
	CreatedBy     uint   `json:"created_by"`
	QuestionCount int    `json:"question_count"`
}

func ToTopicResponse(topic *models.Topic) TopicResponse {
	return TopicResponse{
		ID:            topic.ID,
		Name:          topic.Name,
		Description:   topic.Description,
		Status:        string(topic.Status),
		CreatedBy:     topic.CreatedBy,
		QuestionCount: len(topic.Questions),
	}
}
