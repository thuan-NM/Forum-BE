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

func GroupRoutes(db *gorm.DB, authorized *gin.RouterGroup, permService services.PermissionService, redisClient *redis.Client) {
	// Group routes
	groupRepo := repositories.NewGroupRepository(db)
	groupService := services.NewGroupService(groupRepo, redisClient)
	groupController := controllers.NewGroupController(groupService)

	groups := authorized.Group("/groups")
	{
		groups.POST("/", middlewares.CheckPermission(permService, "group", "create"), groupController.CreateGroup)
		groups.GET("/:id", middlewares.CheckPermission(permService, "group", "view"), groupController.GetGroup)
		groups.PUT("/:id", middlewares.CheckPermission(permService, "group", "edit"), groupController.EditGroup)
		groups.DELETE("/:id", middlewares.CheckPermission(permService, "group", "delete"), groupController.DeleteGroup)
		groups.GET("/", middlewares.CheckPermission(permService, "group", "view"), groupController.ListGroups)
	}
}
