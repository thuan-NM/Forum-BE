package services

import (
	"Forum_BE/models"
	"Forum_BE/repositories"
	"Forum_BE/utils"
	"context"
	"encoding/json"
	"fmt"
	"github.com/cloudinary/cloudinary-go/v2"
	"github.com/cloudinary/cloudinary-go/v2/api/uploader"

	"github.com/go-redis/redis/v8"
	"gorm.io/gorm"
	"log"
	"mime/multipart"
	"path/filepath"
	"strings"
	"time"
)

type FileService interface {
	CreateFile(file *multipart.FileHeader, userID uint, entityType string, entityID uint) (*models.Attachment, error)
	GetFileByID(id uint) (*models.Attachment, error)
	DeleteFile(id uint) error
	ListFiles(filters map[string]interface{}) ([]models.Attachment, int, error)
	UpdateFile(attachment *models.Attachment) error // Thêm để cập nhật EntityType và EntityID
}

type fileService struct {
	fileRepo     repositories.FileRepository
	cloudinary   *cloudinary.Cloudinary
	uploadPreset string
	redisClient  *redis.Client
	db           *gorm.DB
}

func NewFileService(fileRepo repositories.FileRepository, cld *cloudinary.Cloudinary, uploadPreset string, redisClient *redis.Client, db *gorm.DB) FileService {
	return &fileService{
		fileRepo:     fileRepo,
		cloudinary:   cld,
		uploadPreset: uploadPreset,
		redisClient:  redisClient,
		db:           db,
	}
}

func (s *fileService) CreateFile(file *multipart.FileHeader, userID uint, entityType string, entityID uint) (*models.Attachment, error) {
	if file.Filename == "" {
		return nil, fmt.Errorf("file name is required")
	}

	ctx := context.Background()
	src, err := file.Open()
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %v", err)
	}
	defer src.Close()

	resp, err := s.cloudinary.Upload.Upload(ctx, src, uploader.UploadParams{
		PublicID:     fmt.Sprintf("%s_%d", strings.ReplaceAll(file.Filename, " ", "_"), time.Now().UnixNano()),
		UploadPreset: s.uploadPreset,
		ResourceType: "auto",
	})
	if err != nil {
		return nil, fmt.Errorf("failed to upload to Cloudinary: %v", err)
	}

	var thumbnailURL string
	if strings.HasPrefix(resp.ResourceType, "image") {
		img, err := s.cloudinary.Image(resp.PublicID)
		if err != nil {
			log.Printf("Failed to create image transformation: %v", err)
		} else {
			img.Transformation = "c_thumb,w_200,h_200"
			thumbnailURL, err = img.String()
			if err != nil {
				log.Printf("Failed to generate thumbnail URL: %v", err)
			}
		}
		img.Transformation = "c_thumb,w_200,h_200"
		thumbnailURL, err = img.String()
		if err != nil {
			log.Printf("Failed to generate thumbnail URL: %v", err)
		}
	}

	attachment := &models.Attachment{
		UserID:       userID,
		EntityType:   entityType,
		EntityID:     entityID,
		URL:          resp.SecureURL,
		ThumbnailURL: thumbnailURL,
		FileName:     file.Filename,
		FileType:     getFileType(filepath.Ext(file.Filename)),
		FileSize:     file.Size,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	if err := s.fileRepo.CreateFile(attachment); err != nil {
		return nil, fmt.Errorf("failed to create file: %v", err)
	}

	s.invalidateCache("files:*")
	return attachment, nil
}

func (s *fileService) GetFileByID(id uint) (*models.Attachment, error) {
	cacheKey := fmt.Sprintf("file:%d", id)
	ctx := context.Background()

	cached, err := s.redisClient.Get(ctx, cacheKey).Result()
	if err == nil {
		var file models.Attachment
		if err := json.Unmarshal([]byte(cached), &file); err == nil {
			log.Printf("Cache hit for file:%d", id)
			return &file, nil
		}
	}
	if err != redis.Nil {
		log.Printf("Redis error for file:%d: %v", id, err)
	}

	file, err := s.fileRepo.GetFileByID(id)
	if err != nil {
		return nil, fmt.Errorf("failed to get file: %v", err)
	}
	if file.DeletedAt.Valid {
		return nil, fmt.Errorf("file not found")
	}

	data, err := json.Marshal(file)
	if err == nil {
		if err := s.redisClient.Set(ctx, cacheKey, data, 2*time.Minute).Err(); err != nil {
			log.Printf("Failed to set cache for file:%d: %v", id, err)
		}
	}

	return file, nil
}

func (s *fileService) DeleteFile(id uint) error {
	file, err := s.fileRepo.GetFileByID(id)
	if err != nil {
		return fmt.Errorf("failed to get file: %v", err)
	}
	if file.DeletedAt.Valid {
		return fmt.Errorf("file not found")
	}

	if err := s.fileRepo.DeleteFile(id); err != nil {
		return fmt.Errorf("failed to delete file: %v", err)
	}

	s.invalidateCache(fmt.Sprintf("file:%d", id))
	s.invalidateCache("files:*")
	return nil
}

func (s *fileService) UpdateFile(attachment *models.Attachment) error {
	if err := s.fileRepo.UpdateFile(attachment); err != nil {
		return fmt.Errorf("failed to update file: %v", err)
	}
	s.invalidateCache(fmt.Sprintf("file:%d", attachment.ID))
	s.invalidateCache("files:*")
	return nil
}

type FileListResponse struct {
	Files []models.Attachment `json:"files"`
	Total int                 `json:"total"`
}

func (s *fileService) ListFiles(filters map[string]interface{}) ([]models.Attachment, int, error) {
	cacheKey := utils.GenerateCacheKey("files", 0, filters)
	ctx := context.Background()

	cached, err := s.redisClient.Get(ctx, cacheKey).Result()
	if err == nil {
		var response FileListResponse
		if err := json.Unmarshal([]byte(cached), &response); err == nil {
			log.Printf("Cache hit for %s", cacheKey)
			return response.Files, response.Total, nil
		}
	}
	if err != redis.Nil {
		log.Printf("Redis error for %s: %v", cacheKey, err)
	}

	files, total, err := s.fileRepo.ListFiles(filters)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list files: %v", err)
	}

	response := FileListResponse{
		Files: files,
		Total: int(total),
	}

	data, err := json.Marshal(response)
	if err == nil {
		if err := s.redisClient.Set(ctx, cacheKey, data, 2*time.Minute).Err(); err != nil {
			log.Printf("Failed to set cache for %s: %v", cacheKey, err)
		}
	}

	return files, int(total), nil
}

func getFileType(ext string) string {
	ext = strings.ToLower(ext)
	switch ext {
	case ".jpg", ".jpeg", ".png", ".gif":
		return "image"
	case ".pdf", ".doc", ".docx", ".ppt", ".pptx", ".xls", ".xlsx":
		return "document"
	case ".mp4", ".avi", ".mov":
		return "video"
	case ".mp3", ".wav", ".m4a":
		return "audio"
	default:
		return "other"
	}
}

func (s *fileService) invalidateCache(pattern string) {
	ctx := context.Background()
	keys, err := s.redisClient.Keys(ctx, pattern).Result()
	if err != nil {
		log.Printf("Failed to get cache keys for pattern %s: %v", pattern, err)
		return
	}
	if len(keys) > 0 {
		if err := s.redisClient.Del(ctx, keys...).Err(); err != nil {
			log.Printf("Failed to delete cache keys %v: %v", keys, err)
		}
	}
}
