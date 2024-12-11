package services

import (
	"Forum_BE/models"
	"Forum_BE/repositories"
	"errors"
)

type FollowService interface {
	FollowUser(followerID, followingID uint) (*models.Follow, error)
	UnfollowUser(followerID, followingID uint) error
	GetFollowers(userID uint) ([]models.Follow, error)
	GetFollowing(userID uint) ([]models.Follow, error)
}

type followService struct {
	followRepo repositories.FollowRepository
}

func NewFollowService(fRepo repositories.FollowRepository) FollowService {
	return &followService{followRepo: fRepo}
}

func (s *followService) FollowUser(followerID, followingID uint) (*models.Follow, error) {
	if followerID == followingID {
		return nil, errors.New("cannot follow yourself")
	}

	// Kiểm tra xem đã theo dõi chưa
	existingFollow, err := s.followRepo.GetFollow(followerID, followingID)
	if err == nil && existingFollow != nil {
		return nil, errors.New("already following this user")
	}

	follow := &models.Follow{
		FollowerID:  followerID,
		FollowingID: followingID,
	}

	if err := s.followRepo.CreateFollow(follow); err != nil {
		return nil, err
	}

	return follow, nil
}

func (s *followService) UnfollowUser(followerID, followingID uint) error {
	// Kiểm tra xem đã theo dõi chưa
	existingFollow, err := s.followRepo.GetFollow(followerID, followingID)
	if err != nil || existingFollow == nil {
		return errors.New("follow relationship does not exist")
	}

	return s.followRepo.DeleteFollow(followerID, followingID)
}

func (s *followService) GetFollowers(userID uint) ([]models.Follow, error) {
	return s.followRepo.ListFollowers(userID)
}

func (s *followService) GetFollowing(userID uint) ([]models.Follow, error) {
	return s.followRepo.ListFollowing(userID)
}
