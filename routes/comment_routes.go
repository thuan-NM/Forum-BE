package routes

import (
	"Forum_BE/controllers"
	"Forum_BE/middlewares"
	"Forum_BE/repositories"
	"Forum_BE/services"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func CommentRoutes(db *gorm.DB, authorized *gin.RouterGroup, permService services.PermissionService) {
	// Comment routes
	voteRepo := repositories.NewVoteRepository(db)
	voteService := services.NewVoteService(voteRepo)
	questionRepo := repositories.NewQuestionRepository(db)
	answerRepo := repositories.NewAnswerRepository(db)
	commentRepo := repositories.NewCommentRepository(db)
	commentService := services.NewCommentService(commentRepo, questionRepo, answerRepo)
	commentController := controllers.NewCommentController(commentService, voteService)

	comments := authorized.Group("/comments")
	{
		comments.POST("/", middlewares.CheckPermission(permService, "comment", "create"), commentController.CreateComment)
		comments.GET("/:id", middlewares.CheckPermission(permService, "comment", "view"), commentController.GetComment)
		comments.PUT("/:id", middlewares.CheckPermission(permService, "comment", "edit"), commentController.EditComment)
		comments.DELETE("/:id", middlewares.CheckPermission(permService, "comment", "delete"), commentController.DeleteComment)
		comments.GET("/", middlewares.CheckPermission(permService, "comment", "view"), commentController.ListComments)
	}
}
