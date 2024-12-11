package controllers

import (
	"Forum_BE/responses"
	"Forum_BE/services"
	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
)

type QuestionController struct {
	questionService services.QuestionService
	voteService     services.VoteService
}

func NewQuestionController(q services.QuestionService, v services.VoteService) *QuestionController {
	return &QuestionController{questionService: q, voteService: v}
}

// CreateQuestion xử lý yêu cầu tạo question mới
func (qc *QuestionController) CreateQuestion(c *gin.Context) {
	var req struct {
		Title   string `json:"title" binding:"required"`
		Content string `json:"content" binding:"required"`
		GroupID uint   `json:"group_id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID := c.GetUint("user_id") // Middleware đã thêm user_id vào context

	question, err := qc.questionService.CreateQuestion(req.Title, req.Content, userID, req.GroupID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":  "Question created successfully and pending approval",
		"question": responses.ToQuestionResponse(question, 0),
	})
}

// GetQuestion xử lý yêu cầu lấy question theo ID
func (qc *QuestionController) GetQuestion(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.ParseUint(idParam, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid question id"})
		return
	}

	question, err := qc.questionService.GetQuestionByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "question not found"})
		return
	}

	// Lấy số lượng vote
	voteCount, err := qc.voteService.GetVoteCount("question", question.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get vote count"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"question": responses.ToQuestionResponse(question, voteCount),
	})
}

// UpdateQuestion xử lý yêu cầu cập nhật question
func (qc *QuestionController) UpdateQuestion(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.ParseUint(idParam, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid question id"})
		return
	}

	var req struct {
		Title   string `json:"title"`
		Content string `json:"content"`
		GroupID uint   `json:"group_id"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	question, err := qc.questionService.UpdateQuestion(uint(id), req.Title, req.Content, req.GroupID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":  "Question updated successfully",
		"question": responses.ToQuestionResponse(question, 0),
	})
}

// DeleteQuestion xử lý yêu cầu xóa question
func (qc *QuestionController) DeleteQuestion(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.ParseUint(idParam, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid question id"})
		return
	}

	if err := qc.questionService.DeleteQuestion(uint(id)); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Question deleted successfully",
	})
}

// ListQuestions xử lý yêu cầu liệt kê tất cả các question với các bộ lọc
func (qc *QuestionController) ListQuestions(c *gin.Context) {
	// Lấy các query parameters để lọc
	filters := make(map[string]interface{})

	status := c.Query("status")
	if status != "" {
		filters["status"] = status
	}

	groupID := c.Query("group_id")
	if groupID != "" {
		groupIDUint, err := strconv.ParseUint(groupID, 10, 64)
		if err == nil {
			filters["group_id"] = uint(groupIDUint)
		}
	}

	search := c.Query("search")
	if search != "" {
		filters["title_search"] = search
	}

	questions, err := qc.questionService.ListQuestions(filters)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list questions"})
		return
	}

	var responseQuestions []responses.QuestionResponse
	for _, question := range questions {
		// Lấy số lượng vote cho từng câu hỏi
		voteCount, err := qc.voteService.GetVoteCount("question", question.ID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get vote count"})
			return
		}

		responseQuestions = append(responseQuestions, responses.ToQuestionResponse(&question, voteCount))
	}

	c.JSON(http.StatusOK, gin.H{
		"questions": responseQuestions,
	})
}

// ApproveQuestion xử lý yêu cầu duyệt question
func (qc *QuestionController) ApproveQuestion(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.ParseUint(idParam, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid question id"})
		return
	}

	question, err := qc.questionService.ApproveQuestion(uint(id))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":  "Question approved successfully",
		"question": responses.ToQuestionResponse(question, 0),
	})
}

// RejectQuestion xử lý yêu cầu từ chối question
func (qc *QuestionController) RejectQuestion(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.ParseUint(idParam, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid question id"})
		return
	}

	question, err := qc.questionService.RejectQuestion(uint(id))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":  "Question rejected successfully",
		"question": responses.ToQuestionResponse(question, 0),
	})
}
