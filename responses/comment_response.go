package responses

import (
	"Forum_BE/models"
	"Forum_BE/utils"
	"encoding/json"
	"time"
)

type CommentResponse struct {
	ID          uint        `json:"id"`
	Content     string      `json:"content"`
	User        models.User `json:"author"`
	PostID      *uint       `json:"postId,omitempty"`
	AnswerID    *uint       `json:"answerId,omitempty"`
	PostTitle   string      `json:"postTitle,omitempty"`
	AnswerTitle string      `json:"answerTitle,omitempty"`
	Status      string      `json:"status"`
	HasReply    bool        `json:"has_replies"`
	CreatedAt   string      `json:"createdAt"`
	UpdatedAt   string      `json:"updatedAt"`
	ParentTitle string      `json:"parentTitle,omitempty"`
}

func ToCommentResponse(comment *models.Comment) CommentResponse {
	var hasReply bool
	if comment.Metadata != nil {
		var metadata struct {
			HasReplies bool `json:"has_replies"`
		}
		if err := json.Unmarshal(comment.Metadata, &metadata); err == nil {
			hasReply = metadata.HasReplies
		}
	}

	var postTitle, AnswerTitle, parentTitle string
	if comment.Post != nil {
		postTitle = utils.StripHTML(comment.Post.Title)
	}
	if comment.Answer != nil {
		AnswerTitle = utils.StripHTML(comment.Answer.Content)
	}
	if comment.Parent != nil {
		parentTitle = utils.StripHTML(comment.Parent.Content)
	}

	return CommentResponse{
		ID:          comment.ID,
		Content:     comment.Content,
		User:        comment.User,
		PostID:      comment.PostID,
		AnswerID:    comment.AnswerID,
		PostTitle:   postTitle,
		AnswerTitle: AnswerTitle,
		Status:      comment.Status,
		HasReply:    hasReply,
		ParentTitle: parentTitle,
		CreatedAt:   comment.CreatedAt.Format(time.RFC3339),
		UpdatedAt:   comment.UpdatedAt.Format(time.RFC3339),
	}
}
