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

// FollowUser xử lý yêu cầu theo dõi người dùng
func (fc *FollowController) FollowUser(c *gin.Context) {
	var req struct {
		FollowingID uint `json:"following_id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	followerID := c.GetUint("user_id") // Middleware đã thêm user_id vào context

	follow, err := fc.followService.FollowUser(followerID, req.FollowingID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Followed successfully",
		"follow":  responses.ToFollowResponse(follow),
	})
}

// UnfollowUser xử lý yêu cầu hủy theo dõi người dùng
func (fc *FollowController) UnfollowUser(c *gin.Context) {
	followingIDParam := c.Param("following_id")
	followingID, err := strconv.ParseUint(followingIDParam, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid following_id"})
		return
	}

	followerID := c.GetUint("user_id") // Middleware đã thêm user_id vào context

	if err := fc.followService.UnfollowUser(followerID, uint(followingID)); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Unfollowed successfully",
	})
}

// GetFollowers lấy danh sách người theo dõi
func (fc *FollowController) GetFollowers(c *gin.Context) {
	userIDParam := c.Param("user_id")
	userID, err := strconv.ParseUint(userIDParam, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user_id"})
		return
	}

	followers, err := fc.followService.GetFollowers(uint(userID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get followers"})
		return
	}

	var responseFollowers []responses.FollowResponse
	for _, follow := range followers {
		responseFollowers = append(responseFollowers, responses.ToFollowResponse(&follow))
	}

	c.JSON(http.StatusOK, gin.H{
		"followers": responseFollowers,
	})
}

// GetFollowing lấy danh sách người đang theo dõi
func (fc *FollowController) GetFollowing(c *gin.Context) {
	userIDParam := c.Param("user_id")
	userID, err := strconv.ParseUint(userIDParam, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user_id"})
		return
	}

	followings, err := fc.followService.GetFollowing(uint(userID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get following"})
		return
	}

	var responseFollowings []responses.FollowResponse
	for _, follow := range followings {
		responseFollowings = append(responseFollowings, responses.ToFollowResponse(&follow))
	}

	c.JSON(http.StatusOK, gin.H{
		"following": responseFollowings,
	})
}
