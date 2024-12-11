package routes

import (
	"Forum_BE/controllers"
	"Forum_BE/middlewares"
	"Forum_BE/models"
	"Forum_BE/repositories"
	"Forum_BE/services"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"log"
)

func SetupRoutes(r *gin.Engine, db *gorm.DB, jwtSecret string) {
	// Initialize Services
	userRepo := repositories.NewUserRepository(db)
	userService := services.NewUserService(userRepo)
	authService := services.NewAuthService(userService, jwtSecret)

	questionRepo := repositories.NewQuestionRepository(db)
	questionService := services.NewQuestionService(questionRepo)

	answerRepo := repositories.NewAnswerRepository(db)
	answerService := services.NewAnswerService(answerRepo, questionRepo)

	commentRepo := repositories.NewCommentRepository(db)
	commentService := services.NewCommentService(commentRepo, questionRepo, answerRepo)

	tagRepo := repositories.NewTagRepository(db)
	tagService := services.NewTagService(tagRepo)

	followRepo := repositories.NewFollowRepository(db)
	followService := services.NewFollowService(followRepo)

	groupRepo := repositories.NewGroupRepository(db)
	groupService := services.NewGroupService(groupRepo)

	voteRepo := repositories.NewVoteRepository(db)
	voteService := services.NewVoteService(voteRepo)

	permissionRepo := repositories.NewPermissionRepository(db)
	permService := services.NewPermissionService(permissionRepo, userRepo)

	// Initialize Controllers
	authController := controllers.NewAuthController(authService)
	userController := controllers.NewUserController(userService)
	questionController := controllers.NewQuestionController(questionService, voteService)
	answerController := controllers.NewAnswerController(answerService, voteService)
	commentController := controllers.NewCommentController(commentService, voteService)
	tagController := controllers.NewTagController(tagService)
	followController := controllers.NewFollowController(followService)
	groupController := controllers.NewGroupController(groupService)
	voteController := controllers.NewVoteController(voteService)
	permissionController := controllers.NewPermissionController(permService)
	permissions := []models.Permission{
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
	// Public routes
	r.POST("/register", authController.Register)
	r.POST("/login", authController.Login)

	// Protected routes
	authMiddleware := middlewares.AuthMiddleware(jwtSecret)
	authorized := r.Group("/")
	authorized.Use(authMiddleware)
	{
		// User routes
		users := authorized.Group("/users")
		{
			users.POST("/", middlewares.CheckPermission(permService, "user", "create"), userController.CreateUser)
			users.GET("/:id", middlewares.CheckPermission(permService, "user", "view"), userController.GetUser)
			users.PUT("/:id", middlewares.CheckPermission(permService, "user", "edit"), userController.UpdateUser)
			users.DELETE("/:id", middlewares.CheckPermission(permService, "user", "delete"), userController.DeleteUser)
			users.GET("/", middlewares.CheckPermission(permService, "user", "view"), userController.ListUsers)
		}

		// Question routes
		questions := authorized.Group("/questions")
		{
			questions.POST("/", middlewares.CheckPermission(permService, "question", "create"), questionController.CreateQuestion)
			questions.GET("/:id", middlewares.CheckPermission(permService, "question", "view"), questionController.GetQuestion)
			questions.PUT("/:id", middlewares.CheckPermission(permService, "question", "edit"), questionController.UpdateQuestion)
			questions.DELETE("/:id", middlewares.CheckPermission(permService, "question", "delete"), questionController.DeleteQuestion)
			questions.GET("/", middlewares.CheckPermission(permService, "question", "view"), questionController.ListQuestions)

			// Approval routes
			questions.POST("/:id/approve", middlewares.CheckPermission(permService, "question", "approve"), questionController.ApproveQuestion)
			questions.POST("/:id/reject", middlewares.CheckPermission(permService, "question", "reject"), questionController.RejectQuestion)
		}

		// Answer routes
		answers := authorized.Group("/answers")
		{
			answers.POST("/", middlewares.CheckPermission(permService, "answer", "create"), answerController.CreateAnswer)
			answers.GET("/:id", middlewares.CheckPermission(permService, "answer", "view"), answerController.GetAnswer)
			answers.PUT("/:id", middlewares.CheckPermission(permService, "answer", "edit"), answerController.EditAnswer)
			answers.DELETE("/:id", middlewares.CheckPermission(permService, "answer", "delete"), answerController.DeleteAnswer)
			answers.GET("/", middlewares.CheckPermission(permService, "answer", "view"), answerController.ListAnswers)
		}

		// Comment routes
		comments := authorized.Group("/comments")
		{
			comments.POST("/", middlewares.CheckPermission(permService, "comment", "create"), commentController.CreateComment)
			comments.GET("/:id", middlewares.CheckPermission(permService, "comment", "view"), commentController.GetComment)
			comments.PUT("/:id", middlewares.CheckPermission(permService, "comment", "edit"), commentController.EditComment)
			comments.DELETE("/:id", middlewares.CheckPermission(permService, "comment", "delete"), commentController.DeleteComment)
			comments.GET("/", middlewares.CheckPermission(permService, "comment", "view"), commentController.ListComments)
		}

		// Tag routes
		tags := authorized.Group("/tags")
		{
			tags.POST("/", middlewares.CheckPermission(permService, "tag", "create"), tagController.CreateTag)
			tags.GET("/:id", middlewares.CheckPermission(permService, "tag", "view"), tagController.GetTag)
			tags.PUT("/:id", middlewares.CheckPermission(permService, "tag", "edit"), tagController.EditTag)
			tags.DELETE("/:id", middlewares.CheckPermission(permService, "tag", "delete"), tagController.DeleteTag)
			tags.GET("/", middlewares.CheckPermission(permService, "tag", "view"), tagController.ListTags)
		}

		// Follow routes
		follows := authorized.Group("/follows")
		{
			follows.POST("/", middlewares.CheckPermission(permService, "follow", "create"), followController.FollowUser)
			follows.DELETE("/:following_id", middlewares.CheckPermission(permService, "follow", "delete"), followController.UnfollowUser)
			follows.GET("/followers/:user_id", middlewares.CheckPermission(permService, "follow", "view"), followController.GetFollowers)
			follows.GET("/following/:user_id", middlewares.CheckPermission(permService, "follow", "view"), followController.GetFollowing)
		}

		// Group routes
		groups := authorized.Group("/groups")
		{
			groups.POST("/", middlewares.CheckPermission(permService, "group", "create"), groupController.CreateGroup)
			groups.GET("/:id", middlewares.CheckPermission(permService, "group", "view"), groupController.GetGroup)
			groups.PUT("/:id", middlewares.CheckPermission(permService, "group", "edit"), groupController.EditGroup)
			groups.DELETE("/:id", middlewares.CheckPermission(permService, "group", "delete"), groupController.DeleteGroup)
			groups.GET("/", middlewares.CheckPermission(permService, "group", "view"), groupController.ListGroups)
		}

		// Vote routes
		votes := authorized.Group("/votes")
		{
			votes.POST("/", middlewares.CheckPermission(permService, "vote", "create"), voteController.CastVote)
			votes.GET("/:id", middlewares.CheckPermission(permService, "vote", "view"), voteController.GetVote)
			votes.PUT("/:id", middlewares.CheckPermission(permService, "vote", "edit"), voteController.UpdateVote)
			//votes.DELETE("/:id", middlewares.CheckPermission(permService, "vote", "delete"), voteController.DeleteVote)
			//votes.GET("/", middlewares.CheckPermission(permService, "vote", "view"), voteController.ListVotes)
		}

		// Permission routes
		permissions := authorized.Group("/permissions")
		{
			//permissions.POST("/", middlewares.CheckPermission(permService, "permission", "create"), permissionController.CreatePermission)
			permissions.PUT("/", middlewares.CheckPermission(permService, "permission", "update"), permissionController.UpdatePermission)
			permissions.GET("/", middlewares.CheckPermission(permService, "permission", "view"), permissionController.ListPermissions)
		}
	}
}
