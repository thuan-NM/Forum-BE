package controllers

import (
	"Forum_BE/models"
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
		Title string `json:"title" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID := c.GetUint("user_id") // Middleware đã thêm user_id vào context

	question, err := qc.questionService.CreateQuestion(req.Title, userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":  "Question created successfully and pending approval",
		"question": responses.ToQuestionResponse(question),
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

	c.JSON(http.StatusOK, gin.H{
		"question": responses.ToQuestionResponse(question),
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
		Title string `json:"title"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	question, err := qc.questionService.UpdateQuestion(uint(id), req.Title)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":  "Question updated successfully",
		"question": responses.ToQuestionResponse(question),
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
	filters := make(map[string]interface{})
	userIDRaw, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	userID := userIDRaw.(uint)

	status := c.Query("status")
	if status != "" {
		filters["status"] = status
	}

	search := c.Query("search")
	if search != "" {
		filters["title_search"] = search
	}
	filters["user_id"] = userID

	tagID := c.Query("tag_id")
	if tagID != "" {
		if id, err := strconv.ParseUint(tagID, 10, 64); err == nil {
			filters["tag_id"] = uint(id)
		}
	}

	questions, err := qc.questionService.ListQuestions(filters)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list questions"})
		return
	}

	var responseQuestions []responses.QuestionResponse
	for _, question := range questions {
		responseQuestions = append(responseQuestions, responses.ToQuestionResponse(&question))
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
		"question": responses.ToQuestionResponse(question),
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
		"question": responses.ToQuestionResponse(question),
	})
}

// SuggestQuestions gợi ý câu hỏi
func (qc *QuestionController) SuggestQuestions(c *gin.Context) {
	//userID := c.GetUint("user_id")
	filters := map[string]interface{}{
		"status": models.StatusApproved,
	}
	sort := c.Query("sort")
	if sort == "popular" {
		filters["sort"] = "follow_count"
	}
	questions, err := qc.questionService.ListQuestions(filters)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to suggest questions"})
		return
	}

	var responseQuestions []responses.QuestionResponse
	for _, question := range questions {
		responseQuestions = append(responseQuestions, responses.ToQuestionResponse(&question))
	}

	c.JSON(http.StatusOK, gin.H{
		"questions": responseQuestions,
	})
}
