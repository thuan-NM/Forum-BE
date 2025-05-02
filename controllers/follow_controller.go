package controllers

import (
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

func (fc *FollowController) FollowQuestion(c *gin.Context) {
	questionIDParam := c.Param("id")
	questionID, err := strconv.ParseUint(questionIDParam, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid question id"})
		return
	}

	userID := c.GetUint("user_id") // Assume middleware sets user_id

	if err := fc.followService.FollowQuestion(userID, uint(questionID)); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Question followed successfully",
	})
}

func (fc *FollowController) UnfollowQuestion(c *gin.Context) {
	questionIDParam := c.Param("id")
	questionID, err := strconv.ParseUint(questionIDParam, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid question id"})
		return
	}

	userID := c.GetUint("user_id")

	if err := fc.followService.UnfollowQuestion(userID, uint(questionID)); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Question unfollowed successfully",
	})
}
func (fc *FollowController) GetFollowers(c *gin.Context) {
	questionIDParam := c.Param("id")
	questionID, err := strconv.ParseUint(questionIDParam, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid question id"})
		return
	}

	follows, err := fc.followService.GetFollowsByQuestionID(uint(questionID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get followers"})
		return
	}

	var followers []struct {
		UserID   uint   `json:"user_id"`
		Username string `json:"username"`
	}
	for _, f := range follows {
		followers = append(followers, struct {
			UserID   uint   `json:"user_id"`
			Username string `json:"username"`
		}{
			UserID:   f.UserID,
			Username: f.User.Username, // Assume User is preloaded
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"followers": followers,
	})
}
func (fc *FollowController) CheckFollowStatus(c *gin.Context) {
	questionIDParam := c.Param("id")
	questionID, err := strconv.ParseUint(questionIDParam, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid question id"})
		return
	}

	userID := c.GetUint("user_id")

	follows, err := fc.followService.GetFollowsByQuestionID(uint(questionID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to check follow status"})
		return
	}

	isFollowing := false
	for _, f := range follows {
		if f.UserID == userID {
			isFollowing = true
			break
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"isFollowing": isFollowing,
	})
}
