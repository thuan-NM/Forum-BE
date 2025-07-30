package config

import (
	"log"
	"time"

	"Forum_BE/infrastructure"
	"Forum_BE/models"
	"Forum_BE/repositories"
	"Forum_BE/services"
)

func InitPermissions() {
func InitPermissions() {
	cfg := LoadConfig()
	// Connect to database
	// Connect to database
	db, err := infrastructure.ConnectMySQL(cfg.DBDSN)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	// Auto migrate all models
	// Auto migrate all models
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
		&models.PassedQuestion{},
		&models.Post{},
		&models.Report{},
		&models.Notification{},
		&models.Attachment{},
		&models.Message{},
		&models.Topic{},
		&models.Reaction{},
		&models.Post{},
		&models.Report{},
		&models.Notification{},
		&models.Attachment{},
		&models.Message{},
		&models.Topic{},
		&models.Reaction{},
	)
	if err != nil {
		log.Fatal("Failed to migrate database:", err)
	}

	// Initialize Repositories and Services
	userRepo := repositories.NewUserRepository(db)
	permissionRepo := repositories.NewPermissionRepository(db)
	permissionService := services.NewPermissionService(permissionRepo, userRepo)

	// Define roles
	roles := []models.Role{
		models.RoleRoot, models.RoleAdmin, models.RoleEmployee, models.RoleUser,
	}

	// Define permissions with allowed: true for specific role-resource-action combinations
	allowedPermissions := map[string]map[string][]models.Role{
		"user": {
			"create": {models.RoleRoot, models.RoleAdmin},
			"view":   {models.RoleRoot, models.RoleAdmin},
			"edit":   {models.RoleRoot, models.RoleAdmin},
			"delete": {models.RoleRoot},
			"ban":    {models.RoleRoot, models.RoleAdmin},
			"unban":  {models.RoleRoot, models.RoleAdmin},
		},
		"question": {
			"create":              {models.RoleRoot, models.RoleAdmin, models.RoleEmployee, models.RoleUser},
			"view":                {models.RoleRoot, models.RoleAdmin, models.RoleEmployee, models.RoleUser},
			"edit":                {models.RoleRoot, models.RoleAdmin, models.RoleEmployee},
			"delete":              {models.RoleRoot, models.RoleAdmin, models.RoleEmployee},
			"change_status":       {models.RoleRoot, models.RoleAdmin, models.RoleEmployee},
			"change_inter_status": {models.RoleRoot, models.RoleAdmin, models.RoleEmployee, models.RoleUser},
		},
		"answer": {
			"create":      {models.RoleRoot, models.RoleAdmin, models.RoleEmployee, models.RoleUser},
			"view":        {models.RoleRoot, models.RoleAdmin, models.RoleEmployee, models.RoleUser},
			"edit":        {models.RoleRoot, models.RoleAdmin, models.RoleEmployee},
			"delete":      {models.RoleRoot, models.RoleAdmin, models.RoleEmployee},
			"edit_status": {models.RoleRoot, models.RoleAdmin, models.RoleEmployee},
			"accept":      {models.RoleRoot, models.RoleAdmin, models.RoleEmployee, models.RoleUser},
		},
		"comment": {
			"create":  {models.RoleRoot, models.RoleAdmin, models.RoleEmployee, models.RoleUser},
			"view":    {models.RoleRoot, models.RoleAdmin, models.RoleEmployee, models.RoleUser},
			"edit":    {models.RoleRoot, models.RoleAdmin, models.RoleEmployee},
			"delete":  {models.RoleRoot, models.RoleAdmin, models.RoleEmployee},
			"approve": {models.RoleRoot, models.RoleAdmin, models.RoleEmployee},
			"reject":  {models.RoleRoot, models.RoleAdmin, models.RoleEmployee},
		},
		"vote": {
			"create": {models.RoleRoot, models.RoleAdmin, models.RoleEmployee, models.RoleUser},
			"view":   {models.RoleRoot, models.RoleAdmin, models.RoleEmployee, models.RoleUser},
			"edit":   {models.RoleRoot},
			"delete": {models.RoleRoot},
		},
		"tag": {
			"create": {models.RoleRoot, models.RoleAdmin, models.RoleEmployee},
			"view":   {models.RoleRoot, models.RoleAdmin, models.RoleEmployee, models.RoleUser},
			"edit":   {models.RoleRoot, models.RoleAdmin},
			"delete": {models.RoleRoot},
			"follow": {models.RoleRoot, models.RoleAdmin, models.RoleEmployee, models.RoleUser},
		},
		"follow": {
			"create": {models.RoleRoot, models.RoleAdmin, models.RoleEmployee, models.RoleUser},
			"view":   {models.RoleRoot, models.RoleAdmin, models.RoleEmployee, models.RoleUser},
			"delete": {models.RoleRoot},
		},
		"group": {
			"create": {models.RoleRoot, models.RoleAdmin},
			"view":   {models.RoleRoot, models.RoleAdmin, models.RoleEmployee, models.RoleUser},
			"edit":   {models.RoleRoot, models.RoleAdmin},
			"delete": {models.RoleRoot},
			"join":   {models.RoleRoot, models.RoleAdmin, models.RoleEmployee, models.RoleUser},
			"leave":  {models.RoleRoot, models.RoleAdmin, models.RoleEmployee, models.RoleUser},
		},
		"post": {
			"create":      {models.RoleRoot, models.RoleAdmin, models.RoleEmployee, models.RoleUser},
			"view":        {models.RoleRoot, models.RoleAdmin, models.RoleEmployee, models.RoleUser},
			"edit":        {models.RoleRoot, models.RoleAdmin, models.RoleEmployee},
			"delete":      {models.RoleRoot, models.RoleAdmin},
			"approve":     {models.RoleRoot, models.RoleAdmin, models.RoleEmployee},
			"reject":      {models.RoleRoot, models.RoleAdmin, models.RoleEmployee},
			"edit_status": {models.RoleRoot, models.RoleAdmin, models.RoleEmployee},
		},
		"report": {
			"create":      {models.RoleRoot, models.RoleAdmin, models.RoleEmployee, models.RoleUser},
			"view":        {models.RoleRoot, models.RoleAdmin, models.RoleEmployee},
			"edit":        {models.RoleRoot, models.RoleAdmin},
			"delete":      {models.RoleRoot},
			"edit_status": {models.RoleRoot, models.RoleAdmin, models.RoleEmployee},
		},
		"notification": {
			"create": {models.RoleRoot, models.RoleAdmin},
			"view":   {models.RoleRoot, models.RoleAdmin, models.RoleEmployee, models.RoleUser},
			"edit":   {models.RoleRoot},
			"delete": {models.RoleRoot},
		},
		"attachment": {
			"create": {models.RoleRoot, models.RoleAdmin, models.RoleEmployee, models.RoleUser},
			"view":   {models.RoleRoot, models.RoleAdmin, models.RoleEmployee, models.RoleUser},
			"edit":   {models.RoleRoot, models.RoleAdmin, models.RoleEmployee},
			"delete": {models.RoleRoot, models.RoleAdmin, models.RoleEmployee},
		},
		"message": {
			"create": {models.RoleRoot, models.RoleAdmin, models.RoleEmployee, models.RoleUser},
			"view":   {models.RoleRoot, models.RoleAdmin, models.RoleEmployee, models.RoleUser},
			"edit":   {models.RoleRoot},
			"delete": {models.RoleRoot},
		},
		"topic": {
			"create": {models.RoleRoot, models.RoleAdmin, models.RoleEmployee},
			"view":   {models.RoleRoot, models.RoleAdmin, models.RoleEmployee, models.RoleUser},
			"edit":   {models.RoleRoot, models.RoleAdmin},
			"delete": {models.RoleRoot},
		},
		"reaction": {
			"create": {models.RoleRoot, models.RoleAdmin, models.RoleEmployee, models.RoleUser},
			"view":   {models.RoleRoot, models.RoleAdmin, models.RoleEmployee, models.RoleUser},
			"edit":   {models.RoleRoot, models.RoleAdmin, models.RoleEmployee, models.RoleUser},
			"delete": {models.RoleRoot, models.RoleAdmin, models.RoleEmployee, models.RoleUser},
		},
		"pass": {
			"create": {models.RoleRoot, models.RoleAdmin, models.RoleEmployee},
			"view":   {models.RoleRoot, models.RoleAdmin, models.RoleEmployee},
			"delete": {models.RoleRoot},
		},
		"permission": {
			"create": {models.RoleRoot},
			"view":   {models.RoleRoot, models.RoleAdmin},
			"edit":   {models.RoleRoot, models.RoleAdmin},
			"delete": {models.RoleRoot},
		},
	}

	// Derive resources from allowedPermissions keys
	resources := make([]string, 0, len(allowedPermissions))
	for resource := range allowedPermissions {
		resources = append(resources, resource)
	}

	// Create permissions for all role-resource-action combinations
	for _, resource := range resources {
		var resourceActions []string
		if actions, exists := allowedPermissions[resource]; exists {
			resourceActions = make([]string, 0, len(actions))
			for action := range actions {
				resourceActions = append(resourceActions, action)
			}
		} else {
			resourceActions = []string{"create", "view", "delete"}
		}

		for _, action := range resourceActions {
			for _, role := range roles {
				allowed := false
				if roles, exists := allowedPermissions[resource][action]; exists {
					for _, r := range roles {
						if r == role {
							allowed = true
							break
						}
					}
				}

				existingPerm, err := permissionService.GetPermission(string(role), resource, action)
				if err == nil && existingPerm != nil {
					log.Printf("Permission already exists: %s:%s:%s", role, resource, action)
					continue
				}

				perm := models.Permission{
					Role:      role,
					Resource:  resource,
					Action:    action,
					Allowed:   allowed,
					CreatedAt: time.Now(),
					UpdatedAt: time.Now(),
				}
				_, err = permissionService.CreatePermission(string(perm.Role), perm.Resource, perm.Action, perm.Allowed)
				if err != nil {
					log.Printf("Failed to create permission %s:%s:%s: %v", role, resource, action, err)
				} else {
					log.Printf("Created permission: %s:%s:%s (allowed: %v)", role, resource, action, allowed)
				}
			}
		}
	}
	log.Println("Permissions initialized successfully")
}
