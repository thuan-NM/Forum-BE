package routes

import (
	"Forum_BE/controllers"
	"Forum_BE/middlewares"
	"Forum_BE/notification"
	"Forum_BE/notification"
	"Forum_BE/repositories"
	"Forum_BE/services"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"gorm.io/gorm"
)

func FollowRoutes(db *gorm.DB, authorized *gin.RouterGroup, permService services.PermissionService, redisClient *redis.Client, novuClient *notification.NovuClient) {
	topicFollowRepo := repositories.NewTopicFollowRepository(db)
	questionFollowRepo := repositories.NewQuestionFollowRepository(db)
	userFollowRepo := repositories.NewUserFollowRepository(db)
	userRepo := repositories.NewUserRepository(db)
	followService := services.NewFollowService(topicFollowRepo, questionFollowRepo, userFollowRepo, userRepo, redisClient, db, novuClient)
func FollowRoutes(db *gorm.DB, authorized *gin.RouterGroup, permService services.PermissionService, redisClient *redis.Client, novuClient *notification.NovuClient) {
	topicFollowRepo := repositories.NewTopicFollowRepository(db)
	questionFollowRepo := repositories.NewQuestionFollowRepository(db)
	userFollowRepo := repositories.NewUserFollowRepository(db)
	userRepo := repositories.NewUserRepository(db)
	followService := services.NewFollowService(topicFollowRepo, questionFollowRepo, userFollowRepo, userRepo, redisClient, db, novuClient)
	followController := controllers.NewFollowController(followService)

	follows := authorized.Group("/follows")
	{
		follows.POST("/topics/:id/follow", middlewares.CheckPermission(permService, "follow", "create"), followController.FollowTopic)
		follows.DELETE("/topics/:id/unfollow", middlewares.CheckPermission(permService, "follow", "delete"), followController.UnfollowTopic)
		follows.GET("/topics/:id/follows", middlewares.CheckPermission(permService, "follow", "view"), followController.GetTopicFollows)
		follows.GET("/topics/:id/status", middlewares.CheckPermission(permService, "follow", "view"), followController.GetTopicFollowStatus)

		follows.POST("/questions/:id/follow", middlewares.CheckPermission(permService, "follow", "create"), followController.FollowQuestion)
		follows.DELETE("/questions/:id/unfollow", middlewares.CheckPermission(permService, "follow", "delete"), followController.UnfollowQuestion)
		follows.GET("/questions/:id/follows", middlewares.CheckPermission(permService, "follow", "view"), followController.GetQuestionFollows)
		follows.GET("/questions/:id/status", middlewares.CheckPermission(permService, "follow", "view"), followController.GetQuestionFollowStatus)

		follows.POST("/users/:id/follow", middlewares.CheckPermission(permService, "follow", "create"), followController.FollowUser)
		follows.DELETE("/users/:id/unfollow", middlewares.CheckPermission(permService, "follow", "delete"), followController.UnfollowUser)
		follows.GET("/users/:id/follows", middlewares.CheckPermission(permService, "follow", "view"), followController.GetUserFollows)
		follows.GET("/users/:id/status", middlewares.CheckPermission(permService, "follow", "view"), followController.GetUserFollowStatus)

		follows.GET("/me/topics", middlewares.CheckPermission(permService, "follow", "view"), followController.GetFollowedTopics)
		follows.GET("/me/user/following", middlewares.CheckPermission(permService, "follow", "view"), followController.GetFollowingUsers)
		follows.GET("/me/user/followed", middlewares.CheckPermission(permService, "follow", "view"), followController.GetFollowedUsers)
	}
}
