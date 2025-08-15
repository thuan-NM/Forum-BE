package routes

import (
	"Forum_BE/controllers"
	"Forum_BE/middlewares"
	"Forum_BE/notification"
	"Forum_BE/repositories"
	"Forum_BE/services"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"gorm.io/gorm"
)

func QuestionRoutes(db *gorm.DB, authorized *gin.RouterGroup, permService services.PermissionService, redisClient *redis.Client, novuClient *notification.NovuClient) {
	topicRepo := repositories.NewTopicRepository(db)
	questionRepo := repositories.NewQuestionRepository(db)
	userRepo := repositories.NewUserRepository(db)
	topicService := services.NewTopicService(topicRepo, redisClient, db)
	questionService := services.NewQuestionService(questionRepo, topicService, redisClient, userRepo, novuClient)

	questionController := controllers.NewQuestionController(questionService)

	questions := authorized.Group("/questions")
	{
		questions.POST("/", middlewares.CheckPermission(permService, "question", "create"), questionController.CreateQuestion)
		questions.GET("/:id", middlewares.CheckPermission(permService, "question", "view"), questionController.GetQuestion)
		questions.PUT("/:id", middlewares.CheckPermission(permService, "question", "edit"), questionController.UpdateQuestion)
		questions.DELETE("/:id", middlewares.CheckPermission(permService, "question", "delete"), questionController.DeleteQuestion)
		questions.GET("/", middlewares.CheckPermission(permService, "question", "view"), questionController.ListQuestions)
		questions.GET("/all", middlewares.CheckPermission(permService, "question", "view"), questionController.GetAllQuestion)
		questions.GET("/suggest", middlewares.CheckPermission(permService, "question", "view"), questionController.SuggestQuestions)
		questions.PUT("/:id/status", middlewares.CheckPermission(permService, "question", "change_status"), questionController.UpdateQuestionStatus)
		questions.PUT("/:id/interaction-status", middlewares.CheckPermission(permService, "question", "change_inter_status"), questionController.UpdateInteractionStatus)
		questions.POST("/sync", middlewares.CheckPermission(permService, "question", "create"), questionController.SyncQuestionsToRAG)
	}
}
