package routes

import (
	"Forum_BE/controllers"
	"Forum_BE/middlewares"
	"Forum_BE/repositories"
	"Forum_BE/services"
	"github.com/cloudinary/cloudinary-go/v2"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"gorm.io/gorm"
)

func CommentRoutes(db *gorm.DB, authorized *gin.RouterGroup, permService services.PermissionService, redisClient *redis.Client, cld *cloudinary.Cloudinary, uploadPreset string) {
	voteRepo := repositories.NewVoteRepository(db)
	voteService := services.NewVoteService(voteRepo)
	postRepo := repositories.NewPostRepository(db)
	answerRepo := repositories.NewAnswerRepository(db)
	commentRepo := repositories.NewCommentRepository(db)
	fileRepo := repositories.NewFileRepository(db)
	fileService := services.NewFileService(fileRepo, cld, uploadPreset, redisClient, db)
	commentService := services.NewCommentService(commentRepo, postRepo, answerRepo, redisClient, db)
	commentController := controllers.NewCommentController(commentService, voteService, fileService)

	comments := authorized.Group("/comments")
	{
		comments.POST("/", middlewares.CheckPermission(permService, "comment", "create"), commentController.CreateComment)
		comments.GET("/:id", middlewares.CheckPermission(permService, "comment", "view"), commentController.GetComment)
		comments.PUT("/:id", middlewares.CheckPermission(permService, "comment", "edit"), commentController.EditComment)
		comments.DELETE("/:id", middlewares.CheckPermission(permService, "comment", "delete"), commentController.DeleteComment)
		comments.GET("/", middlewares.CheckPermission(permService, "comment", "view"), commentController.ListComments)
		comments.GET("/:id/replies", middlewares.CheckPermission(permService, "comment", "view"), commentController.ListReplies)
		comments.GET("/all", middlewares.CheckPermission(permService, "comment", "view"), commentController.GetAllComments)
		comments.PUT("status/:id", middlewares.CheckPermission(permService, "comment", "edit"), commentController.UpdateStatus)
	}
}
