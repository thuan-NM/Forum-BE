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

func AttachmentRoutes(db *gorm.DB, authorized *gin.RouterGroup, permService services.PermissionService, redisClient *redis.Client) {
	attachmentRepo := repositories.NewAttachmentRepository(db)
	attachmentService := services.NewAttachmentService(attachmentRepo, redisClient)
	attachmentController := controllers.NewAttachmentController(attachmentService)

	attachments := authorized.Group("/attachments")
	{
		attachments.POST("/upload", middlewares.CheckPermission(permService, "attachment", "create"), attachmentController.UploadAttachment)
		attachments.GET("/:id", middlewares.CheckPermission(permService, "attachment", "view"), attachmentController.GetAttachment)
		attachments.PUT("/:id", middlewares.CheckPermission(permService, "attachment", "edit"), attachmentController.UpdateAttachment)
		attachments.DELETE("/:id", middlewares.CheckPermission(permService, "attachment", "delete"), attachmentController.DeleteAttachment)
		attachments.GET("/", middlewares.CheckPermission(permService, "attachment", "view"), attachmentController.ListAttachments)
	}
}
