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

func QuestionRoutes(db *gorm.DB, authorized *gin.RouterGroup, permService services.PermissionService, redisClient *redis.Client) {
	// Question routes
	voteRepo := repositories.NewVoteRepository(db)
	voteService := services.NewVoteService(voteRepo)
	topicRepo := repositories.NewTopicRepository(db)
	topicService := services.NewTopicService(topicRepo, redisClient)
	questionRepo := repositories.NewQuestionRepository(db)
	questionService := services.NewQuestionService(questionRepo, topicService, redisClient)
	questionController := controllers.NewQuestionController(questionService, voteService)

	questions := authorized.Group("/questions")
	{
		questions.POST("/", middlewares.CheckPermission(permService, "question", "create"), questionController.CreateQuestion)
		questions.GET("/:id", middlewares.CheckPermission(permService, "question", "view"), questionController.GetQuestion)
		questions.PUT("/:id", middlewares.CheckPermission(permService, "question", "edit"), questionController.UpdateQuestion)
		questions.DELETE("/:id", middlewares.CheckPermission(permService, "question", "delete"), questionController.DeleteQuestion)
		questions.GET("/", middlewares.CheckPermission(permService, "question", "view"), questionController.ListQuestions)
		questions.GET("/suggest", middlewares.CheckPermission(permService, "question", "view"), questionController.SuggestQuestions)
		// Approval routes
		questions.POST("/:id/approve", middlewares.CheckPermission(permService, "question", "approve"), questionController.ApproveQuestion)
		questions.POST("/:id/reject", middlewares.CheckPermission(permService, "question", "reject"), questionController.RejectQuestion)
	}
}
