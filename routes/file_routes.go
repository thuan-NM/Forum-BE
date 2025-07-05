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

func FileRoutes(db *gorm.DB, authorized *gin.RouterGroup, permService services.PermissionService, redisClient *redis.Client, cld *cloudinary.Cloudinary, uploadPreset string) {
	fileRepo := repositories.NewFileRepository(db)
	fileService := services.NewFileService(fileRepo, cld, uploadPreset, redisClient, db)
	fileController := controllers.NewFileController(fileService, cld, uploadPreset)

	files := authorized.Group("/files")
	{
		files.POST("/", middlewares.CheckPermission(permService, "file", "create"), fileController.CreateFile)
		files.GET("/:id", middlewares.CheckPermission(permService, "file", "view"), fileController.GetFile)
		files.DELETE("/:id", middlewares.CheckPermission(permService, "file", "delete"), fileController.DeleteFile)
		files.GET("/", middlewares.CheckPermission(permService, "file", "view"), fileController.ListFiles)
		files.GET("/:id/download", middlewares.CheckPermission(permService, "file", "view"), fileController.DownloadFile)
	}
}
