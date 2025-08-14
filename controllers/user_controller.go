package controllers

import (
	"Forum_BE/repositories" // Add this import
	"Forum_BE/responses"
	"Forum_BE/services"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
)

type Response struct {
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

type UserController struct {
	userService services.UserService
}

func NewUserController(u services.UserService) *UserController {
	return &UserController{userService: u}
}

func (uc *UserController) CreateUser(c *gin.Context) {
	var req struct {
		Username      string `json:"username" binding:"required"`
		Email         string `json:"email" binding:"required,email"`
		Password      string `json:"password" binding:"required,min=6"`
		FullName      string `json:"full_name" binding:"required"`
		EmailVerified bool   `json:"emailVerified" binding:"required" default:"false"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, Response{Message: err.Error()})
		return
	}

	user, err := uc.userService.CreateUser(req.Username, req.Email, req.Password, req.FullName, req.EmailVerified)
	if err != nil {
		c.JSON(http.StatusBadRequest, Response{Message: err.Error()})
		return
	}

	c.JSON(http.StatusCreated, Response{
		Message: "Tạo người dùng thành công",
		Data:    responses.ToUserResponse(user),
	})
}

func (uc *UserController) GetUser(c *gin.Context) {
	id, err := parseID(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, Response{Message: "Id không hợp lệ"})
		return
	}

	user, err := uc.userService.GetUserByID(id)
	if err != nil {
		if errors.Is(err, repositories.ErrNotFound) { // Use repositories.ErrNotFound
			c.JSON(http.StatusNotFound, Response{Message: "Không tìm thấy người dùng"})
		} else {
			c.JSON(http.StatusInternalServerError, Response{Message: "Failed to fetch user"})
		}
		return
	}

	c.JSON(http.StatusOK, Response{
		Data: responses.ToUserResponse(user),
	})
}

func (uc *UserController) UpdateUser(c *gin.Context) {
	id, err := parseID(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, Response{Message: "ID người dùng không hợp lệ"})
		return
	}

	var req struct {
		Username      *string `json:"username,omitempty" binding:"omitempty,min=3,max=50"`
		Email         *string `json:"email,omitempty" binding:"omitempty,email"`
		Password      *string `json:"password,omitempty" binding:"omitempty,min=6"`
		Role          *string `json:"role,omitempty" binding:"omitempty,oneof=root admin employee user"`
		Status        *string `json:"status,omitempty" binding:"omitempty,oneof=active inactive banned"`
		FullName      *string `json:"full_name,omitempty" binding:"omitempty,min=1,max=100"`
		Avatar        *string `json:"avatar,omitempty" binding:"omitempty"`
		Bio           *string `json:"bio,omitempty" binding:"omitempty"`
		Location      *string `json:"location,omitempty" binding:"omitempty"`
		EmailVerified *bool   `json:"email_verified,omitempty" binding:"omitempty"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, Response{Message: err.Error()})
		return
	}

	updateDTO := services.UpdateUserDTO{
		Username:      req.Username,
		Email:         req.Email,
		Password:      req.Password,
		Role:          req.Role,
		Status:        req.Status,
		FullName:      req.FullName,
		Avatar:        req.Avatar,
		Bio:           req.Bio,
		Location:      req.Location,
		EmailVerified: req.EmailVerified,
	}

	user, err := uc.userService.UpdateUser(id, updateDTO)
	if err != nil {
		c.JSON(http.StatusBadRequest, Response{Message: err.Error()})
		return
	}

	c.JSON(http.StatusOK, Response{
		Message: "Cập nhật người dùng thành công",
		Data:    responses.ToUserResponse(user),
	})
}

func (uc *UserController) DeleteUser(c *gin.Context) {
	id, err := parseID(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, Response{Message: "Invalid user ID"})
		return
	}

	if err := uc.userService.DeleteUser(id); err != nil {
		if errors.Is(err, repositories.ErrNotFound) { // Use repositories.ErrNotFound
			c.JSON(http.StatusNotFound, Response{Message: "User not found"})
		} else {
			c.JSON(http.StatusInternalServerError, Response{Message: "Failed to delete user"})
		}
		return
	}

	c.JSON(http.StatusOK, Response{Message: "User deleted successfully"})
}

func (uc *UserController) GetAllUsers(c *gin.Context) {
	filters := make(map[string]interface{})
	if search := c.Query("search"); search != "" {
		filters["search"] = search
	}
	if page := c.Query("page"); page != "" {
		if p, err := strconv.Atoi(page); err == nil {
			filters["page"] = p
		}
	}
	if limit := c.Query("limit"); limit != "" {
		if l, err := strconv.Atoi(limit); err == nil {
			filters["limit"] = l
		}
	}
	users, total, err := uc.userService.GetAllUsers(filters)
	if err != nil {
		c.JSON(http.StatusInternalServerError, Response{Message: "Failed to get all users"})
		return
	}

	var responseUsers []responses.UserResponse
	for _, user := range users {
		responseUsers = append(responseUsers, responses.ToUserResponse(&user))
	}
	c.JSON(http.StatusOK, gin.H{
		"users": responseUsers,
		"total": total,
	})
}

func (uc *UserController) ModifyUserStatus(c *gin.Context) {
	id, err := parseID(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, Response{Message: "Invalid user ID"})
		return
	}

	var req struct {
		Status string `json:"status" binding:"required,oneof=ban unban"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, Response{Message: err.Error()})
		return
	}
	fmt.Println(req.Status)
	user, err := uc.userService.ModifyUserStatus(id, req.Status)
	if err != nil {
		if errors.Is(err, repositories.ErrNotFound) { // Use repositories.ErrNotFound
			c.JSON(http.StatusNotFound, Response{Message: "User not found"})
		} else {
			c.JSON(http.StatusBadRequest, Response{Message: err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, Response{
		Message: "User status updated successfully",
		Data:    responses.ToUserResponse(user),
	})
}

func parseID(idParam string) (uint, error) {
	id, err := strconv.ParseUint(idParam, 10, 64)
	return uint(id), err
}
