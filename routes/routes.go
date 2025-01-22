package routes

import (
	"Forum_BE/config"
	"Forum_BE/middlewares"
	"Forum_BE/models"
	"Forum_BE/repositories"
	"Forum_BE/services"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"log"
)

func SetupRoutes(r *gin.Engine, db *gorm.DB, jwtSecret string) {
	userRepo := repositories.NewUserRepository(db)
	permissionRepo := repositories.NewPermissionRepository(db)
	permService := services.NewPermissionService(permissionRepo, userRepo)

	var permissions []models.Permission
	config.InitPermissions(&permissions)

	for _, perm := range permissions {
		existingPerm, err := permService.GetPermission(string(perm.Role), perm.Resource, perm.Action)
		if err == nil && existingPerm != nil {
			continue // Permission đã tồn tại
		}

		_, err = permService.CreatePermission(string(perm.Role), perm.Resource, perm.Action, perm.Allowed)
		if err != nil {
			log.Printf("Failed to create permission %v: %v", perm, err)
		} else {
			log.Printf("Created permission: %+v", perm)
		}
	}
	AuthRoutes(r, db, jwtSecret)
	// Protected routes
	authMiddleware := middlewares.AuthMiddleware(jwtSecret)
	authorized := r.Group("/api")
	authorized.Use(authMiddleware)
	{
		UserRoutes(db, authorized, permService)
		QuestionRoutes(db, authorized, permService)
		AnswerRoutes(db, authorized, permService)
		CommentRoutes(db, authorized, permService)
		TagRoutes(db, authorized, permService)
		FollowRoutes(db, authorized, permService)
		GroupRoutes(db, authorized, permService)
		VoteRoutes(db, authorized, permService)
		PermissionRoutes(authorized, permService)
	}
}
