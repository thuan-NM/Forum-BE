package routes

import (
	"Forum_BE/controllers"
	"Forum_BE/middlewares"
	"Forum_BE/services"
	"github.com/gin-gonic/gin"
)

func PermissionRoutes(authorized *gin.RouterGroup, permService services.PermissionService) {
	permissionController := controllers.NewPermissionController(permService)

	permissions := authorized.Group("/permissions")
	{
		//permissions.POST("/", middlewares.CheckPermission(permService, "permission", "create"), permissionController.CreatePermission)
		permissions.PUT("/", middlewares.CheckPermission(permService, "permission", "update"), permissionController.UpdatePermission)
		permissions.GET("/", middlewares.CheckPermission(permService, "permission", "view"), permissionController.ListPermissions)
	}
}
