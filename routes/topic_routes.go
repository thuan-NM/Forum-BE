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

func TopicRoutes(db *gorm.DB, authorized *gin.RouterGroup, permService services.PermissionService, redisClient *redis.Client) {
	topicRepo := repositories.NewTopicRepository(db)
	topicService := services.NewTopicService(topicRepo, redisClient)
	topicController := controllers.NewTopicController(topicService)

	topics := authorized.Group("/topics")
	{
		// Người dùng đề xuất Topic
		topics.POST("/propose", middlewares.CheckPermission(permService, "topic", "propose"), topicController.ProposeTopic)
		// Admin tạo Topic trực tiếp
		topics.POST("/", middlewares.CheckPermission(permService, "topic", "create"), topicController.CreateTopic)
		topics.GET("/:id", middlewares.CheckPermission(permService, "topic", "view"), topicController.GetTopic)
		topics.PUT("/:id", middlewares.CheckPermission(permService, "topic", "edit"), topicController.UpdateTopic)
		topics.DELETE("/:id", middlewares.CheckPermission(permService, "topic", "delete"), topicController.DeleteTopic)
		topics.GET("/", middlewares.CheckPermission(permService, "topic", "view"), topicController.ListTopics)

		topics.POST("/:id/questions/:question_id", middlewares.CheckPermission(permService, "topic", "edit"), topicController.AddQuestionToTopic)
		topics.DELETE("/:id/questions/:question_id", middlewares.CheckPermission(permService, "topic", "edit"), topicController.RemoveQuestionFromTopic)

		topics.POST("/:id/follow", middlewares.CheckPermission(permService, "topic", "follow"), topicController.FollowTopic)
		topics.POST("/:id/unfollow", middlewares.CheckPermission(permService, "topic", "follow"), topicController.UnfollowTopic)
	}
}
