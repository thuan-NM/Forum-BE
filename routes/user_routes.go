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

func UserRoutes(db *gorm.DB, authorized *gin.RouterGroup, permService services.PermissionService, redisClient *redis.Client) {
	userRepo := repositories.NewUserRepository(db)
	userService := services.NewUserService(userRepo, redisClient)
	userController := controllers.NewUserController(userService)

	users := authorized.Group("/users")
	{
		users.POST("/", middlewares.CheckPermission(permService, "user", "create"), userController.CreateUser)
		users.GET("/:id", middlewares.CheckPermission(permService, "user", "view"), userController.GetUser)
		users.PUT("/:id", middlewares.CheckPermission(permService, "user", "edit"), userController.UpdateUser)
		users.DELETE("/:id", middlewares.CheckPermission(permService, "user", "delete"), userController.DeleteUser)
		users.GET("/", middlewares.CheckPermission(permService, "user", "view"), userController.GetAllUsers)
		users.PUT("/:id/status", middlewares.CheckPermission(permService, "user", "edit"), userController.ModifyUserStatus)
	}
}
