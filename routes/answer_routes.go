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

func AnswerRoutes(db *gorm.DB, authorized *gin.RouterGroup, permService services.PermissionService, redisClient *redis.Client, novuClient *notification.NovuClient) {
	topicRepo := repositories.NewTopicRepository(db)
	topicService := services.NewTopicService(topicRepo, redisClient, db)
	questionRepo := repositories.NewQuestionRepository(db)
	questionService := services.NewQuestionService(questionRepo, topicService, redisClient)
	answerRepo := repositories.NewAnswerRepository(db)
	userRepo := repositories.NewUserRepository(db)
	answerService := services.NewAnswerService(answerRepo, questionRepo, questionService, userRepo, redisClient, novuClient)
	answerController := controllers.NewAnswerController(answerService)

	answers := authorized.Group("/answers")
	{
		answers.POST("/", middlewares.CheckPermission(permService, "answer", "create"), answerController.CreateAnswer)
		answers.GET("/:id", middlewares.CheckPermission(permService, "answer", "view"), answerController.GetAnswer)
		answers.PUT("/:id", middlewares.CheckPermission(permService, "answer", "edit"), answerController.EditAnswer)
		answers.DELETE("/:id", middlewares.CheckPermission(permService, "answer", "delete"), answerController.DeleteAnswer)
		answers.GET("/questions", middlewares.CheckPermission(permService, "answer", "view"), answerController.ListAnswers)
		answers.GET("/", middlewares.CheckPermission(permService, "answer", "view"), answerController.GetAllAnswers)
		answers.PUT("/:id/status", middlewares.CheckPermission(permService, "answer", "edit_status"), answerController.UpdateAnswerStatus)
		answers.PUT("/:id/accept", middlewares.CheckPermission(permService, "answer", "accept"), answerController.AcceptAnswer)
	}
}
