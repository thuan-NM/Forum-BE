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

func AnswerRoutes(db *gorm.DB, authorized *gin.RouterGroup, permService services.PermissionService, redisClient *redis.Client) {
	// Answer routes
	voteRepo := repositories.NewVoteRepository(db)
	voteService := services.NewVoteService(voteRepo)
	topicRepo := repositories.NewTopicRepository(db)
	topicService := services.NewTopicService(topicRepo, redisClient)
	questionRepo := repositories.NewQuestionRepository(db)
	questionService := services.NewQuestionService(questionRepo, topicService, redisClient)
	answerRepo := repositories.NewAnswerRepository(db)
	answerService := services.NewAnswerService(answerRepo, questionRepo, questionService, redisClient)
	answerController := controllers.NewAnswerController(answerService, voteService)

	answers := authorized.Group("/answers")
	{
		answers.POST("/", middlewares.CheckPermission(permService, "answer", "create"), answerController.CreateAnswer)
		answers.GET("/:id", middlewares.CheckPermission(permService, "answer", "view"), answerController.GetAnswer)
		answers.PUT("/:id", middlewares.CheckPermission(permService, "answer", "edit"), answerController.EditAnswer)
		answers.DELETE("/:id", middlewares.CheckPermission(permService, "answer", "delete"), answerController.DeleteAnswer)
		answers.GET("/", middlewares.CheckPermission(permService, "answer", "view"), answerController.ListAnswers)
	}
}
