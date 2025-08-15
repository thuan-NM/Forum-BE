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
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID chủ đề không hợp lệ"})
		return
	}

	if err := fc.followService.FollowTopic(userID, uint(id)); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Theo dõi chủ đề thành công"})
}

func (fc *FollowController) UnfollowTopic(c *gin.Context) {
	userID := c.GetUint("user_id")
	idParam := c.Param("id")
	id, err := strconv.ParseUint(idParam, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID chủ đề không hợp lệ"})
		return
	}

	if err := fc.followService.UnfollowTopic(userID, uint(id)); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Huỷ theo dõi chủ đề thành công"})
}

func (fc *FollowController) FollowQuestion(c *gin.Context) {
	userID := c.GetUint("user_id")
	idParam := c.Param("id")
	id, err := strconv.ParseUint(idParam, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID câu hỏi không hợp lệ"})
		return
	}

	if err := fc.followService.FollowQuestion(userID, uint(id)); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Theo dõi chủ đề thành công"})
}

func (fc *FollowController) UnfollowQuestion(c *gin.Context) {
	userID := c.GetUint("user_id")
	idParam := c.Param("id")
	id, err := strconv.ParseUint(idParam, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID câu hỏi không hợp lệ"})
		return
	}

	if err := fc.followService.UnfollowQuestion(userID, uint(id)); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Huỷ theo dõi chủ đề thành công"})
}

func (fc *FollowController) GetQuestionFollowStatus(c *gin.Context) {
	userID := c.GetUint("user_id")
	idParam := c.Param("id")
	id, err := strconv.ParseUint(idParam, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID câu hỏi không hợp lệ"})
		return
	}

	isFollowing, err := fc.followService.GetQuestionFollowStatus(userID, uint(id))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"isFollowing": isFollowing,
		"message":     "Khôi phục trạng thái theo dõi thành công",
	})
}

func (fc *FollowController) GetTopicFollowStatus(c *gin.Context) {
	userID := c.GetUint("user_id")
	idParam := c.Param("id")
	id, err := strconv.ParseUint(idParam, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID chủ đề không hợp lệ"})
		return
	}

	isFollowing, err := fc.followService.GetTopicFollowStatus(userID, uint(id))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"isFollowing": isFollowing,
		"message":     "Khôi phục trạng thái theo dõi thành công",
	})
}

func (fc *FollowController) GetUserFollowStatus(c *gin.Context) {
	userID := c.GetUint("user_id")
	idParam := c.Param("id")
	id, err := strconv.ParseUint(idParam, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID người dùng không hợp lệ"})
		return
	}

	isFollowing, err := fc.followService.GetUserFollowStatus(userID, uint(id))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"isFollowing": isFollowing,
		"message":     "Khôi phục trạng thái theo dõi thành công",
	})
}

func (fc *FollowController) FollowUser(c *gin.Context) {
	userID := c.GetUint("user_id")
	idParam := c.Param("id")
	id, err := strconv.ParseUint(idParam, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID người dùng không hợp lệ"})
		return
	}

	if err := fc.followService.FollowUser(userID, uint(id)); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Theo dõi người dùng thành công"})
}

func (fc *FollowController) UnfollowUser(c *gin.Context) {
	userID := c.GetUint("user_id")
	idParam := c.Param("id")
	id, err := strconv.ParseUint(idParam, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID người dùng không hợp lệ"})
		return
	}

	if err := fc.followService.UnfollowUser(userID, uint(id)); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Huỷ theo dõi người dùng thành công"})
}

func (fc *FollowController) GetTopicFollows(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.ParseUint(idParam, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID chủ đề không hợp lệ"})
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
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID câu hỏi không hợp lệ"})
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
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID người dùng không hợp lệ"})
		return
	}

	follows, err := fc.followService.GetUserFollows(uint(id))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"follows": follows})
}

func (fc *FollowController) GetFollowedUsers(c *gin.Context) {
	userID := c.GetUint("user_id")

	users, err := fc.followService.GetFollowedUsers(userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var responseUsers []responses.UserResponse
	for _, user := range users {
		responseUsers = append(responseUsers, responses.ToUserResponse(&user))
	}

	c.JSON(http.StatusOK, gin.H{"users": responseUsers})
}
func (fc *FollowController) GetFollowingUsers(c *gin.Context) {
	userID := c.GetUint("user_id")

	users, err := fc.followService.GetFollowingUsers(userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var responseUsers []responses.UserResponse
	for _, user := range users {
		responseUsers = append(responseUsers, responses.ToUserResponse(&user))
	}

	c.JSON(http.StatusOK, gin.H{"users": responseUsers})
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
