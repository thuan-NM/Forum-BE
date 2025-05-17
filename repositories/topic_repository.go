package repositories

import (
	"Forum_BE/models"
	"gorm.io/gorm"
)

type TopicRepository interface {
	CreateTopic(topic *models.Topic) error
	GetTopicByID(id uint) (*models.Topic, error)
	GetTopicByName(name string) (*models.Topic, error)
	UpdateTopic(topic *models.Topic) error
	DeleteTopic(id uint) error
	ListTopics(filters map[string]interface{}) ([]models.Topic, error)
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

func (r *topicRepository) ListTopics(filters map[string]interface{}) ([]models.Topic, error) {
	var topics []models.Topic
	query := r.db.Preload("Questions")

	if status, ok := filters["status"]; ok {
		query = query.Where("status = ?", status)
	}

	if search, ok := filters["search"]; ok {
		query = query.Where("name LIKE ?", "%"+search.(string)+"%")
	}

	err := query.Find(&topics).Error
	if err != nil {
		return nil, err
	}
	return topics, nil
}

func (r *topicRepository) AddQuestionToTopic(questionID, topicID uint) error {
	return r.db.Exec("INSERT INTO question_topics (question_id, topic_id) VALUES (?, ?)", questionID, topicID).Error
}

func (r *topicRepository) RemoveQuestionFromTopic(questionID, topicID uint) error {
	return r.db.Exec("DELETE FROM question_topics WHERE question_id = ? AND topic_id = ?", questionID, topicID).Error
}
