package controllers

import (
	"Forum_BE/responses"
	"Forum_BE/services"
	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
)

type ReactionController struct {
	reactionService services.ReactionService
}

func NewReactionController(s services.ReactionService) *ReactionController {
	return &ReactionController{reactionService: s}
}

func (rc *ReactionController) CreateReaction(c *gin.Context) {
	var req struct {
		PostID    *uint `json:"post_id"`
		CommentID *uint `json:"comment_id"`
		AnswerID  *uint `json:"answer_id"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request: " + err.Error()})
		return
	}
	userID := c.GetUint("user_id")
	reaction, err := rc.reactionService.CreateReaction(userID, req.PostID, req.CommentID, req.AnswerID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"message":  "Reaction created successfully",
		"reaction": responses.ToReactionResponse(reaction),
	})
}

func (rc *ReactionController) GetReactionByID(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid reaction id"})
		return
	}
	reaction, err := rc.reactionService.GetReactionByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "reaction not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"reaction": responses.ToReactionResponse(reaction),
	})
}

func (rc *ReactionController) UpdateReaction(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid reaction id"})
		return
	}
	var req struct {
		PostID    *uint `json:"post_id"`
		CommentID *uint `json:"comment_id"`
		AnswerID  *uint `json:"answer_id"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request: " + err.Error()})
		return
	}
	userID := c.GetUint("user_id")
	reaction, err := rc.reactionService.UpdateReaction(uint(id), userID, req.PostID, req.CommentID, req.AnswerID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"message":  "Reaction updated successfully",
		"reaction": responses.ToReactionResponse(reaction),
	})
}

func (rc *ReactionController) DeleteReaction(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid reaction id"})
		return
	}
	userID := c.GetUint("user_id")
	if err := rc.reactionService.DeleteReaction(uint(id), userID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"message": "Reaction deleted successfully",
	})
}

func (rc *ReactionController) ListReactions(c *gin.Context) {
	filters := make(map[string]interface{})
	if userID := c.Query("user_id"); userID != "" {
		if uID, err := strconv.ParseUint(userID, 10, 64); err == nil {
			filters["user_id"] = uint(uID)
		}
	}
	if postID := c.Query("post_id"); postID != "" {
		if pID, err := strconv.ParseUint(postID, 10, 64); err == nil {
			filters["post_id"] = uint(pID)
		}
	}
	if commentID := c.Query("comment_id"); commentID != "" {
		if cID, err := strconv.ParseUint(commentID, 10, 64); err == nil {
			filters["comment_id"] = uint(cID)
		}
	}
	if answerID := c.Query("answer_id"); answerID != "" {
		if aID, err := strconv.ParseUint(answerID, 10, 64); err == nil {
			filters["answer_id"] = uint(aID)
		}
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
	reactions, total, err := rc.reactionService.ListReactions(filters)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	var responseReactions []responses.ReactionResponse
	for _, reaction := range reactions {
		responseReactions = append(responseReactions, responses.ToReactionResponse(&reaction))
	}
	c.JSON(http.StatusOK, gin.H{
		"reactions": responseReactions,
		"total":     total,
	})
}

func (rc *ReactionController) GetReactionCount(c *gin.Context) {
	var postID, commentID, answerID *uint
	if pID := c.Query("post_id"); pID != "" {
		if id, err := strconv.ParseUint(pID, 10, 64); err == nil {
			postID = new(uint)
			*postID = uint(id)
		}
	}
	if cID := c.Query("comment_id"); cID != "" {
		if id, err := strconv.ParseUint(cID, 10, 64); err == nil {
			commentID = new(uint)
			*commentID = uint(id)
		}
	}
	if aID := c.Query("answer_id"); aID != "" {
		if id, err := strconv.ParseUint(aID, 10, 64); err == nil {
			answerID = new(uint)
			*answerID = uint(id)
		}
	}
	count, err := rc.reactionService.GetReactionCount(postID, commentID, answerID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"count": count,
	})
}

func (rc *ReactionController) CheckUserReaction(c *gin.Context) {
	userID := c.GetUint("user_id")
	var postID, commentID, answerID *uint
	if pID := c.Query("post_id"); pID != "" {
		if id, err := strconv.ParseUint(pID, 10, 64); err == nil {
			postID = new(uint)
			*postID = uint(id)
		} else {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid post_id"})
			return
		}
	}
	if cID := c.Query("comment_id"); cID != "" {
		if id, err := strconv.ParseUint(cID, 10, 64); err == nil {
			commentID = new(uint)
			*commentID = uint(id)
		} else {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid comment_id"})
			return
		}
	}
	if aID := c.Query("answer_id"); aID != "" {
		if id, err := strconv.ParseUint(aID, 10, 64); err == nil {
			answerID = new(uint)
			*answerID = uint(id)
		} else {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid answer_id"})
			return
		}
	}
	hasReacted, reaction, err := rc.reactionService.CheckUserReaction(userID, postID, commentID, answerID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	response := gin.H{
		"has_reacted": hasReacted,
	}
	if hasReacted {
		response["reaction"] = responses.ToReactionResponse(reaction)
	}
	c.JSON(http.StatusOK, response)
}

func (rc *ReactionController) GetReactionStatus(c *gin.Context) {
	userID := c.GetUint("user_id")
	var postID, commentID, answerID *uint
	if pID := c.Query("post_id"); pID != "" {
		if id, err := strconv.ParseUint(pID, 10, 64); err == nil {
			postID = new(uint)
			*postID = uint(id)
		} else {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid post_id"})
			return
		}
	}
	if cID := c.Query("comment_id"); cID != "" {
		if id, err := strconv.ParseUint(cID, 10, 64); err == nil {
			commentID = new(uint)
			*commentID = uint(id)
		} else {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid comment_id"})
			return
		}
	}
	if aID := c.Query("answer_id"); aID != "" {
		if id, err := strconv.ParseUint(aID, 10, 64); err == nil {
			answerID = new(uint)
			*answerID = uint(id)
		} else {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid answer_id"})
			return
		}
	}

	if err := rc.reactionService.ValidateReactionID(postID, commentID, answerID); err != nil { // Sửa lại gọi qua reactionService
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	hasReacted, reaction, err := rc.reactionService.CheckUserReaction(userID, postID, commentID, answerID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	count, err := rc.reactionService.GetReactionCount(postID, commentID, answerID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	response := gin.H{
		"has_reacted": hasReacted,
		"count":       count,
	}
	if hasReacted {
		response["reaction"] = responses.ToReactionResponse(reaction)
	}
	c.JSON(http.StatusOK, response)
}
