package services

import (
	"Forum_BE/models"
	"Forum_BE/repositories"
	"errors"
)

type AnswerService interface {
	CreateAnswer(content string, userID uint, questionID uint) (*models.Answer, error)
	GetAnswerByID(id uint) (*models.Answer, error)
	UpdateAnswer(id uint, content string) (*models.Answer, error)
	DeleteAnswer(id uint) error
	ListAnswers(filters map[string]interface{}) ([]models.Answer, error)
}

type answerService struct {
	answerRepo   repositories.AnswerRepository
	questionRepo repositories.QuestionRepository
}

func NewAnswerService(aRepo repositories.AnswerRepository, qRepo repositories.QuestionRepository) AnswerService {
	return &answerService{answerRepo: aRepo, questionRepo: qRepo}
}

func (s *answerService) CreateAnswer(content string, userID uint, questionID uint) (*models.Answer, error) {
	if content == "" {
		return nil, errors.New("content is required")
	}

	// Kiểm tra xem câu hỏi có tồn tại và đã được duyệt không
	question, err := s.questionRepo.GetQuestionByID(questionID)
	if err != nil {
		return nil, errors.New("question not found")
	}

	if question.Status != models.StatusApproved {
		return nil, errors.New("cannot answer a question that is not approved")
	}

	answer := &models.Answer{
		Content:    content,
		UserID:     userID,
		QuestionID: questionID,
	}

	if err := s.answerRepo.CreateAnswer(answer); err != nil {
		return nil, err
	}

	return answer, nil
}

func (s *answerService) GetAnswerByID(id uint) (*models.Answer, error) {
	return s.answerRepo.GetAnswerByID(id)
}

func (s *answerService) UpdateAnswer(id uint, content string) (*models.Answer, error) {
	answer, err := s.answerRepo.GetAnswerByID(id)
	if err != nil {
		return nil, err
	}

	if content != "" {
		answer.Content = content
	}

	if err := s.answerRepo.UpdateAnswer(answer); err != nil {
		return nil, err
	}

	return answer, nil
}

func (s *answerService) DeleteAnswer(id uint) error {
	return s.answerRepo.DeleteAnswer(id)
}

func (s *answerService) ListAnswers(filters map[string]interface{}) ([]models.Answer, error) {
	return s.answerRepo.ListAnswers(filters)
}
