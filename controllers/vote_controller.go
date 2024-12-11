package controllers

import (
	"Forum_BE/responses"
	"Forum_BE/services"
	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
)

type VoteController struct {
	voteService services.VoteService
}

func NewVoteController(v services.VoteService) *VoteController {
	return &VoteController{voteService: v}
}

// CastVote xử lý yêu cầu tạo hoặc thay đổi vote
func (vc *VoteController) CastVote(c *gin.Context) {
	var req struct {
		VotableType string `json:"votable_type" binding:"required,oneof=question answer comment"`
		VotableID   uint   `json:"votable_id" binding:"required"`
		VoteType    string `json:"vote_type" binding:"required,oneof=upvote downvote"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID := c.GetUint("user_id") // Middleware đã thêm user_id vào context

	vote, err := vc.voteService.CastVote(userID, req.VotableType, req.VotableID, req.VoteType)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Lấy số lượng vote sau khi cast
	voteCount, err := vc.voteService.GetVoteCount(req.VotableType, req.VotableID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get vote count"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":    "Vote casted successfully",
		"vote":       responses.ToVoteResponse(vote),
		"vote_count": voteCount,
	})
}

// GetVote lấy vote theo ID
func (vc *VoteController) GetVote(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.ParseUint(idParam, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid vote id"})
		return
	}

	vote, err := vc.voteService.GetVoteByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "vote not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"vote": responses.ToVoteResponse(vote),
	})
}

// UpdateVote cập nhật loại vote
func (vc *VoteController) UpdateVote(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.ParseUint(idParam, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid vote id"})
		return
	}

	var req struct {
		VoteType string `json:"vote_type" binding:"required,oneof=upvote downvote"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	vote, err := vc.voteService.UpdateVote(uint(id), req.VoteType)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Lấy số lượng vote sau khi cập nhật
	voteCount, err := vc.voteService.GetVoteCount(vote.VotableType, vote.VotableID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get vote count"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":    "Vote updated successfully",
		"vote":       responses.ToVoteResponse(vote),
		"vote_count": voteCount,
	})
}

// DeleteVote xóa vote theo ID
func (vc *VoteController) DeleteVote(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.ParseUint(idParam, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid vote id"})
		return
	}

	if err := vc.voteService.DeleteVote(uint(id)); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Vote deleted successfully",
	})
}

// ListVotes liệt kê tất cả các vote
func (vc *VoteController) ListVotes(c *gin.Context) {
	votes, err := vc.voteService.ListVotes()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list votes"})
		return
	}

	var responseVotes []responses.VoteResponse
	for _, vote := range votes {
		responseVotes = append(responseVotes, responses.ToVoteResponse(&vote))
	}

	c.JSON(http.StatusOK, gin.H{
		"votes": responseVotes,
	})
}
