package controllers

import (
	"Forum_BE/responses"
	"Forum_BE/services"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"log"
	"net/http"
	"strconv"
)

type TopicController struct {
	topicService  services.TopicService
	followService services.FollowService
}

func NewTopicController(t services.TopicService) *TopicController {
	return &TopicController{topicService: t}
}

func NewTopicControllerWithDB(db *gorm.DB, t services.TopicService) *TopicController {
	return &TopicController{topicService: t}
}

func (tc *TopicController) CreateTopic(c *gin.Context) {
	var req struct {
		Name        string `json:"name" binding:"required"`
		Description string `json:"description"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	topic, err := tc.topicService.CreateTopic(req.Name, req.Description)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Topic created successfully",
		"topic":   responses.ToTopicResponse(topic),
	})
}

func (tc *TopicController) ProposeTopic(c *gin.Context) {
	var req struct {
		Name        string `json:"name" binding:"required"`
		Description string `json:"description"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID := c.GetUint("user_id")

	topic, err := tc.topicService.ProposeTopic(req.Name, req.Description, userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Topic proposed successfully, pending approval",
		"topic":   responses.ToTopicResponse(topic),
	})
}

func (tc *TopicController) GetTopic(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.ParseUint(idParam, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid topic id"})
		return
	}

	topic, err := tc.topicService.GetTopicByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "topic not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"topic": responses.ToTopicResponse(topic),
	})
}

func (tc *TopicController) UpdateTopic(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.ParseUint(idParam, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid topic id"})
		return
	}

	var req struct {
		Name        string `json:"name"`
		Description string `json:"description"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	topic, err := tc.topicService.UpdateTopic(uint(id), req.Name, req.Description)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Topic updated successfully",
		"topic":   responses.ToTopicResponse(topic),
	})
}

func (tc *TopicController) DeleteTopic(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.ParseUint(idParam, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid topic id"})
		return
	}

	if err := tc.topicService.DeleteTopic(uint(id)); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Topic deleted successfully",
	})
}

func (tc *TopicController) ListTopics(c *gin.Context) {
	filters := make(map[string]interface{})

	if status := c.Query("status"); status != "" {
		filters["status"] = status
	}
	if search := c.Query("search"); search != "" {
		filters["search"] = search
	}
	if sort := c.Query("sort"); sort != "" {
		if sort == "asc" || sort == "desc" {
			filters["sort"] = sort
			log.Printf("Sort parameter received: %s", sort)
		} else {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Giá trị sort không hợp lệ, chỉ chấp nhận 'asc' hoặc 'desc'"})
			return
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

	topics, total, err := tc.topicService.ListTopics(filters)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list topics"})
		return
	}

	var responseTopics []responses.TopicResponse
	for _, topic := range topics {
		responseTopics = append(responseTopics, responses.ToTopicResponse(&topic))
	}

	c.JSON(http.StatusOK, gin.H{
		"topics": responseTopics,
		"total":  total,
	})
}

func (tc *TopicController) AddQuestionToTopic(c *gin.Context) {
	questionIDParam := c.Param("question_id")
	questionID, err := strconv.ParseUint(questionIDParam, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid question id"})
		return
	}

	topicIDParam := c.Param("topic_id")
	topicID, err := strconv.ParseUint(topicIDParam, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid topic id"})
		return
	}

	if err := tc.topicService.AddQuestionToTopic(uint(questionID), uint(topicID)); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Question added to topic successfully",
	})
}

func (tc *TopicController) RemoveQuestionFromTopic(c *gin.Context) {
	questionIDParam := c.Param("question_id")
	questionID, err := strconv.ParseUint(questionIDParam, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid question id"})
		return
	}

	topicIDParam := c.Param("topic_id")
	topicID, err := strconv.ParseUint(topicIDParam, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid topic id"})
		return
	}

	if err := tc.topicService.RemoveQuestionFromTopic(uint(questionID), uint(topicID)); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Question removed from topic successfully",
	})
}
