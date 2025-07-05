package routes

import (
	"Forum_BE/config"
	"github.com/cloudinary/cloudinary-go/v2"
	"os"

	// "Forum_BE/config"
	"Forum_BE/middlewares"
	"Forum_BE/models"
	"Forum_BE/repositories"
	"Forum_BE/services"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"gorm.io/gorm"
	"log"
)

func SetupRoutes(r *gin.Engine, db *gorm.DB, jwtSecret string, redisClient *redis.Client) {
	userRepo := repositories.NewUserRepository(db)
	permissionRepo := repositories.NewPermissionRepository(db)
	permService := services.NewPermissionService(permissionRepo, userRepo)

	var permissions []models.Permission
	config.InitPermissions()
	// Khởi tạo Cloudinary
	cld, err := cloudinary.NewFromParams(
		os.Getenv("CLOUDINARY_CLOUD_NAME"),
		os.Getenv("CLOUDINARY_API_KEY"),
		os.Getenv("CLOUDINARY_API_SECRET"),
	)
	if err != nil {
		log.Fatalf("Failed to initialize Cloudinary: %v", err)
	}
	uploadPreset := os.Getenv("CLOUDINARY_UPLOAD_PRESET")
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

	AuthRoutes(r, db, jwtSecret, redisClient)

	authMiddleware := middlewares.AuthMiddleware(jwtSecret)
	authorized := r.Group("/api")
	authorized.Use(authMiddleware)
	{
		UserRoutes(db, authorized, permService, redisClient)
		QuestionRoutes(db, authorized, permService, redisClient)
		PostRoutes(db, authorized, permService, redisClient)
		AnswerRoutes(db, authorized, permService, redisClient)
		CommentRoutes(db, authorized, permService, redisClient, cld, uploadPreset)
		TagRoutes(db, authorized, permService, redisClient)
		TopicRoutes(db, authorized, permService, redisClient)
		FollowRoutes(db, authorized, permService, redisClient)
		GroupRoutes(db, authorized, permService, redisClient)
		VoteRoutes(db, authorized, permService)
		ReportRoutes(db, authorized, permService, redisClient)
		PermissionRoutes(authorized, permService)
		FileRoutes(db, authorized, permService, redisClient, cld, uploadPreset)

		PassRoutes(db, authorized, permService, redisClient)
		ReactionRoutes(db, authorized, permService, redisClient)
	}
}
