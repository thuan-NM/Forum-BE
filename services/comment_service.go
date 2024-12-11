package services

import (
	"Forum_BE/models"
	"Forum_BE/repositories"
	"errors"
)

type CommentService interface {
	CreateComment(content string, userID uint, questionID *uint, answerID *uint) (*models.Comment, error)
	GetCommentByID(id uint) (*models.Comment, error)
	UpdateComment(id uint, content string) (*models.Comment, error)
	DeleteComment(id uint) error
	ListComments(filters map[string]interface{}) ([]models.Comment, error)
}

type commentService struct {
	commentRepo  repositories.CommentRepository
	questionRepo repositories.QuestionRepository
	answerRepo   repositories.AnswerRepository
}

func NewCommentService(cRepo repositories.CommentRepository, qRepo repositories.QuestionRepository, aRepo repositories.AnswerRepository) CommentService {
	return &commentService{
		commentRepo:  cRepo,
		questionRepo: qRepo,
		answerRepo:   aRepo,
	}
}

func (s *commentService) CreateComment(content string, userID uint, questionID *uint, answerID *uint) (*models.Comment, error) {
	if content == "" {
		return nil, errors.New("content is required")
	}

	if questionID == nil && answerID == nil {
		return nil, errors.New("either question_id or answer_id must be provided")
	}

	// Kiểm tra xem liên kết đến câu hỏi hoặc câu trả lời có tồn tại không
	if questionID != nil {
		question, err := s.questionRepo.GetQuestionByID(*questionID)
		if err != nil {
			return nil, errors.New("question not found")
		}
		if question.Status != models.StatusApproved {
			return nil, errors.New("cannot comment on a question that is not approved")
		}
	}

	if answerID != nil {
		_, err := s.answerRepo.GetAnswerByID(*answerID)
		if err != nil {
			return nil, errors.New("answer not found")
		}
	}

	comment := &models.Comment{
		Content:    content,
		UserID:     userID,
		QuestionID: questionID,
		AnswerID:   answerID,
	}

	if err := s.commentRepo.CreateComment(comment); err != nil {
		return nil, err
	}

	return comment, nil
}

func (s *commentService) GetCommentByID(id uint) (*models.Comment, error) {
	return s.commentRepo.GetCommentByID(id)
}

func (s *commentService) UpdateComment(id uint, content string) (*models.Comment, error) {
	comment, err := s.commentRepo.GetCommentByID(id)
	if err != nil {
		return nil, err
	}

	if content != "" {
		comment.Content = content
	}

	if err := s.commentRepo.UpdateComment(comment); err != nil {
		return nil, err
	}

	return comment, nil
}

func (s *commentService) DeleteComment(id uint) error {
	return s.commentRepo.DeleteComment(id)
}

func (s *commentService) ListComments(filters map[string]interface{}) ([]models.Comment, error) {
	return s.commentRepo.ListComments(filters)
}
