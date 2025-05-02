package services

import (
	"Forum_BE/models"
	"Forum_BE/repositories"
	"errors"
)

type FollowService interface {
	FollowQuestion(userID, questionID uint) error
	UnfollowQuestion(userID, questionID uint) error
	GetFollowsByQuestionID(questionID uint) ([]models.Follow, error)
}

type followService struct {
	followRepo repositories.FollowRepository
}

func NewFollowService(fRepo repositories.FollowRepository) FollowService {
	return &followService{followRepo: fRepo}
}

func (s *followService) FollowQuestion(userID, questionID uint) error {
	// Check if already followed
	follows, err := s.followRepo.GetFollowsByQuestionID(questionID)
	if err != nil {
		return err
	}
	for _, f := range follows {
		if f.UserID == userID {
			return errors.New("user already follows this question")
		}
	}

	follow := &models.Follow{
		UserID:     userID,
		QuestionID: questionID,
	}
	return s.followRepo.CreateFollow(follow)
}

func (s *followService) UnfollowQuestion(userID, questionID uint) error {
	return s.followRepo.DeleteFollow(userID, questionID)
}
func (s *followService) GetFollowsByQuestionID(questionID uint) ([]models.Follow, error) {
	var follows []models.Follow
	follows, err := s.followRepo.GetFollowsByQuestionID(questionID)
	if err != nil {
		return nil, err
	}
	return follows, nil
}
