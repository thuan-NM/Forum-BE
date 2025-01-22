package controllers

import (
	"Forum_BE/config"
	"Forum_BE/responses"
	"Forum_BE/services"
	"Forum_BE/utils"
	"github.com/gin-gonic/gin"
	"net/http"
	"strings"
)

type AuthController struct {
	authService services.AuthService
}

func NewAuthController(a services.AuthService) *AuthController {
	return &AuthController{authService: a}
}

// Register xử lý yêu cầu đăng ký người dùng mới
func (ac *AuthController) Register(c *gin.Context) {
	var req struct {
		Username string `json:"username" binding:"required"`
		Email    string `json:"email" binding:"required,email"`
		Password string `json:"password" binding:"required,min=6"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, err := ac.authService.Register(req.Username, req.Email, req.Password)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Registration successful",
		"user":    responses.ToUserResponse(user),
	})
}

// Login xử lý yêu cầu đăng nhập
func (ac *AuthController) Login(c *gin.Context) {
	var req struct {
		Email    string `json:"email" binding:"required"`
		Password string `json:"password" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	token, user, err := ac.authService.Login(req.Email, req.Password)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Login successful",
		"token":   token,
		"user":    user,
	})
}
func (ac *AuthController) ResetToken(c *gin.Context) {
	// Trích xuất token từ header Authorization

	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header missing"})
		return
	}

	tokenString := strings.TrimPrefix(authHeader, "Bearer ")
	claims, err := utils.ParseJWT(tokenString, config.LoadConfig().JWTSecret)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
		return
	}

	userID := claims.UserID // Giả sử bạn đã lưu userID trong claims

	// Reset token
	newToken, err := ac.authService.ResetToken(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to reset token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"token": newToken})
}
