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

func ReactionRoutes(db *gorm.DB, authorized *gin.RouterGroup, permService services.PermissionService, redisClient *redis.Client) {
	reactionRepo := repositories.NewReactionRepository(db)
	reactionService := services.NewReactionService(reactionRepo, redisClient)
	reactionController := controllers.NewReactionController(reactionService)

	reactions := authorized.Group("/reactions")
	{
		reactions.POST("/", middlewares.CheckPermission(permService, "reaction", "create"), reactionController.CreateReaction)
		reactions.GET("/:id", middlewares.CheckPermission(permService, "reaction", "view"), reactionController.GetReactionByID)
		reactions.PUT("/:id", middlewares.CheckPermission(permService, "reaction", "edit"), reactionController.UpdateReaction)
		reactions.DELETE("/:id", middlewares.CheckPermission(permService, "reaction", "delete"), reactionController.DeleteReaction)
		reactions.GET("/", middlewares.CheckPermission(permService, "reaction", "view"), reactionController.ListReactions)
	}
}
