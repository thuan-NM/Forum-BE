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
}

func NewAnswerController(a services.AnswerService) *AnswerController {
	return &AnswerController{answerService: a}
}

func (ac *AnswerController) CreateAnswer(c *gin.Context) {
	var req struct {
		Content    string `json:"content" binding:"required"`
		QuestionID uint   `json:"questionId" binding:"required"`
		Tags       []uint `json:"tags"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID := c.GetUint("user_id") // Middleware đã thêm user_id vào context

	answer, err := ac.answerService.CreateAnswer(req.Content, userID, req.QuestionID, req.Tags)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Answer created successfully",
		"answer":  responses.ToAnswerResponse(answer),
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

	c.JSON(http.StatusOK, gin.H{
		"answer": responses.ToAnswerResponse(answer),
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
		"answer":  responses.ToAnswerResponse(answer),
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

func (ac *AnswerController) ListAnswers(c *gin.Context) {
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

	limitStr := c.Query("limit")
	pageStr := c.Query("page")
	limit := 10 // Default limit
	page := 1   // Default page
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}
	if pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}
	filters["limit"] = limit
	filters["page"] = page

	answers, total, err := ac.answerService.ListAnswers(filters)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list answers"})
		return
	}

	var responseAnswers []responses.AnswerResponse
	for _, answer := range answers {
		responseAnswers = append(responseAnswers, responses.ToAnswerResponse(&answer))
	}

	c.JSON(http.StatusOK, gin.H{
		"answers": responseAnswers,
		"total":   total,
	})
}

func (ac *AnswerController) GetAllAnswers(c *gin.Context) {
	filters := make(map[string]interface{})

	if search := c.Query("search"); search != "" {
		filters["search"] = search
	}
	if status := c.Query("status"); status != "" {
		filters["status"] = status
	}
	if questiontitle := c.Query("questiontitle"); questiontitle != "" {
		filters["questiontitle"] = questiontitle
	}
	if page := c.Query("page"); page != "" {
		if p, err := strconv.Atoi(page); err == nil {
			filters["page"] = p
		}
	}
	if limit := c.Query("limit"); limit != "" {
		if l, err := strconv.Atoi(limit); err == nil {
			filters["limit"] = l
		}
	}

	answers, total, err := ac.answerService.GetAllAnswers(filters)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	var responseAnswers []responses.AnswerResponse
	for _, answer := range answers {
		responseAnswers = append(responseAnswers, responses.ToAnswerResponse(&answer))
	}

	c.JSON(http.StatusOK, gin.H{
		"answers": responseAnswers,
		"total":   total,
	})
}
func (ac *AnswerController) UpdateAnswerStatus(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.ParseUint(idParam, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid answer id"})
		return
	}

	var req struct {
		Status string `json:"status" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	answer, err := ac.answerService.UpdateAnswerStatus(uint(id), req.Status)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Answer status updated successfully",
		"answer":  responses.ToAnswerResponse(answer),
	})
}

func (ac *AnswerController) AcceptAnswer(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.ParseUint(idParam, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid answer id"})
		return
	}

	userID := c.GetUint("user_id") // Lấy từ middleware

	answer, err := ac.answerService.AcceptAnswer(uint(id), userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Answer accepted successfully",
		"answer":  responses.ToAnswerResponse(answer),
	})
}
