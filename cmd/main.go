package main

import (
	"Forum_BE/config"
	"Forum_BE/infrastructure"
	"Forum_BE/models"
	"Forum_BE/routes"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"log"
	"time"
)

func main() {
	// Load configuration
	cfg := config.LoadConfig()

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
	)
	if err != nil {
		log.Fatal("Failed to migrate database:", err)
	}

	// Initialize Permission Service

	// Initialize Gin router
	// Trong main.go
	// Di chuyển middleware CORS lên TRƯỚC khi setup routes

	r := gin.Default()

	// Thêm middleware CORS trước
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:5173"},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	// Sau đó mới setup routes
	routes.SetupRoutes(r, db, cfg.JWTSecret)

	// Start server
	if err := r.Run(":" + cfg.ServerPort); err != nil {
		log.Fatal("Failed to run server:", err)
	}
}
