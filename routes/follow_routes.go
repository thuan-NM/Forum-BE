package routes

import (
	"Forum_BE/controllers"
	"Forum_BE/middlewares"
	"Forum_BE/repositories"
	"Forum_BE/services"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func FollowRoutes(db *gorm.DB, authorized *gin.RouterGroup, permService services.PermissionService) {
	// Follow routes
	followRepo := repositories.NewFollowRepository(db)
	followService := services.NewFollowService(followRepo)
	followController := controllers.NewFollowController(followService)

	follows := authorized.Group("/follows")
	{
		follows.POST("/", middlewares.CheckPermission(permService, "follow", "create"), followController.FollowUser)
		follows.DELETE("/:following_id", middlewares.CheckPermission(permService, "follow", "delete"), followController.UnfollowUser)
		follows.GET("/followers/:user_id", middlewares.CheckPermission(permService, "follow", "view"), followController.GetFollowers)
		follows.GET("/following/:user_id", middlewares.CheckPermission(permService, "follow", "view"), followController.GetFollowing)
	}
}
