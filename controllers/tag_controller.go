package controllers

import (
	"Forum_BE/responses"
	"Forum_BE/services"
	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
)

type TagController struct {
	tagService services.TagService
}

func NewTagController(t services.TagService) *TagController {
	return &TagController{tagService: t}
}

// CreateTag xử lý yêu cầu tạo tag mới
func (tc *TagController) CreateTag(c *gin.Context) {
	var req struct {
		Name string `json:"name" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	tag, err := tc.tagService.CreateTag(req.Name)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Tag created successfully",
		"tag":     responses.ToTagResponse(tag),
	})
}

// GetTag xử lý yêu cầu lấy tag theo ID
func (tc *TagController) GetTag(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.ParseUint(idParam, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid tag id"})
		return
	}

	tag, err := tc.tagService.GetTagByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "tag not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"tag": responses.ToTagResponse(tag),
	})
}

// EditTag xử lý yêu cầu cập nhật tag
func (tc *TagController) EditTag(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.ParseUint(idParam, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid tag id"})
		return
	}

	var req struct {
		Name string `json:"name" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	tag, err := tc.tagService.UpdateTag(uint(id), req.Name)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Tag updated successfully",
		"tag":     responses.ToTagResponse(tag),
	})
}

// DeleteTag xử lý yêu cầu xóa tag
func (tc *TagController) DeleteTag(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.ParseUint(idParam, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid tag id"})
		return
	}

	if err := tc.tagService.DeleteTag(uint(id)); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Tag deleted successfully",
	})
}

// ListTags xử lý yêu cầu liệt kê tất cả các tag
func (tc *TagController) ListTags(c *gin.Context) {
	tags, err := tc.tagService.ListTags()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list tags"})
		return
	}

	var responseTags []responses.TagResponse
	for _, tag := range tags {
		responseTags = append(responseTags, responses.ToTagResponse(&tag))
	}

	c.JSON(http.StatusOK, gin.H{
		"tags": responseTags,
	})
}
