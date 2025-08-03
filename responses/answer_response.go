package responses

import (
	"Forum_BE/models"
	"time"
)

type AnswerResponse struct {
	ID             uint              `json:"id"`
	Content        string            `json:"content"`
	Title          string            `json:"title"`
	QuestionID     uint              `json:"questionId"`
	CreatedAt      string            `json:"createdAt"`
	UpdatedAt      string            `json:"updatedAt"`
	Comments       []CommentResponse `json:"comments,omitempty"`
	Status         string            `gorm:"type:ENUM('approved','pending','rejected');default:'pending'" json:"status"`
	Accepted       bool              `gorm:"default:false" json:"isAccepted"`
	RootCommentID  *uint             `json:"root_comment_id,omitempty" gorm:"index"`
	HasEditHistory bool              `gorm:"default:false" json:"has_edit_history"`
	Author         models.User       `json:"author"`
	Question       QuestionResponse  `json:"question"`
	ReactionCount  int               `json:"reactionsCount"`
	Tags           []TagResponse     `json:"tags,omitempty"` // Thêm trường Tags
}

func ToAnswerResponse(answer *models.Answer) AnswerResponse {
	var comments []CommentResponse
	for _, comment := range answer.Comments {
		comments = append(comments, ToCommentResponse(&comment))
	}
	var tags []TagResponse
	for _, tag := range answer.Tags {
		tags = append(tags, TagResponse{ID: tag.ID, Name: tag.Name}) // Giả sử TagResponse có ID và Name
	}
	var question = ToQuestionResponse(&answer.Question)
	return AnswerResponse{
		ID:             answer.ID,
		Title:          answer.Title,
		Content:        answer.Content,
		QuestionID:     answer.QuestionID,
		CreatedAt:      answer.CreatedAt.Format(time.RFC3339),
		UpdatedAt:      answer.UpdatedAt.Format(time.RFC3339),
		Accepted:       answer.Accepted,
		RootCommentID:  answer.RootCommentID,
		HasEditHistory: answer.HasEditHistory,
		ReactionCount:  len(answer.Reactions),
		Comments:       comments,
		Status:         answer.Status,
		Author:         answer.User,
		Question:       question,
		Tags:           tags,
	}
}
