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

func FollowRoutes(db *gorm.DB, authorized *gin.RouterGroup, permService services.PermissionService, redisClient *redis.Client) {
	// Follow routes
	followRepo := repositories.NewFollowRepository(db)
	followService := services.NewFollowService(followRepo, redisClient)
	followController := controllers.NewFollowController(followService)

	follows := authorized.Group("/follows")
	{
		follows.GET("/:id/follow-status", middlewares.CheckPermission(permService, "follow", "view"), followController.CheckFollowStatus)
		follows.GET("/:id/followers", middlewares.CheckPermission(permService, "follow", "view"), followController.GetFollowers)
		follows.PUT("/:id/follow", middlewares.CheckPermission(permService, "follow", "create"), followController.FollowQuestion)
		follows.DELETE("/:id/unfollow", middlewares.CheckPermission(permService, "follow", "delete"), followController.UnfollowQuestion)
	}
}
