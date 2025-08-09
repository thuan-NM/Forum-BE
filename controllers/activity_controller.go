package controllers

import (
	"Forum_BE/responses"
	"Forum_BE/services"
	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
)

type ActivityController struct {
	activityService services.ActivityService
}

func NewActivityController(s services.ActivityService) *ActivityController {
	return &ActivityController{activityService: s}
}

func (ac *ActivityController) GetRecentActivities(c *gin.Context) {
	limitStr := c.Query("limit")
	limit, _ := strconv.Atoi(limitStr)
	if limit <= 0 {
		limit = 20
	}

	activities, err := ac.activityService.GetRecentActivities(limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch recent activities"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"activities": responses.ToActivityResponses(activities),
	})
}
