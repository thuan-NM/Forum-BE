package controllers

import (
	"Forum_BE/responses"
	"Forum_BE/services"
	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
)

type GroupController struct {
	groupService services.GroupService
}

func NewGroupController(g services.GroupService) *GroupController {
	return &GroupController{groupService: g}
}

// CreateGroup xử lý yêu cầu tạo group mới
func (gc *GroupController) CreateGroup(c *gin.Context) {
	var req struct {
		Name        string `json:"name" binding:"required"`
		Description string `json:"description"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	group, err := gc.groupService.CreateGroup(req.Name, req.Description)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Tạo nhóm thành công",
		"group":   responses.ToGroupResponse(group),
	})
}

// GetGroup xử lý yêu cầu lấy group theo ID
func (gc *GroupController) GetGroup(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.ParseUint(idParam, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID nhóm không hợp lệ"})
		return
	}

	group, err := gc.groupService.GetGroupByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Không tìm thấy nhóm"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"group": responses.ToGroupResponse(group),
	})
}

// EditGroup xử lý yêu cầu cập nhật group
func (gc *GroupController) EditGroup(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.ParseUint(idParam, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID nhóm không hợp lệ"})
		return
	}

	var req struct {
		Name        string `json:"name"`
		Description string `json:"description"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	group, err := gc.groupService.UpdateGroup(uint(id), req.Name, req.Description)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Cập nhật nhóm thành công",
		"group":   responses.ToGroupResponse(group),
	})
}

// DeleteGroup xử lý yêu cầu xóa group
func (gc *GroupController) DeleteGroup(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.ParseUint(idParam, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID nhóm không hợp lệ"})
		return
	}

	if err := gc.groupService.DeleteGroup(uint(id)); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Xoá nhóm thành công",
	})
}

// ListGroups xử lý yêu cầu liệt kê tất cả các group
func (gc *GroupController) ListGroups(c *gin.Context) {
	groups, err := gc.groupService.ListGroups()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Không thể liệt kê danh sách nhóm"})
		return
	}

	var responseGroups []responses.GroupResponse
	for _, group := range groups {
		responseGroups = append(responseGroups, responses.ToGroupResponse(&group))
	}

	c.JSON(http.StatusOK, gin.H{
		"groups": responseGroups,
	})
}
