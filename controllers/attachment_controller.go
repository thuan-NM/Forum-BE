package controllers

import (
	"Forum_BE/responses"
	"Forum_BE/services"
	"encoding/json"
	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
)

type AttachmentController struct {
	attachmentService services.AttachmentService
}

func NewAttachmentController(as services.AttachmentService) *AttachmentController {
	return &AttachmentController{attachmentService: as}
}

func (ac *AttachmentController) UploadAttachment(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "File is required"})
		return
	}

	userID := c.GetUint("user_id")

	attachment, err := ac.attachmentService.UploadAttachment(file, userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":    "File uploaded successfully",
		"attachment": responses.ToAttachmentResponse(attachment),
	})
}

func (ac *AttachmentController) GetAttachment(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.ParseUint(idParam, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid attachment id"})
		return
	}

	attachment, err := ac.attachmentService.GetAttachmentByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Attachment not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"attachment": responses.ToAttachmentResponse(attachment),
	})
}

func (ac *AttachmentController) UpdateAttachment(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.ParseUint(idParam, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid attachment id"})
		return
	}

	var req struct {
		Metadata json.RawMessage `json:"metadata"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	attachment, err := ac.attachmentService.UpdateAttachment(uint(id), req.Metadata)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":    "Attachment updated successfully",
		"attachment": responses.ToAttachmentResponse(attachment),
	})
}

func (ac *AttachmentController) DeleteAttachment(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.ParseUint(idParam, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid attachment id"})
		return
	}

	userID := c.GetUint("user_id")
	attachment, err := ac.attachmentService.GetAttachmentByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Attachment not found"})
		return
	}
	if attachment.UserID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "You do not own this attachment"})
		return
	}

	if err := ac.attachmentService.DeleteAttachment(uint(id)); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Attachment deleted successfully",
	})
}

func (ac *AttachmentController) ListAttachments(c *gin.Context) {
	filters := make(map[string]interface{})

	userID := c.Query("user_id")
	if userID != "" {
		uID, err := strconv.ParseUint(userID, 10, 64)
		if err == nil {
			filters["user_id"] = uint(uID)
		}
	}

	fileType := c.Query("file_type")
	if fileType != "" {
		filters["file_type"] = fileType
	}

	limitStr := c.Query("limit")
	pageStr := c.Query("page")
	limit := 10 // Default limit
	page := 1   // Default page
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}
	if pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}
	filters["limit"] = limit
	filters["page"] = page

	attachments, total, err := ac.attachmentService.ListAttachments(filters)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list attachments"})
		return
	}

	var responseAttachments []responses.AttachmentResponse
	for _, attachment := range attachments {
		responseAttachments = append(responseAttachments, responses.ToAttachmentResponse(&attachment))
	}

	c.JSON(http.StatusOK, gin.H{
		"attachments": responseAttachments,
		"total":       total,
	})
}
