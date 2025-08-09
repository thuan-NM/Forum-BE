package routes

import (
	"Forum_BE/controllers"
	"Forum_BE/middlewares"
	"Forum_BE/repositories"
	"Forum_BE/services"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"gorm.io/gorm"
)

func RecentActivityRoutes(
	db *gorm.DB,
	authorized *gin.RouterGroup,
	permService services.PermissionService,
	redisClient *redis.Client,
) {
	activityRepo := repositories.NewActivityRepository(db)
	activityService := services.NewActivityService(activityRepo, redisClient)
	activityCtrl := controllers.NewActivityController(activityService)

	group := authorized.Group("/activities")
	{
		group.GET("/recent", middlewares.CheckPermission(permService, "activities", "view"), activityCtrl.GetRecentActivities)
	}
}
