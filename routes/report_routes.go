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

func ReportRoutes(db *gorm.DB, authorized *gin.RouterGroup, permService services.PermissionService, redisClient *redis.Client) {
	userRepo := repositories.NewUserRepository(db)
	postRepo := repositories.NewPostRepository(db)
	commentRepo := repositories.NewCommentRepository(db)
	questionRepo := repositories.NewQuestionRepository(db)
	answerRepo := repositories.NewAnswerRepository(db)
	reportRepo := repositories.NewReportRepository(db)
	reportService := services.NewReportService(reportRepo, userRepo, postRepo, commentRepo, questionRepo, answerRepo, redisClient)
	reportController := controllers.NewReportController(reportService)

	reports := authorized.Group("/reports")
	{
		reports.POST("/", middlewares.CheckPermission(permService, "report", "create"), reportController.CreateReport)
		reports.GET("/:id", middlewares.CheckPermission(permService, "report", "view"), reportController.GetReportById)
		reports.PUT("/:id", middlewares.CheckPermission(permService, "report", "edit"), reportController.UpdateReport)
		reports.PUT("/:id/status", middlewares.CheckPermission(permService, "report", "edit_status"), reportController.UpdateReportStatus)
		reports.DELETE("/:id", middlewares.CheckPermission(permService, "report", "delete"), reportController.DeleteReport)
		reports.POST("/batch-delete", middlewares.CheckPermission(permService, "report", "delete"), reportController.BatchDeleteReports)
		reports.GET("/", middlewares.CheckPermission(permService, "report", "view"), reportController.ListReports)
	}
}
