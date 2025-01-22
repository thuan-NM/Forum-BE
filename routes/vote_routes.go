package routes

import (
	"Forum_BE/controllers"
	"Forum_BE/middlewares"
	"Forum_BE/repositories"
	"Forum_BE/services"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func VoteRoutes(db *gorm.DB, authorized *gin.RouterGroup, permService services.PermissionService) {
	// Vote routes
	voteRepo := repositories.NewVoteRepository(db)
	voteService := services.NewVoteService(voteRepo)
	voteController := controllers.NewVoteController(voteService)

	votes := authorized.Group("/votes")
	{
		votes.POST("/", middlewares.CheckPermission(permService, "vote", "create"), voteController.CastVote)
		votes.GET("/:id", middlewares.CheckPermission(permService, "vote", "view"), voteController.GetVote)
		votes.PUT("/:id", middlewares.CheckPermission(permService, "vote", "edit"), voteController.UpdateVote)
		votes.DELETE("/:id", middlewares.CheckPermission(permService, "vote", "delete"), voteController.DeleteVote)
		votes.GET("/", middlewares.CheckPermission(permService, "vote", "view"), voteController.ListVotes)
	}
}
