package controllers

import (
	"Forum_BE/responses"
	"Forum_BE/services"
	"github.com/cloudinary/cloudinary-go/v2"
	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
)

type FileController struct {
	fileService  services.FileService
	cloudinary   *cloudinary.Cloudinary
	uploadPreset string
}

func NewFileController(fileService services.FileService, cld *cloudinary.Cloudinary, uploadPreset string) *FileController {
	return &FileController{
		fileService:  fileService,
		cloudinary:   cld,
		uploadPreset: uploadPreset,
	}
}

func (fc *FileController) CreateFile(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "file is required"})
		return
	}

	var req struct {
		EntityType string `form:"entity_type" binding:"oneof= post answer comment"`
		EntityID   uint   `form:"entity_id"`
	}
	if err := c.ShouldBind(&req); err != nil {
		req.EntityType = ""
		req.EntityID = 0
	}

	userID := c.GetUint("user_id")
	attachment, err := fc.fileService.CreateFile(file, userID, req.EntityType, req.EntityID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "File uploaded successfully",
		"file":    responses.ToFileResponse(attachment),
	})
}

func (fc *FileController) GetFile(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid file id"})
		return
	}

	file, err := fc.fileService.GetFileByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "file not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"file": responses.ToFileResponse(file),
	})
}

func (fc *FileController) DeleteFile(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid file id"})
		return
	}

	if err := fc.fileService.DeleteFile(uint(id)); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "File deleted successfully",
	})
}

func (fc *FileController) ListFiles(c *gin.Context) {
	filters := make(map[string]interface{})

	if search := c.Query("search"); search != "" {
		filters["search"] = search
	}
	if fileType := c.Query("type"); fileType != "" {
		filters["file_type"] = fileType
	}
	if entityType := c.Query("entity_type"); entityType != "" {
		filters["entity_type"] = entityType
	}
	if entityID := c.Query("entity_id"); entityID != "" {
		if id, err := strconv.ParseUint(entityID, 10, 64); err == nil {
			filters["entity_id"] = uint(id)
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

	files, total, err := fc.fileService.ListFiles(filters)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	var responseFiles []responses.FileResponse
	for _, file := range files {
		responseFiles = append(responseFiles, responses.ToFileResponse(&file))
	}

	c.JSON(http.StatusOK, gin.H{
		"files": responseFiles,
		"total": total,
	})
}

func (fc *FileController) DownloadFile(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid file id"})
		return
	}

	file, err := fc.fileService.GetFileByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "file not found"})
		return
	}

	c.Redirect(http.StatusTemporaryRedirect, file.URL)
}
