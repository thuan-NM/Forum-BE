package main

import (
	"Forum_BE/config"
	"Forum_BE/infrastructure"
	"Forum_BE/models"
	"Forum_BE/routes"
	"log"

	"github.com/gin-gonic/gin"
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
	r := gin.Default()

	// Setup all routes
	routes.SetupRoutes(r, db, cfg.JWTSecret)

	// Start server
	if err := r.Run(":" + cfg.ServerPort); err != nil {
		log.Fatal("Failed to run server:", err)
	}
}
