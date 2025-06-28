package routes

import (
	"Forum_BE/controllers"
	"Forum_BE/repositories"
	"Forum_BE/services"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"gorm.io/gorm"
)

func AuthRoutes(r *gin.Engine, db *gorm.DB, jwtSecret string, redisClient *redis.Client) {
	userRepo := repositories.NewUserRepository(db)
	userService := services.NewUserService(userRepo, redisClient)
	authService := services.NewAuthService(userService, jwtSecret, redisClient, "smtp.gmail.com", "ititblog8@gmail.com", "ukkn bntd vykr yefq", 587)
	authController := controllers.NewAuthController(authService)

	r.POST("/api/register", authController.Register)
	r.POST("/api/login", authController.Login)
	r.POST("/api/reset-token", authController.ResetToken)
	r.POST("/api/logout", authController.Logout)
	r.GET("/api/verify-email", authController.VerifyEmail)
	r.POST("/api/resend-verification", authController.ResendVerificationEmail)
	r.POST("/api/google-login", authController.GoogleLoginWithToken)
	//r.GET("/auth/facebook/callback", authController.FacebookCallback)
	r.GET("/api/me", authController.GetUser)
}
