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

func AnalysticRoutes(
	db *gorm.DB,
	authorized *gin.RouterGroup,
	permService services.PermissionService,
	redisClient *redis.Client,
) {
	dashboardRepo := repositories.NewDashboardRepository(db)
	dashboardService := services.NewDashboardService(dashboardRepo, redisClient, db)
	dashboardCtrl := controllers.NewDashboardController(dashboardService)

	group := authorized.Group("/analystic")
	{
		group.GET("/", middlewares.CheckPermission(permService, "analystic", "view"), dashboardCtrl.GetDashboard) // With auth middleware	}
	}
}
