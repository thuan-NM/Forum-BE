package repositories

import (
	"Forum_BE/models"
	"gorm.io/gorm"
	"log"
)

type TagRepository interface {
	CreateTag(tag *models.Tag) error
	GetTagByID(id uint) (*models.Tag, error)
	GetTagByName(name string) (*models.Tag, error)
	UpdateTag(tag *models.Tag) error
	DeleteTag(id uint) error
	ListTags(filters map[string]interface{}) ([]models.Tag, int, error)
	GetTagsByPostID(postID uint) ([]models.Tag, error)
	GetTagsByAnswerID(answerID uint) ([]models.Tag, error)
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
	err := r.db.Preload("Posts").Preload("Answers").First(&tag, id).Error
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

func (r *tagRepository) ListTags(filters map[string]interface{}) ([]models.Tag, int, error) {
	var tags []models.Tag

	// Process pagination parameters
	page, okPage := filters["page"].(int)
	limit, okLimit := filters["limit"].(int)
	if !okPage || page < 1 {
		page = 1
	}
	if !okLimit || limit < 1 {
		limit = 10
	}
	offset := (page - 1) * limit

	// Query for counting total
	countQuery := r.db.Model(&models.Tag{})
	if search, ok := filters["search"]; ok {
		countQuery = countQuery.Where("name LIKE ? OR description LIKE ?", "%"+search.(string)+"%", "%"+search.(string)+"%")
	}
	var total int64
	if err := countQuery.Count(&total).Error; err != nil {
		log.Printf("Error counting tags: %v", err)
		return nil, 0, err
	}

	// Apply filters and pagination
	query := r.db.Preload("Posts").Preload("Answers")
	if search, ok := filters["search"]; ok {
		query = query.Where("name LIKE ? OR description LIKE ?", "%"+search.(string)+"%", "%"+search.(string)+"%")
	}

	query = query.Offset(offset).Limit(limit)
	err := query.Find(&tags).Error
	if err != nil {
		log.Printf("Error fetching tags: %v", err)
		return nil, 0, err
	}

	log.Printf("Found %d tags with total %d", len(tags), total)
	return tags, int(total), nil
}

func (r *tagRepository) GetTagsByPostID(postID uint) ([]models.Tag, error) {
	var tags []models.Tag
	err := r.db.Joins("JOIN post_tags ON post_tags.tag_id = tags.id").
		Where("post_tags.post_id = ?", postID).
		Preload("Posts").Preload("Answers").Find(&tags).Error
	if err != nil {
		return nil, err
	}
	return tags, nil
}

func (r *tagRepository) GetTagsByAnswerID(answerID uint) ([]models.Tag, error) {
	var tags []models.Tag
	err := r.db.Joins("JOIN answer_tags ON answer_tags.tag_id = tags.id").
		Where("answer_tags.answer_id = ?", answerID).
		Preload("Posts").Preload("Answers").Find(&tags).Error
	if err != nil {
		return nil, err
	}
	return tags, nil
}
