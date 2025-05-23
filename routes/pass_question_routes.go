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

func PassRoutes(db *gorm.DB, authorized *gin.RouterGroup, permService services.PermissionService, redisClient *redis.Client) {
	// Khởi tạo repo, service, controller
	passRepo := repositories.NewPassRepository(db)
	passService := services.NewPassService(passRepo, redisClient)
	passController := controllers.NewPassController(passService)

	passes := authorized.Group("/passes")
	{
		passes.PUT("/:id/pass", middlewares.CheckPermission(permService, "pass", "create"), passController.PassQuestion)
		passes.GET("/passed-ids", middlewares.CheckPermission(permService, "pass", "view"), passController.GetPassedQuestionIDs)
	}
}
