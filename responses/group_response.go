package responses

import (
	"Forum_BE/models"
	"time"
)

type GroupResponse struct {
	ID          uint               `json:"id"`
	Name        string             `json:"name"`
	Description string             `json:"description,omitempty"`
	CreatedAt   string             `json:"created_at"`
	UpdatedAt   string             `json:"updated_at"`
	Questions   []QuestionResponse `json:"questions,omitempty"`
}

func ToGroupResponse(group *models.Group) GroupResponse {
	var questions []QuestionResponse
	for _, question := range group.Questions {
		// Giả sử bạn đã có số lượng vote cho câu hỏi
		questions = append(questions, QuestionResponse{
			ID:        question.ID,
			Title:     question.Title,
			Content:   question.Content,
			UserID:    question.UserID,
			GroupID:   question.GroupID,
			Status:    string(question.Status),
			CreatedAt: question.CreatedAt.Format(time.RFC3339),
			UpdatedAt: question.UpdatedAt.Format(time.RFC3339),
			// VoteCount: 0, // Bạn có thể thêm số lượng vote nếu cần
		})
	}

	return GroupResponse{
		ID:          group.ID,
		Name:        group.Name,
		Description: group.Description,
		CreatedAt:   group.CreatedAt.Format(time.RFC3339),
		UpdatedAt:   group.UpdatedAt.Format(time.RFC3339),
		Questions:   questions,
	}
}
