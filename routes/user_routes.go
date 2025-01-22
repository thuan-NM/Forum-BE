package routes

import (
	"Forum_BE/controllers"
	"Forum_BE/middlewares"
	"Forum_BE/repositories"
	"Forum_BE/services"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func UserRoutes(db *gorm.DB, authorized *gin.RouterGroup, permService services.PermissionService) {
	// User routes
	userRepo := repositories.NewUserRepository(db)
	userService := services.NewUserService(userRepo)
	userController := controllers.NewUserController(userService)

	users := authorized.Group("/users")
	{
		users.POST("/", middlewares.CheckPermission(permService, "user", "create"), userController.CreateUser)
		users.GET("/:id", middlewares.CheckPermission(permService, "user", "view"), userController.GetUser)
		users.PUT("/:id", middlewares.CheckPermission(permService, "user", "edit"), userController.UpdateUser)
		users.DELETE("/:id", middlewares.CheckPermission(permService, "user", "delete"), userController.DeleteUser)
		users.GET("/", middlewares.CheckPermission(permService, "user", "view"), userController.ListUsers)
	}
}
