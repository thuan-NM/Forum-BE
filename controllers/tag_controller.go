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

func (tc *TagController) CreateTag(c *gin.Context) {
	var req struct {
		Name        string `json:"name" binding:"required"`
		Description string `json:"description"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	tag, err := tc.tagService.CreateTag(req.Name, req.Description)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Tạo nhãn thành công",
		"tag":     responses.ToTagResponse(tag),
	})
}

func (tc *TagController) GetTag(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.ParseUint(idParam, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID nhãn không hợp lệ"})
		return
	}

	tag, err := tc.tagService.GetTagByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Không tìm thấy nhãn"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"tag": responses.ToTagResponse(tag),
	})
}

func (tc *TagController) EditTag(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.ParseUint(idParam, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID nhãn không hợp lệ"})
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

	tag, err := tc.tagService.UpdateTag(uint(id), req.Name, req.Description)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Cập nhật nhãn thành công",
		"tag":     responses.ToTagResponse(tag),
	})
}

func (tc *TagController) DeleteTag(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.ParseUint(idParam, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID nhãn không hợp lệ"})
		return
	}

	if err := tc.tagService.DeleteTag(uint(id)); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Xoá nhãn thành công",
	})
}

func (tc *TagController) ListTags(c *gin.Context) {
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

	tags, total, err := tc.tagService.ListTags(filters)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Không thể liệt kê các nhãn"})
		return
	}

	var responseTags []responses.TagResponse
	for _, tag := range tags {
		responseTags = append(responseTags, responses.ToTagResponse(&tag))
	}

	c.JSON(http.StatusOK, gin.H{
		"tags":  responseTags,
		"total": total,
	})
}

func (tc *TagController) GetTagsByPostID(c *gin.Context) {
	postIDParam := c.Param("post_id")
	postID, err := strconv.ParseUint(postIDParam, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID bài đăng không hợp lệ"})
		return
	}

	tags, err := tc.tagService.GetTagsByPostID(uint(postID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Không thể lấy nhãn hoặc bài đăng"})
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

func (tc *TagController) GetTagsByAnswerID(c *gin.Context) {
	answerIDParam := c.Param("answer_id")
	answerID, err := strconv.ParseUint(answerIDParam, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID câu trả lời không hợp lệ"})
		return
	}

	tags, err := tc.tagService.GetTagsByAnswerID(uint(answerID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Không thể lấy nhãn hoặc câu trả lời"})
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
