package repositories

import (
	"Forum_BE/models"
	"fmt"
	"gorm.io/gorm"
	"strings"
)

type FileRepository interface {
	CreateFile(file *models.Attachment) error
	GetFileByID(id uint) (*models.Attachment, error)
	DeleteFile(id uint) error
	ListFiles(filters map[string]interface{}) ([]models.Attachment, int64, error)
	UpdateFile(file *models.Attachment) error
}

type fileRepository struct {
	db *gorm.DB
}

func NewFileRepository(db *gorm.DB) FileRepository {
	return &fileRepository{db: db}
}

func (r *fileRepository) CreateFile(file *models.Attachment) error {
	return r.db.Create(file).Error
}

func (r *fileRepository) GetFileByID(id uint) (*models.Attachment, error) {
	var file models.Attachment
	err := r.db.Preload("User").Preload("Post").Preload("Answer").Preload("Comment").
		Where("deleted_at IS NULL").
		First(&file, id).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get file: %v", err)
	}
	return &file, nil
}

func (r *fileRepository) DeleteFile(id uint) error {
	return r.db.Model(&models.Attachment{}).
		Where("id = ?", id).
		Update("deleted_at", gorm.Expr("NOW()")).Error
}

func (r *fileRepository) UpdateFile(file *models.Attachment) error {
	return r.db.Model(&models.Attachment{}).
		Where("id = ?", file.ID).
		Updates(map[string]interface{}{
			"entity_type": file.EntityType,
			"entity_id":   file.EntityID,
			"updated_at":  gorm.Expr("NOW()"),
		}).Error
}

func (r *fileRepository) ListFiles(filters map[string]interface{}) ([]models.Attachment, int64, error) {
	var files []models.Attachment
	query := r.db.Preload("User").Preload("Post").Preload("Answer").Preload("Comment").
		Where("deleted_at IS NULL")

	page, okPage := filters["page"].(int)
	limit, okLimit := filters["limit"].(int)
	if !okPage || page < 1 {
		page = 1
	}
	if !okLimit || limit < 1 {
		limit = 12 // Phù hợp với rowsPerPage trong frontend
	}

	if search, ok := filters["search"].(string); ok && search != "" {
		query = query.Where("file_name LIKE ?", "%"+strings.ToLower(search)+"%")
	}
	if fileType, ok := filters["file_type"].(string); ok && fileType != "all" {
		query = query.Where("file_type = ?", fileType)
	}
	if entityType, ok := filters["entity_type"].(string); ok && entityType != "" {
		query = query.Where("entity_type = ?", entityType)
	}
	if entityID, ok := filters["entity_id"].(uint); ok && entityID != 0 {
		query = query.Where("entity_id = ?", entityID)
	}

	var total int64
	if err := query.Model(&models.Attachment{}).Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count files: %v", err)
	}

	offset := (page - 1) * limit
	err := query.Limit(limit).Offset(offset).Order("created_at DESC").Find(&files).Error
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list files: %v", err)
	}

	return files, total, nil
}
