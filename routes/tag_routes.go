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

func TagRoutes(db *gorm.DB, authorized *gin.RouterGroup, permService services.PermissionService, redisClient *redis.Client) {
	tagRepo := repositories.NewTagRepository(db)
	tagService := services.NewTagService(tagRepo, redisClient)
	tagController := controllers.NewTagController(tagService)

	tags := authorized.Group("/tags")
	{
		tags.POST("/", middlewares.CheckPermission(permService, "tag", "create"), tagController.CreateTag)
		tags.GET("/:id", middlewares.CheckPermission(permService, "tag", "view"), tagController.GetTag)
		tags.PUT("/:id", middlewares.CheckPermission(permService, "tag", "edit"), tagController.EditTag)
		tags.DELETE("/:id", middlewares.CheckPermission(permService, "tag", "delete"), tagController.DeleteTag)
		tags.GET("/", middlewares.CheckPermission(permService, "tag", "view"), tagController.ListTags)
		tags.GET("/post/:post_id", middlewares.CheckPermission(permService, "tag", "view"), tagController.GetTagsByPostID)
		tags.GET("/answer/:answer_id", middlewares.CheckPermission(permService, "tag", "view"), tagController.GetTagsByAnswerID)
	}
}
