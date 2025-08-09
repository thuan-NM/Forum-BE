package controllers

import (
	"Forum_BE/services"
	"github.com/gin-gonic/gin"
	"net/http"
)

type DashboardController struct {
	dashService services.DashboardService
}

func NewDashboardController(s services.DashboardService) *DashboardController {
	return &DashboardController{dashService: s}
}

func (dc *DashboardController) GetDashboard(c *gin.Context) {
	period := c.Query("period") // "7days", "30days", "90days"
	if period == "" {
		period = "7days"
	}

	data, err := dc.dashService.GetDashboardData(period)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load dashboard data"})
		return
	}

	c.JSON(http.StatusOK, data)
}
