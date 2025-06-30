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
		ReactableID   uint   `json:"reactable_id" binding:"required"`
		ReactableType string `json:"reactable_type" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID := c.GetUint("user_id")
	reaction, err := rc.reactionService.CreateReaction(userID, req.ReactableID, req.ReactableType)
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
	idParam := c.Param("id")
	id, err := strconv.ParseUint(idParam, 10, 64)
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
	idParam := c.Param("id")
	id, err := strconv.ParseUint(idParam, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid reaction id"})
		return
	}

	var req struct {
		ReactableID   uint   `json:"reactable_id" binding:"required"`
		ReactableType string `json:"reactable_type" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID := c.GetUint("user_id")
	reaction, err := rc.reactionService.UpdateReaction(uint(id), userID, req.ReactableID, req.ReactableType)
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
	idParam := c.Param("id")
	id, err := strconv.ParseUint(idParam, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid reaction id"})
		return
	}

	if err := rc.reactionService.DeleteReaction(uint(id)); err != nil {
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
	if reactableID := c.Query("reactable_id"); reactableID != "" {
		if rID, err := strconv.ParseUint(reactableID, 10, 64); err == nil {
			filters["reactable_id"] = uint(rID)
		}
	}
	if reactableType := c.Query("reactable_type"); reactableType != "" {
		filters["reactable_type"] = reactableType
	}

	reactions, total, err := rc.reactionService.ListReactions(filters)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list reactions"})
		return
	}

	var response []responses.ReactionResponse
	for _, r := range reactions {
		response = append(response, responses.ToReactionResponse(&r))
	}

	c.JSON(http.StatusOK, gin.H{
		"reactions": response,
		"total":     total,
	})
}
