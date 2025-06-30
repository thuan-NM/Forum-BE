package controllers

import (
	"Forum_BE/responses"
	"Forum_BE/services"
	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
)

type FollowController struct {
	followService services.FollowService
}

func NewFollowController(f services.FollowService) *FollowController {
	return &FollowController{followService: f}
}

func (fc *FollowController) FollowTopic(c *gin.Context) {
	userID := c.GetUint("user_id")
	idParam := c.Param("id")
	id, err := strconv.ParseUint(idParam, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid topic id"})
		return
	}

	if err := fc.followService.FollowTopic(userID, uint(id)); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Followed topic successfully"})
}

func (fc *FollowController) UnfollowTopic(c *gin.Context) {
	userID := c.GetUint("user_id")
	idParam := c.Param("id")
	id, err := strconv.ParseUint(idParam, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid topic id"})
		return
	}

	if err := fc.followService.UnfollowTopic(userID, uint(id)); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Unfollowed topic successfully"})
}

func (fc *FollowController) FollowQuestion(c *gin.Context) {
	userID := c.GetUint("user_id")
	idParam := c.Param("id")
	id, err := strconv.ParseUint(idParam, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid question id"})
		return
	}

	if err := fc.followService.FollowQuestion(userID, uint(id)); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Followed question successfully"})
}

func (fc *FollowController) UnfollowQuestion(c *gin.Context) {
	userID := c.GetUint("user_id")
	idParam := c.Param("id")
	id, err := strconv.ParseUint(idParam, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid question id"})
		return
	}

	if err := fc.followService.UnfollowQuestion(userID, uint(id)); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Unfollowed question successfully"})
}

func (fc *FollowController) GetQuestionFollowStatus(c *gin.Context) {
	userID := c.GetUint("user_id")
	idParam := c.Param("id")
	id, err := strconv.ParseUint(idParam, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid question id"})
		return
	}

	isFollowing, err := fc.followService.GetQuestionFollowStatus(userID, uint(id))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"isFollowing": isFollowing,
		"message":     "Follow status retrieved successfully",
	})
}

func (fc *FollowController) FollowUser(c *gin.Context) {
	userID := c.GetUint("user_id")
	idParam := c.Param("id")
	id, err := strconv.ParseUint(idParam, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user id"})
		return
	}

	if err := fc.followService.FollowUser(userID, uint(id)); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Followed user successfully"})
}

func (fc *FollowController) UnfollowUser(c *gin.Context) {
	userID := c.GetUint("user_id")
	idParam := c.Param("id")
	id, err := strconv.ParseUint(idParam, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user id"})
		return
	}

	if err := fc.followService.UnfollowUser(userID, uint(id)); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Unfollowed user successfully"})
}

func (fc *FollowController) GetTopicFollows(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.ParseUint(idParam, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid topic id"})
		return
	}

	follows, err := fc.followService.GetTopicFollows(uint(id))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"follows": follows})
}

func (fc *FollowController) GetQuestionFollows(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.ParseUint(idParam, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid question id"})
		return
	}

	follows, err := fc.followService.GetQuestionFollows(uint(id))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"follows": follows})
}

func (fc *FollowController) GetUserFollows(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.ParseUint(idParam, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user id"})
		return
	}

	follows, err := fc.followService.GetUserFollows(uint(id))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"follows": follows})
}

func (fc *FollowController) GetFollowedTopics(c *gin.Context) {
	userID := c.GetUint("user_id")

	topics, err := fc.followService.GetFollowedTopics(userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var responseTopics []responses.TopicResponse
	for _, topic := range topics {
		responseTopics = append(responseTopics, responses.ToTopicResponse(&topic))
	}

	c.JSON(http.StatusOK, gin.H{"topics": responseTopics})
}
