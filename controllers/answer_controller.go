package controllers

import (
	"Forum_BE/responses"
	"Forum_BE/services"
	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
)

type AnswerController struct {
	answerService services.AnswerService
	voteService   services.VoteService
}

func NewAnswerController(a services.AnswerService, v services.VoteService) *AnswerController {
	return &AnswerController{answerService: a, voteService: v}
}

// CreateAnswer xử lý yêu cầu tạo answer mới
func (ac *AnswerController) CreateAnswer(c *gin.Context) {
	var req struct {
		Content    string `json:"content" binding:"required"`
		QuestionID uint   `json:"question_id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID := c.GetUint("user_id") // Middleware đã thêm user_id vào context

	answer, err := ac.answerService.CreateAnswer(req.Content, userID, req.QuestionID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Answer created successfully",
		"answer":  responses.ToAnswerResponse(answer, 0),
	})
}

// GetAnswer xử lý yêu cầu lấy answer theo ID
func (ac *AnswerController) GetAnswer(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.ParseUint(idParam, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid answer id"})
		return
	}

	answer, err := ac.answerService.GetAnswerByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "answer not found"})
		return
	}

	// Lấy số lượng vote
	voteCount, err := ac.voteService.GetVoteCount("answer", answer.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get vote count"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"answer": responses.ToAnswerResponse(answer, voteCount),
	})
}

func (ac *AnswerController) EditAnswer(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.ParseUint(idParam, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid answer id"})
		return
	}

	var req struct {
		Content string `json:"content"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	answer, err := ac.answerService.UpdateAnswer(uint(id), req.Content)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Answer updated successfully",
		"answer":  responses.ToAnswerResponse(answer, 0),
	})
}

// DeleteAnswer xử lý yêu cầu xóa answer
func (ac *AnswerController) DeleteAnswer(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.ParseUint(idParam, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid answer id"})
		return
	}

	if err := ac.answerService.DeleteAnswer(uint(id)); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Answer deleted successfully",
	})
}

// ListAnswers xử lý yêu cầu liệt kê tất cả các answer với các bộ lọc
func (ac *AnswerController) ListAnswers(c *gin.Context) {
	// Lấy các query parameters để lọc
	filters := make(map[string]interface{})

	questionID := c.Query("question_id")
	if questionID != "" {
		qID, err := strconv.ParseUint(questionID, 10, 64)
		if err == nil {
			filters["question_id"] = uint(qID)
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

	answers, err := ac.answerService.ListAnswers(filters)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list answers"})
		return
	}

	var responseAnswers []responses.AnswerResponse
	for _, answer := range answers {
		// Lấy số lượng vote cho từng câu trả lời
		voteCount, err := ac.voteService.GetVoteCount("answer", answer.ID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get vote count"})
			return
		}

		responseAnswers = append(responseAnswers, responses.ToAnswerResponse(&answer, voteCount))
	}

	c.JSON(http.StatusOK, gin.H{
		"answers": responseAnswers,
	})
}
