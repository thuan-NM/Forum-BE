package routes

import (
	"Forum_BE/controllers"
	"Forum_BE/repositories"
	"Forum_BE/services"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func AuthRoutes(r *gin.Engine, db *gorm.DB, jwtSecret string) {
	userRepo := repositories.NewUserRepository(db)
	userService := services.NewUserService(userRepo)
	authService := services.NewAuthService(userService, jwtSecret)
	authController := controllers.NewAuthController(authService)

	// Public routes
	r.POST("/api/register", authController.Register)
	r.POST("/api/login", authController.Login)
	r.POST("/api/reset-token", authController.ResetToken)
}
