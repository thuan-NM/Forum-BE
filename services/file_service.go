package services

import (
	"Forum_BE/models"
	"Forum_BE/repositories"
	"Forum_BE/utils"
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/cloudinary/cloudinary-go/v2"
	"github.com/cloudinary/cloudinary-go/v2/api/uploader"
	"github.com/go-redis/redis/v8"
	"gorm.io/gorm"
	"log"
	"mime/multipart"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

type FileService interface {
	CreateFile(tx *gorm.DB, file *multipart.FileHeader, userID uint) (*models.Attachment, error)
	ProcessImageTags(tx *gorm.DB, content string, files []*multipart.FileHeader, userID uint) (string, []uint64, error)
	GetFileByID(id uint) (*models.Attachment, error)
	DeleteFile(id uint) error
	ListFiles(filters map[string]interface{}) ([]models.Attachment, int, error)
	UpdateFile(attachment *models.Attachment) error
	CleanupOrphanedFiles() error
	GetDB() *gorm.DB
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

func (s *fileService) GetDB() *gorm.DB {
	return s.db
}

func (s *fileService) CreateFile(tx *gorm.DB, file *multipart.FileHeader, userID uint) (*models.Attachment, error) {
	if file.Filename == "" {
		log.Printf("File name is empty")
		return nil, fmt.Errorf("file name is required")
	}

	log.Printf("Uploading file: %s (size: %d)", file.Filename, file.Size)
	ctx := context.Background()
	src, err := file.Open()
	if err != nil {
		log.Printf("Failed to open file %s: %v", file.Filename, err)
		return nil, fmt.Errorf("failed to open file: %v", err)
	}
	defer src.Close()

	resp, err := s.cloudinary.Upload.Upload(ctx, src, uploader.UploadParams{
		PublicID:     fmt.Sprintf("%s_%d", strings.ReplaceAll(file.Filename, " ", "_"), time.Now().UnixNano()),
		UploadPreset: s.uploadPreset,
		ResourceType: "image",
	})
	if err != nil {
		log.Printf("Failed to upload file %s to Cloudinary: %v", file.Filename, err)
		return nil, fmt.Errorf("failed to upload to Cloudinary: %v", err)
	}
	log.Printf("Uploaded file %s to Cloudinary, URL: %s", file.Filename, resp.SecureURL)

	var thumbnailURL string
	img, err := s.cloudinary.Image(resp.PublicID)
	if err != nil {
		log.Printf("Failed to create image transformation for %s: %v", file.Filename, err)
	} else {
		img.Transformation = "c_thumb,w_200,h_200"
		thumbnailURL, err = img.String()
		if err != nil {
			log.Printf("Failed to generate thumbnail URL for %s: %v", file.Filename, err)
		}
	}

	attachment := &models.Attachment{
		UserID:       userID,
		URL:          resp.SecureURL,
		ThumbnailURL: thumbnailURL,
		FileName:     file.Filename,
		FileType:     getFileType(filepath.Ext(file.Filename)),
		FileSize:     file.Size,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	if err := s.fileRepo.CreateFileWithTx(tx, attachment); err != nil {
		log.Printf("Failed to create file record for %s in database: %v", file.Filename, err)
		return nil, fmt.Errorf("failed to create file: %v", err)
	}
	log.Printf("Created file record for %s, ID: %d", file.Filename, attachment.ID)

	s.invalidateCache("files:*")
	return attachment, nil
}

func (s *fileService) ProcessImageTags(tx *gorm.DB, content string, files []*multipart.FileHeader, userID uint) (string, []uint64, error) {
	// Tìm tất cả thẻ <img> trong content
	imgTagRegex := regexp.MustCompile(`<img[^>]+src=["'](.*?)["'][^>]*>`)
	matches := imgTagRegex.FindAllStringSubmatch(content, -1)
	var attachmentIDs []uint64
	updatedContent := content

	// Tạo map để ánh xạ filename với file từ form data
	fileMap := make(map[string]*multipart.FileHeader)
	for _, file := range files {
		fileMap[file.Filename] = file
	}

	for _, match := range matches {
		imgTag := match[0]
		src := match[1]

		var fileName string
		var fileData interface{}
		var fileSize int64
		var err error

		if strings.HasPrefix(src, "data:image/") {
			// Xử lý base64
			parts := strings.Split(src, ",")
			if len(parts) != 2 {
				log.Printf("Invalid base64 image format in tag %s", imgTag)
				continue
			}
			data, err := base64.StdEncoding.DecodeString(parts[1])
			if err != nil {
				log.Printf("Failed to decode base64 image: %v", err)
				continue
			}
			fileType := strings.Split(strings.Split(parts[0], ";")[0], "/")[1]
			fileName = fmt.Sprintf("image_%d_%d.%s", userID, time.Now().UnixNano(), fileType)
			fileData = bytes.NewReader(data)
			fileSize = int64(len(data))
		} else if !strings.HasPrefix(src, "http") {
			// Xử lý đường dẫn cục bộ
			fileName = filepath.Base(src)
			file, exists := fileMap[fileName]
			if !exists {
				log.Printf("File %s not found in form data", fileName)
				continue
			}
			fileData = file
			fileSize = file.Size
		} else {
			// Bỏ qua nếu src đã là URL công khai
			continue
		}

		// Upload ảnh lên Cloudinary
		ctx := context.Background()
		var resp *uploader.UploadResult
		switch data := fileData.(type) {
		case *multipart.FileHeader:
			src, err := data.Open()
			if err != nil {
				log.Printf("Failed to open file %s: %v", fileName, err)
				continue
			}
			defer src.Close()
			resp, err = s.cloudinary.Upload.Upload(ctx, src, uploader.UploadParams{
				PublicID:     fmt.Sprintf("%s_%d", strings.ReplaceAll(fileName, " ", "_"), time.Now().UnixNano()),
				UploadPreset: s.uploadPreset,
				ResourceType: "image",
			})
		case *bytes.Reader:
			resp, err = s.cloudinary.Upload.Upload(ctx, data, uploader.UploadParams{
				PublicID:     fmt.Sprintf("%s_%d", strings.ReplaceAll(fileName, " ", "_"), time.Now().UnixNano()),
				UploadPreset: s.uploadPreset,
				ResourceType: "image",
			})
		}
		if err != nil {
			log.Printf("Failed to upload image %s to Cloudinary: %v", fileName, err)
			return "", nil, fmt.Errorf("failed to upload image: %v", err)
		}
		log.Printf("Uploaded image %s to Cloudinary, URL: %s", fileName, resp.SecureURL)

		// Tạo thumbnail
		var thumbnailURL string
		img, err := s.cloudinary.Image(resp.PublicID)
		if err != nil {
			log.Printf("Failed to create image transformation for %s: %v", fileName, err)
		} else {
			img.Transformation = "c_thumb,w_200,h_200"
			thumbnailURL, err = img.String()
			if err != nil {
				log.Printf("Failed to generate thumbnail URL for %s: %v", fileName, err)
			}
		}

		// Lưu attachment vào database
		attachment := &models.Attachment{
			UserID:       userID,
			URL:          resp.SecureURL,
			ThumbnailURL: thumbnailURL,
			FileName:     fileName,
			FileType:     "image",
			FileSize:     fileSize,
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		}

		if err := s.fileRepo.CreateFileWithTx(tx, attachment); err != nil {
			log.Printf("Failed to create file record for %s in database: %v", fileName, err)
			return "", nil, fmt.Errorf("failed to create file: %v", err)
		}
		log.Printf("Created file record for %s, ID: %d", fileName, attachment.ID)

		// Thay thế src trong thẻ <img>
		newImgTag := strings.Replace(imgTag, src, resp.SecureURL, 1)
		updatedContent = strings.Replace(updatedContent, imgTag, newImgTag, 1)
		attachmentIDs = append(attachmentIDs, uint64(attachment.ID))
	}

	s.invalidateCache("files:*")
	return updatedContent, attachmentIDs, nil
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

func (s *fileService) CleanupOrphanedFiles() error {
	var files []models.Attachment
	err := s.db.Find(&files).Error
	if err != nil {
		return fmt.Errorf("failed to find files: %v", err)
	}

	ctx := context.Background()
	for _, file := range files {
		var count int64
		err := s.db.Model(&models.Comment{}).
			Where("JSON_CONTAINS(attachment_ids, ?)", fmt.Sprintf(`"%d"`, file.ID)).
			Count(&count).Error
		if err != nil {
			log.Printf("Failed to check comment attachment for file %d: %v", file.ID, err)
			continue
		}
		if count == 0 {
			_, err := s.cloudinary.Upload.Destroy(ctx, uploader.DestroyParams{
				PublicID:     strings.TrimPrefix(file.URL, "https://res.cloudinary.com/your_cloud_name/"),
				ResourceType: getCloudinaryResourceType(file.FileType),
			})
			if err != nil {
				log.Printf("Failed to delete file %s from Cloudinary: %v", file.URL, err)
				continue
			}

			if err := s.fileRepo.DeleteFile(file.ID); err != nil {
				log.Printf("Failed to delete file %d from database: %v", file.ID, err)
				continue
			}

			s.invalidateCache(fmt.Sprintf("file:%d", file.ID))
			log.Printf("Deleted orphaned file %d", file.ID)
		}
	}

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

func getCloudinaryResourceType(fileType string) string {
	switch fileType {
	case "image":
		return "image"
	case "video":
		return "video"
	case "audio":
		return "raw"
	default:
		return "raw"
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
