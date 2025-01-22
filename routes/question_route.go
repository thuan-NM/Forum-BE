package routes

import (
	"Forum_BE/controllers"
	"Forum_BE/middlewares"
	"Forum_BE/repositories"
	"Forum_BE/services"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func QuestionRoutes(db *gorm.DB, authorized *gin.RouterGroup, permService services.PermissionService) {
	// Question routes
	voteRepo := repositories.NewVoteRepository(db)
	voteService := services.NewVoteService(voteRepo)
	questionRepo := repositories.NewQuestionRepository(db)
	questionService := services.NewQuestionService(questionRepo)
	questionController := controllers.NewQuestionController(questionService, voteService)

	questions := authorized.Group("/questions")
	{
		questions.POST("/", middlewares.CheckPermission(permService, "question", "create"), questionController.CreateQuestion)
		questions.GET("/:id", middlewares.CheckPermission(permService, "question", "view"), questionController.GetQuestion)
		questions.PUT("/:id", middlewares.CheckPermission(permService, "question", "edit"), questionController.UpdateQuestion)
		questions.DELETE("/:id", middlewares.CheckPermission(permService, "question", "delete"), questionController.DeleteQuestion)
		questions.GET("/", middlewares.CheckPermission(permService, "question", "view"), questionController.ListQuestions)

		// Approval routes
		questions.POST("/:id/approve", middlewares.CheckPermission(permService, "question", "approve"), questionController.ApproveQuestion)
		questions.POST("/:id/reject", middlewares.CheckPermission(permService, "question", "reject"), questionController.RejectQuestion)
	}
}
