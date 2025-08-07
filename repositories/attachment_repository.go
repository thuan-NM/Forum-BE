package repositories

import (
	"Forum_BE/models"
	"gorm.io/gorm"
	"log"
	"strings"
)

type AttachmentRepository interface {
	CreateAttachment(attachment *models.Attachment) error
	GetAttachmentByID(id uint) (*models.Attachment, error)
	UpdateAttachment(attachment *models.Attachment) error
	DeleteAttachment(id uint) error
	ListAttachments(filters map[string]interface{}) ([]models.Attachment, int, error)
}

type attachmentRepository struct {
	db *gorm.DB
}

func NewAttachmentRepository(db *gorm.DB) AttachmentRepository {
	return &attachmentRepository{db: db}
}

func (r *attachmentRepository) CreateAttachment(attachment *models.Attachment) error {
	return r.db.Create(attachment).Error
}

func (r *attachmentRepository) GetAttachmentByID(id uint) (*models.Attachment, error) {
	var attachment models.Attachment
	err := r.db.Preload("User").First(&attachment, id).Error
	if err != nil {
		log.Printf("Failed to get attachment %d: %v", id, err)
		return nil, err
	}
	return &attachment, nil
}

func (r *attachmentRepository) UpdateAttachment(attachment *models.Attachment) error {
	return r.db.Save(attachment).Error
}

func (r *attachmentRepository) DeleteAttachment(id uint) error {
	return r.db.Delete(&models.Attachment{}, id).Error
}

func (r *attachmentRepository) ListAttachments(filters map[string]interface{}) ([]models.Attachment, int, error) {
	var attachments []models.Attachment
	query := r.db.Model(&models.Attachment{})

	if userID, ok := filters["user_id"].(uint); ok {
		query = query.Where("user_id = ?", userID)
	}
	if fileType, ok := filters["file_type"].(string); ok {
		query = query.Where("file_type = ?", fileType)
	}
	search, ok := filters["search"].(string)

	var total int64
	if err := query.Count(&total).Error; err != nil {
		log.Printf("Error counting attachments: %v", err)
		return nil, 0, err
	}

	limit, okLimit := filters["limit"].(int)
	page, okPage := filters["page"].(int)
	if okLimit && limit > 0 {
		query = query.Limit(limit)
	}
	if okPage && page > 0 {
		offset := (page - 1) * limit
		query = query.Offset(offset)
	}
	if ok && search != "" {
		search = strings.ToLower(search)
		query = query.Where("file_name LIKE ?", "%"+search+"%")
	}
	err := query.Preload("User").Find(&attachments).Error
	if err != nil {
		log.Printf("Error fetching attachments: %v", err)
		return nil, 0, err
	}

	return attachments, int(total), nil
}
