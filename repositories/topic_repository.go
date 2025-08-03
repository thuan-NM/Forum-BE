package repositories

import (
	"Forum_BE/models"
	"gorm.io/gorm"
	"log"
)

type TopicRepository interface {
	CreateTopic(topic *models.Topic) error
	GetTopicByID(id uint) (*models.Topic, error)
	GetTopicByName(name string) (*models.Topic, error)
	UpdateTopic(topic *models.Topic) error
	DeleteTopic(id uint) error
	ListTopics(filters map[string]interface{}) ([]models.Topic, int, error)
	AddQuestionToTopic(questionID, topicID uint) error
	RemoveQuestionFromTopic(questionID, topicID uint) error
}

type topicRepository struct {
	db *gorm.DB
}

func NewTopicRepository(db *gorm.DB) TopicRepository {
	return &topicRepository{db: db}
}

func (r *topicRepository) CreateTopic(topic *models.Topic) error {
	return r.db.Create(topic).Error
}

func (r *topicRepository) GetTopicByID(id uint) (*models.Topic, error) {
	var topic models.Topic
	err := r.db.Preload("Questions").First(&topic, id).Error
	if err != nil {
		return nil, err
	}
	return &topic, nil
}

func (r *topicRepository) GetTopicByName(name string) (*models.Topic, error) {
	var topic models.Topic
	err := r.db.Where("name = ?", name).First(&topic).Error
	if err != nil {
		return nil, err
	}
	return &topic, nil
}

func (r *topicRepository) UpdateTopic(topic *models.Topic) error {
	return r.db.Save(topic).Error
}

func (r *topicRepository) DeleteTopic(id uint) error {
	return r.db.Delete(&models.Topic{}, id).Error
}

func (r *topicRepository) ListTopics(filters map[string]interface{}) ([]models.Topic, int, error) {
	var topics []models.Topic

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

	// Process sort parameter
	sortOrder := "created_at DESC"
	if sort, ok := filters["sort"].(string); ok {
		if sort == "asc" {
			sortOrder = "created_at ASC"
		} else if sort == "desc" {
			sortOrder = "created_at DESC"
		}
		log.Printf("Applying sort order: %s", sortOrder)
	}

	// Query for counting total
	countQuery := r.db.Model(&models.Topic{})
	if search, ok := filters["search"]; ok {
		countQuery = countQuery.Where("name LIKE ? OR description LIKE ?", "%"+search.(string)+"%", "%"+search.(string)+"%")
	}
	if status, ok := filters["status"]; ok {
		countQuery = countQuery.Where("status = ?", status)
	}
	var total int64
	if err := countQuery.Count(&total).Error; err != nil {
		log.Printf("Error counting topics: %v", err)
		return nil, 0, err
	}

	// Apply filters and pagination
	query := r.db.Preload("Questions")
	if search, ok := filters["search"]; ok {
		query = query.Where("name LIKE ? OR description LIKE ?", "%"+search.(string)+"%", "%"+search.(string)+"%")
	}
	if status, ok := filters["status"]; ok {
		query = query.Where("status = ?", status)
	}

	query = query.Offset(offset).Limit(limit).Order(sortOrder)
	err := query.Find(&topics).Error
	if err != nil {
		log.Printf("Error fetching topics: %v", err)
		return nil, 0, err
	}

	log.Printf("Found %d topics with total %d, sort order: %s", len(topics), total, sortOrder)
	return topics, int(total), nil
}

func (r *topicRepository) AddQuestionToTopic(questionID, topicID uint) error {
	return r.db.Exec("INSERT INTO question_topics (question_id, topic_id) VALUES (?, ?)", questionID, topicID).Error
}

func (r *topicRepository) RemoveQuestionFromTopic(questionID, topicID uint) error {
	return r.db.Exec("DELETE FROM question_topics WHERE question_id = ? AND topic_id = ?", questionID, topicID).Error
}
