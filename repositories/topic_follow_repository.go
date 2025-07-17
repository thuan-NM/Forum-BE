package repositories

import (
	"Forum_BE/models"
	"gorm.io/gorm"
)

type TopicFollowRepository interface {
	CreateFollow(follow *models.TopicFollow) error
	DeleteFollow(userID, topicID uint) error
	GetFollowsByTopic(topicID uint) ([]models.TopicFollow, error)
	ExistsByTopicAndUser(topicID, userID uint) (bool, error) // Thêm method
}

type topicFollowRepository struct {
	db *gorm.DB
}

func NewTopicFollowRepository(db *gorm.DB) TopicFollowRepository {
	return &topicFollowRepository{db: db}
}

func (r *topicFollowRepository) CreateFollow(follow *models.TopicFollow) error {
	return r.db.Create(follow).Error
}

func (r *topicFollowRepository) DeleteFollow(userID, topicID uint) error {
	return r.db.Unscoped().Where("user_id = ? AND topic_id = ?", userID, topicID).Delete(&models.TopicFollow{}).Error
}

func (r *topicFollowRepository) GetFollowsByTopic(topicID uint) ([]models.TopicFollow, error) {
	var follows []models.TopicFollow
	err := r.db.Where("topic_id = ?", topicID).Find(&follows).Error
	return follows, err
}

func (r *topicFollowRepository) ExistsByTopicAndUser(topicID, userID uint) (bool, error) {
	var count int64
	err := r.db.Model(&models.TopicFollow{}).Where("topic_id = ? AND user_id = ?", topicID, userID).Count(&count).Error
	return count > 0, err
}
