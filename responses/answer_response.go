package responses

import (
	"Forum_BE/models"
	"time"
)

type AnswerResponse struct {
	ID             uint              `json:"id"`
	Content        string            `json:"content"`
	QuestionID     uint              `json:"questionId"`
	CreatedAt      string            `json:"createdAt"`
	UpdatedAt      string            `json:"updatedAt"`
	Comments       []CommentResponse `json:"comments,omitempty"`
	Status         string            `gorm:"type:ENUM('approved','pending','rejected');default:'pending'" json:"status"`
	Accepted       bool              `gorm:"default:false" json:"isAccepted"`
	RootCommentID  *uint             `json:"root_comment_id,omitempty" gorm:"index"`
	HasEditHistory bool              `gorm:"default:false" json:"has_edit_history"`
	Author         models.User       `json:"author"`
	Question       models.Question   `json:"question"`
}

func ToAnswerResponse(answer *models.Answer) AnswerResponse {
	var comments []CommentResponse
	for _, comment := range answer.Comments {
		comments = append(comments, ToCommentResponse(&comment))
	}

	return AnswerResponse{
		ID:             answer.ID,
		Content:        answer.Content,
		QuestionID:     answer.QuestionID,
		CreatedAt:      answer.CreatedAt.Format(time.RFC3339),
		UpdatedAt:      answer.UpdatedAt.Format(time.RFC3339),
		Accepted:       answer.Accepted,
		RootCommentID:  answer.RootCommentID,
		HasEditHistory: answer.HasEditHistory,
		Comments:       comments,
		Status:         answer.Status,
		Author:         answer.User,
		Question:       answer.Question,
	}
}
