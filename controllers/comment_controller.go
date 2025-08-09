package controllers

import (
	"Forum_BE/responses"
	"Forum_BE/services"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
	"strconv"
)

type CommentController struct {
	commentService services.CommentService
	voteService    services.VoteService
}

func NewCommentController(c services.CommentService, v services.VoteService) *CommentController {
	return &CommentController{commentService: c, voteService: v}
}

func (cc *CommentController) CreateComment(c *gin.Context) {
	var req struct {
		Content  string `json:"content" binding:"required"`
		PostID   *uint  `json:"post_id"`
		AnswerID *uint  `json:"answer_id"`
		ParentID *uint  `json:"parent_id"`
		Status   string `json:"status"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request: " + err.Error()})
		return
	}
	userID := c.GetUint("user_id")
	comment, err := cc.commentService.CreateComment(req.Content, userID, req.PostID, req.AnswerID, req.ParentID, req.Status)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Comment created successfully",
		"comment": responses.ToCommentResponse(comment),
	})
}

func (cc *CommentController) GetComment(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid comment id"})
		return
	}

	comment, err := cc.commentService.GetCommentByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "comment not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"comment": responses.ToCommentResponse(comment),
	})
}

func (cc *CommentController) EditComment(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid comment id"})
		return
	}

	var req struct {
		Content string `json:"content" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request: " + err.Error()})
		return
	}

	comment, err := cc.commentService.UpdateComment(uint(id), req.Content)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Comment updated successfully",
		"comment": responses.ToCommentResponse(comment),
	})
}

func (cc *CommentController) DeleteComment(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid comment id"})
		return
	}

	if err := cc.commentService.DeleteComment(uint(id)); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Comment deleted successfully",
	})
}

func (cc *CommentController) ListComments(c *gin.Context) {
	filters := make(map[string]interface{})

	postID := c.Query("post_id")
	if postID != "" {
		if postID, err := strconv.ParseUint(postID, 10, 64); err == nil {
			filters["post_id"] = uint(postID)
		}
	}

	answerID := c.Query("answer_id")
	if answerID != "" {
		if aID, err := strconv.ParseUint(answerID, 10, 64); err == nil {
			filters["answer_id"] = uint(aID)
		}
	}

	userID := c.Query("user_id")
	if userID != "" {
		if uID, err := strconv.ParseUint(userID, 10, 64); err == nil {
			filters["user_id"] = uint(uID)
		}
	}
	if status := c.Query("status"); status != "" {
		filters["status"] = status
	}
	if search := c.Query("search"); search != "" {
		filters["content"] = search
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

	comments, total, err := cc.commentService.ListComments(filters)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	var responseComments []responses.CommentResponse
	for _, comment := range comments {
		responseComments = append(responseComments, responses.ToCommentResponse(&comment))
	}

	c.JSON(http.StatusOK, gin.H{
		"comments": responseComments,
		"total":    total,
	})
}

func (cc *CommentController) ListReplies(c *gin.Context) {
	filters := make(map[string]interface{})

	parentID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid comment id"})
		return
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
	if status := c.Query("status"); status != "" {
		filters["status"] = status
	}
	replies, total, err := cc.commentService.ListReplies(uint(parentID), filters)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	var responseReplies []responses.CommentResponse
	for _, reply := range replies {
		responseReplies = append(responseReplies, responses.ToCommentResponse(&reply))
	}

	c.JSON(http.StatusOK, gin.H{
		"replies": responseReplies,
		"total":   total,
	})
}

// Thêm vào comment_controller.go

func (cc *CommentController) GetAllComments(c *gin.Context) {
	filters := make(map[string]interface{})

	if search := c.Query("search"); search != "" {
		filters["search"] = search
	}

	if status := c.Query("status"); status != "" {
		filters["status"] = status
	}
	if typefilter := c.Query("typefilter"); typefilter != "" {
		filters["typefilter"] = typefilter
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

	comments, total, err := cc.commentService.GetAllComments(filters)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	var responseComments []responses.CommentResponse
	for _, comment := range comments {
		responseComments = append(responseComments, responses.ToCommentResponse(&comment))
	}

	c.JSON(http.StatusOK, gin.H{
		"comments": responseComments,
		"total":    total,
	})
}
func (cc *CommentController) UpdateStatus(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.ParseUint(idParam, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid comment id"})
	}
	var req struct {
		Status string `json:"status" binding:"required"`
	}
	log.Printf(req.Status)

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	comment, err := cc.commentService.UpdateCommentStatus(uint(id), req.Status)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"message": "Comment updated successfully",
		"comment": responses.ToCommentResponse(comment),
	})
}
