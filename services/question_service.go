package services

import (
	"Forum_BE/models"
	"Forum_BE/repositories"
	"errors"
)

type QuestionService interface {
	CreateQuestion(title, content string, userID uint, groupID uint) (*models.Question, error)
	GetQuestionByID(id uint) (*models.Question, error)
	UpdateQuestion(id uint, title, content string, groupID uint) (*models.Question, error)
	DeleteQuestion(id uint) error
	ListQuestions(filters map[string]interface{}) ([]models.Question, error)
	ApproveQuestion(id uint) (*models.Question, error)
	RejectQuestion(id uint) (*models.Question, error)
}

type questionService struct {
	questionRepo repositories.QuestionRepository
}

func NewQuestionService(qRepo repositories.QuestionRepository) QuestionService {
	return &questionService{questionRepo: qRepo}
}

func (s *questionService) CreateQuestion(title, content string, userID uint, groupID uint) (*models.Question, error) {
	if title == "" || content == "" {
		return nil, errors.New("title and content are required")
	}

	question := &models.Question{
		Title:   title,
		Content: content,
		UserID:  userID,
		GroupID: groupID,
		Status:  models.StatusPending,
	}

	if err := s.questionRepo.CreateQuestion(question); err != nil {
		return nil, err
	}

	return question, nil
}

func (s *questionService) GetQuestionByID(id uint) (*models.Question, error) {
	return s.questionRepo.GetQuestionByID(id)
}

func (s *questionService) UpdateQuestion(id uint, title, content string, groupID uint) (*models.Question, error) {
	question, err := s.questionRepo.GetQuestionByID(id)
	if err != nil {
		return nil, err
	}

	if title != "" {
		question.Title = title
	}

	if content != "" {
		question.Content = content
	}

	if groupID != 0 {
		question.GroupID = groupID
	}

	if err := s.questionRepo.UpdateQuestion(question); err != nil {
		return nil, err
	}

	return question, nil
}

func (s *questionService) DeleteQuestion(id uint) error {
	return s.questionRepo.DeleteQuestion(id)
}

func (s *questionService) ListQuestions(filters map[string]interface{}) ([]models.Question, error) {
	return s.questionRepo.ListQuestions(filters)
}

func (s *questionService) ApproveQuestion(id uint) (*models.Question, error) {
	question, err := s.questionRepo.GetQuestionByID(id)
	if err != nil {
		return nil, err
	}

	if question.Status != models.StatusPending {
		return nil, errors.New("question is not pending")
	}

	question.Status = models.StatusApproved

	if err := s.questionRepo.UpdateQuestion(question); err != nil {
		return nil, err
	}

	return question, nil
}

func (s *questionService) RejectQuestion(id uint) (*models.Question, error) {
	question, err := s.questionRepo.GetQuestionByID(id)
	if err != nil {
		return nil, err
	}

	if question.Status != models.StatusPending {
		return nil, errors.New("question is not pending")
	}

	question.Status = models.StatusRejected

	if err := s.questionRepo.UpdateQuestion(question); err != nil {
		return nil, err
	}

	return question, nil
}
