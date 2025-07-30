package services

import (
	config "Forum_BE/cloudinaryconfig"
	"Forum_BE/models"
	"Forum_BE/repositories"
	"Forum_BE/utils"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/cloudinary/cloudinary-go/v2"
	"github.com/cloudinary/cloudinary-go/v2/api/uploader"
	"github.com/go-redis/redis/v8"
	"log"
	"mime/multipart"
	"os"
	"strings"
	"time"
)

type AttachmentService interface {
	UploadAttachment(file *multipart.FileHeader, userID uint) (*models.Attachment, error)
	GetAttachmentByID(id uint) (*models.Attachment, error)
	UpdateAttachment(id uint, metadata json.RawMessage) (*models.Attachment, error)
	DeleteAttachment(id uint) error
	ListAttachments(filters map[string]interface{}) ([]models.Attachment, int, error)
}

type attachmentService struct {
	attachmentRepo repositories.AttachmentRepository
	cloudinary     *cloudinary.Cloudinary
	redisClient    *redis.Client
	uploadPreset   string
}

func NewAttachmentService(repo repositories.AttachmentRepository, redisClient *redis.Client) AttachmentService {
	cld, err := config.NewCloudinaryClient()
	if err != nil {
		log.Fatalf("Failed to initialize Cloudinary: %v", err)
	}
	uploadPreset := os.Getenv("CLOUDINARY_UPLOAD_PRESET")
	return &attachmentService{
		attachmentRepo: repo,
		cloudinary:     cld,
		redisClient:    redisClient,
		uploadPreset:   uploadPreset,
	}
}

func (s *attachmentService) UploadAttachment(file *multipart.FileHeader, userID uint) (*models.Attachment, error) {
	// Validate file size (10MB limit)
	if file.Size > 10*1024*1024 {
		return nil, errors.New("file size exceeds 10MB")
	}

	// Validate Content-Type early
	contentType := file.Header.Get("Content-Type")
	if !strings.HasPrefix(contentType, "image/") {
		return nil, errors.New("only image files are supported")
	}

	// Open file with buffered reader
	ctx := context.Background()
	start := time.Now()
	fileContent, err := file.Open()
	if err != nil {
		log.Printf("Failed to open file: %v", err.Error())
		return nil, err
	}
	defer fileContent.Close()
	log.Printf("File open took %v", time.Since(start))

	// Upload to Cloudinary with optimized parameters
	uploadParams := uploader.UploadParams{
		Folder:         "forum_attachments",
		Transformation: "q_auto,f_auto,w_1000,h_1000,c_limit",
		Eager:          "w_300,h_300,c_fill,g_auto,f_auto", // Single string for eager transformation
	}
	if s.uploadPreset != "" {
		uploadParams.UploadPreset = s.uploadPreset
	}

	start = time.Now()
	uploadResult, err := s.cloudinary.Upload.Upload(ctx, fileContent, uploadParams)
	if err != nil {
		log.Printf("Failed to upload to Cloudinary: %v", err.Error())
		return nil, err
	}
	log.Printf("Cloudinary upload took %v", time.Since(start))

	// Get thumbnail URL from eager transformation
	thumbnailURL := ""
	if len(uploadResult.Eager) > 0 {
		thumbnailURL = uploadResult.Eager[0].SecureURL
	} else {
		// Fallback: Generate thumbnail URL manually
		asset, err := s.cloudinary.Image(uploadResult.PublicID)
		if err != nil {
			log.Printf("Failed to create image asset: %v", err.Error())
			// Rollback Cloudinary upload
			_, _ = s.cloudinary.Upload.Destroy(ctx, uploader.DestroyParams{PublicID: uploadResult.PublicID})
			return nil, err
		}
		asset.Transformation = "w_300,h_300,c_fill,g_auto,f_auto"
		thumbnailURL, err = asset.String()
		if err != nil {
			log.Printf("Failed to generate thumbnail URL: %v", err.Error())
			// Rollback Cloudinary upload
			_, _ = s.cloudinary.Upload.Destroy(ctx, uploader.DestroyParams{PublicID: uploadResult.PublicID})
			return nil, err
		}
	}

	// Create metadata
	start = time.Now()
	metadata, err := json.Marshal(map[string]interface{}{
		"width":  uploadResult.Width,
		"height": uploadResult.Height,
		"format": uploadResult.Format,
	})
	if err != nil {
		log.Printf("Failed to marshal metadata: %v", err.Error())
		// Rollback Cloudinary upload
		_, _ = s.cloudinary.Upload.Destroy(ctx, uploader.DestroyParams{PublicID: uploadResult.PublicID})
		return nil, err
	}
	log.Printf("Metadata creation took %v", time.Since(start))

	// Create Attachment record
	attachment := &models.Attachment{
		UserID:       userID,
		URL:          uploadResult.SecureURL,
		ThumbnailURL: thumbnailURL,
		FileName:     file.Filename,
		FileType:     utils.DetermineFileType(contentType),
		FileSize:     file.Size,
		Metadata:     metadata,
	}

	// Save to database
	start = time.Now()
	if err := s.attachmentRepo.CreateAttachment(attachment); err != nil {
		log.Printf("Failed to create attachment: %v", err.Error())
		// Rollback Cloudinary upload
		_, _ = s.cloudinary.Upload.Destroy(ctx, uploader.DestroyParams{PublicID: uploadResult.PublicID})
		return nil, err
	}
	log.Printf("Database save took %v", time.Since(start))

	// Invalidate cache asynchronously
	go func() {
		start := time.Now()
		s.invalidateCache("attachments:*")
		log.Printf("Cache invalidation took %v", time.Since(start))
	}()

	return attachment, nil
}

