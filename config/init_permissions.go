package config

import (
	"Forum_BE/infrastructure"
	"Forum_BE/models"
	"Forum_BE/repositories"
	"Forum_BE/services"
	"fmt"
	"log"
)

func InitPermissions(permissions *[]models.Permission) {
	// Load configuration
	cfg := LoadConfig()
	fmt.Println(cfg.DBDSN)
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
		&models.QuestionTag{},
	)
	if err != nil {
		log.Fatal("Failed to migrate database:", err)
	}

	// Initialize Repositories and Services
	userRepo := repositories.NewUserRepository(db)
	permissionRepo := repositories.NewPermissionRepository(db)
	permissionService := services.NewPermissionService(permissionRepo, userRepo)

	// Define initial permissions
	*permissions = []models.Permission{
		// Root Permissions
		{Role: models.RoleRoot, Resource: "user", Action: "create", Allowed: true},
		{Role: models.RoleRoot, Resource: "user", Action: "view", Allowed: true},
		{Role: models.RoleRoot, Resource: "user", Action: "edit", Allowed: true},
		{Role: models.RoleRoot, Resource: "user", Action: "delete", Allowed: true},
		{Role: models.RoleRoot, Resource: "question", Action: "create", Allowed: true},
		{Role: models.RoleRoot, Resource: "question", Action: "view", Allowed: true},
		{Role: models.RoleRoot, Resource: "question", Action: "edit", Allowed: true},
		{Role: models.RoleRoot, Resource: "question", Action: "delete", Allowed: true},
		{Role: models.RoleRoot, Resource: "question", Action: "approve", Allowed: true},
		{Role: models.RoleRoot, Resource: "question", Action: "reject", Allowed: true},
		{Role: models.RoleRoot, Resource: "group", Action: "create", Allowed: true},
		{Role: models.RoleRoot, Resource: "group", Action: "view", Allowed: true},
		{Role: models.RoleRoot, Resource: "group", Action: "edit", Allowed: true},
		{Role: models.RoleRoot, Resource: "group", Action: "delete", Allowed: true},
		{Role: models.RoleRoot, Resource: "vote", Action: "create", Allowed: true},
		{Role: models.RoleRoot, Resource: "vote", Action: "view", Allowed: true},
		{Role: models.RoleRoot, Resource: "vote", Action: "edit", Allowed: true},
		{Role: models.RoleRoot, Resource: "vote", Action: "delete", Allowed: true},

		// Admin Permissions
		{Role: models.RoleAdmin, Resource: "user", Action: "create", Allowed: true},
		{Role: models.RoleAdmin, Resource: "user", Action: "view", Allowed: true},
		{Role: models.RoleAdmin, Resource: "user", Action: "edit", Allowed: true},
		{Role: models.RoleAdmin, Resource: "user", Action: "delete", Allowed: true},
		{Role: models.RoleAdmin, Resource: "question", Action: "create", Allowed: true},
		{Role: models.RoleAdmin, Resource: "question", Action: "view", Allowed: true},
		{Role: models.RoleAdmin, Resource: "question", Action: "edit", Allowed: true},
		{Role: models.RoleAdmin, Resource: "question", Action: "delete", Allowed: true},
		{Role: models.RoleAdmin, Resource: "question", Action: "approve", Allowed: true},
		{Role: models.RoleAdmin, Resource: "question", Action: "reject", Allowed: true},
		// Employee Permissions
		{Role: models.RoleEmployee, Resource: "question", Action: "create", Allowed: true},
		{Role: models.RoleEmployee, Resource: "question", Action: "view", Allowed: true},
		{Role: models.RoleEmployee, Resource: "question", Action: "edit", Allowed: true},
		{Role: models.RoleEmployee, Resource: "question", Action: "delete", Allowed: true},
		// User Permissions
		{Role: models.RoleUser, Resource: "question", Action: "create", Allowed: true},
		{Role: models.RoleUser, Resource: "question", Action: "view", Allowed: true},
		{Role: models.RoleUser, Resource: "answer", Action: "create", Allowed: true},
		{Role: models.RoleUser, Resource: "answer", Action: "view", Allowed: true},
		{Role: models.RoleUser, Resource: "comment", Action: "create", Allowed: true},
		{Role: models.RoleUser, Resource: "comment", Action: "view", Allowed: true},
		{Role: models.RoleUser, Resource: "vote", Action: "create", Allowed: true},
		{Role: models.RoleUser, Resource: "vote", Action: "view", Allowed: true},
		{Role: models.RoleUser, Resource: "follow", Action: "create", Allowed: true},
		{Role: models.RoleUser, Resource: "follow", Action: "view", Allowed: true},
	}

	for _, perm := range *permissions {
		existingPerm, err := permissionService.GetPermission(string(perm.Role), perm.Resource, perm.Action)
		if err == nil && existingPerm != nil {
			continue // Permission đã tồn tại
		}

		_, err = permissionService.CreatePermission(string(perm.Role), perm.Resource, perm.Action, perm.Allowed)
		if err != nil {
			log.Printf("Failed to create permission %v: %v", perm, err)
		} else {
			log.Printf("Created permission: %+v", perm)
		}
	}
	log.Println("Permissions initialized successfully")
}
