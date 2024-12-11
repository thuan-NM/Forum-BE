package middlewares

import (
	"Forum_BE/services"
	"github.com/gin-gonic/gin"
	"net/http"
)

func CheckPermission(permService services.PermissionService, resource, action string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetUint("user_id")
		if userID == 0 {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
			c.Abort()
			return
		}

		// Lấy role của user
		role, err := permService.GetUserRole(userID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get user role"})
			c.Abort()
			return
		}

		// Kiểm tra permission
		permission, err := permService.GetPermission(string(role), resource, action)
		if err != nil || permission == nil || !permission.Allowed {
			c.JSON(http.StatusForbidden, gin.H{"error": "Forbidden"})
			c.Abort()
			return
		}

		c.Next()
	}
}
