package repositories

import (
	"Forum_BE/models"
	"gorm.io/gorm"
)

type FileRepository interface {
	CreateFile(attachment *models.Attachment) error
	CreateFileWithTx(tx *gorm.DB, attachment *models.Attachment) error
	GetFileByID(id uint) (*models.Attachment, error)
	UpdateFile(attachment *models.Attachment) error
	DeleteFile(id uint) error
	ListFiles(filters map[string]interface{}) ([]models.Attachment, int64, error)
}

type fileRepository struct {
	db *gorm.DB
}

func NewFileRepository(db *gorm.DB) FileRepository {
	return &fileRepository{db: db}
}

func (r *fileRepository) CreateFile(attachment *models.Attachment) error {
	return r.db.Create(attachment).Error
}

func (r *fileRepository) CreateFileWithTx(tx *gorm.DB, attachment *models.Attachment) error {
	return tx.Create(attachment).Error
}

func (r *fileRepository) GetFileByID(id uint) (*models.Attachment, error) {
	var attachment models.Attachment
	err := r.db.First(&attachment, id).Error
	if err != nil {
		return nil, err
	}
	return &attachment, nil
}

func (r *fileRepository) UpdateFile(attachment *models.Attachment) error {
	return r.db.Save(attachment).Error
}

func (r *fileRepository) DeleteFile(id uint) error {
	return r.db.Delete(&models.Attachment{}, id).Error
}

func (r *fileRepository) ListFiles(filters map[string]interface{}) ([]models.Attachment, int64, error) {
	var files []models.Attachment
	query := r.db.Model(&models.Attachment{})

	if search, ok := filters["search"].(string); ok {
		query = query.Where("file_name LIKE ?", "%"+search+"%")
	}
	if fileType, ok := filters["file_type"].(string); ok {
		query = query.Where("file_type = ?", fileType)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if page, ok := filters["page"].(int); ok && page > 0 {
		query = query.Offset((page - 1) * 10)
	}
	if limit, ok := filters["limit"].(int); ok && limit > 0 {
		query = query.Limit(limit)
	}

	err := query.Find(&files).Error
	if err != nil {
		return nil, 0, err
	}

	return files, total, nil
}
