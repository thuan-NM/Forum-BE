package controllers

import (
	"Forum_BE/responses"
	"Forum_BE/services"
	"github.com/gin-gonic/gin"
	"net/http"
)

type PermissionController struct {
	permissionService services.PermissionService
}

func NewPermissionController(p services.PermissionService) *PermissionController {
	return &PermissionController{permissionService: p}
}

// UpdatePermission cập nhật quyền cho một role và action cụ thể
func (pc *PermissionController) UpdatePermission(c *gin.Context) {
	var req struct {
		Role     string `json:"role" binding:"required,oneof=root admin employee user"`
		Resource string `json:"resource" binding:"required"`
		Action   string `json:"action" binding:"required"`
		Allowed  bool   `json:"allowed" `
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	permission, err := pc.permissionService.UpdatePermission(req.Role, req.Resource, req.Action, req.Allowed)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":    "Cập nhật quyền thành công",
		"permission": responses.ToPermissionResponse(permission),
	})
}

// ListPermissions liệt kê tất cả các permissions
func (pc *PermissionController) ListPermissions(c *gin.Context) {
	permissions, err := pc.permissionService.ListPermissions()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Không thể liệt kê các quyền"})
		return
	}

	var responsePermissions []responses.PermissionResponse
	for _, permission := range permissions {
		responsePermissions = append(responsePermissions, responses.ToPermissionResponse(&permission))
	}

	c.JSON(http.StatusOK, gin.H{
		"permissions": responsePermissions,
	})
}
