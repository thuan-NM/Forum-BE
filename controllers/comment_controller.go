package controllers

import (
	"Forum_BE/responses"
	"Forum_BE/services"
	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
)

type CommentController struct {
	commentService services.CommentService
	voteService    services.VoteService
}

func NewCommentController(c services.CommentService, v services.VoteService) *CommentController {
	return &CommentController{commentService: c, voteService: v}
}

// CreateComment xử lý yêu cầu tạo comment mới
func (cc *CommentController) CreateComment(c *gin.Context) {
	var req struct {
		Content    string `json:"content" binding:"required"`
		QuestionID *uint  `json:"question_id"`
		AnswerID   *uint  `json:"answer_id"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID := c.GetUint("user_id") // Middleware đã thêm user_id vào context

	comment, err := cc.commentService.CreateComment(req.Content, userID, req.QuestionID, req.AnswerID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Comment created successfully",
		"comment": responses.ToCommentResponse(comment, 0),
	})
}

// GetComment xử lý yêu cầu lấy comment theo ID
func (cc *CommentController) GetComment(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.ParseUint(idParam, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid comment id"})
		return
	}

	comment, err := cc.commentService.GetCommentByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "comment not found"})
		return
	}

	// Lấy số lượng vote
	voteCount, err := cc.voteService.GetVoteCount("comment", comment.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get vote count"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"comment": responses.ToCommentResponse(comment, voteCount),
	})
}

// EditComment xử lý yêu cầu cập nhật comment
func (cc *CommentController) EditComment(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.ParseUint(idParam, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid comment id"})
		return
	}

	var req struct {
		Content string `json:"content" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	comment, err := cc.commentService.UpdateComment(uint(id), req.Content)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Comment updated successfully",
		"comment": responses.ToCommentResponse(comment, 0),
	})
}

// DeleteComment xử lý yêu cầu xóa comment
func (cc *CommentController) DeleteComment(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.ParseUint(idParam, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid comment id"})
		return
	}

	if err := cc.commentService.DeleteComment(uint(id)); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Comment deleted successfully",
	})
}

// ListComments xử lý yêu cầu liệt kê tất cả các comment với các bộ lọc
func (cc *CommentController) ListComments(c *gin.Context) {
	// Lấy các query parameters để lọc
	filters := make(map[string]interface{})

	questionID := c.Query("question_id")
	if questionID != "" {
		qID, err := strconv.ParseUint(questionID, 10, 64)
		if err == nil {
			filters["question_id"] = uint(qID)
		}
	}

	answerID := c.Query("answer_id")
	if answerID != "" {
		aID, err := strconv.ParseUint(answerID, 10, 64)
		if err == nil {
			filters["answer_id"] = uint(aID)
		}
	}

	userID := c.Query("user_id")
	if userID != "" {
		uID, err := strconv.ParseUint(userID, 10, 64)
		if err == nil {
			filters["user_id"] = uint(uID)
		}
	}

	search := c.Query("search")
	if search != "" {
		filters["content LIKE ?"] = "%" + search + "%"
	}

	comments, err := cc.commentService.ListComments(filters)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list comments"})
		return
	}

	var responseComments []responses.CommentResponse
	for _, comment := range comments {
		// Lấy số lượng vote cho từng bình luận
		voteCount, err := cc.voteService.GetVoteCount("comment", comment.ID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get vote count"})
			return
		}

		responseComments = append(responseComments, responses.ToCommentResponse(&comment, voteCount))
	}

	c.JSON(http.StatusOK, gin.H{
		"comments": responseComments,
	})
}