func (s *attachmentService) GetAttachmentByID(id uint) (*models.Attachment, error) {
	cacheKey := fmt.Sprintf("attachment:%d", id)
	ctx := context.Background()

	// Kiểm tra cache
	var attachment *models.Attachment
	cached, err := s.redisClient.Get(ctx, cacheKey).Result()
	if err == nil {
		if err := json.Unmarshal([]byte(cached), &attachment); err == nil {
			log.Printf("Cache hit for attachment:%d", id)
			return attachment, nil
		}
	}

	// Lấy từ database
	attachment, err = s.attachmentRepo.GetAttachmentByID(id)
	if err != nil {
		log.Printf("Failed to get attachment %d: %v", id, err)
		return nil, err
	}

	// Lưu vào cache
	data, err := json.Marshal(attachment)
	if err == nil {
		if err := s.redisClient.Set(ctx, cacheKey, data, 2*time.Minute).Err(); err != nil {
			log.Printf("Failed to set cache for attachment:%d: %v", id, err)
		}
	}

	return attachment, nil
}

func (s *attachmentService) UpdateAttachment(id uint, metadata json.RawMessage) (*models.Attachment, error) {
	attachment, err := s.attachmentRepo.GetAttachmentByID(id)
	if err != nil {
		log.Printf("Failed to get attachment %d: %v", id, err)
		return nil, err
	}

	if metadata != nil {
		attachment.Metadata = metadata
	}

	if err := s.attachmentRepo.UpdateAttachment(attachment); err != nil {
		log.Printf("Failed to update attachment %d: %v", id, err)
		return nil, err
	}

	// Invalidate cache
	s.invalidateCache(fmt.Sprintf("attachment:%d", id))
	s.invalidateCache("attachments:*")
	return attachment, nil
}

func (s *attachmentService) DeleteAttachment(id uint) error {
	attachment, err := s.attachmentRepo.GetAttachmentByID(id)
	if err != nil {
		log.Printf("Failed to get attachment %d: %v", id, err)
		return err
	}

	// Xóa file trên Cloudinary
	publicID := utils.ExtractPublicID(attachment.URL)
	ctx := context.Background()
	_, err = s.cloudinary.Upload.Destroy(ctx, uploader.DestroyParams{PublicID: publicID})
	if err != nil {
		log.Printf("Failed to delete file from Cloudinary: %v", err)
	}

	// Xóa record trong database
	if err := s.attachmentRepo.DeleteAttachment(id); err != nil {
		log.Printf("Failed to delete attachment %d: %v", id, err)
		return err
	}

	// Invalidate cache
	s.invalidateCache(fmt.Sprintf("attachment:%d", id))
	s.invalidateCache("attachments:*")
	return nil
}

func (s *attachmentService) ListAttachments(filters map[string]interface{}) ([]models.Attachment, int, error) {
	cacheKey := utils.GenerateCacheKey("attachments", 0, filters)
	ctx := context.Background()

	// Kiểm tra cache
	var attachments []models.Attachment
	cached, err := s.redisClient.Get(ctx, cacheKey).Result()
	if err == nil {
		var cachedData struct {
			Attachments []models.Attachment
			Total       int
		}
		if err := json.Unmarshal([]byte(cached), &cachedData); err == nil {
			log.Printf("Cache hit for attachments")
			return cachedData.Attachments, cachedData.Total, nil
		}
	}

	// Lấy từ database
	attachments, total, err := s.attachmentRepo.ListAttachments(filters)
	if err != nil {
		log.Printf("Failed to list attachments: %v", err)
		return nil, 0, err
	}

	// Lưu vào cache
	cacheData := struct {
		Attachments []models.Attachment
		Total       int
	}{Attachments: attachments, Total: total}
	data, err := json.Marshal(cacheData)
	if err == nil {
		if err := s.redisClient.Set(ctx, cacheKey, data, 2*time.Minute).Err(); err != nil {
			log.Printf("Failed to set cache for attachments: %v", err)
		}
	}

	return attachments, total, nil
}

func (s *attachmentService) invalidateCache(pattern string) {
	ctx := context.Background()
	var keys []string
	var cursor uint64
	for {
		batch, nextCursor, err := s.redisClient.Scan(ctx, cursor, pattern, 100).Result()
		if err != nil {
			log.Printf("Failed to scan cache keys for pattern %s: %v", pattern, err)
			return
		}
		keys = append(keys, batch...)
		cursor = nextCursor
		if cursor == 0 {
			break
		}
	}
	if len(keys) > 0 {
		if err := s.redisClient.Del(ctx, keys...).Err(); err != nil {
			log.Printf("Failed to delete cache keys %v: %v", keys, err)
		} else {
			log.Printf("Deleted cache keys %v", keys)
		}
	}
}
