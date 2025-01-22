package routes

import (
	"Forum_BE/controllers"
	"Forum_BE/middlewares"
	"Forum_BE/repositories"
	"Forum_BE/services"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func AnswerRoutes(db *gorm.DB, authorized *gin.RouterGroup, permService services.PermissionService) {
	// Answer routes
	voteRepo := repositories.NewVoteRepository(db)
	voteService := services.NewVoteService(voteRepo)
	questionRepo := repositories.NewQuestionRepository(db)
	answerRepo := repositories.NewAnswerRepository(db)
	answerService := services.NewAnswerService(answerRepo, questionRepo)
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
