package services

import (
	"Forum_BE/models"
	"Forum_BE/repositories"
	"errors"
)

type QuestionService interface {
	CreateQuestion(title string, userID uint) (*models.Question, error)
	GetQuestionByID(id uint) (*models.Question, error)
	UpdateQuestion(id uint, title string) (*models.Question, error)
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

func (s *questionService) CreateQuestion(title string, userID uint) (*models.Question, error) {
	if title == "" {
		return nil, errors.New("title is required")
	}

	question := &models.Question{
		Title:  title,
		UserID: userID,
		Status: models.StatusPending,
	}

	if err := s.questionRepo.CreateQuestion(question); err != nil {
		return nil, err
	}

	return question, nil
}

func (s *questionService) GetQuestionByID(id uint) (*models.Question, error) {
	return s.questionRepo.GetQuestionByID(id)
}

func (s *questionService) UpdateQuestion(id uint, title string) (*models.Question, error) {
	question, err := s.questionRepo.GetQuestionByID(id)
	if err != nil {
		return nil, err
	}

	if title != "" {
		question.Title = title
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
	questions, err := s.questionRepo.ListQuestions(filters)
	if err != nil {
		return nil, err
	}

	// Lọc theo tag nếu có
	if tagID, ok := filters["tag_id"]; ok {
		var filtered []models.Question
		for _, q := range questions {
			for _, t := range q.Tags {
				if t.ID == tagID.(uint) {
					filtered = append(filtered, q)
					break
				}
			}
		}
		questions = filtered
	}

	if userIDRaw, ok := filters["user_id"]; ok {
		userID := userIDRaw.(uint)

		// Lấy các ID câu hỏi bị user này ẩn
		passedIDs, err := s.questionRepo.GetPassedQuestionIDs(userID)
		if err != nil {
			return nil, err
		}

		// Tạo map để tra nhanh
		passedMap := make(map[uint]bool)
		for _, id := range passedIDs {
			passedMap[id] = true
		}

		var visibleQuestions []models.Question
		for _, q := range questions {
			if !passedMap[q.ID] {
				visibleQuestions = append(visibleQuestions, q)
			}
		}
		questions = visibleQuestions
	}

	return questions, nil
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
