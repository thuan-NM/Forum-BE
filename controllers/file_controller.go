package controllers

import (
	"Forum_BE/models"
	"Forum_BE/responses"
	"Forum_BE/services"
	"github.com/cloudinary/cloudinary-go/v2"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"net/http"
	"path/filepath"
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
		c.JSON(http.StatusBadRequest, gin.H{"error": "bắt buộc phải có tệp"})
		return
	}

	userID := c.GetUint("user_id")
	attachment, err := fc.fileService.CreateFile(fc.fileService.GetDB(), file, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Tải tệp lên thành công",
		"file":    responses.ToFileResponse(attachment),
	})
}

func (fc *FileController) GetFile(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID tệp không hợp lệ"})
		return
	}

	file, err := fc.fileService.GetFileByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Không tìm thấy tệp"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"file": responses.ToFileResponse(file),
	})
}

func (fc *FileController) DeleteFile(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID tệp không hợp lệ"})
		return
	}

	if err := fc.fileService.DeleteFile(uint(id)); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Xoá tệp thành công",
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
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID tệp không hợp lệ"})
		return
	}

	file, err := fc.fileService.GetFileByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Không tìm thấy tệp"})
		return
	}

	c.Redirect(http.StatusTemporaryRedirect, file.URL)
}

type FileResponse struct {
	Message string `json:"message"`
	File    struct {
		ID        uint   `json:"id"`
		URL       string `json:"url"`
		Thumbnail string `json:"thumbnail_url,omitempty"`
		Name      string `json:"file_name"`
		Type      string `json:"file_type"`
		Size      int64  `json:"file_size"`
	} `json:"file"`
}

func UploadFile(c *gin.Context, db *gorm.DB) {
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Không có tệp được tải lên"})
		return
	}

	// Generate a unique file path (adjust based on your storage setup)
	filename := filepath.Base(file.Filename)
	filePath := "uploads/" + filename // Adjust to your storage path
	if err := c.SaveUploadedFile(file, filePath); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Lưu tệp thất bại"})
		return
	}

	// Create Attachment record
	attachment := models.Attachment{
		FileName:     filename,
		FileType:     file.Header.Get("Content-Type"),
		FileSize:     file.Size,
		URL:          "/uploads/" + filename, // Adjust URL based on your server setup
		ThumbnailURL: "",                     // Add logic for thumbnail if needed
		UserID:       c.GetUint("user_id"),   // Assuming user ID from auth middleware
	}

	if err := db.Create(&attachment).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Lưu tệp đính kèm thất bại"})
		return
	}

	// Return FileResponse
	response := FileResponse{
		Message: "Tải tệp lên thành công",
	}
	response.File.ID = attachment.ID
	response.File.URL = attachment.URL
	response.File.Thumbnail = attachment.ThumbnailURL
	response.File.Name = attachment.FileName
	response.File.Type = attachment.FileType
	response.File.Size = attachment.FileSize

	c.JSON(http.StatusOK, response)
}
