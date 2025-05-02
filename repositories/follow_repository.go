package repositories

import (
	"Forum_BE/models"
	"gorm.io/gorm"
)

type FollowRepository interface {
	CreateFollow(follow *models.Follow) error
	DeleteFollow(userID, questionID uint) error
	GetFollowsByQuestionID(questionID uint) ([]models.Follow, error)
}

type followRepository struct {
	db *gorm.DB
}

func NewFollowRepository(db *gorm.DB) FollowRepository {
	return &followRepository{db: db}
}

func (r *followRepository) CreateFollow(follow *models.Follow) error {
	return r.db.Create(follow).Error
}

func (r *followRepository) DeleteFollow(userID, questionID uint) error {
	return r.db.Unscoped().Where("user_id = ? AND question_id = ?", userID, questionID).Delete(&models.Follow{}).Error
}

func (r *followRepository) GetFollowsByQuestionID(questionID uint) ([]models.Follow, error) {
	var follows []models.Follow
	err := r.db.Preload("User").Where("question_id = ?", questionID).Find(&follows).Error
	if err != nil {
		return nil, err
	}
	return follows, nil
}
