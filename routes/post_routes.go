package routes

import (
	"Forum_BE/controllers"
	"Forum_BE/middlewares"
	"Forum_BE/repositories"
	"Forum_BE/services"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func PostRoutes(db *gorm.DB, authorized *gin.RouterGroup, permService services.PermissionService) {
	// Post routes
	postRepo := repositories.NewPostRepository(db)
	postService := services.NewPostService(postRepo)
	postController := controllers.NewPostController(postService)

	posts := authorized.Group("/posts")
	{
		posts.POST("/", middlewares.CheckPermission(permService, "post", "create"), postController.CreatePost)
		posts.GET("/:id", middlewares.CheckPermission(permService, "post", "view"), postController.GetPostById)
		posts.PUT("/:id", middlewares.CheckPermission(permService, "post", "edit"), postController.UpdatePost)
		posts.DELETE("/:id", middlewares.CheckPermission(permService, "post", "delete"), postController.DeletePost)
		posts.GET("/", middlewares.CheckPermission(permService, "post", "view"), postController.ListPosts)

	}
}
