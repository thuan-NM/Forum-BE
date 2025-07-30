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
}

func NewQuestionController(q services.QuestionService) *QuestionController {
	return &QuestionController{questionService: q}
}

func (qc *QuestionController) CreateQuestion(c *gin.Context) {
	var req struct {
		Title       string `json:"title" binding:"required"`
		Description string `json:"description"`
		TopicID     uint   `json:"topicId" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID := c.GetUint("user_id")

	question, err := qc.questionService.CreateQuestion(req.Title, req.Description, userID, req.TopicID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":  "Câu hỏi được tạo thành công và đang chờ phê duyệt",
		"question": responses.ToQuestionResponse(question),
	})
}

func (qc *QuestionController) GetQuestion(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.ParseUint(idParam, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID câu hỏi không hợp lệ"})
		return
	}

	question, err := qc.questionService.GetQuestionByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Không tìm thấy câu hỏi"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"question": responses.ToQuestionResponse(question),
	})
}

func (qc *QuestionController) UpdateQuestion(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.ParseUint(idParam, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID câu hỏi không hợp lệ"})
		return
	}

	var req struct {
		Title       string `json:"title"`
		Description string `json:"description"`
		TopicID     uint   `json:"topic_id"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	question, err := qc.questionService.UpdateQuestion(uint(id), req.Title, req.Description, req.TopicID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":  "Câu hỏi được cập nhật thành công",
		"question": responses.ToQuestionResponse(question),
	})
}

func (qc *QuestionController) DeleteQuestion(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.ParseUint(idParam, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID câu hỏi không hợp lệ"})
		return
	}

	if err := qc.questionService.DeleteQuestion(uint(id)); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Câu hỏi đã được xóa thành công",
	})
}

func (qc *QuestionController) ListQuestions(c *gin.Context) {
	filters := make(map[string]interface{})
	userIDRaw, ok := c.Get("user_id")
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Không được phép"})
		return
	}
	userID := userIDRaw.(uint)

	if search := c.Query("search"); search != "" {
		filters["title_search"] = search
	}
	if status := c.Query("status"); status != "" {
		filters["status"] = status
	}
	if interstatus := c.Query("interstatus"); interstatus != "" {
		filters["interstatus"] = interstatus
	}
	if topicID := c.Query("topic_id"); topicID != "" {
		if id, err := strconv.ParseUint(topicID, 10, 64); err == nil {
			filters["topic_id"] = uint(id)
		}
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
	filters["user_id"] = userID

	questions, total, err := qc.questionService.ListQuestions(filters)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Không thể liệt kê danh sách câu hỏi"})
		return
	}

	var responseQuestions []responses.QuestionResponse
	for _, question := range questions {
		responseQuestions = append(responseQuestions, responses.ToQuestionResponse(&question))
	}

	c.JSON(http.StatusOK, gin.H{
		"questions": responseQuestions,
		"total":     total,
	})
}

func (qc *QuestionController) GetAllQuestion(c *gin.Context) {
	filters := make(map[string]interface{})
	if userID := c.Query("user_id"); userID != "" {
		if uID, err := strconv.ParseUint(userID, 10, 64); err == nil {
			filters["user_id"] = uint(uID)
		}
	}

	if search := c.Query("search"); search != "" {
		filters["title_search"] = search
	}
	if status := c.Query("status"); status != "" {
		filters["status"] = status
	}
	if interstatus := c.Query("interstatus"); interstatus != "" {
		filters["interstatus"] = interstatus
	}
	if topicID := c.Query("topic_id"); topicID != "" {
		if id, err := strconv.ParseUint(topicID, 10, 64); err == nil {
			filters["topic_id"] = uint(id)
		}
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

	questions, total, err := qc.questionService.GetAllQuestion(filters)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Không thể liệt kê danh sách câu hỏi"})
		return
	}

	var responseQuestions []responses.QuestionResponse
	for _, question := range questions {
		responseQuestions = append(responseQuestions, responses.ToQuestionResponse(&question))
	}

	c.JSON(http.StatusOK, gin.H{
		"questions": responseQuestions,
		"total":     total,
	})
}
func (qc *QuestionController) UpdateQuestionStatus(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.ParseUint(idParam, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID câu hỏi không hợp lệ"})
		return
	}

	var req struct {
		Status string `json:"status" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	question, err := qc.questionService.UpdateQuestionStatus(uint(id), req.Status)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":  "Trạng thái phê duyệt được cập nhật thành công",
		"question": responses.ToQuestionResponse(question),
	})
}

func (qc *QuestionController) UpdateInteractionStatus(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.ParseUint(idParam, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID câu hỏi không hợp lệ"})
		return
	}

	var req struct {
		InteractionStatus string `json:"interaction_status" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID := c.GetUint("user_id")

	question, err := qc.questionService.UpdateInteractionStatus(uint(id), models.InteractionStatus(req.InteractionStatus), userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":  "Trạng thái tương tác được cập nhật thành công",
		"question": responses.ToQuestionResponse(question),
	})
}

func (qc *QuestionController) SuggestQuestions(c *gin.Context) {
	filters := map[string]interface{}{
		"status": models.StatusApproved,
	}
	userIDRaw, ok := c.Get("user_id")
	if ok {
		filters["user_id"] = userIDRaw.(uint)
	}
	sort := c.Query("sort")
	if sort == "popular" {
		filters["sort"] = "follow_count"
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

	questions, total, err := qc.questionService.ListQuestions(filters)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Không thể gợi ý câu hỏi"})
		return
	}

	var responseQuestions []responses.QuestionResponse
	for _, question := range questions {
		responseQuestions = append(responseQuestions, responses.ToQuestionResponse(&question))
	}

	c.JSON(http.StatusOK, gin.H{
		"questions": responseQuestions,
		"total":     total,
	})
}
