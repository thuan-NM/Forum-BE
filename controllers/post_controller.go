package controllers

import (
	"Forum_BE/models"
	"Forum_BE/responses"
	"Forum_BE/services"
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
)

type PostController struct {
	postService services.PostService
}

func NewPostController(p services.PostService) *PostController {
	return &PostController{p}
}

func (pc *PostController) CreatePost(c *gin.Context) {
	var req struct {
		Title   string `json:"title"`
		Content string `json:"content" binding:"required"`
		Tags    []uint `json:"tags"`
	}

	fmt.Println("hello")
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID := c.GetUint("user_id")

	post, err := pc.postService.CreatePost(req.Content, userID, req.Title, req.Tags)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Post created successfully",
		"post":    responses.ToPostResponse(post),
	})
}

func (pc *PostController) GetPostById(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.ParseUint(idParam, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid post id"})
		return
	}

	post, err := pc.postService.GetPostByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "post not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"post": responses.ToPostResponse(post),
	})
}

func (pc *PostController) DeletePost(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.ParseUint(idParam, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid post id"})
		return
	}

	if err := pc.postService.DeletePost(uint(id)); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"message": "Post deleted successfully",
	})
}

func (pc *PostController) UpdatePost(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.ParseUint(idParam, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid post id"})
		return
	}

	var req struct {
		Title   string `json:"title"`
		Content string `json:"content" binding:"required"`
		Status  string `json:"status"`
		Tags    []uint `json:"tags"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	post, err := pc.postService.UpdatePost(uint(id), req.Content, models.PostStatus(req.Status), req.Tags)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Post updated successfully",
		"post":    responses.ToPostResponse(post),
	})
}

func (pc *PostController) UpdatePostStatus(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.ParseUint(idParam, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid post id"})
		return
	}

	var req struct {
		Status string `json:"status" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	post, err := pc.postService.UpdatePostStatus(uint(id), req.Status)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Post status updated successfully",
		"post":    responses.ToPostResponse(post),
	})
}

func (pc *PostController) ListPosts(c *gin.Context) {
	filters := make(map[string]interface{})

	if userID := c.Query("user_id"); userID != "" {
		if uID, err := strconv.ParseUint(userID, 10, 64); err == nil {
			filters["user_id"] = uint(uID)
		}
	}
	if status := c.Query("status"); status != "" {
		filters["status"] = status
	}
	if search := c.Query("search"); search != "" {
		filters["search"] = search
	}
	if tagfilter := c.Query("tagfilter"); tagfilter != "" {
		filters["tagfilter"] = tagfilter
	}
	//if title := c.Query("title"); title != "" {
	//	filters["title"] = title
	//}
	if page := c.Query("page"); page != "" {
		if p, err := strconv.Atoi(page); err == nil && p > 0 {
			filters["page"] = p
		}
	}
	if limit := c.Query("limit"); limit != "" {
		if l, err := strconv.Atoi(limit); err == nil && l > 0 {
			filters["limit"] = l
		}
	}

	posts, total, err := pc.postService.ListPosts(filters)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list posts"})
		return
	}

	var response []responses.PostResponse
	for _, post := range posts {
		response = append(response, responses.ToPostResponse(&post))
	}

	c.JSON(http.StatusOK, gin.H{
		"posts": response,
		"total": total,
	})
}

func (pc *PostController) GetAllPosts(c *gin.Context) {
	filters := make(map[string]interface{})

	if search := c.Query("search"); search != "" {
		filters["search"] = search
	}
	if status := c.Query("status"); status != "" {
		filters["status"] = status
	}
	if page := c.Query("page"); page != "" {
		if p, err := strconv.Atoi(page); err == nil && p > 0 {
			filters["page"] = p
		}
	}
	if userID := c.Query("user_id"); userID != "" {
		if uID, err := strconv.ParseUint(userID, 10, 64); err == nil {
			filters["user_id"] = uint(uID)
		}
	}
	//if tagfilter := c.Query("tagfilter"); tagfilter != "" {
	//	filters["tagfilter"] = tagfilter
	//}
	if tagfilter := c.Query("tagfilter"); tagfilter != "" {
		if tagfilter, err := strconv.ParseUint(tagfilter, 10, 64); err == nil {
			filters["tagfilter"] = uint(tagfilter)
		}
	}
	if limit := c.Query("limit"); limit != "" {
		if l, err := strconv.Atoi(limit); err == nil && l > 0 {
			filters["limit"] = l
		}
	}

	posts, total, err := pc.postService.GetAllPosts(filters)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get all posts"})
		return
	}

	var response []responses.PostResponse
	for _, post := range posts {
		response = append(response, responses.ToPostResponse(&post))
	}

	c.JSON(http.StatusOK, gin.H{
		"posts": response,
		"total": total,
	})
}
