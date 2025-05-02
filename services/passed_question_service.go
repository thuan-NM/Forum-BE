package services

import "Forum_BE/repositories"

type PassService interface {
	PassQuestion(userID, questionID uint) error
	IsQuestionPassed(userID, questionID uint) (bool, error)
	GetPassedIDs(userID uint) ([]uint, error)
}

type passService struct {
	repo repositories.PassRepository
}

func NewPassService(r repositories.PassRepository) PassService {
	return &passService{repo: r}
}

func (s *passService) PassQuestion(userID, questionID uint) error {
	return s.repo.PassQuestion(userID, questionID)
}

func (s *passService) IsQuestionPassed(userID, questionID uint) (bool, error) {
	return s.repo.IsPassed(userID, questionID)
}

func (s *passService) GetPassedIDs(userID uint) ([]uint, error) {
	return s.repo.GetPassedQuestionIDs(userID)
}
