package repositories

import (
	"Forum_BE/models"
	"gorm.io/gorm"
)

type TagRepository interface {
	CreateTag(tag *models.Tag) error
	GetTagByID(id uint) (*models.Tag, error)
	GetTagByName(name string) (*models.Tag, error)
	UpdateTag(tag *models.Tag) error
	DeleteTag(id uint) error
	ListTags() ([]models.Tag, error)
}

type tagRepository struct {
	db *gorm.DB
}

func NewTagRepository(db *gorm.DB) TagRepository {
	return &tagRepository{db: db}
}

func (r *tagRepository) CreateTag(tag *models.Tag) error {
	return r.db.Create(tag).Error
}

func (r *tagRepository) GetTagByID(id uint) (*models.Tag, error) {
	var tag models.Tag
	err := r.db.First(&tag, id).Error
	if err != nil {
		return nil, err
	}
	return &tag, nil
}

func (r *tagRepository) GetTagByName(name string) (*models.Tag, error) {
	var tag models.Tag
	err := r.db.Where("name = ?", name).First(&tag).Error
	if err != nil {
		return nil, err
	}
	return &tag, nil
}

func (r *tagRepository) UpdateTag(tag *models.Tag) error {
	return r.db.Save(tag).Error
}

func (r *tagRepository) DeleteTag(id uint) error {
	return r.db.Delete(&models.Tag{}, id).Error
}

func (r *tagRepository) ListTags() ([]models.Tag, error) {
	var tags []models.Tag
	err := r.db.Find(&tags).Error
	if err != nil {
		return nil, err
	}
	return tags, nil
}
