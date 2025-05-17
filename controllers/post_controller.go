package controllers

import (
	"Forum_BE/models"
	"Forum_BE/responses"
	"Forum_BE/services"
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
		Content string `json:"content" binding:"required"`
		Status  string `json:"status"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID := c.GetUint("user_id")
	status := req.Status
	if status == "" {
		status = string(models.Pending)
	}

	post, err := pc.postService.CreatePost(req.Content, userID, models.PostStatus(status))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Post created successfully", "post": responses.ToPostResponse(post)})
}

func (pc *PostController) GetPostById(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	post, err := pc.postService.GetPostByID(uint(id))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Post not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"post": responses.ToPostResponse(post)})
}

func (pc *PostController) DeletePost(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := pc.postService.DeletePost(uint(id)); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Post deleted successfully"})
}

func (pc *PostController) UpdatePost(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var req struct {
		Content string `json:"content" binding:"required"`
		Status  string `json:"status"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	post, err := pc.postService.UpdatePost(uint(id), req.Content, models.PostStatus(req.Status))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Post update successfully", "post": responses.ToPostResponse(post)})
}

func (pc *PostController) ListPosts(c *gin.Context) {
	filters := make(map[string]interface{})

	userID := c.Query("user_id")
	if userID != "" {
		uID, err := strconv.Atoi(userID)
		if err == nil {
			filters["user_id"] = uint(uID)
		}
	}
	status := c.Query("status")
	if status != "" {
		filters["status"] = status
	}
	search := c.Query("search")
	if search != "" {
		filters["content LIKE ?"] = "%" + search + "%"
	}

	posts, err := pc.postService.ListPosts(filters)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	}

	var response []responses.PostResponse
	for _, post := range posts {
		response = append(response, responses.ToPostResponse(&post))
	}

	c.JSON(http.StatusOK, gin.H{"posts": response})
}
