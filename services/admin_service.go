package services

import (
	"Forum_BE/models"
	"Forum_BE/repositories"
	"errors"
)

type AdminService interface {
	ApproveQuestion(questionID uint) error
	RejectQuestion(questionID uint) error
}

type adminService struct {
	questionRepo repositories.QuestionRepository
}

func NewAdminService(qRepo repositories.QuestionRepository) AdminService {
	return &adminService{questionRepo: qRepo}
}

func (s *adminService) ApproveQuestion(questionID uint) error {
	question, err := s.questionRepo.GetQuestionByID(questionID)
	if err != nil {
		return errors.New("question not found")
	}

	if question.Status != models.StatusPending {
		return errors.New("question is not pending approval")
	}

	question.Status = models.StatusApproved
	return s.questionRepo.UpdateQuestion(question)
}

func (s *adminService) RejectQuestion(questionID uint) error {
	question, err := s.questionRepo.GetQuestionByID(questionID)
	if err != nil {
		return errors.New("question not found")
	}

	if question.Status != models.StatusPending {
		return errors.New("question is not pending approval")
	}

	question.Status = models.StatusRejected
	return s.questionRepo.UpdateQuestion(question)
}
