package main

import (
	"Forum_BE/config"
	"Forum_BE/infrastructure"
	"Forum_BE/models"
	"Forum_BE/routes"
	"log"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {
	// Load configuration
	cfg := config.LoadConfig()
	redisClient := config.InitRedis()

	// Connect to MySQL
	db, err := infrastructure.ConnectMySQL(cfg.DBDSN)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	// Auto migrate models
	err = db.AutoMigrate(
		&models.User{},
		&models.Permission{},
		&models.Question{},
		&models.Answer{},
		&models.Comment{},
		&models.Vote{},
		&models.Tag{},
		&models.Follow{},
		&models.Group{},
		&models.Post{},
		&models.PassedQuestion{},
	)
	if err != nil {
		log.Fatal("Failed to migrate database:", err)
	}

	// Initialize Gin router
	r := gin.Default()

	// Thêm middleware CORS trước
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:5173"},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
	}))

	// Setup routes với redisClient
	routes.SetupRoutes(r, db, cfg.JWTSecret, redisClient)

	// Start server
	if err := r.Run(":" + cfg.ServerPort); err != nil {
		log.Fatal("Failed to run server:", err)
	}
}
