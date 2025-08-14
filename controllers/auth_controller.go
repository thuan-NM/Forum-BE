package controllers

import (
	"Forum_BE/config"
	"Forum_BE/responses"
	"Forum_BE/services"
	"Forum_BE/utils"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
	"strings"
)

type AuthController struct {
	authService services.AuthService
}

func NewAuthController(a services.AuthService) *AuthController {
	return &AuthController{authService: a}
}

func (ac *AuthController) Register(c *gin.Context) {
	var req struct {
		Username      string `json:"username" binding:"required"`
		Email         string `json:"email" binding:"required,email"`
		Password      string `json:"password" binding:"required,min=6"`
		FullName      string `json:"fullName" binding:"required"`
		EmailVerified bool   `json:"emailVerified"  default:"false"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, err := ac.authService.Register(req.Username, req.Email, req.Password, req.FullName, req.EmailVerified)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Đăng ký thành công",
		"user":    responses.ToUserResponse(user),
	})
}

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
		"message": "Đăng nhập thành công",
		"token":   token,
		"user":    responses.ToUserResponse(user),
	})
}

func (ac *AuthController) Logout(c *gin.Context) {
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Chưa được xác thực"})
		c.Abort()
		return
	}

	tokenString := strings.TrimPrefix(authHeader, "Bearer ")
	claims, err := utils.ParseJWT(tokenString, config.LoadConfig().JWTSecret)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Token không hợp lệ"})
		return
	}

	userID := claims.UserID

	if err := ac.authService.Logout(userID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Đăng xuất thất bại"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Đăng xuất thành công"})
}

func (ac *AuthController) ResetToken(c *gin.Context) {
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Chưa được xác thực"})
		return
	}

	tokenString := strings.TrimPrefix(authHeader, "Bearer ")
	claims, err := utils.ParseJWT(tokenString, config.LoadConfig().JWTSecret)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Token không hợp lệ"})
		return
	}

	userID := claims.UserID

	newToken, err := ac.authService.ResetToken(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Không thể reset token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"token": newToken})
}
func (ac *AuthController) VerifyEmail(c *gin.Context) {
	token := c.Query("token")
	if token == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Thiếu token xác thực"})
		return
	}
	log.Printf(token)
	_, err := ac.authService.VerifyEmailToken(token)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Xác thực email thành công, bạn có thể đăng nhập ngay."})
}
func (ac *AuthController) ResendVerificationEmail(c *gin.Context) {
	var req struct {
		Email string `json:"email" binding:"required,email"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Email không hợp lệ"})
		return
	}

	err := ac.authService.ResendVerificationEmail(req.Email)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Gửi lại email xác thực thành công"})
}
func (ac *AuthController) GoogleLoginWithToken(c *gin.Context) {
	var req struct {
		IDToken string `json:"idToken" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request: " + err.Error()})
		return
	}

	token, user, err := ac.authService.HandleGoogleIDToken(req.IDToken)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Đăng nhập Google thành công",
		"token":   token,
		"user":    responses.ToUserResponse(user),
	})

}

func (ac *AuthController) GetUser(c *gin.Context) {
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Chưa xác thực"})
		return
	}

	token := strings.TrimPrefix(authHeader, "Bearer ")
	user, err := ac.authService.GetUserFromToken(token)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"user": user})
}
func (ac *AuthController) ChangePassWord(c *gin.Context) {
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Chưa xác thực"})
		return
	}

	token := strings.TrimPrefix(authHeader, "Bearer ")
	user, err := ac.authService.GetUserFromToken(token)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	var req struct {
		NewPassword string `json:"newPassword" binding:"required,min=6"`
		OldPassword string `json:"oldPassword" binding:"required,min=6"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Mật khẩu không hợp lệ: " + err.Error()})
		return
	}

	updatedUser, err := ac.authService.ChangePassword(user.ID, req.OldPassword, req.NewPassword)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Đổi mật khẩu thất bại: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Đổi mật khẩu thành công",
		"user":    responses.ToUserResponse(updatedUser),
	})
}
