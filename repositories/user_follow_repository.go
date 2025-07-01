package repositories

import (
	"Forum_BE/models"
	"gorm.io/gorm"
)

type UserFollowRepository interface {
	CreateFollow(follow *models.UserFollow) error
	DeleteFollow(userID, followedUserID uint) error
	GetFollowsByUser(followedUserID uint) ([]models.UserFollow, error)
}

type userFollowRepository struct {
	db *gorm.DB
}

func NewUserFollowRepository(db *gorm.DB) UserFollowRepository {
	return &userFollowRepository{db: db}
}

func (r *userFollowRepository) CreateFollow(follow *models.UserFollow) error {
	return r.db.Create(follow).Error
}

func (r *userFollowRepository) DeleteFollow(userID, followedUserID uint) error {
	return r.db.Where("user_id = ? AND followed_user_id = ?", userID, followedUserID).Delete(&models.UserFollow{}).Error
}

func (r *userFollowRepository) GetFollowsByUser(followedUserID uint) ([]models.UserFollow, error) {
	var follows []models.UserFollow
	err := r.db.Where("followed_user_id = ?", followedUserID).Find(&follows).Error
	return follows, err
}
