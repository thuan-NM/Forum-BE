package repositories

import (
	"Forum_BE/models"
	"gorm.io/gorm"
)

type FollowRepository interface {
	CreateFollow(follow *models.Follow) error
	GetFollow(followerID, followingID uint) (*models.Follow, error)
	DeleteFollow(followerID, followingID uint) error
	ListFollowers(userID uint) ([]models.Follow, error)
	ListFollowing(userID uint) ([]models.Follow, error)
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

func (r *followRepository) GetFollow(followerID, followingID uint) (*models.Follow, error) {
	var follow models.Follow
	err := r.db.Where("follower_id = ? AND following_id = ?", followerID, followingID).First(&follow).Error
	if err != nil {
		return nil, err
	}
	return &follow, nil
}

func (r *followRepository) DeleteFollow(followerID, followingID uint) error {
	return r.db.Where("follower_id = ? AND following_id = ?", followerID, followingID).Delete(&models.Follow{}).Error
}

func (r *followRepository) ListFollowers(userID uint) ([]models.Follow, error) {
	var followers []models.Follow
	err := r.db.Preload("Follower").Where("following_id = ?", userID).Find(&followers).Error
	if err != nil {
		return nil, err
	}
	return followers, nil
}

func (r *followRepository) ListFollowing(userID uint) ([]models.Follow, error) {
	var followings []models.Follow
	err := r.db.Preload("Following").Where("follower_id = ?", userID).Find(&followings).Error
	if err != nil {
		return nil, err
	}
	return followings, nil
}
