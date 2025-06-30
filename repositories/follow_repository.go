package repositories

import (
	"Forum_BE/models"
	"gorm.io/gorm"
)

type FollowRepository interface {
	CreateFollow(follow *models.Follow) error
	DeleteFollow(userID, followableID uint, followableType string) error
	GetFollowsByFollowable(followableID uint, followableType string) ([]models.Follow, error)
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

func (r *followRepository) DeleteFollow(userID, followableID uint, followableType string) error {
	return r.db.Unscoped().Where("user_id = ? AND followable_id = ? AND followable_type = ?", userID, followableID, followableType).Delete(&models.Follow{}).Error
}

func (r *followRepository) GetFollowsByFollowable(followableID uint, followableType string) ([]models.Follow, error) {
	var follows []models.Follow
	query := r.db.Preload("User")
	if followableType == "question" {
		query = query.Preload("Question")
	} else if followableType == "user" {
		query = query.Preload("FollowedUser")
	} else if followableType == "topic" {
		query = query.Preload("Topic")
	}
	err := query.Where("followable_id = ? AND followable_type = ?", followableID, followableType).Find(&follows).Error
	if err != nil {
		return nil, err
	}
	return follows, nil
}
